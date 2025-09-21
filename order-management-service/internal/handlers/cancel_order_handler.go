package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vdntruong/dddcqrs/shared/domain/entities"
)

type CancelOrderHandler struct {
    Service *CommandService
}

type CancelOrderRequest struct {
    Reason string `json:"reason"`
}

func (h *CancelOrderHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    orderID := entities.OrderID(vars["id"])
    
    var req CancelOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    if err := h.Service.CancelOrder(r.Context(), orderID, req.Reason); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Order cancelled successfully"))
}
