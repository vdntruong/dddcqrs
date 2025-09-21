package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vdntruong/dddcqrs/shared/domain/events"
)

type OutboxRepository interface {
    SaveEvent(ctx context.Context, event events.DomainEvent) error
    SaveEventWithTx(ctx context.Context, tx *sql.Tx, event events.DomainEvent) error
    GetUnprocessedEvents(ctx context.Context) ([]OutboxEvent, error)
    MarkAsProcessed(ctx context.Context, eventID string) error
}

type OutboxEvent struct {
    ID        string    `json:"id"`
    EventType string    `json:"event_type"`
    EventData []byte    `json:"event_data"`
    CreatedAt time.Time `json:"created_at"`
    Processed bool      `json:"processed"`
}

type outboxRepository struct {
    db *sql.DB
}

func NewOutboxRepository(db *sql.DB) OutboxRepository {
    return &outboxRepository{db: db}
}

func (r *outboxRepository) SaveEvent(ctx context.Context, event events.DomainEvent) error {
    eventData, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    
    query := `
        INSERT INTO outbox_events (id, event_type, event_data, created_at, processed)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    _, err = r.db.ExecContext(ctx, query,
        uuid.New().String(),
        event.Type(),
        eventData,
        time.Now(),
        false,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save event to outbox: %w", err)
    }
    
    return nil
}

func (r *outboxRepository) SaveEventWithTx(ctx context.Context, tx *sql.Tx, event events.DomainEvent) error {
    eventData, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }
    
    query := `
        INSERT INTO outbox_events (id, event_type, event_data, created_at, processed)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    _, err = tx.ExecContext(ctx, query,
        uuid.New().String(),
        event.Type(),
        eventData,
        time.Now(),
        false,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save event to outbox: %w", err)
    }
    
    return nil
}

func (r *outboxRepository) GetUnprocessedEvents(ctx context.Context) ([]OutboxEvent, error) {
    query := `
        SELECT id, event_type, event_data, created_at, processed
        FROM outbox_events
        WHERE processed = false
        ORDER BY created_at ASC
        LIMIT 100
    `
    
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to query unprocessed events: %w", err)
    }
    defer rows.Close()
    
    var events []OutboxEvent
    for rows.Next() {
        var event OutboxEvent
        err := rows.Scan(
            &event.ID,
            &event.EventType,
            &event.EventData,
            &event.CreatedAt,
            &event.Processed,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan outbox event: %w", err)
        }
        events = append(events, event)
    }
    
    return events, nil
}

func (r *outboxRepository) MarkAsProcessed(ctx context.Context, eventID string) error {
    query := `
        UPDATE outbox_events
        SET processed = true
        WHERE id = $1
    `
    
    result, err := r.db.ExecContext(ctx, query, eventID)
    if err != nil {
        return fmt.Errorf("failed to mark event as processed: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("event not found: %s", eventID)
    }
    
    return nil
}
