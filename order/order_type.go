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
