package ingest

import (
	"fmt"

	"github.com/felipemonteiro/mintlog/internal/bus"
	"github.com/felipemonteiro/mintlog/pkg/logmodel"
)

type LogPublisher struct {
	pub *bus.Publisher
}

func NewLogPublisher(pub *bus.Publisher) *LogPublisher {
	return &LogPublisher{pub: pub}
}

func (lp *LogPublisher) Publish(tenantID string, event *logmodel.LogEvent) error {
	subject := fmt.Sprintf("logs.raw.%s", tenantID)
	return lp.pub.Publish(subject, event)
}
