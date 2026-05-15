package order

import (
	"context"
	"errors"
	"orders-service/internal/logging"
	"time"
)

func (s *service) getAssembler() OrderViewAssembler {
	if s.assembler != nil {
		return s.assembler
	}

	return nil
}

func (s *service) PrepareOrdersData(
	ctx context.Context,
	orders []FormattedOrder,
	f WarningFilter,
) ([]OrderView, error) {
	totalStarted := time.Now()
	var statusChangeMS int64
	var buildViewsMS int64

	seen := make(map[int64]struct{}, len(orders))
	assembler := s.getAssembler()
	if assembler == nil {
		return nil, errors.New("order view assembler is not configured")
	}

	uniqueOrders := make([]FormattedOrder, 0, len(orders))
	for _, o := range orders {
		if _, ok := seen[o.OrderID]; ok {
			continue
		}
		seen[o.OrderID] = struct{}{}
		uniqueOrders = append(uniqueOrders, o)
	}

	started := time.Now()
	statusChangeTimes, err := s.loadStatusChangeTimes(ctx, uniqueOrders)
	statusChangeMS = time.Since(started).Milliseconds()
	if err != nil {
		logging.Error(ctx, "prepare orders status changes failed", err, "duration_ms", statusChangeMS)
		return nil, err
	}

	started = time.Now()
	if batchAssembler, ok := assembler.(OrderViewsAssembler); ok {
		result, err := batchAssembler.BuildOrderViews(ctx, uniqueOrders, f, statusChangeTimes)
		buildViewsMS = time.Since(started).Milliseconds()
		if err != nil {
			logging.Error(ctx, "prepare orders build views failed", err, "duration_ms", buildViewsMS, "batch", true)
			return nil, err
		}

		logging.Info(ctx, "prepare orders timings",
			"total_ms", time.Since(totalStarted).Milliseconds(),
			"status_change_ms", statusChangeMS,
			"build_views_ms", buildViewsMS,
			"orders_count", len(orders),
			"unique_orders_count", len(uniqueOrders),
			"prepared_count", len(result),
			"batch_assembler", true,
		)
		return result, nil
	}

	result := make([]OrderView, 0, len(uniqueOrders))
	for _, o := range uniqueOrders {
		prepared, err := assembler.BuildOrderView(ctx, o, f, statusChangeTimes)
		if err != nil {
			buildViewsMS = time.Since(started).Milliseconds()
			logging.Error(ctx, "prepare orders build views failed", err, "duration_ms", buildViewsMS, "batch", false)
			return nil, err
		}
		result = append(result, prepared)
	}
	buildViewsMS = time.Since(started).Milliseconds()

	logging.Info(ctx, "prepare orders timings",
		"total_ms", time.Since(totalStarted).Milliseconds(),
		"status_change_ms", statusChangeMS,
		"build_views_ms", buildViewsMS,
		"orders_count", len(orders),
		"unique_orders_count", len(uniqueOrders),
		"prepared_count", len(result),
		"batch_assembler", false,
	)

	return result, nil
}

func resolveSummaryCost(o FormattedOrder, group string) any {
	summaryCost := any(o.PredvPrice)
	if o.PredvPriceNoDiscount > 0 {
		summaryCost = o.PredvPriceNoDiscount
	}

	if group == "completed" {
		switch {
		case o.SummaryCostNoDiscount != nil:
			summaryCost = *o.SummaryCostNoDiscount
		case o.SummaryCost != nil:
			summaryCost = *o.SummaryCost
		}
	}

	return summaryCost
}

func formatOrderTimeForSort(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05")
}

func formatOrderTime(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("02.01.06 15:04")
}

func (s *service) loadStatusChangeTimes(
	ctx context.Context,
	orders []FormattedOrder,
) (map[StatusKey]int64, error) {
	keysMap := make(map[StatusKey]struct{}, len(orders))

	for _, o := range orders {
		keysMap[StatusKey{
			OrderID:  o.OrderID,
			StatusID: o.StatusID,
		}] = struct{}{}
	}

	keys := make([]StatusKey, 0, len(keysMap))
	for k := range keysMap {
		keys = append(keys, k)
	}

	return s.statusChangeReader.GetStatusChangeTimes(ctx, keys)
}

func getTimeOrderStatusChanged(
	orderID int64,
	statusID int64,
	statusTime int64,
	statusChangeTimes map[StatusKey]int64,
) int64 {
	key := StatusKey{
		OrderID:  orderID,
		StatusID: statusID,
	}

	if t, ok := statusChangeTimes[key]; ok {
		return t
	}

	return statusTime
}

func ShowCodeOrID(showCode bool, orderCode string, orderNumber int64) any {
	if showCode {
		if orderCode != "" {
			return orderCode
		}
		return orderNumber
	}
	return orderNumber
}
