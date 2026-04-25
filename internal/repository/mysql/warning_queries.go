package mysql

import (
	"context"
	"log"
	"orders-service/internal/app/order"
	"strings"
	"time"
)

func (r *OrdersRepository) FetchUnpaid(
	ctx context.Context,
	f order.UnpaidFilter,
) ([]int64, error) {
	var sb strings.Builder
	args := []any{}

	sb.WriteString(`
SELECT o.order_id
FROM tbl_order o
WHERE ( 1=1
`)

	r.buildBaseQuery(&sb, &args, f.BaseFilter, true)
	sb.WriteString(") ")
	sb.WriteString("  AND o.status_id = ?\n")
	args = append(args, f.StatusCompletedNotPaid)

	r.appendOrderBy(&sb, f.BaseFilter)

	return r.executeQuery(ctx, sb.String(), args)
}

func (r *OrdersRepository) FetchBadReview(
	ctx context.Context,
	f order.BadReviewFilter,
) ([]int64, error) {
	var sb strings.Builder
	args := []any{}

	sb.WriteString(`
SELECT o.order_id
FROM tbl_order o
LEFT JOIN tbl_client_review cr ON o.order_id = cr.order_id
WHERE ( 1=1
`)

	r.buildBaseQuery(&sb, &args, f.BaseFilter, true)
	sb.WriteString(") ")
	sb.WriteString("  AND cr.rating BETWEEN 1 AND ?\n")
	args = append(args, f.BadRatingMax)

	r.appendOrderBy(&sb, f.BaseFilter)

	return r.executeQuery(ctx, sb.String(), args)
}

func (r *OrdersRepository) FetchExceededPrice(
	ctx context.Context,
	f order.ExceededPriceFilter,
) ([]int64, error) {
	var sb strings.Builder
	var args []any

	sb.WriteString(`
SELECT o.order_id
FROM tbl_order o
WHERE ( 1=1
`)

	r.buildBaseQuery(&sb, &args, f.BaseFilter, true)
	sb.WriteString(") ")
	if len(f.FinishedStatus) > 0 {
		sb.WriteString("  AND o.status_id NOT IN (")
		for i, st := range f.FinishedStatus {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			args = append(args, st)
		}
		sb.WriteString(")\n")
	}

	sb.WriteString("  AND o.realtime_price > ?\n")
	args = append(args, f.MinRealPrice)

	r.appendOrderBy(&sb, f.BaseFilter)

	return r.executeQuery(ctx, sb.String(), args)
}

func (r *OrdersRepository) CountOrdersWithWarning(
	ctx context.Context,
	f order.BaseFilter,
	warningIDs []int64,
) (int64, error) {
	var sb strings.Builder
	var args []any

	sb.WriteString(`
SELECT COUNT(*)
FROM tbl_order o
WHERE ( 1=1
`)

	r.buildBaseQuery(&sb, &args, f)
	sb.WriteString(") ")
	if len(warningIDs) > 0 {
		sb.WriteString(" OR o.order_id IN (")
		for i, id := range warningIDs {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			args = append(args, id)
		}
		sb.WriteString(")\n")
	}

	start := time.Now()
	row := r.db.QueryRowContext(ctx, sb.String(), args...)
	log.Println("BASE REQUEST TIME:", time.Since(start))

	var cnt int64
	if err := row.Scan(&cnt); err != nil {
		return 0, err
	}

	return cnt, nil
}
