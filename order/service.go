package order

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

type SortOrder string

type BaseFilter struct {
	TenantID       int64
	CityIDs        []int64
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

type Repository interface {
	FetchUnpaid(ctx context.Context, filter UnpaidFilter) ([]int64, error)
	FetchBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error)
	FetchExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
	FetchWarningStatus(ctx context.Context, f WarningFilter) ([]int64, error)
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
	FetchOrdersByStatusGroup(
		ctx context.Context,
		f BaseFilter,
	) ([]int64, error)
}

type Service interface {
	GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error)
	GetOrdersByGroup(
		ctx context.Context,
		f WarningFilter,
		page,
		pageSize int,
	) (int64, []FullOrder, error)
	GetOrdersForTabs(
		ctx context.Context,
		f WarningFilter,
	) (GroupOrdersResult, error)
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

func (s *service) GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error) {

	g, ctx := errgroup.WithContext(ctx)

	var (
		statusIDs []int64
		unpaidIDs []int64
		badIDs    []int64
		realIDs   []int64
	)

	// 1) unpaid
	g.Go(func() error {
		ids, err := s.repo.FetchUnpaid(ctx, UnpaidFilter{
			BaseFilter:             f.BaseFilter,
			StatusCompletedNotPaid: f.StatusCompletedNotPaid,
		})
		if err != nil {
			return err
		}
		unpaidIDs = ids
		return nil
	})

	// 2) bad reviews
	g.Go(func() error {
		ids, err := s.repo.FetchBadReview(ctx, BadReviewFilter{
			BaseFilter:   f.BaseFilter,
			BadRatingMax: f.BadRatingMax,
		})
		if err != nil {
			return err
		}
		badIDs = ids
		return nil
	})

	// 3) real > predv
	g.Go(func() error {
		ids, err := s.repo.FetchExceededPrice(ctx, ExceededPriceFilter{
			BaseFilter:     f.BaseFilter,
			MinRealPrice:   f.MinRealPrice,
			FinishedStatus: f.FinishedStatus,
		})
		if err != nil {
			return err
		}
		realIDs = ids
		return nil
	})

	// Ждём завершения всех горутин
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// ---------- ОБЪЕДИНЕНИЕ РЕЗУЛЬТАТОВ ----------

	idSet := make(map[int64]struct{},
		len(statusIDs)+len(unpaidIDs)+len(badIDs)+len(realIDs),
	)

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

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result, nil
}

func (s *service) GetOrdersByGroup(
	ctx context.Context,
	f WarningFilter,
	page, pageSize int,
) (int64, []FullOrder, error) {

	var (
		ordersCount     int64
		ordersPaginated []FullOrder
	)
	start := time.Now()
	// Если это "warning" группа — нужно учитывать warningOrderIDs (OR o.order_id IN (...))
	// В PHP: для STATUS_GROUP_7 -> if empty(warningOrderIds) ? count() : orFilterWhere(...)->count()
	if f.BaseFilter.Group == "warning" {

		warningOrderIDs, err := s.GetWarningOrder(ctx, f)
		if err != nil {
			return 0, nil, err
		}
		// Если warningOrderIDs пуст — это просто обычный подсчёт/пагинация по baseFilter
		// В противном случае используем их как дополнительный OR (CountOrdersWithWarning / FetchOrdersWithWarning реализуют это).

		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			cnt, err := s.repo.CountOrdersWithWarning(ctx, f.BaseFilter, warningOrderIDs)
			if err != nil {
				return err
			}
			ordersCount = cnt
			return nil
		})

		g.Go(func() error {
			ords, err := s.repo.FetchOrdersWithWarning(ctx, f.BaseFilter, warningOrderIDs, page, pageSize)
			if err != nil {
				return err
			}
			ordersPaginated = ords
			return nil
		})

		if err := g.Wait(); err != nil {
			return 0, nil, err
		}

		log.Println("Execution took:", time.Since(start))
		return ordersCount, ordersPaginated, nil
	}
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		cnt, err := s.repo.CountOrdersWithWarning(ctx, f.BaseFilter, nil)
		if err != nil {
			return err
		}
		ordersCount = cnt
		return nil
	})

	g.Go(func() error {
		ords, err := s.repo.FetchOrdersWithWarning(ctx, f.BaseFilter, nil, page, pageSize)
		if err != nil {
			return err
		}
		ordersPaginated = ords
		return nil
	})

	if err := g.Wait(); err != nil {
		return 0, nil, err
	}
	log.Println("Execution took:", time.Since(start))
	return ordersCount, ordersPaginated, nil
}

type StatusGroup string

const (
	StatusGroup0 StatusGroup = "new"
	StatusGroup6 StatusGroup = "pre_order"
	StatusGroup7 StatusGroup = "warning" // warning
	StatusGroup8 StatusGroup = "works"
)

type GroupOrdersResult struct {
	GroupCounts     map[StatusGroup]int
	OrdersForSignal map[StatusGroup][]int64
}

var orderGroupIds = map[StatusGroup][]int64{
	StatusGroup0: {
		1, 2, 3, 4, 5, 52, 108, 109, 115, 127, 128, 130, 131,
	},
	StatusGroup6: {
		6, 7, 16, 111, 112, 116, 117, 118, 119,
	},
	StatusGroup7: {
		5, 10, 16, 27, 30, 38, 45, 46, 47, 48,
		52, 54, 117, 118, 129, 135,
	},
	StatusGroup8: {
		17, 26, 27, 29, 30, 36, 54, 55,
		106, 110, 113, 114,
		132, 133, 134, 135, 136,
	},
}

func (s *service) GetOrdersForTabs(
	ctx context.Context,
	f WarningFilter,
) (GroupOrdersResult, error) {
	// ---------- ЭТАП 1: базовые группы ----------
	groupOrders := make(map[StatusGroup][]int64, 4)
	var mu sync.Mutex

	g, groupCtx := errgroup.WithContext(ctx)

	for group, statusIDs := range orderGroupIds {
		group := group
		statusIDs := statusIDs

		bf := f.BaseFilter
		bf.Status = statusIDs
		if group == StatusGroup7 {
			bf.SelectForDate = true
		} else {
			bf.SelectForDate = false
		}
		g.Go(func() error {
			ids, err := s.repo.FetchOrdersByStatusGroup(
				groupCtx,
				bf,
			)
			if err != nil {
				return err
			}

			mu.Lock()
			groupOrders[group] = ids
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return GroupOrdersResult{}, err
	}

	// ---------- ЭТАП 2: WARNING (НОВЫЙ КОНТЕКСТ) ----------
	f.BaseFilter.SelectForDate = true
	warningIDs, err := s.GetWarningOrder(ctx, f)
	if err != nil {
		return GroupOrdersResult{}, err
	}

	// merge warning → group 7
	idSet := make(map[int64]struct{})
	for _, id := range groupOrders[StatusGroup7] {
		idSet[id] = struct{}{}
	}
	for _, id := range warningIDs {
		idSet[id] = struct{}{}
	}

	merged := make([]int64, 0, len(idSet))
	for id := range idSet {
		merged = append(merged, id)
	}
	groupOrders[StatusGroup7] = merged

	// ---------- COUNTS ----------
	groupCounts := make(map[StatusGroup]int, len(groupOrders))
	for g, ids := range groupOrders {
		groupCounts[g] = len(ids)
	}

	// ---------- SIGNAL ----------
	ordersForSignal := map[StatusGroup][]int64{
		StatusGroup0: groupOrders[StatusGroup0],
		StatusGroup6: groupOrders[StatusGroup6],
	}

	return GroupOrdersResult{
		GroupCounts:     groupCounts,
		OrdersForSignal: ordersForSignal,
	}, nil
}
