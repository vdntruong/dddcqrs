package events

import "time"

type DomainEvent interface {
    Type() string
    AggregateID() string
    OccurredAt() time.Time
}

type BaseDomainEvent struct {
    EventType       string    `json:"event_type"`
    AggregateIDValue string    `json:"aggregate_id"`
    OccurredAtTime   time.Time `json:"occurred_at"`
}

func (e BaseDomainEvent) Type() string {
    return e.EventType
}

func (e BaseDomainEvent) AggregateID() string {
    return e.AggregateIDValue
}

func (e BaseDomainEvent) OccurredAt() time.Time {
    return e.OccurredAtTime
}
