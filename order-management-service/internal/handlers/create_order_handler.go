package handlers

import (
	"encoding/json"
	"net/http"
)

type CreateOrderHandler struct {
    Service *CommandService
}

func (h *CreateOrderHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    var cmd CreateOrderCommand
    
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    order, err := h.Service.CreateOrder(r.Context(), cmd)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    
    response := map[string]interface{}{
        "id":         order.ID,
        "customer_id": order.CustomerID,
        "status":     order.Status.String(),
        "total_amount": order.TotalAmount,
        "created_at": order.CreatedAt,
    }
    
    json.NewEncoder(w).Encode(response)
}
