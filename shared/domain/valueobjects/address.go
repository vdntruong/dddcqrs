package valueobjects

import (
	"errors"
	"strings"
)

type Address struct {
    Street string `json:"street"`
    City   string `json:"city"`
    State  string `json:"state"`
    Zip    string `json:"zip"`
    Country string `json:"country"`
}

func NewAddress(street, city, state, zip, country string) Address {
    return Address{
        Street:  street,
        City:    city,
        State:   state,
        Zip:     zip,
        Country: country,
    }
}

func (a Address) Validate() error {
    if strings.TrimSpace(a.Street) == "" {
        return errors.New("street cannot be empty")
    }
    
    if strings.TrimSpace(a.City) == "" {
        return errors.New("city cannot be empty")
    }
    
    if strings.TrimSpace(a.State) == "" {
        return errors.New("state cannot be empty")
    }
    
    if strings.TrimSpace(a.Zip) == "" {
        return errors.New("zip cannot be empty")
    }
    
    if strings.TrimSpace(a.Country) == "" {
        return errors.New("country cannot be empty")
    }
    
    return nil
}

func (a Address) String() string {
    return strings.Join([]string{a.Street, a.City, a.State, a.Zip, a.Country}, ", ")
}

func (a Address) IsEmpty() bool {
    return a.Street == "" && a.City == "" && a.State == "" && a.Zip == "" && a.Country == ""
}
