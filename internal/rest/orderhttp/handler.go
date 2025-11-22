package orderhttp

import (
	"encoding/json"
	"net/http"

	"orders-service/order"
)

// Handler — HTTP-обёртка над доменным сервисом заказов.
type Handler struct {
	service order.Service
}

// NewHandler — конструктор хендлера.
func NewHandler(service order.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetUnpaidOrders — HTTP-эндпоинт /orderhttp/unpaid.
func (h *Handler) GetUnpaidOrders(w http.ResponseWriter, r *http.Request) {
	var req UnpaidOrdersRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error":       "invalid_request_body",
			"description": err.Error(),
		})
		return
	}
	filter := order.UnpaidOrdersFilter{
		TenantID:               req.TenantID,
		CityIDs:                req.CityIDs,
		Date:                   req.Date,
		StatusTimeFrom:         req.StatusTimeFrom,
		StatusTimeTo:           req.StatusTimeTo,
		Status:                 req.Status,
		Tariffs:                req.Tariffs,
		UserPositions:          req.UserPositions,
		SortField:              req.Sort,
		SortOrder:              order.SortOrder(req.Order),
		StatusCompletedNotPaid: req.StatusCompletedNotPaid,
	}

	ctx := r.Context()

	ids, err := h.service.GetUnpaidOrderIDs(ctx, filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error":       "invalid_request_body",
			"description": err.Error(),
		})
		return
	}

	resp := UnpaidOrdersResponse{
		UnpaidOrderIDs: ids,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
