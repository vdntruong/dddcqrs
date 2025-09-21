package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vdntruong/dddcqrs/shared/domain/entities"
)

type ConfirmOrderHandler struct {
    Service *CommandService
}

func (h *ConfirmOrderHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    orderID := entities.OrderID(vars["id"])
    
    if err := h.Service.ConfirmOrder(r.Context(), orderID); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Order confirmed successfully"))
}
