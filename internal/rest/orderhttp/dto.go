package orderhttp

type OrderBaseRequest struct {
	TenantID       int64    `json:"tenant_id"`
	CityIDs        []int64  `json:"city_ids"`
	Date           *string  `json:"date"`
	StatusTimeFrom *int64   `json:"status_time_from"`
	StatusTimeTo   *int64   `json:"status_time_to"`
	Status         []string `json:"status"`
	Tariffs        []int64  `json:"tariffs"`
	UserPositions  []int64  `json:"user_positions"`
	SortField      string   `json:"sort_field"`
	SortOrder      string   `json:"sort_order"`
}

type UnpaidRequest struct {
	OrderBaseRequest             //джсон сам мапиться в соответствии с полями
	StatusCompletedNotPaid int64 `json:"status_completed_not_paid"`
}

type BadReviewRequest struct {
	OrderBaseRequest
	BadRatingMax int64 `json:"bad_rating_max"`
}

type ExceededPriceRequest struct {
	OrderBaseRequest
	MinRealPrice   float64 `json:"min_real_price"`
	FinishedStatus []int64 `json:"finished_status"`
}

type WarningRequest struct {
	OrderBaseRequest
	WarningStatus          []int64 `json:"warning_status"`
	StatusCompletedNotPaid int64   `json:"status_completed_not_paid"`
	BadRatingMax           int64   `json:"bad_rating_max"`
	MinRealPrice           float64 `json:"min_real_price"`
	FinishedStatus         []int64 `json:"finished_status"`
}

type UnpaidResponse struct {
	UnpaidIDs []int64 `json:"unpaid_order_ids"`
}

type BadReviewResponse struct {
	BadReviewIDs []int64 `json:"bad_review_ids"`
}

type ExceededPriceResponse struct {
	PriceIDs []int64 `json:"price_ids"`
}

type WarningResponse struct {
	WarningOrderIDs []int64 `json:"warning_order_ids"`
}
