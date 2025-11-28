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

type ExceededPriceFilter struct {
	BaseFilter BaseFilter

	MinRealPrice   float64
	FinishedStatus []int64
}

type Repository interface {
	FetchUnpaid(ctx context.Context, filter UnpaidFilter) ([]int64, error)
	FetchBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error)
	FetchExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
}

type Service interface {
	GetUnpaid(ctx context.Context, f UnpaidFilter) ([]int64, error)
	GetBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error)
	GetExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
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
