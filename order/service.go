package order

import (
	"context"
	"sort"
)

type SortOrder string

type BaseFilter struct {
	TenantID       int64
	CityIDs        []int64
	Status         []int64
	Date           *string
	StatusTimeFrom *int64
	StatusTimeTo   *int64
	Tariffs        []int64
	UserPositions  []int64

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

type Repository interface {
	FetchUnpaid(ctx context.Context, filter UnpaidFilter) ([]int64, error)
	FetchBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error)
	FetchExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
	FetchWarningStatus(ctx context.Context, f WarningFilter) ([]int64, error)
	CountOrdersWithWarning(ctx context.Context, f BaseFilter, warningIDs []int64) (int64, error)
	FetchOrdersWithWarning(ctx context.Context, f BaseFilter, warningIDs []int64, page, pageSize int) ([]FullOrder, error)
}

type Service interface {
	GetUnpaid(ctx context.Context, f UnpaidFilter) ([]int64, error)
	GetBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error)
	GetExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
	GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error)
	GetWarningGroupOrders(ctx context.Context, base BaseFilter, warningIDs []int64, page, pageSize int) (WarningGroupResult, error)
}

type WarningGroupResult struct {
	TotalCount int64       `json:"total_count"`
	Orders     []FullOrder `json:"orders"`
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetUnpaid(ctx context.Context, filter UnpaidFilter) ([]int64, error) {
	return s.repo.FetchUnpaid(ctx, filter)
}

func (s *service) GetBadReview(
	ctx context.Context, filter BadReviewFilter,
) ([]int64, error) {
	return s.repo.FetchBadReview(ctx, filter)
}

func (s *service) GetExceededPrice(
	ctx context.Context,
	filter ExceededPriceFilter,
) ([]int64, error) {
	return s.repo.FetchExceededPrice(ctx, filter)
}

func (s *service) GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error) {
	// 1) ID по warning-статусам
	statusIDs, err := s.repo.FetchWarningStatus(ctx, f)
	if err != nil {
		return nil, err
	}

	// 2) unpaid
	unpaidIDs, err := s.repo.FetchUnpaid(ctx, UnpaidFilter{
		BaseFilter:             f.BaseFilter,
		StatusCompletedNotPaid: f.StatusCompletedNotPaid,
	})
	if err != nil {
		return nil, err
	}

	// 3) bad reviews
	badIDs, err := s.repo.FetchBadReview(ctx, BadReviewFilter{
		BaseFilter:   f.BaseFilter,
		BadRatingMax: f.BadRatingMax,
	})
	if err != nil {
		return nil, err
	}

	// 4) real price > predv
	realIDs, err := s.repo.FetchExceededPrice(ctx, ExceededPriceFilter{
		BaseFilter:     f.BaseFilter,
		MinRealPrice:   f.MinRealPrice,
		FinishedStatus: f.FinishedStatus,
	})
	if err != nil {
		return nil, err
	}

	idSet := make(map[int64]struct{})

	for _, id := range statusIDs {
		idSet[id] = struct{}{}
	}
	for _, id := range unpaidIDs {
		idSet[id] = struct{}{}
	}
	for _, id := range badIDs {
		idSet[id] = struct{}{}
	}
	for _, id := range realIDs {
		idSet[id] = struct{}{}
	}

	result := make([]int64, 0, len(idSet))
	for id := range idSet {
		result = append(result, id)
	}

	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })

	return result, nil
}

func (s *service) GetWarningGroupOrders(
	ctx context.Context,
	base BaseFilter,
	warningIDs []int64,
	page, pageSize int,
) (WarningGroupResult, error) {

	cnt, err := s.repo.CountOrdersWithWarning(ctx, base, warningIDs)
	if err != nil {
		return WarningGroupResult{}, err
	}

	orders, err := s.repo.FetchOrdersWithWarning(ctx, base, warningIDs, page, pageSize)
	if err != nil {
		return WarningGroupResult{}, err
	}

	return WarningGroupResult{
		TotalCount: cnt,
		Orders:     orders,
	}, nil
}
