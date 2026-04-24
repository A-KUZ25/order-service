package mysql

import (
	"context"
	"orders-service/order"
	"strings"
)

func (r *OrdersRepository) GetStatusChangeTimes(
	ctx context.Context,
	keys []order.StatusKey,
) (map[order.StatusKey]int64, error) {
	if len(keys) == 0 {
		return map[order.StatusKey]int64{}, nil
	}

	var (
		args   []any
		values []string
	)

	for _, k := range keys {
		values = append(values, "(?, ?)")
		args = append(args, k.OrderID, k.StatusID)
	}

	query := `
		SELECT order_id, change_val AS status_id, change_time
		FROM tbl_order_change_data
		WHERE change_field = 'status_id'
		  AND (order_id, change_val) IN (` + strings.Join(values, ",") + `)
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[order.StatusKey]int64, len(keys))
	for rows.Next() {
		var (
			orderID  int64
			statusID int64
			timeVal  int64
		)

		if err := rows.Scan(&orderID, &statusID, &timeVal); err != nil {
			return nil, err
		}

		result[order.StatusKey{
			OrderID:  orderID,
			StatusID: statusID,
		}] = timeVal
	}

	return result, nil
}
