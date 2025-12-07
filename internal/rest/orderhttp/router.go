package orderhttp

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler *Handler) {
	r.Post("/orders/unpaid", handler.Unpaid)
	r.Post("/orders/bad-review", handler.BadReview)
	r.Post("/orders/exceeded-price", handler.ExceededPrice)
	r.Post("/orders/warning", handler.GetWarningOrders)
	r.Post("/orders/warning-paginated", handler.GetWarningPaginated)
}
