package events

import (
	"time"

	"github.com/vdntruong/dddcqrs/shared/domain/entities"
	"github.com/vdntruong/dddcqrs/shared/domain/valueobjects"
)

type OrderCreatedEvent struct {
    BaseDomainEvent
    CustomerID      string                `json:"customer_id"`
    Items           []OrderItemData       `json:"items"`
    TotalAmount     valueobjects.Money    `json:"total_amount"`
    ShippingAddress valueobjects.Address `json:"shipping_address"`
}

type OrderItemData struct {
    ProductID string              `json:"product_id"`
    Quantity  int                `json:"quantity"`
    Price     valueobjects.Money `json:"price"`
}

func NewOrderCreatedEvent(order *entities.Order) OrderCreatedEvent {
    items := make([]OrderItemData, len(order.Items))
    for i, item := range order.Items {
        items[i] = OrderItemData{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            Price:     item.Price,
        }
    }
    
    return OrderCreatedEvent{
        BaseDomainEvent: BaseDomainEvent{
            EventType:   "OrderCreated",
            AggregateID: string(order.ID),
            OccurredAt:  time.Now(),
        },
        CustomerID:      order.CustomerID,
        Items:           items,
        TotalAmount:     order.TotalAmount,
        ShippingAddress: order.ShippingAddress,
    }
}

type OrderConfirmedEvent struct {
    BaseDomainEvent
    CustomerID string `json:"customer_id"`
}

func NewOrderConfirmedEvent(order *entities.Order) OrderConfirmedEvent {
    return OrderConfirmedEvent{
        BaseDomainEvent: BaseDomainEvent{
            EventType:   "OrderConfirmed",
            AggregateID: string(order.ID),
            OccurredAt:  time.Now(),
        },
        CustomerID: order.CustomerID,
    }
}

type OrderShippedEvent struct {
    BaseDomainEvent
    CustomerID string `json:"customer_id"`
}

func NewOrderShippedEvent(order *entities.Order) OrderShippedEvent {
    return OrderShippedEvent{
        BaseDomainEvent: BaseDomainEvent{
            EventType:   "OrderShipped",
            AggregateID: string(order.ID),
            OccurredAt:  time.Now(),
        },
        CustomerID: order.CustomerID,
    }
}

type OrderDeliveredEvent struct {
    BaseDomainEvent
    CustomerID string `json:"customer_id"`
}

func NewOrderDeliveredEvent(order *entities.Order) OrderDeliveredEvent {
    return OrderDeliveredEvent{
        BaseDomainEvent: BaseDomainEvent{
            EventType:   "OrderDelivered",
            AggregateID: string(order.ID),
            OccurredAt:  time.Now(),
        },
        CustomerID: order.CustomerID,
    }
}

type OrderCancelledEvent struct {
    BaseDomainEvent
    CustomerID string `json:"customer_id"`
    Reason     string `json:"reason"`
}

func NewOrderCancelledEvent(order *entities.Order, reason string) OrderCancelledEvent {
    return OrderCancelledEvent{
        BaseDomainEvent: BaseDomainEvent{
            EventType:   "OrderCancelled",
            AggregateID: string(order.ID),
            OccurredAt:  time.Now(),
        },
        CustomerID: order.CustomerID,
        Reason:     reason,
    }
}

type OrderItemAddedEvent struct {
    BaseDomainEvent
    ProductID string              `json:"product_id"`
    Quantity  int                `json:"quantity"`
    Price     valueobjects.Money `json:"price"`
}

func NewOrderItemAddedEvent(order *entities.Order, productID string, quantity int, price valueobjects.Money) OrderItemAddedEvent {
    return OrderItemAddedEvent{
        BaseDomainEvent: BaseDomainEvent{
            EventType:   "OrderItemAdded",
            AggregateID: string(order.ID),
            OccurredAt:  time.Now(),
        },
        ProductID: productID,
        Quantity:  quantity,
        Price:     price,
    }
}

type OrderItemRemovedEvent struct {
    BaseDomainEvent
    ProductID string `json:"product_id"`
}

func NewOrderItemRemovedEvent(order *entities.Order, productID string) OrderItemRemovedEvent {
    return OrderItemRemovedEvent{
        BaseDomainEvent: BaseDomainEvent{
            EventType:   "OrderItemRemoved",
            AggregateID: string(order.ID),
            OccurredAt:  time.Now(),
        },
        ProductID: productID,
    }
}
