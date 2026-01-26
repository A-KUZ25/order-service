package orderhttp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"orders-service/order"

	"golang.org/x/sync/errgroup"
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
	start := time.Now()
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
		SelectForDate:  req.SelectForDate,
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

	count, orders, err := h.service.GetFormattedOrdersByGroup(ctx, f, page, pageSize)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	preparedOrder, err := h.service.PrepareOrdersData(ctx, orders, f)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Формируем ответ (пример структуры)
	resp := struct {
		TotalCount int64                 `json:"total_count"`
		Orders     []order.PreparedOrder `json:"orders"`
		Page       int                   `json:"page"`
		PageSize   int                   `json:"page_size"`
	}{
		TotalCount: count,
		Orders:     preparedOrder,
		Page:       page,
		PageSize:   pageSize,
	}
	log.Println("Execution OrdersByGroup took:", time.Since(start))

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) OrderForTabs(w http.ResponseWriter, r *http.Request) {
	var req WarningFullRequest
	start := time.Now()
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
		SelectForDate:  req.SelectForDate,
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
	ctx := r.Context()
	result, err := h.service.GetOrdersForTabs(ctx, f)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		OrderCounts     map[order.StatusGroup]int     `json:"order_counts"`
		OrdersForSignal map[order.StatusGroup][]int64 `json:"orders_for_signal"`
	}{
		OrderCounts:     result.GroupCounts,
		OrdersForSignal: result.OrdersForSignal,
	}
	log.Println("Execution OrderForTabs took:", time.Since(start))
	writeJSON(w, http.StatusOK, resp)

}

type ordersResponse struct {
	OrderTotalCount int64                         `json:"orderTotalCount"`
	OrdersForSignal map[order.StatusGroup][]int64 `json:"ordersForSignal"`
	OrderCounts     map[order.StatusGroup]int     `json:"orderCounts"`
	CountPerPage    int                           `json:"countPerPage"`
	Orders          []order.PreparedOrder         `json:"orders"`
}

func (h *Handler) Orders(w http.ResponseWriter, r *http.Request) {
	var req WarningFullRequest
	start := time.Now()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	f := buildWarningFilter(req)

	page := req.Page
	if page < 0 {
		page = 0
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	ctx := r.Context()

	var (
		totalCount int64
		prepared   []order.PreparedOrder
		tabs       order.GroupOrdersResult
	)

	g, gctx := errgroup.WithContext(ctx)

	// Ветка 1: orders
	g.Go(func() error {
		t0 := time.Now()

		count, formatted, err := h.service.GetFormattedOrdersByGroup(gctx, f, page, pageSize)
		if err != nil {
			return err
		}

		p, err := h.service.PrepareOrdersData(gctx, formatted, f)
		if err != nil {
			return err
		}

		totalCount = count
		prepared = p

		log.Println("Orders branch took:", time.Since(t0))
		return nil
	})

	// Ветка 2: tabs
	g.Go(func() error {
		t0 := time.Now()

		res, err := h.service.GetOrdersForTabs(gctx, f)
		if err != nil {
			return err
		}

		tabs = res

		log.Println("Tabs branch took:", time.Since(t0))
		return nil
	})

	if err := g.Wait(); err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := ordersResponse{
		OrderTotalCount: totalCount,
		OrdersForSignal: tabs.OrdersForSignal,
		OrderCounts:     tabs.GroupCounts,
		CountPerPage:    pageSize,
		Orders:          prepared,
	}

	log.Println("Execution Orders took:", time.Since(start))
	writeJSON(w, http.StatusOK, resp)
}

func buildWarningFilter(req WarningFullRequest) order.WarningFilter {
	base := order.BaseFilter{
		TenantID:       req.TenantID,
		CityIDs:        req.CityIDs,
		Date:           req.Date,
		StatusTimeFrom: req.StatusTimeFrom,
		StatusTimeTo:   req.StatusTimeTo,
		SelectForDate:  req.SelectForDate,
		Tariffs:        req.Tariffs,
		UserPositions:  req.UserPositions,
		SortField:      req.SortField,
		SortOrder:      req.SortOrder,
		Status:         req.Status,
		Group:          req.Group,
	}

	return order.WarningFilter{
		BaseFilter:             base,
		WarningStatus:          req.WarningStatus,
		FinishedStatus:         req.FinishedStatus,
		BadRatingMax:           req.BadRatingMax,
		StatusCompletedNotPaid: req.StatusCompletedNotPaid,
		MinRealPrice:           req.MinRealPrice,
	}
}
