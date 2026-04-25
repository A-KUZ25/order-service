package order

import (
	"context"
	"sort"

	"golang.org/x/sync/errgroup"
)

func (s *service) GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error) {
	g, ctx := errgroup.WithContext(ctx)

	var (
		unpaidIDs []int64
		badIDs    []int64
		realIDs   []int64
	)

	g.Go(func() error {
		ids, err := s.warningReader.FetchUnpaid(ctx, UnpaidFilter{
			BaseFilter:             f.BaseFilter,
			StatusCompletedNotPaid: f.StatusCompletedNotPaid,
		})
		if err != nil {
			return err
		}
		unpaidIDs = ids
		return nil
	})

	g.Go(func() error {
		ids, err := s.warningReader.FetchBadReview(ctx, BadReviewFilter{
			BaseFilter:   f.BaseFilter,
			BadRatingMax: f.BadRatingMax,
		})
		if err != nil {
			return err
		}
		badIDs = ids
		return nil
	})

	g.Go(func() error {
		ids, err := s.warningReader.FetchExceededPrice(ctx, ExceededPriceFilter{
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

	if err := g.Wait(); err != nil {
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

	if f.BaseFilter.Group == "warning" {
		warningOrderIDs, err := s.GetWarningOrder(ctx, f)
		if err != nil {
			return 0, nil, err
		}

		g, ctx := errgroup.WithContext(ctx)

		g.Go(func() error {
			cnt, err := s.orderListReader.CountOrdersWithWarning(ctx, f.BaseFilter, warningOrderIDs)
			if err != nil {
				return err
			}
			ordersCount = cnt
			return nil
		})

		g.Go(func() error {
			ords, err := s.orderListReader.FetchOrdersWithWarning(ctx, f.BaseFilter, warningOrderIDs, page, pageSize)
			if err != nil {
				return err
			}
			ordersPaginated = ords
			return nil
		})

		if err := g.Wait(); err != nil {
			return 0, nil, err
		}

		return ordersCount, ordersPaginated, nil
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		cnt, err := s.orderListReader.CountOrdersWithWarning(ctx, f.BaseFilter, nil)
		if err != nil {
			return err
		}
		ordersCount = cnt
		return nil
	})

	g.Go(func() error {
		ords, err := s.orderListReader.FetchOrdersWithWarning(ctx, f.BaseFilter, nil, page, pageSize)
		if err != nil {
			return err
		}
		ordersPaginated = ords
		return nil
	})

	if err := g.Wait(); err != nil {
		return 0, nil, err
	}

	return ordersCount, ordersPaginated, nil
}
