package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type Searcher struct {
	client *opensearchapi.Client
}

func NewSearcher(client *opensearchapi.Client) *Searcher {
	return &Searcher{client: client}
}

type SearchResult struct {
	Hits  []json.RawMessage `json:"hits"`
	Total int               `json:"total"`
}

func (s *Searcher) Search(ctx context.Context, indices []string, query map[string]any) (*SearchResult, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	resp, err := s.client.Search(ctx, &opensearchapi.SearchReq{
		Indices: indices,
		Body:    strings.NewReader(string(body)),
	})
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	result := &SearchResult{
		Total: resp.Hits.Total.Value,
		Hits:  make([]json.RawMessage, 0, len(resp.Hits.Hits)),
	}
	for _, hit := range resp.Hits.Hits {
		result.Hits = append(result.Hits, hit.Source)
	}

	return result, nil
}

func (s *Searcher) Aggregate(ctx context.Context, indices []string, query map[string]any) (json.RawMessage, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("marshal query: %w", err)
	}

	resp, err := s.client.Search(ctx, &opensearchapi.SearchReq{
		Indices: indices,
		Body:    strings.NewReader(string(body)),
	})
	if err != nil {
		return nil, fmt.Errorf("aggregate: %w", err)
	}

	return resp.Aggregations, nil
}
