package bus

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	js nats.JetStreamContext
}

func NewPublisher(js nats.JetStreamContext) *Publisher {
	return &Publisher{js: js}
}

func (p *Publisher) Publish(subject string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = p.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("publish %s: %w", subject, err)
	}
	return nil
}
