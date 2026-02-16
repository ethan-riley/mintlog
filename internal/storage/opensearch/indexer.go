package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"

	"github.com/felipemonteiro/mintlog/pkg/logmodel"
)

const (
	defaultBatchSize  = 500
	defaultFlushEvery = 2 * time.Second
)

type Indexer struct {
	client *opensearchapi.Client
	js     nats.JetStreamContext
	sub    *nats.Subscription

	mu     sync.Mutex
	buffer []bulkItem
	cancel context.CancelFunc
}

type bulkItem struct {
	index string
	id    string
	doc   []byte
	msg   *nats.Msg
}

func NewIndexer(client *opensearchapi.Client, js nats.JetStreamContext) *Indexer {
	return &Indexer{
		client: client,
		js:     js,
		buffer: make([]bulkItem, 0, defaultBatchSize),
	}
}

func (idx *Indexer) Start(ctx context.Context) error {
	ctx, idx.cancel = context.WithCancel(ctx)

	sub, err := idx.js.QueueSubscribe(
		"logs.parsed.>",
		"opensearch-indexers",
		idx.handleMessage,
		nats.Durable("opensearch-indexer"),
		nats.ManualAck(),
		nats.AckWait(30_000_000_000),
		nats.MaxDeliver(3),
	)
	if err != nil {
		return fmt.Errorf("subscribe logs.parsed: %w", err)
	}
	idx.sub = sub

	go idx.flushLoop(ctx)

	slog.Info("opensearch indexer started")
	return nil
}

func (idx *Indexer) Stop() {
	if idx.cancel != nil {
		idx.cancel()
	}
	if idx.sub != nil {
		idx.sub.Unsubscribe()
	}
	idx.flush()
}

func (idx *Indexer) handleMessage(msg *nats.Msg) {
	var event logmodel.LogEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		slog.Error("indexer: unmarshal failed", "error", err)
		msg.Nak()
		return
	}

	indexName := IndexName(event.TenantID, event.Timestamp)
	doc, _ := json.Marshal(event)

	idx.mu.Lock()
	idx.buffer = append(idx.buffer, bulkItem{
		index: indexName,
		id:    event.ID,
		doc:   doc,
		msg:   msg,
	})
	shouldFlush := len(idx.buffer) >= defaultBatchSize
	idx.mu.Unlock()

	if shouldFlush {
		idx.flush()
	}
}

func (idx *Indexer) flushLoop(ctx context.Context) {
	ticker := time.NewTicker(defaultFlushEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			idx.flush()
		}
	}
}

func (idx *Indexer) flush() {
	idx.mu.Lock()
	if len(idx.buffer) == 0 {
		idx.mu.Unlock()
		return
	}
	items := idx.buffer
	idx.buffer = make([]bulkItem, 0, defaultBatchSize)
	idx.mu.Unlock()

	var body strings.Builder
	for _, item := range items {
		meta := fmt.Sprintf(`{"index":{"_index":"%s","_id":"%s"}}`, item.index, item.id)
		body.WriteString(meta)
		body.WriteByte('\n')
		body.Write(item.doc)
		body.WriteByte('\n')
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := idx.client.Bulk(ctx, opensearchapi.BulkReq{
		Body: strings.NewReader(body.String()),
	})

	if err != nil {
		slog.Error("bulk index failed", "error", err, "count", len(items))
		for _, item := range items {
			item.msg.Nak()
		}
		return
	}

	if resp.Errors {
		slog.Warn("bulk index had errors", "count", len(items))
	}

	for _, item := range items {
		item.msg.Ack()
	}

	slog.Debug("bulk indexed", "count", len(items))
}
