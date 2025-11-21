package orders

type UnpaidOrdersRequest struct {
	CityID        *int64   `json:"city_id"`
	Date          *string  `json:"date"`
	Status        []string `json:"status"`
	Tariffs       []int64  `json:"tariffs"`
	UserPositions []int64  `json:"user_positions"`
	Sort          string   `json:"sort"`
	Order         string   `json:"order"`
}

type UnpaidOrdersResponse struct {
	UnpaidOrderIDs []int64 `json:"unpaid_order_ids"`
}
