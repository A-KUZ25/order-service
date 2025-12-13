package orderhttp

import (
	"encoding/json"
	"net/http"

	"orders-service/order"
)

type Handler struct {
	service order.Service
}

func NewHandler(service order.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) OrdersByGroup(w http.ResponseWriter, r *http.Request) {
	var req WarningFullRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	base := order.BaseFilter{
		TenantID:       req.TenantID,
		CityIDs:        req.CityIDs,
		Date:           req.Date,
		StatusTimeFrom: req.StatusTimeFrom,
		StatusTimeTo:   req.StatusTimeTo,
		Tariffs:        req.Tariffs,
		UserPositions:  req.UserPositions,
		SortField:      req.SortField,
		SortOrder:      req.SortOrder,
		Status:         req.Status,
		Group:          req.Group,
	}

	f := order.WarningFilter{
		BaseFilter:             base,
		WarningStatus:          req.WarningStatus,
		FinishedStatus:         req.FinishedStatus,
		BadRatingMax:           req.BadRatingMax,
		StatusCompletedNotPaid: req.StatusCompletedNotPaid,
		MinRealPrice:           req.MinRealPrice,
	}

	page := req.Page
	if page < 0 {
		page = 0
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	ctx := r.Context()

	count, orders, err := h.service.GetOrdersByGroup(ctx, f, page, pageSize)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Формируем ответ (пример структуры)
	resp := struct {
		TotalCount int64             `json:"total_count"`
		Orders     []order.FullOrder `json:"orders"`
		Page       int               `json:"page"`
		PageSize   int               `json:"page_size"`
	}{
		TotalCount: count,
		Orders:     orders,
		Page:       page,
		PageSize:   pageSize,
	}

	writeJSON(w, http.StatusOK, resp)
}
