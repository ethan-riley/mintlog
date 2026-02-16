package ingest

import (
	"fmt"

	"github.com/felipemonteiro/mintlog/pkg/logmodel"
)

const (
	maxMessageLen = 65536
	maxFieldsLen  = 100
	maxBatchSize  = 1000
)

func ValidateBatch(req *logmodel.IngestRequest) error {
	if len(req.Events) == 0 {
		return fmt.Errorf("events array is empty")
	}
	if len(req.Events) > maxBatchSize {
		return fmt.Errorf("batch size %d exceeds maximum %d", len(req.Events), maxBatchSize)
	}
	return nil
}

func ValidateEvent(e *logmodel.IngestEvent) error {
	if e.Message == "" {
		return fmt.Errorf("message is required")
	}
	if len(e.Message) > maxMessageLen {
		return fmt.Errorf("message exceeds maximum length %d", maxMessageLen)
	}
	if e.Service == "" {
		return fmt.Errorf("service is required")
	}
	if len(e.Fields) > maxFieldsLen {
		return fmt.Errorf("fields count %d exceeds maximum %d", len(e.Fields), maxFieldsLen)
	}
	return nil
}
