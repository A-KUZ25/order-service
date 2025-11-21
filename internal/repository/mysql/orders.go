package mysql

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"orders-service/order"
)

// OrdersRepository — реализация доменного интерфейса orderhttp.Repository для MySQL.
type OrdersRepository struct {
	db *sql.DB

	stmtUnpaidBase *sql.Stmt
}

// NewOrdersRepository — конструктор репозитория заказов.
func NewOrdersRepository(db *sql.DB) (*OrdersRepository, error) {
	stmt, err := db.Prepare(QueryUnpaidOrdersBase)
	if err != nil {
		return nil, err
	}

	return &OrdersRepository{
		db:             db,
		stmtUnpaidBase: stmt,
	}, nil
}

// Compile-time проверка: OrdersRepository реализует orderhttp.Repository.
var _ order.Repository = (*OrdersRepository)(nil)

func (r *OrdersRepository) FetchUnpaidOrderIDs(
	ctx context.Context,
	filter order.UnpaidOrdersFilter,
) ([]int64, error) {

	// 1) Выполняем базовый подготовленный запрос
	rows, err := r.stmtUnpaidBase.QueryContext(ctx,
		filter.TenantID,
		filter.StatusCompletedNotPaid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 2) Собираем ID, прошедшие базовые условия
	baseIDs := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		baseIDs = append(baseIDs, id)
	}

	// Если нет базовых ID → нечего фильтровать далее
	if len(baseIDs) == 0 {
		return []int64{}, nil
	}

	// 3) Строим динамический SQL для дополнительной фильтрации
	var (
		sb   strings.Builder
		args []any
	)

	sb.WriteString(
		"SELECT o.order_id\n" +
			"FROM tbl_order o\n" +
			"WHERE o.order_id IN (",
	)
	for i, id := range baseIDs {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("?")
		args = append(args, id)
	}
	sb.WriteString(")\n")

	// фильтр по city_id (список городов)
	if len(filter.CityIDs) > 0 {
		sb.WriteString("  AND o.city_id IN (")
		placeholders := make([]string, 0, len(filter.CityIDs))

		for _, id := range filter.CityIDs {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}

		sb.WriteString(strings.Join(placeholders, ","))
		sb.WriteString(")\n")
	}

	// status_time BETWEEN
	if filter.StatusTimeFrom != nil && filter.StatusTimeTo != nil {
		sb.WriteString("AND o.status_time BETWEEN ? AND ?\n")
		args = append(args, *filter.StatusTimeFrom, *filter.StatusTimeTo)
	}

	// tariffs
	if len(filter.Tariffs) > 0 {
		sb.WriteString("AND o.tariff_id IN (")
		for i, t := range filter.Tariffs {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			args = append(args, t)
		}
		sb.WriteString(")\n")
	}

	// positions
	if len(filter.UserPositions) > 0 {
		sb.WriteString("AND o.position_id IN (")
		for i, p := range filter.UserPositions {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			args = append(args, p)
		}
		sb.WriteString(")\n")
	}

	// sort
	orderField := filter.SortField
	if orderField == "" {
		orderField = "o.status_time"
	}
	orderDir := "DESC"
	if strings.ToLower(string(filter.SortOrder)) == "asc" {
		orderDir = "ASC"
	}
	sb.WriteString("ORDER BY " + orderField + " " + orderDir + "\n")

	// 4) Выполняем второй SQL как обычный QueryContext
	finalSQL := sb.String()
	log.Printf("\nGO SQL:\n%s\nARGS: %+v\n", finalSQL, args)
	rows2, err := r.db.QueryContext(ctx, finalSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	result := []int64{}
	for rows2.Next() {
		var id int64
		if err := rows2.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}

	return result, rows2.Err()

}
