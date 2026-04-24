package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"orders-service/order"
	"strings"
	"time"
)

type OrdersRepository struct {
	db *sql.DB
}

func NewOrdersRepository(db *sql.DB) (*OrdersRepository, error) {
	return &OrdersRepository{db: db}, nil
}

// строит общую часть WHERE (tenant, active, date-range, city, tariffs…)
func (r *OrdersRepository) buildBaseQuery(sb *strings.Builder, args *[]any, f order.BaseFilter, warn ...bool) {
	warning := false
	if len(warn) > 0 {
		warning = true
	}
	sb.WriteString(" AND o.tenant_id = ? ")
	*args = append(*args, f.TenantID)

	sb.WriteString(" AND o.active = 1 ")

	// date
	if f.SelectForDate && f.StatusTimeFrom != nil && f.StatusTimeTo != nil {
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

	// status
	if !warning {
		if len(f.Status) > 0 {
			sb.WriteString(" AND o.status_id IN (")
			for i, st := range f.Status {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString("?")
				*args = append(*args, st)
			}
			sb.WriteString(")\n")
		}
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

	start := time.Now()
	rows, err := r.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	log.Println("BASE REQUEST TIME:", time.Since(start))

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

func formatArg(a any) string {
	switch v := a.(type) {
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case []byte:
		return "'" + strings.ReplaceAll(string(v), "'", "''") + "'"
	case time.Time:
		return "'" + v.Format("2006-01-02 15:04:05") + "'"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}
