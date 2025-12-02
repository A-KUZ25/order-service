package mysql

import (
	"context"
	"database/sql"
	"orders-service/order"
	"strings"
)

type OrdersRepository struct {
	db *sql.DB
}

func NewOrdersRepository(db *sql.DB) (*OrdersRepository, error) {
	return &OrdersRepository{db: db}, nil
}

// строит общую часть WHERE (tenant, active, date-range, city, tariffs…)
func (r *OrdersRepository) buildBaseQuery(sb *strings.Builder, args *[]any, f order.BaseFilter) {
	sb.WriteString(`
WHERE o.tenant_id = ?
  AND o.active = 1
`)
	*args = append(*args, f.TenantID)

	// date
	if f.StatusTimeFrom != nil && f.StatusTimeTo != nil {
		sb.WriteString("  AND o.status_time BETWEEN ? AND ?\n")
		*args = append(*args, *f.StatusTimeFrom, *f.StatusTimeTo)
	}

	// city
	if len(f.CityIDs) > 0 {
		sb.WriteString("  AND o.city_id IN (")
		for i, id := range f.CityIDs {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			*args = append(*args, id)
		}
		sb.WriteString(")\n")
	}

	// tariffs
	if len(f.Tariffs) > 0 {
		sb.WriteString("  AND o.tariff_id IN (")
		for i, t := range f.Tariffs {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			*args = append(*args, t)
		}
		sb.WriteString(")\n")
	}

	// positions
	if len(f.UserPositions) > 0 {
		sb.WriteString("  AND o.position_id IN (")
		for i, p := range f.UserPositions {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			*args = append(*args, p)
		}
		sb.WriteString(")\n")
	}
}

func (r *OrdersRepository) appendOrderBy(sb *strings.Builder, f order.BaseFilter) {
	orderField := f.SortField
	if orderField == "" {
		orderField = "o.status_time"
	}

	orderDir := "DESC"
	if strings.ToLower(string(f.SortOrder)) == "asc" {
		orderDir = "ASC"
	}

	sb.WriteString("ORDER BY " + orderField + " " + orderDir + "\n")
}

func (r *OrdersRepository) executeQuery(
	ctx context.Context,
	sql string,
	args []any,
) ([]int64, error) {

	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *OrdersRepository) FetchUnpaid(
	ctx context.Context,
	f order.UnpaidFilter,
) ([]int64, error) {

	var sb strings.Builder
	args := []any{}

	sb.WriteString(`
SELECT o.order_id
FROM tbl_order o
`)

	r.buildBaseQuery(&sb, &args, f.BaseFilter)

	// Специфичная часть unpaid
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
`)

	r.buildBaseQuery(&sb, &args, f.BaseFilter)

	// специфично bad reviews
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
`)

	r.buildBaseQuery(&sb, &args, f.BaseFilter)

	//статус не должен быть финальным
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

	//realtime_price > MinRealPrice
	sb.WriteString("  AND o.realtime_price > ?\n")
	args = append(args, f.MinRealPrice)

	r.appendOrderBy(&sb, f.BaseFilter)

	return r.executeQuery(ctx, sb.String(), args)
}

func (r *OrdersRepository) FetchWarningStatus(
	ctx context.Context,
	f order.WarningFilter,
) ([]int64, error) {

	var sb strings.Builder
	args := []any{}

	sb.WriteString(`
SELECT o.order_id
FROM tbl_order o
`)

	r.buildBaseQuery(&sb, &args, f.BaseFilter)

	if len(f.WarningStatus) > 0 {
		sb.WriteString("  AND o.status_id IN (")
		for i, st := range f.WarningStatus {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			args = append(args, st)
		}
		sb.WriteString(")\n")
	}

	r.appendOrderBy(&sb, f.BaseFilter)

	return r.executeQuery(ctx, sb.String(), args)
}
