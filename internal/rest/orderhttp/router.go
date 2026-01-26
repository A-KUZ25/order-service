package orderhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Post("/orders/order-by-group", handler.OrdersByGroup)
	r.Post("/orders/order-for-tabs", handler.OrderForTabs)
	r.Post("/orders", handler.Orders)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
}
