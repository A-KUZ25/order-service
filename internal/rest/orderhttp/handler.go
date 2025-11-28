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

func (h *Handler) Unpaid(w http.ResponseWriter, r *http.Request) {
	var req UnpaidRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	filter := order.UnpaidFilter{
		BaseFilter: order.BaseFilter{
			TenantID:       req.TenantID,
			CityIDs:        req.CityIDs,
			Date:           req.Date,
			StatusTimeFrom: req.StatusTimeFrom,
			StatusTimeTo:   req.StatusTimeTo,
			Status:         req.Status,
			Tariffs:        req.Tariffs,
			UserPositions:  req.UserPositions,
			SortField:      req.SortField,
			SortOrder:      req.SortOrder,
		},
		StatusCompletedNotPaid: req.StatusCompletedNotPaid,
	}

	ctx := r.Context()

	ids, err := h.service.GetUnpaid(ctx, filter)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, UnpaidResponse{
		UnpaidIDs: ids,
	})

}

func (h *Handler) BadReview(w http.ResponseWriter, r *http.Request) {
	var req BadReviewRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	filter := order.BadReviewFilter{
		BaseFilter: order.BaseFilter{
			TenantID:       req.TenantID,
			CityIDs:        req.CityIDs,
			Date:           req.Date,
			StatusTimeFrom: req.StatusTimeFrom,
			StatusTimeTo:   req.StatusTimeTo,
			Status:         req.Status,
			Tariffs:        req.Tariffs,
			UserPositions:  req.UserPositions,
			SortField:      req.SortField,
			SortOrder:      req.SortOrder,
		},
		BadRatingMax: req.BadRatingMax,
	}

	ctx := r.Context()

	ids, err := h.service.GetBadReview(ctx, filter)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, BadReviewResponse{
		BadReviewIDs: ids,
	})

}

func (h *Handler) ExceededPrice(w http.ResponseWriter, r *http.Request) {
	var req ExceededPriceRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	filter := order.ExceededPriceFilter{
		BaseFilter: order.BaseFilter{
			TenantID:       req.TenantID,
			CityIDs:        req.CityIDs,
			Date:           req.Date,
			StatusTimeFrom: req.StatusTimeFrom,
			StatusTimeTo:   req.StatusTimeTo,
			Status:         req.Status,
			Tariffs:        req.Tariffs,
			UserPositions:  req.UserPositions,
			SortField:      req.SortField,
			SortOrder:      req.SortOrder,
		},
		MinRealPrice:   req.MinRealPrice,
		FinishedStatus: req.FinishedStatus,
	}

	ids, err := h.service.GetExceededPrice(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, ExceededPriceResponse{
		PriceIDs: ids,
	})
}
