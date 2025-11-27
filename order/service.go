package order

import (
	"context"
)

type SortOrder string

type BaseFilter struct {
	TenantID       int64
	CityIDs        []int64
	Status         []string
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

type RealPriceFilter struct {
	BaseFilter BaseFilter

	MinRealPrice   float64
	FinishedStatus []int64
}

type Repository interface {
	FetchUnpaidOrderIDs(ctx context.Context, filter UnpaidFilter) ([]int64, error)
	FetchBadReviewOrderIDs(ctx context.Context, f BadReviewFilter) ([]int64, error)
	FetchRealPriceMorePredvPrice(ctx context.Context, f RealPriceFilter) ([]int64, error)
}

type Service interface {
	GetUnpaidOrderIDs(ctx context.Context, f UnpaidFilter) ([]int64, error)
	GetBadReviewOrderIDs(ctx context.Context, f BadReviewFilter) ([]int64, error)
	GetRealPriceMorePredvPrice(ctx context.Context, f RealPriceFilter) ([]int64, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetUnpaidOrderIDs(ctx context.Context, filter UnpaidFilter) ([]int64, error) {
	return s.repo.FetchUnpaidOrderIDs(ctx, filter)
}

func (s *service) GetBadReviewOrderIDs(ctx context.Context, filter BadReviewFilter) ([]int64, error) {
	return s.repo.FetchBadReviewOrderIDs(ctx, filter)
}

func (s *service) GetRealPriceMorePredvPrice(
	ctx context.Context,
	filter RealPriceFilter,
) ([]int64, error) {
	return s.repo.FetchRealPriceMorePredvPrice(ctx, filter)
}
