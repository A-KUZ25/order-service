package mysql

import (
	"context"
	"database/sql"
	"orders-service/internal/app/order"
	"orders-service/internal/logging"
	"strings"
	"time"
)

func (r *OrdersRepository) GetOptionsForOrders(
	ctx context.Context,
	orderIDs []int64,
) (map[int64][]order.OptionDTO, error) {
	result := make(map[int64][]order.OptionDTO)

	if len(orderIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, 0, len(orderIDs))
	args := make([]any, 0, len(orderIDs))
	for _, id := range orderIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := `
		SELECT
			oho.order_id,
			oho.option_id,
			co.name,
			oho.quantity
		FROM tbl_order_has_option oho
		LEFT JOIN tbl_car_option co
		       ON co.option_id = oho.option_id
		WHERE oho.order_id IN (` + strings.Join(placeholders, ",") + `)
	`

	totalStarted := time.Now()
	queryStarted := time.Now()
	rows, err := r.db.QueryContext(ctx, query, args...)
	queryMS := time.Since(queryStarted).Milliseconds()
	if err != nil {
		logging.Error(ctx, "mysql options query failed", err,
			"query_ms", queryMS,
			"order_ids_count", len(orderIDs),
		)
		return nil, err
	}
	defer rows.Close()

	scanStarted := time.Now()
	rowCount := 0
	for rows.Next() {
		var (
			orderID  int64
			optionID int64
			name     sql.NullString
			quantity sql.NullInt64
		)

		if err := rows.Scan(&orderID, &optionID, &name, &quantity); err != nil {
			return nil, err
		}

		result[orderID] = append(result[orderID], order.OptionDTO{
			OptionID: optionID,
			Name:     name.String,
			Quantity: quantity.Int64,
		})
		rowCount++
	}
	scanMS := time.Since(scanStarted).Milliseconds()
	totalMS := time.Since(totalStarted).Milliseconds()

	if err := rows.Err(); err != nil {
		logging.Error(ctx, "mysql options rows failed", err,
			"query_ms", queryMS,
			"scan_ms", scanMS,
			"total_ms", totalMS,
			"order_ids_count", len(orderIDs),
			"row_count", rowCount,
		)
		return nil, err
	}

	logging.Info(ctx, "mysql options timings",
		"query_ms", queryMS,
		"scan_ms", scanMS,
		"total_ms", totalMS,
		"order_ids_count", len(orderIDs),
		"orders_with_options_count", len(result),
		"row_count", rowCount,
	)

	return result, nil
}
