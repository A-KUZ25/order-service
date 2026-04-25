package order

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type StatusGroup string

const (
	StatusGroup0 StatusGroup = "new"
	StatusGroup6 StatusGroup = "pre_order"
	StatusGroup7 StatusGroup = "warning"
	StatusGroup8 StatusGroup = "works"
)

type GroupOrdersResult struct {
	GroupCounts     map[StatusGroup]int
	OrdersForSignal map[StatusGroup][]int64
}

var orderGroupIDs = map[StatusGroup][]int64{
	StatusGroup0: {1, 2, 3, 4, 5, 52, 108, 109, 115, 127, 128, 130, 131},
	StatusGroup6: {6, 7, 16, 111, 112, 116, 117, 118, 119},
	StatusGroup7: {5, 10, 16, 27, 30, 38, 45, 46, 47, 48, 52, 54, 117, 118, 129, 135},
	StatusGroup8: {17, 26, 27, 29, 30, 36, 54, 55, 106, 110, 113, 114, 132, 133, 134, 135, 136},
}

func (s *service) GetOrdersForTabs(
	ctx context.Context,
	f WarningFilter,
) (GroupOrdersResult, error) {
	groupOrders := make(map[StatusGroup][]int64, 4)
	var mu sync.Mutex

	g, groupCtx := errgroup.WithContext(ctx)

	for group, statusIDs := range orderGroupIDs {
		group := group
		statusIDs := statusIDs

		bf := f.BaseFilter
		bf.Status = statusIDs
		bf.SelectForDate = group == StatusGroup7

		g.Go(func() error {
			ids, err := s.groupOrderReader.FetchOrdersByStatusGroup(groupCtx, bf)
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

	f.BaseFilter.SelectForDate = true
	warningIDs, err := s.GetWarningOrder(ctx, f)
	if err != nil {
		return GroupOrdersResult{}, err
	}

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

	groupCounts := make(map[StatusGroup]int, len(groupOrders))
	for g, ids := range groupOrders {
		groupCounts[g] = len(ids)
	}

	return GroupOrdersResult{
		GroupCounts: groupCounts,
		OrdersForSignal: map[StatusGroup][]int64{
			StatusGroup0: groupOrders[StatusGroup0],
			StatusGroup6: groupOrders[StatusGroup6],
		},
	}, nil
}
