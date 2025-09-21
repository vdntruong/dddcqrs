package readmodels

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vdntruong/dddcqrs/shared/domain/valueobjects"
)

type OrderReadModel interface {
    GetOrder(ctx context.Context, orderID string) (*OrderDTO, error)
    CreateOrder(ctx context.Context, order *OrderDTO) error
    UpdateOrder(ctx context.Context, order *OrderDTO) error
    DeleteOrder(ctx context.Context, orderID string) error
    ListOrders(ctx context.Context, customerID string, limit, offset int) ([]*OrderDTO, error)
    GetOrderAnalytics(ctx context.Context, period string) (*OrderAnalyticsDTO, error)
}

type OrderDTO struct {
    ID              string                `json:"id"`
    CustomerID      string                `json:"customer_id"`
    Status          string                `json:"status"`
    TotalAmount     valueobjects.Money    `json:"total_amount"`
    ShippingAddress valueobjects.Address  `json:"shipping_address"`
    Items           []OrderItemDTO        `json:"items"`
    CreatedAt       time.Time             `json:"created_at"`
    UpdatedAt       time.Time             `json:"updated_at"`
}

type OrderItemDTO struct {
    ProductID string              `json:"product_id"`
    Quantity  int                `json:"quantity"`
    Price     valueobjects.Money  `json:"price"`
}

type OrderAnalyticsDTO struct {
    TotalOrders     int64   `json:"total_orders"`
    TotalRevenue    int64   `json:"total_revenue"`
    AverageOrderValue int64 `json:"average_order_value"`
    OrdersByStatus  map[string]int64 `json:"orders_by_status"`
}

type orderReadModel struct {
    db    *sql.DB
    redis *redis.Client
}

func NewOrderReadModel(db *sql.DB, redis *redis.Client) OrderReadModel {
    return &orderReadModel{
        db:    db,
        redis: redis,
    }
}

func (rm *orderReadModel) GetOrder(ctx context.Context, orderID string) (*OrderDTO, error) {
    // Try cache first
    cacheKey := "order:" + orderID
    cached, err := rm.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var order OrderDTO
        if err := json.Unmarshal([]byte(cached), &order); err == nil {
            return &order, nil
        }
    }
    
    // Fallback to database
    query := `
        SELECT id, customer_id, status, total_amount, shipping_address, items, created_at, updated_at
        FROM order_read_models
        WHERE id = $1
    `
    
    var order OrderDTO
    var shippingAddressJSON, itemsJSON string
    
    err = rm.db.QueryRowContext(ctx, query, orderID).Scan(
        &order.ID,
        &order.CustomerID,
        &order.Status,
        &order.TotalAmount.Amount,
        &shippingAddressJSON,
        &itemsJSON,
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
    
    // Parse items
    if err := json.Unmarshal([]byte(itemsJSON), &order.Items); err != nil {
        return nil, fmt.Errorf("failed to unmarshal items: %w", err)
    }
    
    // Set currency
    order.TotalAmount.Currency = "USD"
    for i := range order.Items {
        order.Items[i].Price.Currency = "USD"
    }
    
    // Cache the result
    orderData, _ := json.Marshal(order)
    rm.redis.Set(ctx, cacheKey, orderData, 1*time.Hour)
    
    return &order, nil
}

func (rm *orderReadModel) CreateOrder(ctx context.Context, order *OrderDTO) error {
    shippingAddressJSON, err := json.Marshal(order.ShippingAddress)
    if err != nil {
        return fmt.Errorf("failed to marshal shipping address: %w", err)
    }
    
    itemsJSON, err := json.Marshal(order.Items)
    if err != nil {
        return fmt.Errorf("failed to marshal items: %w", err)
    }
    
    query := `
        INSERT INTO order_read_models (id, customer_id, status, total_amount, shipping_address, items, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (id) DO UPDATE SET
            customer_id = $2,
            status = $3,
            total_amount = $4,
            shipping_address = $5,
            items = $6,
            updated_at = $8
    `
    
    _, err = rm.db.ExecContext(ctx, query,
        order.ID,
        order.CustomerID,
        order.Status,
        order.TotalAmount.Amount,
        shippingAddressJSON,
        itemsJSON,
        order.CreatedAt,
        order.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save order: %w", err)
    }
    
    // Cache the result
    orderData, _ := json.Marshal(order)
    rm.redis.Set(ctx, "order:"+order.ID, orderData, 1*time.Hour)
    
    return nil
}

func (rm *orderReadModel) UpdateOrder(ctx context.Context, order *OrderDTO) error {
    return rm.CreateOrder(ctx, order) // Same as create due to UPSERT
}

func (rm *orderReadModel) DeleteOrder(ctx context.Context, orderID string) error {
    query := `DELETE FROM order_read_models WHERE id = $1`
    
    _, err := rm.db.ExecContext(ctx, query, orderID)
    if err != nil {
        return fmt.Errorf("failed to delete order: %w", err)
    }
    
    // Remove from cache
    rm.redis.Del(ctx, "order:"+orderID)
    
    return nil
}

func (rm *orderReadModel) ListOrders(ctx context.Context, customerID string, limit, offset int) ([]*OrderDTO, error) {
    query := `
        SELECT id, customer_id, status, total_amount, shipping_address, items, created_at, updated_at
        FROM order_read_models
        WHERE customer_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `
    
    rows, err := rm.db.QueryContext(ctx, query, customerID, limit, offset)
    if err != nil {
        return nil, fmt.Errorf("failed to query orders: %w", err)
    }
    defer rows.Close()
    
    var orders []*OrderDTO
    for rows.Next() {
        var order OrderDTO
        var shippingAddressJSON, itemsJSON string
        
        err := rows.Scan(
            &order.ID,
            &order.CustomerID,
            &order.Status,
            &order.TotalAmount.Amount,
            &shippingAddressJSON,
            &itemsJSON,
            &order.CreatedAt,
            &order.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan order: %w", err)
        }
        
        // Parse JSON fields
        json.Unmarshal([]byte(shippingAddressJSON), &order.ShippingAddress)
        json.Unmarshal([]byte(itemsJSON), &order.Items)
        
        order.TotalAmount.Currency = "USD"
        for i := range order.Items {
            order.Items[i].Price.Currency = "USD"
        }
        
        orders = append(orders, &order)
    }
    
    return orders, nil
}

func (rm *orderReadModel) GetOrderAnalytics(ctx context.Context, period string) (*OrderAnalyticsDTO, error) {
    // This is a simplified analytics query
    // In production, you might want to use a separate analytics database or data warehouse
    
    var whereClause string
    switch period {
    case "daily":
        whereClause = "created_at >= CURRENT_DATE"
    case "weekly":
        whereClause = "created_at >= CURRENT_DATE - INTERVAL '7 days'"
    case "monthly":
        whereClause = "created_at >= CURRENT_DATE - INTERVAL '30 days'"
    default:
        whereClause = "1=1" // All time
    }
    
    // Get total orders and revenue
    query := fmt.Sprintf(`
        SELECT 
            COUNT(*) as total_orders,
            COALESCE(SUM(total_amount), 0) as total_revenue,
            COALESCE(AVG(total_amount), 0) as average_order_value
        FROM order_read_models
        WHERE %s
    `, whereClause)
    
    var analytics OrderAnalyticsDTO
    err := rm.db.QueryRowContext(ctx, query).Scan(
        &analytics.TotalOrders,
        &analytics.TotalRevenue,
        &analytics.AverageOrderValue,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to get analytics: %w", err)
    }
    
    // Get orders by status
    statusQuery := fmt.Sprintf(`
        SELECT status, COUNT(*)
        FROM order_read_models
        WHERE %s
        GROUP BY status
    `, whereClause)
    
    rows, err := rm.db.QueryContext(ctx, statusQuery)
    if err != nil {
        return nil, fmt.Errorf("failed to get status analytics: %w", err)
    }
    defer rows.Close()
    
    analytics.OrdersByStatus = make(map[string]int64)
    for rows.Next() {
        var status string
        var count int64
        if err := rows.Scan(&status, &count); err != nil {
            return nil, fmt.Errorf("failed to scan status: %w", err)
        }
        analytics.OrdersByStatus[status] = count
    }
    
    return &analytics, nil
}
