package main

import (
	"context"
	"fmt"

	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
)

type publisher struct {
	bus   broker.Publisher
	topic string
}

func (p *publisher) publishEvents(ctx context.Context, events []schema.LogEvent) error {
	for i, ev := range events {
		if err := p.bus.Publish(ctx, p.topic, ev); err != nil {
			return fmt.Errorf("publish event[%d]: %w", i, err)
		}
	}
	return nil
}
