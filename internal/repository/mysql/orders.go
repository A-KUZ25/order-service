package mysql

import (
	"context"
	"database/sql"
	"strings"

	"orders-service/order"
)

type OrdersRepository struct {
	db *sql.DB
}

func NewOrdersRepository(db *sql.DB) (*OrdersRepository, error) {
	return &OrdersRepository{db: db}, nil
}

// -----------------------------------------------------------
//  COMMON SQL BUILDERS
// -----------------------------------------------------------

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

// добавляем ORDER BY
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

// выполняет SQL и возвращает IDs
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

// -----------------------------------------------------------
//  UNPAID
// -----------------------------------------------------------

func (r *OrdersRepository) FetchUnpaidOrderIDs(
	ctx context.Context,
	f order.UnpaidFilter,
) ([]int64, error) {

	var sb strings.Builder
	args := []any{}

	sb.WriteString(`
SELECT o.order_id
FROM tbl_order o
`)

	// Общая часть
	r.buildBaseQuery(&sb, &args, f.BaseFilter)

	// Специфичная часть unpaid
	sb.WriteString("  AND o.status_id = ?\n")
	args = append(args, f.StatusCompletedNotPaid)

	// OrderBy
	r.appendOrderBy(&sb, f.BaseFilter)

	// Final
	return r.executeQuery(ctx, sb.String(), args)
}

// -----------------------------------------------------------
//  BAD REVIEWS
// -----------------------------------------------------------

func (r *OrdersRepository) FetchBadReviewOrderIDs(
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

	// Общая часть
	r.buildBaseQuery(&sb, &args, f.BaseFilter)

	// специфично bad reviews
	sb.WriteString("  AND cr.rating BETWEEN 1 AND ?\n")
	args = append(args, f.BadRatingMax)

	r.appendOrderBy(&sb, f.BaseFilter)

	return r.executeQuery(ctx, sb.String(), args)
}
