package mysql

import (
	"strings"
	"testing"

	"orders-service/internal/app/order"

	"github.com/stretchr/testify/require"
)

func TestNormalizeSortField(t *testing.T) {
	require.Equal(t, "o.status_time", normalizeSortField(""))
	require.Equal(t, "o.order_id", normalizeSortField("o.order_id"))
	require.Equal(t, "o.order_time", normalizeSortField("o.order_time"))
	require.Equal(t, "o.status_time", normalizeSortField("o.order_id; DROP TABLE tbl_order;"))
}

func TestNormalizeSortDirection(t *testing.T) {
	require.Equal(t, "ASC", normalizeSortDirection("asc"))
	require.Equal(t, "ASC", normalizeSortDirection("ASC"))
	require.Equal(t, "DESC", normalizeSortDirection("desc"))
	require.Equal(t, "DESC", normalizeSortDirection("desc; drop table"))
}

func TestOrdersRepository_AppendOrderBySanitizesInput(t *testing.T) {
	repo := &OrdersRepository{}
	var sb strings.Builder

	repo.appendOrderBy(&sb, order.BaseFilter{
		SortField: "o.order_id; DROP TABLE tbl_order;",
		SortOrder: "asc; DROP TABLE tbl_order;",
	})

	require.Equal(t, "ORDER BY o.status_time DESC\n", sb.String())
}
