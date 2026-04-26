package orderhttp

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"orders-service/internal/app/order"

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
		prepared   []order.OrderView
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

	resp := buildOrdersResponse(totalCount, pageSize, tabs, prepared)

	log.Println("Execution Orders took:", time.Since(start))
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) AllOrders(w http.ResponseWriter, r *http.Request) {
	var req GetAllOrdersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	page := req.Page
	if page < 0 {
		page = 0
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}

	attributes := make([]order.SearchAttribute, 0, len(req.Attributes))
	for _, attribute := range req.Attributes {
		attributes = append(attributes, order.SearchAttribute{
			Attribute:    attribute.Attribute,
			SearchString: attribute.SearchString,
		})
	}

	result, err := h.service.GetAllOrders(r.Context(), order.GetAllOrdersFilter{
		BaseFilter: order.BaseFilter{
			TenantID:  req.TenantID,
			CityIDs:   req.CityIDs,
			Language:  req.Language,
			Date:      req.Date,
			Tariffs:   req.Tariffs,
			SortField: req.SortField,
			SortOrder: req.SortOrder,
		},
		Page:         page,
		PageSize:     pageSize,
		SearchStatus: req.SearchStatus,
		Attributes:   attributes,
		SearchString: req.SearchString,
		ShopIDs:      req.ShopIDs,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, buildAllOrdersResponse(result))
}

func buildWarningFilter(req WarningFullRequest) order.WarningFilter {
	base := order.BaseFilter{
		TenantID:       req.TenantID,
		CityIDs:        req.CityIDs,
		Language:       req.Language,
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
