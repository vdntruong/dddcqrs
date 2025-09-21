package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/vdntruong/dddcqrs/order-management-service/internal/repositories"
	"github.com/vdntruong/dddcqrs/shared/domain/events"
	"github.com/vdntruong/dddcqrs/shared/infrastructure/eventbus"
)

type EventPublisher struct {
    OutboxRepo repositories.OutboxRepository
    EventBus   eventbus.EventBus
}

func (ep *EventPublisher) ProcessEvents(ctx context.Context) error {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := ep.processBatch(ctx); err != nil {
                log.Printf("Error processing event batch: %v", err)
            }
        }
    }
}

func (ep *EventPublisher) processBatch(ctx context.Context) error {
    // Get unprocessed events
    outboxEvents, err := ep.OutboxRepo.GetUnprocessedEvents(ctx)
    if err != nil {
        return err
    }
    
    if len(outboxEvents) == 0 {
        return nil
    }
    
    log.Printf("Processing %d events from outbox", len(outboxEvents))
    
    for _, outboxEvent := range outboxEvents {
        if err := ep.processEvent(ctx, outboxEvent); err != nil {
            log.Printf("Error processing event %s: %v", outboxEvent.ID, err)
            continue
        }
        
        // Mark as processed
        if err := ep.OutboxRepo.MarkAsProcessed(ctx, outboxEvent.ID); err != nil {
            log.Printf("Error marking event %s as processed: %v", outboxEvent.ID, err)
        }
    }
    
    return nil
}

func (ep *EventPublisher) processEvent(ctx context.Context, outboxEvent repositories.OutboxEvent) error {
    // Parse the event
    event, err := ep.parseEvent(outboxEvent.EventType, outboxEvent.EventData)
    if err != nil {
        return err
    }
    
    // Publish to Kafka
    if err := ep.EventBus.Publish(ctx, event); err != nil {
        return err
    }
    
    log.Printf("Successfully published event %s for aggregate %s", event.Type(), event.AggregateID())
    return nil
}

func (ep *EventPublisher) parseEvent(eventType string, eventData []byte) (events.DomainEvent, error) {
    switch eventType {
    case "OrderCreated":
        var event events.OrderCreatedEvent
        if err := json.Unmarshal(eventData, &event); err != nil {
            return nil, err
        }
        return event, nil
    case "OrderConfirmed":
        var event events.OrderConfirmedEvent
        if err := json.Unmarshal(eventData, &event); err != nil {
            return nil, err
        }
        return event, nil
    case "OrderShipped":
        var event events.OrderShippedEvent
        if err := json.Unmarshal(eventData, &event); err != nil {
            return nil, err
        }
        return event, nil
    case "OrderDelivered":
        var event events.OrderDeliveredEvent
        if err := json.Unmarshal(eventData, &event); err != nil {
            return nil, err
        }
        return event, nil
    case "OrderCancelled":
        var event events.OrderCancelledEvent
        if err := json.Unmarshal(eventData, &event); err != nil {
            return nil, err
        }
        return event, nil
    case "OrderItemAdded":
        var event events.OrderItemAddedEvent
        if err := json.Unmarshal(eventData, &event); err != nil {
            return nil, err
        }
        return event, nil
    case "OrderItemRemoved":
        var event events.OrderItemRemovedEvent
        if err := json.Unmarshal(eventData, &event); err != nil {
            return nil, err
        }
        return event, nil
    default:
        return nil, errors.New("unknown event type: " + eventType)
    }
}
