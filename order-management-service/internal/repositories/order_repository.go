package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/vdntruong/dddcqrs/shared/domain/entities"
)

type OrderRepository interface {
    Save(ctx context.Context, order *entities.Order) error
    FindByID(ctx context.Context, id entities.OrderID) (*entities.Order, error)
    Update(ctx context.Context, order *entities.Order) error
    Delete(ctx context.Context, id entities.OrderID) error
}

type orderRepository struct {
    db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
    return &orderRepository{db: db}
}

func (r *orderRepository) Save(ctx context.Context, order *entities.Order) error {
    query := `
        INSERT INTO orders (id, customer_id, status, total_amount, shipping_address, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE SET
            customer_id = $2,
            status = $3,
            total_amount = $4,
            shipping_address = $5,
            updated_at = $7
    `
    
    shippingAddressJSON, err := json.Marshal(order.ShippingAddress)
    if err != nil {
        return fmt.Errorf("failed to marshal shipping address: %w", err)
    }
    
    _, err = r.db.ExecContext(ctx, query,
        order.ID,
        order.CustomerID,
        order.Status.String(),
        order.TotalAmount.Amount,
        shippingAddressJSON,
        order.CreatedAt,
        order.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    // Save order items
    return r.saveOrderItems(ctx, order)
}

func (r *orderRepository) FindByID(ctx context.Context, id entities.OrderID) (*entities.Order, error) {
    query := `
        SELECT id, customer_id, status, total_amount, shipping_address, created_at, updated_at
        FROM orders
        WHERE id = $1
    `
    
    var order entities.Order
    var shippingAddressJSON string
    
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &order.ID,
        &order.CustomerID,
        &order.Status,
        &order.TotalAmount.Amount,
        &shippingAddressJSON,
        &order.CreatedAt,
        &order.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("order not found")
        }
        return nil, fmt.Errorf("failed to find order: %w", err)
    }
    
    // Parse shipping address
    if err := json.Unmarshal([]byte(shippingAddressJSON), &order.ShippingAddress); err != nil {
        return nil, fmt.Errorf("failed to unmarshal shipping address: %w", err)
    }
    
    // Set currency
    order.TotalAmount.Currency = "USD"
    
    // Load order items
    items, err := r.findOrderItems(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to load order items: %w", err)
    }
    order.Items = items
    
    return &order, nil
}

func (r *orderRepository) Update(ctx context.Context, order *entities.Order) error {
    return r.Save(ctx, order)
}

func (r *orderRepository) Delete(ctx context.Context, id entities.OrderID) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    // Delete order items first
    _, err = tx.ExecContext(ctx, "DELETE FROM order_items WHERE order_id = $1", id)
    if err != nil {
        return fmt.Errorf("failed to delete order items: %w", err)
    }
    
    // Delete order
    _, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE id = $1", id)
    if err != nil {
        return fmt.Errorf("failed to delete order: %w", err)
    }
    
    return tx.Commit()
}

func (r *orderRepository) saveOrderItems(ctx context.Context, order *entities.Order) error {
    // Delete existing items
    _, err := r.db.ExecContext(ctx, "DELETE FROM order_items WHERE order_id = $1", order.ID)
    if err != nil {
        return fmt.Errorf("failed to delete existing order items: %w", err)
    }
    
    // Insert new items
    for _, item := range order.Items {
        query := `
            INSERT INTO order_items (order_id, product_id, quantity, price_amount, price_currency)
            VALUES ($1, $2, $3, $4, $5)
        `
        
        _, err = r.db.ExecContext(ctx, query,
            order.ID,
            item.ProductID,
            item.Quantity,
            item.Price.Amount,
            item.Price.Currency,
        )
        
        if err != nil {
            return fmt.Errorf("failed to save order item: %w", err)
        }
    }
    
    return nil
}

func (r *orderRepository) findOrderItems(ctx context.Context, orderID entities.OrderID) ([]entities.OrderItem, error) {
    query := `
        SELECT product_id, quantity, price_amount, price_currency
        FROM order_items
        WHERE order_id = $1
        ORDER BY product_id
    `
    
    rows, err := r.db.QueryContext(ctx, query, orderID)
    if err != nil {
        return nil, fmt.Errorf("failed to query order items: %w", err)
    }
    defer rows.Close()
    
    var items []entities.OrderItem
    for rows.Next() {
        var item entities.OrderItem
        err := rows.Scan(
            &item.ProductID,
            &item.Quantity,
            &item.Price.Amount,
            &item.Price.Currency,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan order item: %w", err)
        }
        items = append(items, item)
    }
    
    return items, nil
}
