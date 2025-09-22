package handlers

import (
	"context"
	"log"

	"github.com/vdntruong/dddcqrs/order-reporting-service/internal/readmodels"
	"github.com/vdntruong/dddcqrs/shared/domain/events"
)

type OrderProjectionHandler struct {
    OrderReadModel    readmodels.OrderReadModel
    CustomerReadModel readmodels.CustomerReadModel
}

func (h *OrderProjectionHandler) Handle(ctx context.Context, event events.DomainEvent) error {
    switch e := event.(type) {
    case events.OrderCreatedEvent:
        return h.handleOrderCreated(ctx, e)
    case events.OrderConfirmedEvent:
        return h.handleOrderConfirmed(ctx, e)
    case events.OrderShippedEvent:
        return h.handleOrderShipped(ctx, e)
    case events.OrderDeliveredEvent:
        return h.handleOrderDelivered(ctx, e)
    case events.OrderCancelledEvent:
        return h.handleOrderCancelled(ctx, e)
    case events.OrderItemAddedEvent:
        return h.handleOrderItemAdded(ctx, e)
    case events.OrderItemRemovedEvent:
        return h.handleOrderItemRemoved(ctx, e)
    default:
        log.Printf("Unknown event type: %T", event)
        return nil
    }
}

func (h *OrderProjectionHandler) handleOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error {
    // Convert items
    items := make([]readmodels.OrderItemDTO, len(event.Items))
    for i, item := range event.Items {
        items[i] = readmodels.OrderItemDTO{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     item.Price,
        }
    }
    
    order := &readmodels.OrderDTO{
        ID:              event.AggregateID(),
        CustomerID:      event.CustomerID,
        Status:          "draft",
        TotalAmount:     event.TotalAmount,
        ShippingAddress: event.ShippingAddress,
        Items:           items,
        CreatedAt:       event.OccurredAt(),
        UpdatedAt:       event.OccurredAt(),
    }
    
    return h.OrderReadModel.CreateOrder(ctx, order)
}

func (h *OrderProjectionHandler) handleOrderConfirmed(ctx context.Context, event events.OrderConfirmedEvent) error {
    // Get existing order
    order, err := h.OrderReadModel.GetOrder(ctx, event.AggregateID())
    if err != nil {
        return err
    }
    
    // Update status
    order.Status = "confirmed"
    order.UpdatedAt = event.OccurredAt()
    
    return h.OrderReadModel.UpdateOrder(ctx, order)
}

func (h *OrderProjectionHandler) handleOrderShipped(ctx context.Context, event events.OrderShippedEvent) error {
    // Get existing order
    order, err := h.OrderReadModel.GetOrder(ctx, event.AggregateID())
    if err != nil {
        return err
    }
    
    // Update status
    order.Status = "shipped"
    order.UpdatedAt = event.OccurredAt()
    
    return h.OrderReadModel.UpdateOrder(ctx, order)
}

func (h *OrderProjectionHandler) handleOrderDelivered(ctx context.Context, event events.OrderDeliveredEvent) error {
    // Get existing order
    order, err := h.OrderReadModel.GetOrder(ctx, event.AggregateID())
    if err != nil {
        return err
    }
    
    // Update status
    order.Status = "delivered"
    order.UpdatedAt = event.OccurredAt()
    
    return h.OrderReadModel.UpdateOrder(ctx, order)
}

func (h *OrderProjectionHandler) handleOrderCancelled(ctx context.Context, event events.OrderCancelledEvent) error {
    // Get existing order
    order, err := h.OrderReadModel.GetOrder(ctx, event.AggregateID())
    if err != nil {
        return err
    }
    
    // Update status
    order.Status = "cancelled"
    order.UpdatedAt = event.OccurredAt()
    
    return h.OrderReadModel.UpdateOrder(ctx, order)
}

func (h *OrderProjectionHandler) handleOrderItemAdded(ctx context.Context, event events.OrderItemAddedEvent) error {
    // Get existing order
    order, err := h.OrderReadModel.GetOrder(ctx, event.AggregateID())
    if err != nil {
        return err
    }
    
    // Add item
    newItem := readmodels.OrderItemDTO{
        ProductID: event.ProductID,
        Quantity:  event.Quantity,
        Price:     event.Price,
    }
    
    order.Items = append(order.Items, newItem)
    order.UpdatedAt = event.OccurredAt()
    
    // Recalculate total (simplified)
    total := int64(0)
    for _, item := range order.Items {
        total += item.Price.Amount * int64(item.Quantity)
    }
    order.TotalAmount.Amount = total
    
    return h.OrderReadModel.UpdateOrder(ctx, order)
}

func (h *OrderProjectionHandler) handleOrderItemRemoved(ctx context.Context, event events.OrderItemRemovedEvent) error {
    // Get existing order
    order, err := h.OrderReadModel.GetOrder(ctx, event.AggregateID())
    if err != nil {
        return err
    }
    
    // Remove item
    for i, item := range order.Items {
        if item.ProductID == event.ProductID {
            order.Items = append(order.Items[:i], order.Items[i+1:]...)
            break
        }
    }
    
    order.UpdatedAt = event.OccurredAt()
    
    // Recalculate total (simplified)
    total := int64(0)
    for _, item := range order.Items {
        total += item.Price.Amount * int64(item.Quantity)
    }
    order.TotalAmount.Amount = total
    
    return h.OrderReadModel.UpdateOrder(ctx, order)
}
