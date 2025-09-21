package handlers

import (
	"context"
	"log"

	"github.com/vdntruong/dddcqrs/shared/domain/events"
	"github.com/vdntruong/dddcqrs/shared/infrastructure/eventbus"
)

type EventConsumer struct {
    ProjectionHandler *OrderProjectionHandler
    EventBus         eventbus.EventBus
}

func (ec *EventConsumer) Start(ctx context.Context) error {
    topic := "orders"
    
    log.Printf("Starting event consumer for topic: %s", topic)
    
    return ec.EventBus.Subscribe(ctx, topic, ec.handleEvent)
}

func (ec *EventConsumer) handleEvent(event events.DomainEvent) error {
    log.Printf("Processing event: %s for aggregate: %s", event.Type(), event.AggregateID())
    
    if err := ec.ProjectionHandler.Handle(context.Background(), event); err != nil {
        log.Printf("Error processing event %s: %v", event.Type(), err)
        return err
    }
    
    log.Printf("Successfully processed event: %s", event.Type())
    return nil
}
