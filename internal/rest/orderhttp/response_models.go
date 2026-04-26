package orderhttp

type ordersResponse struct {
	OrderTotalCount int64               `json:"orderTotalCount"`
	OrdersForSignal map[string][]int64  `json:"ordersForSignal"`
	OrderCounts     map[string]int      `json:"orderCounts"`
	CountPerPage    int                 `json:"countPerPage"`
	Orders          []orderViewResponse `json:"orders"`
}

type allOrdersResponse struct {
	OrderTotalCount int64               `json:"orderTotalCount"`
	CountPerPage    int                 `json:"countPerPage"`
	Orders          []orderViewResponse `json:"orders"`
}

type orderViewResponse struct {
	ID             int64             `json:"id"`
	OrderNumber    any               `json:"order_number"`
	OrderIDForSort int64             `json:"orderId_forSort"`
	Status         statusResponse    `json:"status"`
	DateForSort    string            `json:"dateForSort"`
	Date           string            `json:"date"`
	Address        []addressResponse `json:"address"`
	CityID         int64             `json:"cityId"`
	Phone          string            `json:"phone"`
	Device         string            `json:"device"`
	DeviceName     string            `json:"deviceName"`
	Client         clientResponse    `json:"client"`
	Dispatcher     any               `json:"dispatcher"`
	Worker         *workerResponse   `json:"worker"`
	Car            *carResponse      `json:"car"`
	Tariff         tariffResponse    `json:"tariff"`
	Options        []optionResponse  `json:"options"`
	Comment        *string           `json:"comment"`
	SummaryCost    any               `json:"summaryCost"`
	StatusTime     int64             `json:"status_time"`
	TimeToClient   *int64            `json:"time_to_client"`
	WaitTime       int64             `json:"wait_time"`
	CreateTime     int64             `json:"create_time"`
	OrderTime      int64             `json:"order_time"`
	PositionID     int64             `json:"positionId"`
	UnitQuantity   *float64          `json:"unit_quantity,omitempty"`
}

type statusResponse struct {
	StatusID int64  `json:"statusId"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Color    string `json:"color"`
}

type clientResponse struct {
	ClientID int64   `json:"clientId"`
	Phone    *string `json:"phone"`
	Name     *string `json:"name"`
	LastName *string `json:"lastName"`
}

type addressResponse struct {
	ID      *string `json:"id"`
	City    *string `json:"city"`
	Street  *string `json:"street"`
	Label   *string `json:"label"`
	House   *string `json:"house"`
	Apt     *string `json:"apt"`
	Parking *string `json:"parking"`
	Type    string  `json:"type"`
}

type workerResponse struct {
	WorkerID int64   `json:"workerId"`
	Callsign *int64  `json:"callsign"`
	Name     string  `json:"name"`
	Phone    *string `json:"phone"`
}

type carResponse struct {
	CarID  int64   `json:"carId"`
	Name   *string `json:"name"`
	Color  *int64  `json:"color"`
	Number *string `json:"number"`
}

type tariffResponse struct {
	TariffID          int64    `json:"tariffId"`
	Name              string   `json:"name"`
	QuantitativeTitle *string  `json:"quantitative_title,omitempty"`
	PriceForUnit      *float64 `json:"price_for_unit,omitempty"`
	UnitName          *string  `json:"unit_name,omitempty"`
}

type optionResponse struct {
	OptionID int64  `json:"option_id"`
	Name     string `json:"name"`
	Quantity int64  `json:"quantity"`
}
