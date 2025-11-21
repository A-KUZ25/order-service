package orderhttp

type UnpaidOrdersRequest struct {
	TenantID               int64    `json:"tenant_id"`
	CityIDs                []int64  `json:"city_ids"`
	Date                   *string  `json:"date"`
	StatusTimeFrom         *int64   `json:"status_time_from"`
	StatusTimeTo           *int64   `json:"status_time_to"`
	Status                 []string `json:"status"`
	Tariffs                []int64  `json:"tariffs"`
	UserPositions          []int64  `json:"user_positions"`
	Sort                   string   `json:"sort"`
	Order                  string   `json:"order"`
	StatusCompletedNotPaid int64    `json:"status_completed_not_paid"`
}

type UnpaidOrdersResponse struct {
	UnpaidOrderIDs []int64 `json:"unpaid_order_ids"`
}
