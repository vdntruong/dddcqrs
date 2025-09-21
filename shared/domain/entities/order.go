package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vdntruong/dddcqrs/shared/domain/valueobjects"
)

type OrderID string

type Order struct {
    ID              OrderID
    CustomerID      string
    Items           []OrderItem
    Status          valueobjects.OrderStatus
    TotalAmount     valueobjects.Money
    ShippingAddress valueobjects.Address
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type OrderItem struct {
    ProductID string
    Quantity  int
    Price     valueobjects.Money
}

func NewOrder(customerID string, shippingAddress valueobjects.Address) *Order {
    return &Order{
        ID:              OrderID(uuid.New().String()),
        CustomerID:      customerID,
        Items:           []OrderItem{},
        Status:          valueobjects.OrderStatusDraft,
        TotalAmount:     valueobjects.Money{},
        ShippingAddress: shippingAddress,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }
}

func (o *Order) AddItem(productID string, quantity int, price valueobjects.Money) error {
    if o.Status != valueobjects.OrderStatusDraft {
        return errors.New("cannot modify order that is not in draft status")
    }
    
    if quantity <= 0 {
        return errors.New("quantity must be greater than zero")
    }
    
    item := OrderItem{
        ProductID: productID,
        Quantity:  quantity,
        Price:     price,
    }
    
    o.Items = append(o.Items, item)
    o.recalculateTotal()
    o.UpdatedAt = time.Now()
    
    return nil
}

func (o *Order) RemoveItem(productID string) error {
    if o.Status != valueobjects.OrderStatusDraft {
        return errors.New("cannot modify order that is not in draft status")
    }
    
    for i, item := range o.Items {
        if item.ProductID == productID {
            o.Items = append(o.Items[:i], o.Items[i+1:]...)
            o.recalculateTotal()
            o.UpdatedAt = time.Now()
            return nil
        }
    }
    
    return errors.New("item not found")
}

func (o *Order) Confirm() error {
    if o.Status != valueobjects.OrderStatusDraft {
        return errors.New("can only confirm draft orders")
    }
    
    if len(o.Items) == 0 {
        return errors.New("cannot confirm order without items")
    }
    
    o.Status = valueobjects.OrderStatusConfirmed
    o.UpdatedAt = time.Now()
    
    return nil
}

func (o *Order) Cancel() error {
    if o.Status == valueobjects.OrderStatusShipped || o.Status == valueobjects.OrderStatusDelivered {
        return errors.New("cannot cancel shipped or delivered orders")
    }
    
    o.Status = valueobjects.OrderStatusCancelled
    o.UpdatedAt = time.Now()
    
    return nil
}

func (o *Order) Ship() error {
    if o.Status != valueobjects.OrderStatusConfirmed {
        return errors.New("can only ship confirmed orders")
    }
    
    o.Status = valueobjects.OrderStatusShipped
    o.UpdatedAt = time.Now()
    
    return nil
}

func (o *Order) Deliver() error {
    if o.Status != valueobjects.OrderStatusShipped {
        return errors.New("can only deliver shipped orders")
    }
    
    o.Status = valueobjects.OrderStatusDelivered
    o.UpdatedAt = time.Now()
    
    return nil
}

func (o *Order) recalculateTotal() {
    total := int64(0)
    for _, item := range o.Items {
        itemTotal := item.Price.Amount * int64(item.Quantity)
        total += itemTotal
    }
    
    o.TotalAmount = valueobjects.Money{
        Amount:   total,
        Currency: "USD",
    }
}
