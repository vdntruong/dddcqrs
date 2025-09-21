package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vdntruong/dddcqrs/shared/domain/entities"
)

type UpdateOrderHandler struct {
    Service *CommandService
}

func (h *UpdateOrderHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    orderID := entities.OrderID(vars["id"])
    
    var cmd UpdateOrderCommand
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    cmd.OrderID = string(orderID)
    
    // For now, we'll implement a simple update that replaces all items
    // In a real application, you might want more granular updates
    
    // Load existing order
    order, err := h.Service.OrderRepo.FindByID(r.Context(), orderID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    
    // Clear existing items
    order.Items = []entities.OrderItem{}
    
    // Add new items
    for _, item := range cmd.Items {
        if err := order.AddItem(item.ProductID, item.Quantity, item.Price); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
    }
    
    // Update shipping address
    order.ShippingAddress = cmd.ShippingAddress
    
    // Save updated order
    if err := h.Service.OrderRepo.Update(r.Context(), order); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Order updated successfully"))
}
