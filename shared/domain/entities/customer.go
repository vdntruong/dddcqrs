package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vdntruong/dddcqrs/shared/domain/valueobjects"
)

type CustomerID string

type Customer struct {
    ID        CustomerID
    Email     string
    Name      string
    Addresses []valueobjects.Address
    CreatedAt time.Time
    UpdatedAt time.Time
}

func NewCustomer(email, name string) *Customer {
    return &Customer{
        ID:        CustomerID(uuid.New().String()),
        Email:     email,
        Name:      name,
        Addresses: []valueobjects.Address{},
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}

func (c *Customer) UpdateEmail(email string) error {
    if email == "" {
        return errors.New("email cannot be empty")
    }
    
    // Basic email validation
    if !isValidEmail(email) {
        return errors.New("invalid email format")
    }
    
    c.Email = email
    c.UpdatedAt = time.Now()
    
    return nil
}

func (c *Customer) UpdateName(name string) error {
    if name == "" {
        return errors.New("name cannot be empty")
    }
    
    c.Name = name
    c.UpdatedAt = time.Now()
    
    return nil
}

func (c *Customer) AddAddress(address valueobjects.Address) error {
    if err := address.Validate(); err != nil {
        return err
    }
    
    c.Addresses = append(c.Addresses, address)
    c.UpdatedAt = time.Now()
    
    return nil
}

func (c *Customer) RemoveAddress(index int) error {
    if index < 0 || index >= len(c.Addresses) {
        return errors.New("invalid address index")
    }
    
    c.Addresses = append(c.Addresses[:index], c.Addresses[index+1:]...)
    c.UpdatedAt = time.Now()
    
    return nil
}

func isValidEmail(email string) bool {
    // Simple email validation - in production, use a proper email validation library
    return len(email) > 0 && len(email) < 255
}
