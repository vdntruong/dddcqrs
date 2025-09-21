package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vdntruong/dddcqrs/order-reporting-service/internal/readmodels"
)

type GetOrderHandler struct {
    ReadModel readmodels.OrderReadModel
}

func (h *GetOrderHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    orderID := vars["id"]
    
    order, err := h.ReadModel.GetOrder(r.Context(), orderID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}
