package valueobjects

import "errors"

type OrderStatus string

const (
    OrderStatusDraft     OrderStatus = "draft"
    OrderStatusConfirmed OrderStatus = "confirmed"
    OrderStatusShipped   OrderStatus = "shipped"
    OrderStatusDelivered OrderStatus = "delivered"
    OrderStatusCancelled OrderStatus = "cancelled"
)

func (s OrderStatus) String() string {
    return string(s)
}

func (s OrderStatus) IsValid() bool {
    switch s {
    case OrderStatusDraft, OrderStatusConfirmed, OrderStatusShipped, OrderStatusDelivered, OrderStatusCancelled:
        return true
    default:
        return false
    }
}

func (s OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
    switch s {
    case OrderStatusDraft:
        return newStatus == OrderStatusConfirmed || newStatus == OrderStatusCancelled
    case OrderStatusConfirmed:
        return newStatus == OrderStatusShipped || newStatus == OrderStatusCancelled
    case OrderStatusShipped:
        return newStatus == OrderStatusDelivered
    case OrderStatusDelivered, OrderStatusCancelled:
        return false
    default:
        return false
    }
}

func (s OrderStatus) Validate() error {
    if !s.IsValid() {
        return errors.New("invalid order status")
    }
    return nil
}

func ParseOrderStatus(status string) (OrderStatus, error) {
    orderStatus := OrderStatus(status)
    if !orderStatus.IsValid() {
        return "", errors.New("invalid order status")
    }
    return orderStatus, nil
}
