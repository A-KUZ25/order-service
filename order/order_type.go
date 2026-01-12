package order

import "database/sql"

type FullOrder struct {
	// tbl_order (o.*)
	OrderID                    int64
	TenantID                   int64
	WorkerID                   sql.NullInt64
	CarID                      sql.NullInt64
	CityID                     sql.NullInt64
	TariffID                   int64
	UserCreate                 sql.NullInt64
	StatusID                   int64
	UserModified               sql.NullInt64
	CompanyID                  sql.NullInt64
	ParkingID                  sql.NullInt64
	Address                    string
	Comment                    sql.NullString
	PredvPrice                 sql.NullFloat64
	PredvPriceNoDiscount       sql.NullFloat64
	Device                     sql.NullString
	OrderNumber                int64
	Payment                    sql.NullString
	ShowPhone                  sql.NullInt64
	CreateTime                 sql.NullInt64
	StatusTime                 int64
	TimeToClient               sql.NullInt64
	ClientDeviceToken          sql.NullString
	AppID                      sql.NullInt64
	OrderTime                  sql.NullInt64
	PredvDistance              sql.NullFloat64
	PredvTime                  sql.NullInt64
	CallWarningID              sql.NullInt64
	Phone                      sql.NullString
	ClientID                   int64
	BonusPayment               sql.NullInt64
	CurrencyID                 int64
	TimeOffset                 sql.NullInt64
	IsFix                      int64
	UpdateTime                 sql.NullInt64
	DenyRefuseOrder            sql.NullInt64
	PositionID                 int64
	PromoCodeID                sql.NullInt64
	TenantCompanyID            sql.NullInt64
	Mark                       sql.NullInt64
	ProcessedExchangeProgramID sql.NullInt64
	ClientPassengerID          sql.NullInt64
	ClientPassengerPhone       sql.NullString
	Active                     sql.NullInt64
	IsPreOrder                 sql.NullInt64
	AppVersion                 sql.NullString
	AgentCommission            sql.NullFloat64
	IsFixByDispatcher          sql.NullInt64
	FinishTime                 sql.NullInt64
	CommentForDispatcher       sql.NullString
	WorkerManualSurcharge      sql.NullFloat64
	RealtimePrice              sql.NullFloat64
	UnitQuantity               sql.NullFloat64
	ShopID                     sql.NullInt64
	RequirePrepayment          sql.NullInt64
	OrderCode                  sql.NullString
	ClientOfferedPrice         sql.NullFloat64
	IdempotentKey              sql.NullString
	AdditionalTariffID         sql.NullInt64
	InitialPrice               sql.NullFloat64
	TimeToOrder                sql.NullInt64
	Sort                       sql.NullInt64

	// tbl_order_detail_cost (d)
	SummaryCost           sql.NullString
	SummaryCostNoDiscount sql.NullString

	// tbl_order_status (s)
	StatusStatusID int64
	StatusName     string

	// tbl_worker (w)
	WorkerWorkerID   sql.NullInt64
	WorkerCallsign   sql.NullInt64
	WorkerName       sql.NullString
	WorkerLastName   sql.NullString
	WorkerSecondName sql.NullString
	WorkerPhone      sql.NullString

	// tbl_client (cl)
	ClientClientID   sql.NullInt64
	ClientPhone      sql.NullString
	ClientName       sql.NullString
	ClientLastName   sql.NullString
	ClientSecondName sql.NullString

	// tbl_car (car)
	CarCarID     sql.NullInt64
	CarName      sql.NullString
	CarColor     sql.NullInt64
	CarGosNumber sql.NullString

	// tbl_taxi_tariff (t)
	TariffTariffID          sql.NullInt64
	TariffType              sql.NullString
	TariffName              sql.NullString
	TariffQuantitativeTitle sql.NullString
	TariffPriceForUnit      sql.NullFloat64
	TariffUnitName          sql.NullString

	// tbl_user (u)
	UserUserID     sql.NullInt64
	UserName       sql.NullString
	UserLastName   sql.NullString
	UserSecondName sql.NullString

	// tbl_currency (curr)
	CurrencyName   sql.NullString
	CurrencyCode   sql.NullString
	CurrencySymbol sql.NullString
}

type FormattedOrder struct {
	// ===== БАЗОВЫЕ ПОЛЯ =====
	OrderID      int64  `json:"order_id"`
	TenantID     int64  `json:"tenant_id"`
	WorkerID     *int64 `json:"worker_id"`
	CarID        *int64 `json:"car_id"`
	CityID       int64  `json:"city_id"`
	TariffID     int64  `json:"tariff_id"`
	UserCreate   int64  `json:"user_create"`
	StatusID     int64  `json:"status_id"`
	UserModified int64  `json:"user_modifed"`
	CompanyID    *int64 `json:"company_id"`
	ParkingID    *int64 `json:"parking_id"`

	Address any     `json:"address"`
	Comment *string `json:"comment"`

	PredvPrice           float64 `json:"predv_price"`
	PredvPriceNoDiscount float64 `json:"predv_price_no_discount"`

	Device                     string   `json:"device"`
	OrderNumber                int64    `json:"order_number"`
	Payment                    string   `json:"payment"`
	ShowPhone                  int64    `json:"show_phone"`
	CreateTime                 int64    `json:"create_time"`
	StatusTime                 int64    `json:"status_time"`
	TimeToClient               *int64   `json:"time_to_client"`
	ClientDeviceToken          *string  `json:"client_device_token"`
	AppID                      *int64   `json:"app_id"`
	OrderTime                  int64    `json:"order_time"`
	PredvDistance              float64  `json:"predv_distance"`
	PredvTime                  int64    `json:"predv_time"`
	CallWarningID              *int64   `json:"call_warning_id"`
	Phone                      string   `json:"phone"`
	ClientID                   int64    `json:"client_id"`
	BonusPayment               int64    `json:"bonus_payment"`
	CurrencyID                 int64    `json:"currency_id"`
	TimeOffset                 int64    `json:"time_offset"`
	IsFix                      int64    `json:"is_fix"`
	UpdateTime                 int64    `json:"update_time"`
	DenyRefuseOrder            int64    `json:"deny_refuse_order"`
	PositionID                 int64    `json:"position_id"`
	PromoCodeID                *int64   `json:"promo_code_id"`
	TenantCompanyID            *int64   `json:"tenant_company_id"`
	Mark                       int64    `json:"mark"`
	ProcessedExchangeProgramID *int64   `json:"processed_exchange_program_id"`
	ClientPassengerID          *int64   `json:"client_passenger_id"`
	ClientPassengerPhone       *string  `json:"client_passenger_phone"`
	Active                     int64    `json:"active"`
	IsPreOrder                 int64    `json:"is_pre_order"`
	AppVersion                 *string  `json:"app_version"`
	AgentCommission            float64  `json:"agent_commission"`
	IsFixByDispatcher          int64    `json:"is_fix_by_dispatcher"`
	FinishTime                 *int64   `json:"finish_time"`
	CommentForDispatcher       *string  `json:"comment_for_dispatcher"`
	WorkerManualSurcharge      float64  `json:"worker_manual_surcharge"`
	RealtimePrice              *float64 `json:"realtime_price"`
	UnitQuantity               *float64 `json:"unit_quantity"`
	ShopID                     *int64   `json:"shop_id"`
	RequirePrepayment          int64    `json:"require_prepayment"`
	OrderCode                  string   `json:"order_code"`
	ClientOfferedPrice         *float64 `json:"client_offered_price"`
	IdempotentKey              string   `json:"idempotent_key"`
	AdditionalTariffID         *int64   `json:"additional_tariff_id"`
	InitialPrice               *float64 `json:"initial_price"`
	TimeToOrder                *int64   `json:"time_to_order"`
	Sort                       *int64   `json:"sort"`
	SummaryCost                *string  `json:"summary_cost"`
	SummaryCostNoDiscount      *string  `json:"summary_cost_no_discount"`

	// ===== ДУБЛИ СТАТУСА =====
	StatusStatusID int64  `json:"status_status_id"`
	StatusName     string `json:"status_name"`

	// ===== ДАННЫЕ СВЯЗЕЙ =====
	Callsign    *int64  `json:"callsign"`
	WName       *string `json:"wName"`
	WLastName   *string `json:"wLastName"`
	WSecondName *string `json:"wSecondName"`
	WPhone      *string `json:"wPhone"`

	CPhone      *string `json:"cPhone"`
	CName       *string `json:"cName"`
	CLastName   *string `json:"cLastName"`
	CSecondName *string `json:"cSecondName"`

	CarName      *string `json:"car_name"`
	CarColor     *string `json:"car_color"`
	CarGosNumber *string `json:"car_gos_number"`

	TariffType        string  `json:"tariff_type"`
	TName             string  `json:"tName"`
	QuantitativeTitle string  `json:"quantitative_title"`
	PriceForUnit      float64 `json:"price_for_unit"`
	UnitName          string  `json:"unit_name"`

	UserID      int64   `json:"user_id"`
	UName       string  `json:"uName"`
	ULastName   string  `json:"uLastName"`
	USecondName *string `json:"uSecondName"`

	CurrencyName string `json:"currency_name"`
	CurrencyCode string `json:"currency_code"`
	Symbol       string `json:"symbol"`

	// ===== ВЛОЖЕННЫЕ DTO =====
	Status      StatusDTO   `json:"status"`
	Client      ClientDTO   `json:"client"`
	UserCreated UserDTO     `json:"userCreated"`
	Worker      WorkerDTO   `json:"worker"`
	Car         CarDTO      `json:"car"`
	Tariff      TariffDTO   `json:"tariff"`
	Options     []OptionDTO `json:"options"`
	Currency    CurrencyDTO `json:"currency"`
}

type StatusDTO struct {
	StatusID int64  `json:"status_id"`
	Name     string `json:"name"`
}

type ClientDTO struct {
	ClientID   int64   `json:"client_id"`
	Phone      *string `json:"phone"`
	Name       *string `json:"name"`
	LastName   *string `json:"last_name"`
	SecondName *string `json:"second_name"`
}

type UserDTO struct {
	UserID     int64   `json:"user_id"`
	Name       string  `json:"name"`
	LastName   string  `json:"last_name"`
	SecondName *string `json:"second_name"`
}

type WorkerDTO struct {
	WorkerID   int64   `json:"worker_id"`
	Callsign   *int64  `json:"callsign"`
	Name       *string `json:"name"`
	LastName   *string `json:"last_name"`
	SecondName *string `json:"second_name"`
	Phone      *string `json:"phone"`
}

type CarDTO struct {
	CarID     int64   `json:"car_id"`
	Name      *string `json:"name"`
	Color     *int64  `json:"color"`
	GosNumber *string `json:"gos_number"`
}

type TariffDTO struct {
	TariffID          int64   `json:"tariff_id"`
	TariffType        string  `json:"tariff_type"`
	Name              string  `json:"name"`
	QuantitativeTitle string  `json:"quantitative_title"`
	PriceForUnit      float64 `json:"price_for_unit"`
	UnitName          string  `json:"unit_name"`
}

type CurrencyDTO struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

type OptionDTO struct {
	OptionID int64  `json:"option_id"`
	Name     string `json:"name"`
	Quantity int64  `json:"quantity"`
}

func nullableInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func nullableString(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

func nullableFloat64(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	return &v.Float64
}

type PreparedOrder struct {
	ID             int64        `json:"id"`
	OrderNumber    any          `json:"order_number"`
	OrderIDForSort int64        `json:"orderId_forSort"`
	Status         StatusOut    `json:"status"`
	DateForSort    string       `json:"dateForSort"`
	Date           string       `json:"date"`
	Address        []AddressOut `json:"address"`
	CityID         int64        `json:"cityId"`
	Phone          string       `json:"phone"`
	Device         string       `json:"device"`
	DeviceName     string       `json:"deviceName"`
	Client         ClientOut    `json:"client"`
	Dispatcher     any          `json:"dispatcher"`
	Worker         *WorkerOut   `json:"worker"`
	Car            *CarOut      `json:"car"`
	Tariff         TariffOut    `json:"tariff"`
	Options        []OptionDTO  `json:"options"`
	Comment        *string      `json:"comment"`
	SummaryCost    any          `json:"summaryCost"`
	StatusTime     int64        `json:"status_time"`
	TimeToClient   *int64       `json:"time_to_client"`
	WaitTime       int64        `json:"wait_time"`
	CreateTime     int64        `json:"create_time"`
	OrderTime      int64        `json:"order_time"`
	PositionID     int64        `json:"positionId"`
	UnitQuantity   *float64     `json:"unit_quantity,omitempty"`
}

type StatusOut struct {
	StatusID int64  `json:"statusId"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Color    string `json:"color"`
}
type ClientOut struct {
	ClientID int64   `json:"clientId"`
	Phone    *string `json:"phone"`
	Name     *string `json:"name"`
	LastName *string `json:"lastName"`
}

type AddressOut struct {
	ID      string `json:"id"`
	City    string `json:"city"`
	Street  string `json:"street"`
	Label   string `json:"label"`
	House   string `json:"house"`
	Apt     string `json:"apt"`
	Parking string `json:"parking"`
	Type    string `json:"type"`
}

type WorkerOut struct {
	WorkerID int64   `json:"workerId"`
	Callsign *int64  `json:"callsign"`
	Name     string  `json:"name"`
	Phone    *string `json:"phone"`
}

type CarOut struct {
	CarID  int64   `json:"carId"`
	Name   *string `json:"name"`
	Color  *int64  `json:"color"`
	Number *string `json:"number"`
}

type TariffOut struct {
	TariffID          int64    `json:"tariffId"`
	Name              string   `json:"name"`
	QuantitativeTitle *string  `json:"quantitative_title,omitempty"`
	PriceForUnit      *float64 `json:"price_for_unit,omitempty"`
	UnitName          *string  `json:"unit_name,omitempty"`
}

type StatusKey struct {
	OrderID  int64
	StatusID int64
}
