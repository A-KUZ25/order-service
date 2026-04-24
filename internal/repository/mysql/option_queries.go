package mysql

import (
	"context"
	"database/sql"
	"log"
	"orders-service/order"
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

	start := time.Now()
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	log.Println("BASE REQUEST TIME:", time.Since(start))
	defer rows.Close()

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
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
