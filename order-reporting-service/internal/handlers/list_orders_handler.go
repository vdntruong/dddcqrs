package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/vdntruong/dddcqrs/order-reporting-service/internal/readmodels"
)

type ListOrdersHandler struct {
    ReadModel readmodels.OrderReadModel
}

func (h *ListOrdersHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    customerID := r.URL.Query().Get("customer_id")
    if customerID == "" {
        http.Error(w, "customer_id parameter is required", http.StatusBadRequest)
        return
    }
    
    // Parse pagination parameters
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")
    
    limit := 10 // default
    offset := 0  // default
    
    if limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
            limit = l
        }
    }
    
    if offsetStr != "" {
        if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
            offset = o
        }
    }
    
    orders, err := h.ReadModel.ListOrders(r.Context(), customerID, limit, offset)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "orders": orders,
        "pagination": map[string]interface{}{
            "limit":  limit,
            "offset": offset,
            "count":  len(orders),
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
