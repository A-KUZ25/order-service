package order

import "context"

type SortOrder string

type BaseFilter struct {
	TenantID       int64
	CityIDs        []int64
	Language       string
	Status         []int64
	Date           *string
	StatusTimeFrom *int64
	StatusTimeTo   *int64
	SelectForDate  bool
	Tariffs        []int64
	UserPositions  []int64
	Group          string

	SortField string
	SortOrder string
}

type UnpaidFilter struct {
	BaseFilter

	StatusCompletedNotPaid int64
}

type BadReviewFilter struct {
	BaseFilter

	BadRatingMax int64
}

type ExceededPriceFilter struct {
	BaseFilter BaseFilter

	MinRealPrice   float64
	FinishedStatus []int64
}

type WarningFilter struct {
	BaseFilter             BaseFilter
	WarningStatus          []int64
	FinishedStatus         []int64
	BadRatingMax           int64
	StatusCompletedNotPaid int64
	MinRealPrice           float64
}

type WarningOrderReader interface {
	FetchUnpaid(ctx context.Context, filter UnpaidFilter) ([]int64, error)
	FetchBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error)
	FetchExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
}

type OrderListReader interface {
	CountOrdersWithWarning(
		ctx context.Context,
		f BaseFilter,
		warningIDs []int64,
	) (int64, error)
	FetchOrdersWithWarning(
		ctx context.Context,
		f BaseFilter, warningIDs []int64,
		page,
		pageSize int,
	) ([]FullOrder, error)
}

type GroupOrderReader interface {
	FetchOrdersByStatusGroup(
		ctx context.Context,
		f BaseFilter,
	) ([]int64, error)
}

type OrderOptionsReader interface {
	GetOptionsForOrders(
		ctx context.Context,
		orderIDs []int64,
	) (map[int64][]OptionDTO, error)
}

type StatusChangeReader interface {
	GetStatusChangeTimes(
		ctx context.Context,
		keys []StatusKey,
	) (map[StatusKey]int64, error)
}

type Repository interface {
	WarningOrderReader
	OrderListReader
	GroupOrderReader
	OrderOptionsReader
	StatusChangeReader
}

type OrderAddressResolver interface {
	ResolveAddresses(orders []FullOrder) (map[int64][]AddressView, error)
}

type WaitingTimeProvider interface {
	GetWorkerWaitingTime(
		ctx context.Context,
		tenantID, orderID int64,
	) (int64, error)
}

type StatusTranslator interface {
	TranslateStatus(ctx context.Context, language, name string) (string, error)
}

type ShowOrderCodeProvider interface {
	ShouldShowOrderCode(
		ctx context.Context,
		tenantID, cityID, positionID int64,
	) (bool, error)
}

type OrderViewAssembler interface {
	BuildOrderView(
		ctx context.Context,
		o FormattedOrder,
		f WarningFilter,
		statusChangeTimes map[StatusKey]int64,
	) (OrderView, error)
}

type Service interface {
	GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error)
	GetFormattedOrdersByGroup(
		ctx context.Context,
		f WarningFilter,
		page, pageSize int,
	) (int64, []FormattedOrder, error)
	GetOrdersForTabs(
		ctx context.Context,
		f WarningFilter,
	) (GroupOrdersResult, error)
	PrepareOrdersData(
		ctx context.Context,
		orders []FormattedOrder,
		f WarningFilter,
	) ([]OrderView, error)
}

type WarningGroupResult struct {
	WarningOrderIDs []int64
	TotalCount      int64       `json:"total_count"`
	Orders          []FullOrder `json:"orders"`
}

type service struct {
	warningReader      WarningOrderReader
	orderListReader    OrderListReader
	groupOrderReader   GroupOrderReader
	optionsReader      OrderOptionsReader
	statusChangeReader StatusChangeReader
	assembler          OrderViewAssembler
	addressResolver    OrderAddressResolver
}

func NewService(
	repo Repository,
	addressResolver OrderAddressResolver,
	assembler OrderViewAssembler,
) Service {
	return &service{
		warningReader:      repo,
		orderListReader:    repo,
		groupOrderReader:   repo,
		optionsReader:      repo,
		statusChangeReader: repo,
		assembler:          assembler,
		addressResolver:    addressResolver,
	}
}
