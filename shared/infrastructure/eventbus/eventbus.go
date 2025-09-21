package eventbus

import (
	"context"

	"github.com/vdntruong/dddcqrs/shared/domain/events"
)

type EventBus interface {
    Publish(ctx context.Context, event events.DomainEvent) error
    Subscribe(ctx context.Context, topic string, handler func(events.DomainEvent) error) error
    Close() error
}
