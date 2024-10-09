package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/paudarco/orders-db-task/internal/cache"
)

type Handler struct {
	cache *cache.Cache
}

func NewHandler(cache *cache.Cache) *Handler {
	return &Handler{cache: cache}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	order, found := h.cache.Get(orderID)
	if !found {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
