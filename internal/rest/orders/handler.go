package orders

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	//service OrdersService todo добавим позже
}

func (h *Handler) GetUnpaidOrders(w http.ResponseWriter, r *http.Request) {
	var req UnpaidOrdersRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// На данном этапе отдаём заглушку
	resp := UnpaidOrdersResponse{
		UnpaidOrderIDs: []int64{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
