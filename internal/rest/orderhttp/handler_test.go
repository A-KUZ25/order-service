package orderhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"orders-service/order"

	"github.com/stretchr/testify/require"
)

type stubService struct {
	getFormattedOrdersByGroupFunc func(
		ctx context.Context,
		f order.WarningFilter,
		page, pageSize int,
	) (int64, []order.FormattedOrder, error)
	getOrdersForTabsFunc func(
		ctx context.Context,
		f order.WarningFilter,
	) (order.GroupOrdersResult, error)
	prepareOrdersDataFunc func(
		ctx context.Context,
		orders []order.FormattedOrder,
		f order.WarningFilter,
	) ([]order.PreparedOrder, error)
}

func (s stubService) GetWarningOrder(ctx context.Context, f order.WarningFilter) ([]int64, error) {
	return nil, nil
}

func (s stubService) GetFormattedOrdersByGroup(
	ctx context.Context,
	f order.WarningFilter,
	page, pageSize int,
) (int64, []order.FormattedOrder, error) {
	return s.getFormattedOrdersByGroupFunc(ctx, f, page, pageSize)
}

func (s stubService) GetOrdersForTabs(
	ctx context.Context,
	f order.WarningFilter,
) (order.GroupOrdersResult, error) {
	return s.getOrdersForTabsFunc(ctx, f)
}

func (s stubService) PrepareOrdersData(
	ctx context.Context,
	orders []order.FormattedOrder,
	f order.WarningFilter,
) ([]order.PreparedOrder, error) {
	return s.prepareOrdersDataFunc(ctx, orders, f)
}

func TestOrders_BadJSON(t *testing.T) {
	handler := NewHandler(stubService{})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	handler.Orders(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "Bad Request", body["error"])
	require.NotEmpty(t, body["description"])
}

func TestOrders_SuccessUsesDefaultsAndBuildsResponse(t *testing.T) {
	date := "21.03.2026"
	var gotFilter order.WarningFilter
	var gotPage, gotPageSize int

	service := stubService{
		getFormattedOrdersByGroupFunc: func(
			ctx context.Context,
			f order.WarningFilter,
			page, pageSize int,
		) (int64, []order.FormattedOrder, error) {
			gotFilter = f
			gotPage = page
			gotPageSize = pageSize
			return 2, []order.FormattedOrder{{OrderID: 1}}, nil
		},
		getOrdersForTabsFunc: func(
			ctx context.Context,
			f order.WarningFilter,
		) (order.GroupOrdersResult, error) {
			require.Equal(t, "ru", f.BaseFilter.Language)
			return order.GroupOrdersResult{
				GroupCounts: map[order.StatusGroup]int{
					order.StatusGroup0: 3,
				},
				OrdersForSignal: map[order.StatusGroup][]int64{
					order.StatusGroup0: {1, 2},
				},
			}, nil
		},
		prepareOrdersDataFunc: func(
			ctx context.Context,
			orders []order.FormattedOrder,
			f order.WarningFilter,
		) ([]order.PreparedOrder, error) {
			require.Len(t, orders, 1)
			require.Equal(t, "ru", f.BaseFilter.Language)
			return []order.PreparedOrder{{
				ID:          1,
				OrderNumber: "q4ccf",
			}}, nil
		},
	}

	handler := NewHandler(service)
	reqBody := WarningFullRequest{
		OrderBaseRequest: OrderBaseRequest{
			TenantID:      68,
			CityIDs:       []int64{26068},
			Language:      "ru",
			Date:          &date,
			UserPositions: []int64{1, 2},
			SortField:     "o.status_time",
			SortOrder:     "desc",
		},
		Group: "warning",
		Page:  -1,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Orders(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 0, gotPage)
	require.Equal(t, 50, gotPageSize)
	require.Equal(t, int64(68), gotFilter.BaseFilter.TenantID)
	require.Equal(t, []int64{26068}, gotFilter.BaseFilter.CityIDs)
	require.Equal(t, "ru", gotFilter.BaseFilter.Language)
	require.Equal(t, "warning", gotFilter.BaseFilter.Group)
	require.Equal(t, "o.status_time", gotFilter.BaseFilter.SortField)
	require.Equal(t, "desc", gotFilter.BaseFilter.SortOrder)

	var resp ordersResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Equal(t, int64(2), resp.OrderTotalCount)
	require.Equal(t, 50, resp.CountPerPage)
	require.Equal(t, map[order.StatusGroup]int{order.StatusGroup0: 3}, resp.OrderCounts)
	require.Equal(t, map[order.StatusGroup][]int64{order.StatusGroup0: {1, 2}}, resp.OrdersForSignal)
	require.Len(t, resp.Orders, 1)
	require.Equal(t, "q4ccf", resp.Orders[0].OrderNumber)
}

func TestOrders_ReturnsServiceError(t *testing.T) {
	service := stubService{
		getFormattedOrdersByGroupFunc: func(
			ctx context.Context,
			f order.WarningFilter,
			page, pageSize int,
		) (int64, []order.FormattedOrder, error) {
			return 0, nil, errors.New("formatted failed")
		},
		getOrdersForTabsFunc: func(
			ctx context.Context,
			f order.WarningFilter,
		) (order.GroupOrdersResult, error) {
			return order.GroupOrdersResult{}, nil
		},
		prepareOrdersDataFunc: func(
			ctx context.Context,
			orders []order.FormattedOrder,
			f order.WarningFilter,
		) ([]order.PreparedOrder, error) {
			return nil, nil
		},
	}

	handler := NewHandler(service)
	body, err := json.Marshal(WarningFullRequest{
		OrderBaseRequest: OrderBaseRequest{
			TenantID: 68,
			Language: "ru",
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Orders(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Contains(t, rec.Body.String(), "internal error: formatted failed")
}
