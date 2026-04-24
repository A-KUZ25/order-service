package order

type OrderView struct {
	ID             int64           `json:"id"`
	OrderNumber    any             `json:"order_number"`
	OrderIDForSort int64           `json:"orderId_forSort"`
	Status         OrderStatusView `json:"status"`
	DateForSort    string          `json:"dateForSort"`
	Date           string          `json:"date"`
	Address        []AddressView   `json:"address"`
	CityID         int64           `json:"cityId"`
	Phone          string          `json:"phone"`
	Device         string          `json:"device"`
	DeviceName     string          `json:"deviceName"`
	Client         ClientView      `json:"client"`
	Dispatcher     any             `json:"dispatcher"`
	Worker         *WorkerView     `json:"worker"`
	Car            *CarView        `json:"car"`
	Tariff         TariffView      `json:"tariff"`
	Options        []OptionDTO     `json:"options"`
	Comment        *string         `json:"comment"`
	SummaryCost    any             `json:"summaryCost"`
	StatusTime     int64           `json:"status_time"`
	TimeToClient   *int64          `json:"time_to_client"`
	WaitTime       int64           `json:"wait_time"`
	CreateTime     int64           `json:"create_time"`
	OrderTime      int64           `json:"order_time"`
	PositionID     int64           `json:"positionId"`
	UnitQuantity   *float64        `json:"unit_quantity,omitempty"`
}

type OrderStatusView struct {
	StatusID int64  `json:"statusId"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Color    string `json:"color"`
}

type ClientView struct {
	ClientID int64   `json:"clientId"`
	Phone    *string `json:"phone"`
	Name     *string `json:"name"`
	LastName *string `json:"lastName"`
}

type AddressView struct {
	ID      *string `json:"id"`
	City    *string `json:"city"`
	Street  *string `json:"street"`
	Label   *string `json:"label"`
	House   *string `json:"house"`
	Apt     *string `json:"apt"`
	Parking *string `json:"parking"`
	Type    string  `json:"type"`
}

type WorkerView struct {
	WorkerID int64   `json:"workerId"`
	Callsign *int64  `json:"callsign"`
	Name     string  `json:"name"`
	Phone    *string `json:"phone"`
}

type CarView struct {
	CarID  int64   `json:"carId"`
	Name   *string `json:"name"`
	Color  *int64  `json:"color"`
	Number *string `json:"number"`
}

type TariffView struct {
	TariffID          int64    `json:"tariffId"`
	Name              string   `json:"name"`
	QuantitativeTitle *string  `json:"quantitative_title,omitempty"`
	PriceForUnit      *float64 `json:"price_for_unit,omitempty"`
	UnitName          *string  `json:"unit_name,omitempty"`
}
