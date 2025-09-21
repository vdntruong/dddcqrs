package valueobjects

import (
	"errors"
	"fmt"
)

type Money struct {
    Amount   int64  `json:"amount"`
    Currency string `json:"currency"`
}

func NewMoney(amount int64, currency string) Money {
    return Money{
        Amount:   amount,
        Currency: currency,
    }
}

func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, errors.New("cannot add different currencies")
    }
    
    return Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }, nil
}

func (m Money) Subtract(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, errors.New("cannot subtract different currencies")
    }
    
    if m.Amount < other.Amount {
        return Money{}, errors.New("insufficient funds")
    }
    
    return Money{
        Amount:   m.Amount - other.Amount,
        Currency: m.Currency,
    }, nil
}

func (m Money) Multiply(factor int64) Money {
    return Money{
        Amount:   m.Amount * factor,
        Currency: m.Currency,
    }
}

func (m Money) IsZero() bool {
    return m.Amount == 0
}

func (m Money) IsPositive() bool {
    return m.Amount > 0
}

func (m Money) IsNegative() bool {
    return m.Amount < 0
}

func (m Money) String() string {
    return fmt.Sprintf("%.2f %s", float64(m.Amount)/100, m.Currency)
}

func (m Money) Validate() error {
    if m.Currency == "" {
        return errors.New("currency cannot be empty")
    }
    
    if len(m.Currency) != 3 {
        return errors.New("currency must be 3 characters")
    }
    
    return nil
}
