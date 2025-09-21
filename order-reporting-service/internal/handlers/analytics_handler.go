package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/vdntruong/dddcqrs/order-reporting-service/internal/readmodels"
)

type GetOrderAnalyticsHandler struct {
    ReadModel readmodels.OrderReadModel
}

func (h *GetOrderAnalyticsHandler) HandleHTTP(w http.ResponseWriter, r *http.Request) {
    period := r.URL.Query().Get("period")
    if period == "" {
        period = "monthly" // default
    }
    
    // Validate period
    validPeriods := map[string]bool{
        "daily":   true,
        "weekly":  true,
        "monthly": true,
        "all":     true,
    }
    
    if !validPeriods[period] {
        http.Error(w, "Invalid period. Must be one of: daily, weekly, monthly, all", http.StatusBadRequest)
        return
    }
    
    analytics, err := h.ReadModel.GetOrderAnalytics(r.Context(), period)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    response := map[string]interface{}{
        "period": period,
        "analytics": analytics,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
