package orderhttp

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Post("/orders/unpaid", handler.Unpaid)
	r.Post("/orders/bad-review", handler.BadReview)
	r.Post("/orders/exceeded-price", handler.ExceededPrice)
	r.Post("/orders/warning", handler.GetWarningOrders)
	r.Post("/orders/order-by-group", handler.OrdersByGroup)
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
}
