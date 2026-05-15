package order

import (
	"context"
	"orders-service/internal/logging"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"
)

func (s *service) GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error) {
	totalStarted := time.Now()
	g, ctx := errgroup.WithContext(ctx)

	var (
		unpaidIDs []int64
		badIDs    []int64
		realIDs   []int64
		unpaidMS  int64
		badMS     int64
		realMS    int64
	)

	g.Go(func() error {
		started := time.Now()
		ids, err := s.warningReader.FetchUnpaid(ctx, UnpaidFilter{
			BaseFilter:             f.BaseFilter,
			StatusCompletedNotPaid: f.StatusCompletedNotPaid,
		})
		unpaidMS = time.Since(started).Milliseconds()
		if err != nil {
			return err
		}
		unpaidIDs = ids
		return nil
	})

	g.Go(func() error {
		started := time.Now()
		ids, err := s.warningReader.FetchBadReview(ctx, BadReviewFilter{
			BaseFilter:   f.BaseFilter,
			BadRatingMax: f.BadRatingMax,
		})
		badMS = time.Since(started).Milliseconds()
		if err != nil {
			return err
		}
		badIDs = ids
		return nil
	})

	g.Go(func() error {
		started := time.Now()
		ids, err := s.warningReader.FetchExceededPrice(ctx, ExceededPriceFilter{
			BaseFilter:     f.BaseFilter,
			MinRealPrice:   f.MinRealPrice,
			FinishedStatus: f.FinishedStatus,
		})
		realMS = time.Since(started).Milliseconds()
		if err != nil {
			return err
		}
		realIDs = ids
		return nil
	})

	if err := g.Wait(); err != nil {
		logging.Error(ctx, "refresh warning ids failed", err,
			"fetch_unpaid_ms", unpaidMS,
			"fetch_bad_review_ms", badMS,
			"fetch_exceeded_price_ms", realMS,
		)
		return nil, err
	}

	idSet := make(map[int64]struct{}, len(unpaidIDs)+len(badIDs)+len(realIDs))
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

	logging.Info(ctx, "refresh warning ids timings",
		"total_ms", time.Since(totalStarted).Milliseconds(),
		"fetch_unpaid_ms", unpaidMS,
		"fetch_bad_review_ms", badMS,
		"fetch_exceeded_price_ms", realMS,
		"unpaid_count", len(unpaidIDs),
		"bad_review_count", len(badIDs),
		"exceeded_price_count", len(realIDs),
		"merged_count", len(result),
	)

	return result, nil
}

func (s *service) GetOrdersByGroup(
	ctx context.Context,
	f WarningFilter,
	page, pageSize int,
) (int64, []FullOrder, error) {
	totalStarted := time.Now()
	var (
		ordersCount     int64
		ordersPaginated []FullOrder
		warningIDsMS    int64
		countMS         int64
		fetchMS         int64
	)

	if f.BaseFilter.Group == "warning" {
		started := time.Now()
		warningOrderIDs, err := s.GetWarningOrder(ctx, f)
		warningIDsMS = time.Since(started).Milliseconds()
		if err != nil {
			return 0, nil, err
		}

		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			started := time.Now()
			cnt, err := s.orderListReader.CountOrdersWithWarning(ctx, f.BaseFilter, warningOrderIDs)
			countMS = time.Since(started).Milliseconds()
			if err != nil {
				return err
			}
			ordersCount = cnt
			return nil
		})

		g.Go(func() error {
			started := time.Now()
			ords, err := s.orderListReader.FetchOrdersWithWarning(ctx, f.BaseFilter, warningOrderIDs, page, pageSize)
			fetchMS = time.Since(started).Milliseconds()
			if err != nil {
				return err
			}
			ordersPaginated = ords
			return nil
		})

		if err := g.Wait(); err != nil {
			return 0, nil, err
		}

		logging.Info(ctx, "refresh get orders by group timings",
			"total_ms", time.Since(totalStarted).Milliseconds(),
			"warning_ids_ms", warningIDsMS,
			"count_ms", countMS,
			"fetch_ms", fetchMS,
			"group", f.BaseFilter.Group,
			"page", page,
			"page_size", pageSize,
			"total_count", ordersCount,
			"orders_count", len(ordersPaginated),
			"warning_ids_count", len(warningOrderIDs),
		)

		return ordersCount, ordersPaginated, nil
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		started := time.Now()
		cnt, err := s.orderListReader.CountOrdersWithWarning(ctx, f.BaseFilter, nil)
		countMS = time.Since(started).Milliseconds()
		if err != nil {
			return err
		}
		ordersCount = cnt
		return nil
	})

	g.Go(func() error {
		started := time.Now()
		ords, err := s.orderListReader.FetchOrdersWithWarning(ctx, f.BaseFilter, nil, page, pageSize)
		fetchMS = time.Since(started).Milliseconds()
		if err != nil {
			return err
		}
		ordersPaginated = ords
		return nil
	})

	if err := g.Wait(); err != nil {
		return 0, nil, err
	}

	logging.Info(ctx, "refresh get orders by group timings",
		"total_ms", time.Since(totalStarted).Milliseconds(),
		"warning_ids_ms", warningIDsMS,
		"count_ms", countMS,
		"fetch_ms", fetchMS,
		"group", f.BaseFilter.Group,
		"page", page,
		"page_size", pageSize,
		"total_count", ordersCount,
		"orders_count", len(ordersPaginated),
	)

	return ordersCount, ordersPaginated, nil
}
