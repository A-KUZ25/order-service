package orderhttp

type OrderBaseRequest struct {
	TenantID       int64   `json:"tenant_id"`
	CityIDs        []int64 `json:"city_ids"`
	Date           *string `json:"date"`
	StatusTimeFrom *int64  `json:"status_time_from"`
	StatusTimeTo   *int64  `json:"status_time_to"`
	Status         []int64 `json:"status"`
	Tariffs        []int64 `json:"tariffs"`
	UserPositions  []int64 `json:"user_positions"`
	SortField      string  `json:"sort_field"`
	SortOrder      string  `json:"sort_order"`
}

type WarningFullRequest struct {
	OrderBaseRequest
	WarningStatus          []int64 `json:"warning_status"`
	StatusCompletedNotPaid int64   `json:"status_completed_not_paid"`
	BadRatingMax           int64   `json:"bad_rating_max"`
	MinRealPrice           float64 `json:"min_real_price"`
	FinishedStatus         []int64 `json:"finished_status"`
	Page                   int     `json:"page"`
	PageSize               int     `json:"page_size"`
	Group                  string  `json:"group"`
}
