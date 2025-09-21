package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vdntruong/dddcqrs/shared/domain/events"
)

type EventStore interface {
    SaveEvents(ctx context.Context, aggregateID string, events []events.DomainEvent, expectedVersion int) error
    GetEvents(ctx context.Context, aggregateID string) ([]events.DomainEvent, error)
}

type eventStore struct {
    db *sql.DB
}

func NewEventStore(db *sql.DB) EventStore {
    return &eventStore{db: db}
}

func (es *eventStore) SaveEvents(ctx context.Context, aggregateID string, domainEvents []events.DomainEvent, expectedVersion int) error {
    tx, err := es.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    for i, event := range domainEvents {
        version := expectedVersion + i + 1
        
        eventData, err := json.Marshal(event)
        if err != nil {
            return fmt.Errorf("failed to marshal event: %w", err)
        }
        
        query := `
            INSERT INTO events (aggregate_id, event_type, event_data, version, occurred_at)
            VALUES ($1, $2, $3, $4, $5)
        `
        
        _, err = tx.ExecContext(ctx, query,
            aggregateID,
            event.Type(),
            eventData,
            version,
            event.OccurredAt(),
        )
        
        if err != nil {
            return fmt.Errorf("failed to save event: %w", err)
        }
    }
    
    return tx.Commit()
}

func (es *eventStore) GetEvents(ctx context.Context, aggregateID string) ([]events.DomainEvent, error) {
    query := `
        SELECT event_type, event_data, version, occurred_at
        FROM events
        WHERE aggregate_id = $1
        ORDER BY version ASC
    `
    
    rows, err := es.db.QueryContext(ctx, query, aggregateID)
    if err != nil {
        return nil, fmt.Errorf("failed to query events: %w", err)
    }
    defer rows.Close()
    
    var domainEvents []events.DomainEvent
    for rows.Next() {
        var eventType string
        var eventData []byte
        var version int
        var occurredAt string
        
        err := rows.Scan(&eventType, &eventData, &version, &occurredAt)
        if err != nil {
            return nil, fmt.Errorf("failed to scan event: %w", err)
        }
        
        // Parse event based on type
        event, err := es.parseEvent(eventType, eventData)
        if err != nil {
            return nil, fmt.Errorf("failed to parse event: %w", err)
        }
        
        domainEvents = append(domainEvents, event)
    }
    
    return domainEvents, nil
}

func (es *eventStore) parseEvent(eventType string, eventData []byte) (events.DomainEvent, error) {
    // This is a simplified version - in production, you'd use a more sophisticated
    // event deserialization mechanism
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
        return nil, fmt.Errorf("unknown event type: %s", eventType)
    }
}
