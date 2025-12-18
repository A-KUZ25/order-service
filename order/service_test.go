package order

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) FetchUnpaid(ctx context.Context, f UnpaidFilter) ([]int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRepository) FetchBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRepository) FetchExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRepository) FetchWarningStatus(ctx context.Context, f WarningFilter) ([]int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]int64), args.Error(1)
}

func (m *MockRepository) CountOrdersWithWarning(
	ctx context.Context,
	f BaseFilter,
	warningIDs []int64,
) (int64, error) {
	args := m.Called(ctx, f, warningIDs)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) FetchOrdersWithWarning(
	ctx context.Context,
	f BaseFilter,
	warningIDs []int64,
	page,
	pageSize int,
) ([]FullOrder, error) {
	args := m.Called(ctx, f, warningIDs, page, pageSize)
	return args.Get(0).([]FullOrder), args.Error(1)
}

func (m *MockRepository) FetchOrdersByStatusGroup(
	ctx context.Context,
	f BaseFilter,
) ([]int64, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]int64), args.Error(1)
}

type testService struct {
	*service
	getWarningOrderFunc func(ctx context.Context, f WarningFilter) ([]int64, error)
}

func (s *testService) GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error) {
	return s.getWarningOrderFunc(ctx, f)
}

func TestGetOrdersByGroup_NotWarning(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepository)
	svc := &service{repo: repo}

	filter := WarningFilter{
		BaseFilter: BaseFilter{
			Group: "new",
		},
	}

	expectedCount := int64(2)
	expectedOrders := []FullOrder{
		{OrderID: 1},
		{OrderID: 2},
	}

	repo.On(
		"CountOrdersWithWarning",
		mock.Anything,
		filter.BaseFilter,
		([]int64)(nil),
	).Return(expectedCount, nil)

	repo.On(
		"FetchOrdersWithWarning",
		mock.Anything,
		filter.BaseFilter,
		([]int64)(nil),
		1,
		10,
	).Return(expectedOrders, nil)

	count, orders, err := svc.GetOrdersByGroup(ctx, filter, 1, 10)

	require.NoError(t, err)
	require.Equal(t, expectedCount, count)
	require.Len(t, orders, 2)
	require.Equal(t, int64(1), orders[0].OrderID)

	repo.AssertExpectations(t)
}

func TestGetOrdersByGroup_Warning_FullMock(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepository)
	svc := &service{repo: repo}

	filter := WarningFilter{
		BaseFilter: BaseFilter{
			Group: "warning",
		},
	}

	// ---------- МОКИ ДЛЯ GetWarningOrder ----------

	// 1) unpaid
	repo.On(
		"FetchUnpaid",
		mock.Anything,
		mock.AnythingOfType("order.UnpaidFilter"),
	).Return([]int64{6, 7}, nil)

	// 2) bad reviews
	repo.On(
		"FetchBadReview",
		mock.Anything,
		mock.AnythingOfType("order.BadReviewFilter"),
	).Return([]int64{2, 3}, nil)

	// 3) exceeded price
	repo.On(
		"FetchExceededPrice",
		mock.Anything,
		mock.AnythingOfType("order.ExceededPriceFilter"),
	).Return([]int64{2, 6}, nil)

	// ---------- ИТОГОВЫЕ ВЫЗОВЫ ----------

	repo.On(
		"CountOrdersWithWarning",
		mock.Anything,
		filter.BaseFilter,
		[]int64{2, 3, 6, 7},
	).Return(int64(4), nil)

	expectedOrders := []FullOrder{
		{OrderID: 1},
		{OrderID: 2},
	}

	repo.On(
		"FetchOrdersWithWarning",
		mock.Anything,
		filter.BaseFilter,
		[]int64{2, 3, 6, 7},
		1,
		10,
	).Return(expectedOrders, nil)

	// ---------- ВЫЗОВ ----------

	count, orders, err := svc.GetOrdersByGroup(ctx, filter, 1, 10)

	// ---------- ПРОВЕРКИ ----------

	require.NoError(t, err)
	require.Equal(t, int64(4), count)
	require.Len(t, orders, 2)
	require.Equal(t, int64(1), orders[0].OrderID)

	repo.AssertExpectations(t)
}
