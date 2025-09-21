package handlers

import (
	"errors"
	"fmt"

	"github.com/vdntruong/dddcqrs/shared/domain/valueobjects"
)

type CreateOrderCommand struct {
    CustomerID      string                `json:"customer_id"`
    Items           []OrderItemCommand    `json:"items"`
    ShippingAddress valueobjects.Address  `json:"shipping_address"`
}

type OrderItemCommand struct {
    ProductID string              `json:"product_id"`
    Quantity  int                `json:"quantity"`
    Price     valueobjects.Money  `json:"price"`
}

type UpdateOrderCommand struct {
    OrderID         string                `json:"order_id"`
    Items           []OrderItemCommand    `json:"items"`
    ShippingAddress valueobjects.Address  `json:"shipping_address"`
}

type ConfirmOrderCommand struct {
    OrderID string `json:"order_id"`
}

type CancelOrderCommand struct {
    OrderID string `json:"order_id"`
    Reason  string `json:"reason"`
}

func (c CreateOrderCommand) Validate() error {
    if c.CustomerID == "" {
        return errors.New("customer_id is required")
    }
    
    if len(c.Items) == 0 {
        return errors.New("at least one item is required")
    }
    
    if err := c.ShippingAddress.Validate(); err != nil {
        return fmt.Errorf("invalid shipping address: %w", err)
    }
    
    for i, item := range c.Items {
        if err := item.Validate(); err != nil {
            return fmt.Errorf("invalid item at index %d: %w", i, err)
        }
    }
    
    return nil
}

func (i OrderItemCommand) Validate() error {
    if i.ProductID == "" {
        return errors.New("product_id is required")
    }
    
    if i.Quantity <= 0 {
        return errors.New("quantity must be greater than zero")
    }
    
    if err := i.Price.Validate(); err != nil {
        return fmt.Errorf("invalid price: %w", err)
    }
    
    return nil
}

func (c UpdateOrderCommand) Validate() error {
    if c.OrderID == "" {
        return errors.New("order_id is required")
    }
    
    if err := c.ShippingAddress.Validate(); err != nil {
        return fmt.Errorf("invalid shipping address: %w", err)
    }
    
    for i, item := range c.Items {
        if err := item.Validate(); err != nil {
            return fmt.Errorf("invalid item at index %d: %w", i, err)
        }
    }
    
    return nil
}

func (c ConfirmOrderCommand) Validate() error {
    if c.OrderID == "" {
        return errors.New("order_id is required")
    }
    return nil
}

func (c CancelOrderCommand) Validate() error {
    if c.OrderID == "" {
        return errors.New("order_id is required")
    }
    if c.Reason == "" {
        return errors.New("reason is required")
    }
    return nil
}
