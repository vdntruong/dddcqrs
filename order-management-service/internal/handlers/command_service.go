package handlers

import (
	"context"
	"fmt"

	"github.com/vdntruong/dddcqrs/order-management-service/internal/repositories"
	"github.com/vdntruong/dddcqrs/shared/domain/entities"
	"github.com/vdntruong/dddcqrs/shared/domain/events"
	"github.com/vdntruong/dddcqrs/shared/infrastructure/eventbus"
)

type CommandService struct {
    OrderRepo  repositories.OrderRepository
    EventStore repositories.EventStore
    Outbox     repositories.OutboxRepository
    EventBus   eventbus.EventBus
}

func (cs *CommandService) CreateOrder(ctx context.Context, cmd CreateOrderCommand) (*entities.Order, error) {
    // Validate command
    if err := cmd.Validate(); err != nil {
        return nil, fmt.Errorf("invalid command: %w", err)
    }
    
    // Create order aggregate
    order := entities.NewOrder(cmd.CustomerID, cmd.ShippingAddress)
    
    // Add items
    for _, item := range cmd.Items {
        if err := order.AddItem(item.ProductID, item.Quantity, item.Price); err != nil {
            return nil, fmt.Errorf("failed to add item: %w", err)
        }
    }
    
    // Save aggregate to database
    if err := cs.OrderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("failed to save order: %w", err)
    }
    
    // Create domain event
    event := events.NewOrderCreatedEvent(order)
    
    // Save event to event store
    if err := cs.EventStore.SaveEvents(ctx, string(order.ID), []events.DomainEvent{event}, 0); err != nil {
        return nil, fmt.Errorf("failed to save event: %w", err)
    }
    
    // Save event to outbox for reliable publishing
    if err := cs.Outbox.SaveEvent(ctx, event); err != nil {
        return nil, fmt.Errorf("failed to save event to outbox: %w", err)
    }
    
    return order, nil
}

func (cs *CommandService) ConfirmOrder(ctx context.Context, orderID entities.OrderID) error {
    // Load order
    order, err := cs.OrderRepo.FindByID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed to find order: %w", err)
    }
    
    // Confirm order
    if err := order.Confirm(); err != nil {
        return fmt.Errorf("failed to confirm order: %w", err)
    }
    
    // Save updated order
    if err := cs.OrderRepo.Update(ctx, order); err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    // Create domain event
    event := events.NewOrderConfirmedEvent(order)
    
    // Save event to outbox
    if err := cs.Outbox.SaveEvent(ctx, event); err != nil {
        return fmt.Errorf("failed to save event to outbox: %w", err)
    }
    
    return nil
}

func (cs *CommandService) CancelOrder(ctx context.Context, orderID entities.OrderID, reason string) error {
    // Load order
    order, err := cs.OrderRepo.FindByID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed to find order: %w", err)
    }
    
    // Cancel order
    if err := order.Cancel(); err != nil {
        return fmt.Errorf("failed to cancel order: %w", err)
    }
    
    // Save updated order
    if err := cs.OrderRepo.Update(ctx, order); err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    // Create domain event
    event := events.NewOrderCancelledEvent(order, reason)
    
    // Save event to outbox
    if err := cs.Outbox.SaveEvent(ctx, event); err != nil {
        return fmt.Errorf("failed to save event to outbox: %w", err)
    }
    
    return nil
}

func (cs *CommandService) AddOrderItem(ctx context.Context, orderID entities.OrderID, productID string, quantity int, price entities.Money) error {
    // Load order
    order, err := cs.OrderRepo.FindByID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed to find order: %w", err)
    }
    
    // Add item
    if err := order.AddItem(productID, quantity, price); err != nil {
        return fmt.Errorf("failed to add item: %w", err)
    }
    
    // Save updated order
    if err := cs.OrderRepo.Update(ctx, order); err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    // Create domain event
    event := events.NewOrderItemAddedEvent(order, productID, quantity, price)
    
    // Save event to outbox
    if err := cs.Outbox.SaveEvent(ctx, event); err != nil {
        return fmt.Errorf("failed to save event to outbox: %w", err)
    }
    
    return nil
}

func (cs *CommandService) RemoveOrderItem(ctx context.Context, orderID entities.OrderID, productID string) error {
    // Load order
    order, err := cs.OrderRepo.FindByID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("failed to find order: %w", err)
    }
    
    // Remove item
    if err := order.RemoveItem(productID); err != nil {
        return fmt.Errorf("failed to remove item: %w", err)
    }
    
    // Save updated order
    if err := cs.OrderRepo.Update(ctx, order); err != nil {
        return fmt.Errorf("failed to update order: %w", err)
    }
    
    // Create domain event
    event := events.NewOrderItemRemovedEvent(order, productID)
    
    // Save event to outbox
    if err := cs.Outbox.SaveEvent(ctx, event); err != nil {
        return fmt.Errorf("failed to save event to outbox: %w", err)
    }
    
    return nil
}
