package order

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"
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
	GetWarningFull(ctx context.Context, f WarningFilter, page, pageSize int) (WarningGroupResult, error)
}

type WarningGroupResult struct {
	WarningOrderIDs []int64
	TotalCount      int64       `json:"total_count"`
	Orders          []FullOrder `json:"orders"`
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

	type result struct {
		ids []int64
		err error
	}

	chStatus := make(chan result, 1)
	chUnpaid := make(chan result, 1)
	chBad := make(chan result, 1)
	chReal := make(chan result, 1)

	// 1) warning statuses
	go func() {
		ids, err := s.repo.FetchWarningStatus(ctx, f)
		chStatus <- result{ids, err}
	}()

	// 2) unpaid
	go func() {
		ids, err := s.repo.FetchUnpaid(ctx, UnpaidFilter{
			BaseFilter:             f.BaseFilter,
			StatusCompletedNotPaid: f.StatusCompletedNotPaid,
		})
		chUnpaid <- result{ids, err}
	}()

	// 3) bad reviews
	go func() {
		ids, err := s.repo.FetchBadReview(ctx, BadReviewFilter{
			BaseFilter:   f.BaseFilter,
			BadRatingMax: f.BadRatingMax,
		})
		chBad <- result{ids, err}
	}()

	// 4) real > predv
	go func() {
		ids, err := s.repo.FetchExceededPrice(ctx, ExceededPriceFilter{
			BaseFilter:     f.BaseFilter,
			MinRealPrice:   f.MinRealPrice,
			FinishedStatus: f.FinishedStatus,
		})
		chReal <- result{ids, err}
	}()

	resStatus := <-chStatus
	resUnpaid := <-chUnpaid
	resBad := <-chBad
	resReal := <-chReal

	if resStatus.err != nil {
		return nil, resStatus.err
	}
	if resUnpaid.err != nil {
		return nil, resUnpaid.err
	}
	if resBad.err != nil {
		return nil, resBad.err
	}
	if resReal.err != nil {
		return nil, resReal.err
	}

	// Объединяем уникальные ID
	idSet := make(map[int64]struct{}, len(resStatus.ids)+len(resUnpaid.ids)+len(resBad.ids)+len(resReal.ids))

	for _, id := range resStatus.ids {
		idSet[id] = struct{}{}
	}
	for _, id := range resUnpaid.ids {
		idSet[id] = struct{}{}
	}
	for _, id := range resBad.ids {
		idSet[id] = struct{}{}
	}
	for _, id := range resReal.ids {
		idSet[id] = struct{}{}
	}

	resultIDs := make([]int64, 0, len(idSet))
	for id := range idSet {
		resultIDs = append(resultIDs, id)
	}

	sort.Slice(resultIDs, func(i, j int) bool { return resultIDs[i] < resultIDs[j] })

	return resultIDs, nil
}

func (s *service) GetWarningFull(
	ctx context.Context,
	f WarningFilter,
	page, pageSize int,
) (WarningGroupResult, error) {

	start := time.Now()

	// 1) warning IDs — сначала (внутри у тебя уже 4 параллельных SQL)
	warningIDs, err := s.GetWarningOrder(ctx, f)
	if err != nil {
		return WarningGroupResult{}, err
	}

	// 2) параллельно считаем COUNT и PAGINATED
	var (
		cnt    int64
		orders []FullOrder
	)

	var wg sync.WaitGroup
	wg.Add(2)

	errCh := make(chan error, 2)

	// COUNT goroutine
	go func() {
		defer wg.Done()

		c, err := s.repo.CountOrdersWithWarning(ctx, f.BaseFilter, warningIDs)
		if err != nil {
			errCh <- err
			return
		}
		cnt = c
	}()

	// PAGINATED goroutine
	go func() {
		defer wg.Done()

		ords, err := s.repo.FetchOrdersWithWarning(ctx, f.BaseFilter, warningIDs, page, pageSize)
		if err != nil {
			errCh <- err
			return
		}
		orders = ords
	}()

	wg.Wait()
	close(errCh)

	for e := range errCh {
		if e != nil {
			return WarningGroupResult{}, e
		}
	}

	log.Println("Execution took:", time.Since(start))

	return WarningGroupResult{
		WarningOrderIDs: warningIDs,
		TotalCount:      cnt,
		Orders:          orders,
	}, nil
}
