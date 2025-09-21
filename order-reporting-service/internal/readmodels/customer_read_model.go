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

type CustomerReadModel interface {
    GetCustomer(ctx context.Context, customerID string) (*CustomerDTO, error)
    CreateCustomer(ctx context.Context, customer *CustomerDTO) error
    UpdateCustomer(ctx context.Context, customer *CustomerDTO) error
    DeleteCustomer(ctx context.Context, customerID string) error
}

type CustomerDTO struct {
    ID        string                `json:"id"`
    Email     string                `json:"email"`
    Name      string                `json:"name"`
    Addresses []valueobjects.Address `json:"addresses"`
    CreatedAt time.Time             `json:"created_at"`
    UpdatedAt time.Time             `json:"updated_at"`
}

type customerReadModel struct {
    db    *sql.DB
    redis *redis.Client
}

func NewCustomerReadModel(db *sql.DB, redis *redis.Client) CustomerReadModel {
    return &customerReadModel{
        db:    db,
        redis: redis,
    }
}

func (rm *customerReadModel) GetCustomer(ctx context.Context, customerID string) (*CustomerDTO, error) {
    // Try cache first
    cacheKey := "customer:" + customerID
    cached, err := rm.redis.Get(ctx, cacheKey).Result()
    if err == nil {
        var customer CustomerDTO
        if err := json.Unmarshal([]byte(cached), &customer); err == nil {
            return &customer, nil
        }
    }
    
    // Fallback to database
    query := `
        SELECT id, email, name, addresses, created_at, updated_at
        FROM customer_read_models
        WHERE id = $1
    `
    
    var customer CustomerDTO
    var addressesJSON string
    
    err = rm.db.QueryRowContext(ctx, query, customerID).Scan(
        &customer.ID,
        &customer.Email,
        &customer.Name,
        &addressesJSON,
        &customer.CreatedAt,
        &customer.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("customer not found")
        }
        return nil, fmt.Errorf("failed to find customer: %w", err)
    }
    
    // Parse addresses
    if err := json.Unmarshal([]byte(addressesJSON), &customer.Addresses); err != nil {
        return nil, fmt.Errorf("failed to unmarshal addresses: %w", err)
    }
    
    // Cache the result
    customerData, _ := json.Marshal(customer)
    rm.redis.Set(ctx, cacheKey, customerData, 1*time.Hour)
    
    return &customer, nil
}

func (rm *customerReadModel) CreateCustomer(ctx context.Context, customer *CustomerDTO) error {
    addressesJSON, err := json.Marshal(customer.Addresses)
    if err != nil {
        return fmt.Errorf("failed to marshal addresses: %w", err)
    }
    
    query := `
        INSERT INTO customer_read_models (id, email, name, addresses, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE SET
            email = $2,
            name = $3,
            addresses = $4,
            updated_at = $6
    `
    
    _, err = rm.db.ExecContext(ctx, query,
        customer.ID,
        customer.Email,
        customer.Name,
        addressesJSON,
        customer.CreatedAt,
        customer.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to save customer: %w", err)
    }
    
    // Cache the result
    customerData, _ := json.Marshal(customer)
    rm.redis.Set(ctx, "customer:"+customer.ID, customerData, 1*time.Hour)
    
    return nil
}

func (rm *customerReadModel) UpdateCustomer(ctx context.Context, customer *CustomerDTO) error {
    return rm.CreateCustomer(ctx, customer) // Same as create due to UPSERT
}

func (rm *customerReadModel) DeleteCustomer(ctx context.Context, customerID string) error {
    query := `DELETE FROM customer_read_models WHERE id = $1`
    
    _, err := rm.db.ExecContext(ctx, query, customerID)
    if err != nil {
        return fmt.Errorf("failed to delete customer: %w", err)
    }
    
    // Remove from cache
    rm.redis.Del(ctx, "customer:"+customerID)
    
    return nil
}
