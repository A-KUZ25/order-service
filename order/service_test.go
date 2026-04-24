package order

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

type stubRepository struct {
	fetchUnpaidFunc            func(ctx context.Context, f UnpaidFilter) ([]int64, error)
	fetchBadReviewFunc         func(ctx context.Context, f BadReviewFilter) ([]int64, error)
	fetchExceededPriceFunc     func(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
	countOrdersWithWarningFunc func(ctx context.Context, f BaseFilter, warningIDs []int64) (int64, error)
	fetchOrdersWithWarningFunc func(ctx context.Context, f BaseFilter, warningIDs []int64, page, pageSize int) ([]FullOrder, error)
	fetchOrdersByStatusGroup   func(ctx context.Context, f BaseFilter) ([]int64, error)
	getOptionsForOrdersFunc    func(ctx context.Context, orderIDs []int64) (map[int64][]OptionDTO, error)
	getStatusChangeTimesFunc   func(ctx context.Context, keys []StatusKey) (map[StatusKey]int64, error)
}

func (s stubRepository) FetchUnpaid(ctx context.Context, f UnpaidFilter) ([]int64, error) {
	if s.fetchUnpaidFunc == nil {
		return nil, nil
	}
	return s.fetchUnpaidFunc(ctx, f)
}

func (s stubRepository) FetchBadReview(ctx context.Context, f BadReviewFilter) ([]int64, error) {
	if s.fetchBadReviewFunc == nil {
		return nil, nil
	}
	return s.fetchBadReviewFunc(ctx, f)
}

func (s stubRepository) FetchExceededPrice(ctx context.Context, f ExceededPriceFilter) ([]int64, error) {
	if s.fetchExceededPriceFunc == nil {
		return nil, nil
	}
	return s.fetchExceededPriceFunc(ctx, f)
}

func (s stubRepository) CountOrdersWithWarning(
	ctx context.Context,
	f BaseFilter,
	warningIDs []int64,
) (int64, error) {
	if s.countOrdersWithWarningFunc == nil {
		return 0, nil
	}
	return s.countOrdersWithWarningFunc(ctx, f, warningIDs)
}

func (s stubRepository) FetchOrdersWithWarning(
	ctx context.Context,
	f BaseFilter,
	warningIDs []int64,
	page,
	pageSize int,
) ([]FullOrder, error) {
	if s.fetchOrdersWithWarningFunc == nil {
		return nil, nil
	}
	return s.fetchOrdersWithWarningFunc(ctx, f, warningIDs, page, pageSize)
}

func (s stubRepository) FetchOrdersByStatusGroup(
	ctx context.Context,
	f BaseFilter,
) ([]int64, error) {
	if s.fetchOrdersByStatusGroup == nil {
		return nil, nil
	}
	return s.fetchOrdersByStatusGroup(ctx, f)
}

func (s stubRepository) GetOptionsForOrders(
	ctx context.Context,
	orderIDs []int64,
) (map[int64][]OptionDTO, error) {
	if s.getOptionsForOrdersFunc == nil {
		return nil, nil
	}
	return s.getOptionsForOrdersFunc(ctx, orderIDs)
}

func (s stubRepository) GetStatusChangeTimes(
	ctx context.Context,
	keys []StatusKey,
) (map[StatusKey]int64, error) {
	if s.getStatusChangeTimesFunc == nil {
		return nil, nil
	}
	return s.getStatusChangeTimesFunc(ctx, keys)
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

func (m *MockRepository) GetOptionsForOrders(
	ctx context.Context,
	orderIDs []int64,
) (map[int64][]OptionDTO, error) {
	args := m.Called(ctx, orderIDs)
	return args.Get(0).(map[int64][]OptionDTO), args.Error(1)
}

func (m *MockRepository) GetStatusChangeTimes(
	ctx context.Context,
	keys []StatusKey,
) (map[StatusKey]int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(map[StatusKey]int64), args.Error(1)
}

type testService struct {
	*service
	getWarningOrderFunc func(ctx context.Context, f WarningFilter) ([]int64, error)
}

func (s *testService) GetWarningOrder(ctx context.Context, f WarningFilter) ([]int64, error) {
	return s.getWarningOrderFunc(ctx, f)
}

type stubAddressParser struct {
	parseFunc func(raw string) ([]AddressOut, error)
}

func (s stubAddressParser) ParseAddress(raw string) ([]AddressOut, error) {
	return s.parseFunc(raw)
}

type stubWaitingTimeProvider struct {
	getFunc func(ctx context.Context, tenantID, orderID int64) (int64, error)
}

func (s stubWaitingTimeProvider) GetWorkerWaitingTime(
	ctx context.Context,
	tenantID, orderID int64,
) (int64, error) {
	return s.getFunc(ctx, tenantID, orderID)
}

type stubStatusTranslator struct {
	translateFunc func(ctx context.Context, language, name string) (string, error)
}

func (s stubStatusTranslator) TranslateStatus(
	ctx context.Context,
	language, name string,
) (string, error) {
	return s.translateFunc(ctx, language, name)
}

type stubShowOrderCodeProvider struct {
	shouldFunc func(ctx context.Context, tenantID, cityID, positionID int64) (bool, error)
}

func (s stubShowOrderCodeProvider) ShouldShowOrderCode(
	ctx context.Context,
	tenantID, cityID, positionID int64,
) (bool, error) {
	return s.shouldFunc(ctx, tenantID, cityID, positionID)
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

func TestGetWarningOrder_MergesDeduplicatesAndSorts(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepository)
	svc := &service{repo: repo}

	filter := WarningFilter{
		BaseFilter: BaseFilter{TenantID: 68},
	}

	repo.On(
		"FetchUnpaid",
		mock.Anything,
		UnpaidFilter{BaseFilter: filter.BaseFilter, StatusCompletedNotPaid: 0},
	).Return([]int64{6, 2, 7}, nil)
	repo.On(
		"FetchBadReview",
		mock.Anything,
		BadReviewFilter{BaseFilter: filter.BaseFilter, BadRatingMax: 0},
	).Return([]int64{3, 2}, nil)
	repo.On(
		"FetchExceededPrice",
		mock.Anything,
		ExceededPriceFilter{BaseFilter: filter.BaseFilter, MinRealPrice: 0, FinishedStatus: nil},
	).Return([]int64{7, 1}, nil)

	ids, err := svc.GetWarningOrder(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, []int64{1, 2, 3, 6, 7}, ids)
	repo.AssertExpectations(t)
}

func TestGetWarningOrder_ReturnsError(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepository)
	svc := &service{repo: repo}
	filter := WarningFilter{BaseFilter: BaseFilter{TenantID: 68}}

	repo.On("FetchUnpaid", mock.Anything, mock.AnythingOfType("order.UnpaidFilter")).
		Return([]int64(nil), errors.New("boom"))
	repo.On("FetchBadReview", mock.Anything, mock.AnythingOfType("order.BadReviewFilter")).
		Return([]int64{}, nil)
	repo.On("FetchExceededPrice", mock.Anything, mock.AnythingOfType("order.ExceededPriceFilter")).
		Return([]int64{}, nil)

	ids, err := svc.GetWarningOrder(ctx, filter)

	require.Error(t, err)
	require.Nil(t, ids)
}

func TestGetFormattedOrdersByGroup_ParsesAddressesAndLoadsOptions(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)

	svc := &service{
		repo: repo,
		addressParser: stubAddressParser{
			parseFunc: func(raw string) ([]AddressOut, error) {
				return []AddressOut{{Street: strPtr("parsed " + raw), Type: "house"}}, nil
			},
		},
	}

	filter := WarningFilter{BaseFilter: BaseFilter{Group: "new"}}
	orders := []FullOrder{
		{OrderID: 1, Address: "a:1", StatusName: "New order", StatusStatusID: 1},
		{OrderID: 2, Address: "", StatusName: "Completed", StatusStatusID: 2},
	}

	repo.On("CountOrdersWithWarning", mock.Anything, filter.BaseFilter, ([]int64)(nil)).
		Return(int64(2), nil)
	repo.On("FetchOrdersWithWarning", mock.Anything, filter.BaseFilter, ([]int64)(nil), 0, 50).
		Return(orders, nil)
	repo.On("GetOptionsForOrders", mock.Anything, []int64{1, 2}).
		Return(map[int64][]OptionDTO{
			1: {{OptionID: 10, Name: "wifi", Quantity: 1}},
		}, nil)

	count, formatted, err := svc.GetFormattedOrdersByGroup(ctx, filter, 0, 50)

	require.NoError(t, err)
	require.Equal(t, int64(2), count)
	require.Len(t, formatted, 2)
	require.Equal(t, "parsed a:1", derefString(formatted[0].Address[0].Street))
	require.Nil(t, formatted[1].Address)
	require.Equal(t, []OptionDTO{{OptionID: 10, Name: "wifi", Quantity: 1}}, formatted[0].Options)
	repo.AssertExpectations(t)
}

func TestGetFormattedOrdersByGroup_ReturnsAddressParserError(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	svc := &service{
		repo: repo,
		addressParser: stubAddressParser{
			parseFunc: func(raw string) ([]AddressOut, error) {
				return nil, errors.New("bad address")
			},
		},
	}

	filter := WarningFilter{BaseFilter: BaseFilter{Group: "new"}}
	orders := []FullOrder{{OrderID: 1, Address: "a:1"}}

	repo.On("CountOrdersWithWarning", mock.Anything, filter.BaseFilter, ([]int64)(nil)).
		Return(int64(1), nil)
	repo.On("FetchOrdersWithWarning", mock.Anything, filter.BaseFilter, ([]int64)(nil), 0, 50).
		Return(orders, nil)

	count, formatted, err := svc.GetFormattedOrdersByGroup(ctx, filter, 0, 50)

	require.Error(t, err)
	require.Zero(t, count)
	require.Nil(t, formatted)
}

func TestPrepareOrdersData_AppliesTranslationWaitTimeStatusTimeAndShowCode(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	svc := &service{
		repo: repo,
		waitingTimeProvider: stubWaitingTimeProvider{
			getFunc: func(ctx context.Context, tenantID, orderID int64) (int64, error) {
				require.Equal(t, int64(68), tenantID)
				require.Equal(t, int64(11), orderID)
				return 180, nil
			},
		},
		statusTranslator: stubStatusTranslator{
			translateFunc: func(ctx context.Context, language, name string) (string, error) {
				require.Equal(t, "ru", language)
				require.Equal(t, "New order", name)
				return "Новый заказ", nil
			},
		},
		showOrderCode: stubShowOrderCodeProvider{
			shouldFunc: func(ctx context.Context, tenantID, cityID, positionID int64) (bool, error) {
				require.Equal(t, int64(68), tenantID)
				require.Equal(t, int64(26068), cityID)
				require.Equal(t, int64(1), positionID)
				return true, nil
			},
		},
	}

	repo.On("GetStatusChangeTimes", mock.Anything, []StatusKey{{OrderID: 11, StatusID: 1}}).
		Return(map[StatusKey]int64{{OrderID: 11, StatusID: 1}: 777}, nil)

	orders := []FormattedOrder{
		{
			OrderID:              11,
			TenantID:             68,
			CityID:               26068,
			StatusID:             1,
			Device:               DeviceAndroid,
			StatusTime:           555,
			OrderTime:            1711111111,
			TimeOffset:           3600,
			Phone:                "79990000000",
			PositionID:           1,
			OrderNumber:          123456,
			OrderCode:            "ab12",
			PredvPrice:           100,
			PredvPriceNoDiscount: 150,
			CreateTime:           444,
			Status:               StatusDTO{StatusID: 1, Name: "New order"},
			Client:               ClientDTO{ClientID: 88},
			Address:              []AddressOut{{Street: strPtr("Lenina"), Type: "house"}},
		},
		{
			OrderID:     11,
			TenantID:    68,
			CityID:      26068,
			StatusID:    1,
			Device:      DeviceAndroid,
			StatusTime:  555,
			OrderTime:   1711111111,
			TimeOffset:  3600,
			Phone:       "79990000000",
			PositionID:  1,
			OrderNumber: 123456,
			OrderCode:   "ab12",
			PredvPrice:  100,
			CreateTime:  444,
			Status:      StatusDTO{StatusID: 1, Name: "New order"},
			Client:      ClientDTO{ClientID: 88},
		},
	}

	prepared, err := svc.PrepareOrdersData(ctx, orders, WarningFilter{
		BaseFilter: BaseFilter{Language: "ru", Group: "new"},
	})

	require.NoError(t, err)
	require.Len(t, prepared, 1)
	require.Equal(t, "Новый заказ", prepared[0].Status.Name)
	require.Equal(t, int64(777), prepared[0].StatusTime)
	require.Equal(t, int64(180), prepared[0].WaitTime)
	require.Equal(t, "ab12", prepared[0].OrderNumber)
	require.Equal(t, float64(150), prepared[0].SummaryCost)
	require.Equal(t, int64(1711111111-3600), prepared[0].OrderTime)
}

func TestPrepareOrdersData_UsesCompletedSummaryCostAndOrderNumberFallback(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	svc := &service{
		repo: repo,
		statusTranslator: stubStatusTranslator{
			translateFunc: func(ctx context.Context, language, name string) (string, error) {
				return "", nil
			},
		},
		showOrderCode: stubShowOrderCodeProvider{
			shouldFunc: func(ctx context.Context, tenantID, cityID, positionID int64) (bool, error) {
				return false, nil
			},
		},
	}

	repo.On("GetStatusChangeTimes", mock.Anything, []StatusKey{{OrderID: 21, StatusID: 38}}).
		Return(map[StatusKey]int64{}, nil)

	summaryNoDiscount := "900"
	orders := []FormattedOrder{{
		OrderID:               21,
		TenantID:              68,
		CityID:                26068,
		StatusID:              38,
		Device:                DeviceDispatcher,
		StatusTime:            600,
		OrderTime:             1711111111,
		Phone:                 "79990000000",
		PositionID:            2,
		OrderNumber:           999,
		OrderCode:             "code-999",
		PredvPrice:            100,
		Status:                StatusDTO{StatusID: 38, Name: "Completed, paid"},
		Client:                ClientDTO{ClientID: 90},
		SummaryCostNoDiscount: &summaryNoDiscount,
	}}

	prepared, err := svc.PrepareOrdersData(ctx, orders, WarningFilter{
		BaseFilter: BaseFilter{Language: "ru", Group: "completed"},
	})

	require.NoError(t, err)
	require.Len(t, prepared, 1)
	require.Equal(t, "Completed, paid", prepared[0].Status.Name)
	require.Equal(t, int64(999), prepared[0].OrderNumber)
	require.Equal(t, summaryNoDiscount, prepared[0].SummaryCost)
	require.Equal(t, map[string]any{"device": "Диспетчер", "user": map[string]any{
		"userId": int64(0), "name": "", "lastName": "", "secondName": (*string)(nil),
	}}, prepared[0].Dispatcher)
}

func TestPrepareOrdersData_ReturnsDependencyErrors(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	repo.On("GetStatusChangeTimes", mock.Anything, []StatusKey{{OrderID: 1, StatusID: 1}}).
		Return(map[StatusKey]int64{}, nil)

	baseOrder := FormattedOrder{
		OrderID:     1,
		TenantID:    68,
		CityID:      26068,
		StatusID:    1,
		Device:      DeviceAndroid,
		StatusTime:  10,
		OrderTime:   20,
		Phone:       "79990000000",
		PositionID:  1,
		OrderNumber: 100,
		Status:      StatusDTO{StatusID: 1, Name: "New order"},
		Client:      ClientDTO{ClientID: 1},
	}

	t.Run("translator error", func(t *testing.T) {
		svc := &service{
			repo: repo,
			statusTranslator: stubStatusTranslator{
				translateFunc: func(ctx context.Context, language, name string) (string, error) {
					return "", errors.New("translate failed")
				},
			},
		}

		prepared, err := svc.PrepareOrdersData(ctx, []FormattedOrder{baseOrder}, WarningFilter{
			BaseFilter: BaseFilter{Language: "ru"},
		})
		require.Error(t, err)
		require.Nil(t, prepared)
	})

	t.Run("wait time error", func(t *testing.T) {
		svc := &service{
			repo: repo,
			waitingTimeProvider: stubWaitingTimeProvider{
				getFunc: func(ctx context.Context, tenantID, orderID int64) (int64, error) {
					return 0, errors.New("wait failed")
				},
			},
		}

		prepared, err := svc.PrepareOrdersData(ctx, []FormattedOrder{baseOrder}, WarningFilter{
			BaseFilter: BaseFilter{Language: "ru"},
		})
		require.Error(t, err)
		require.Nil(t, prepared)
	})

	t.Run("show code error", func(t *testing.T) {
		svc := &service{
			repo: repo,
			showOrderCode: stubShowOrderCodeProvider{
				shouldFunc: func(ctx context.Context, tenantID, cityID, positionID int64) (bool, error) {
					return false, errors.New("show code failed")
				},
			},
		}

		prepared, err := svc.PrepareOrdersData(ctx, []FormattedOrder{baseOrder}, WarningFilter{
			BaseFilter: BaseFilter{Language: "ru"},
		})
		require.Error(t, err)
		require.Nil(t, prepared)
	})
}

func TestGetOrdersForTabs_MergesWarningAndBuildsSignalGroups(t *testing.T) {
	ctx := context.Background()
	repo := stubRepository{
		fetchOrdersByStatusGroup: func(ctx context.Context, f BaseFilter) ([]int64, error) {
			switch {
			case f.SelectForDate == false && requireStatusSet(f.Status, orderGroupIds[StatusGroup0]):
				return []int64{1, 2}, nil
			case f.SelectForDate == false && requireStatusSet(f.Status, orderGroupIds[StatusGroup6]):
				return []int64{6}, nil
			case f.SelectForDate == true && requireStatusSet(f.Status, orderGroupIds[StatusGroup7]):
				return []int64{7, 8}, nil
			case f.SelectForDate == false && requireStatusSet(f.Status, orderGroupIds[StatusGroup8]):
				return []int64{10, 11, 12}, nil
			default:
				t.Fatalf("unexpected filter: %+v", f)
				return nil, nil
			}
		},
		fetchUnpaidFunc: func(ctx context.Context, f UnpaidFilter) ([]int64, error) {
			return []int64{8, 9}, nil
		},
		fetchBadReviewFunc: func(ctx context.Context, f BadReviewFilter) ([]int64, error) {
			return []int64{9, 13}, nil
		},
		fetchExceededPriceFunc: func(ctx context.Context, f ExceededPriceFilter) ([]int64, error) {
			return []int64{14}, nil
		},
	}
	svc := &service{repo: repo}

	filter := WarningFilter{
		BaseFilter: BaseFilter{
			TenantID:      68,
			CityIDs:       []int64{26068},
			SelectForDate: false,
		},
	}

	result, err := svc.GetOrdersForTabs(ctx, filter)

	require.NoError(t, err)
	require.Equal(t, map[StatusGroup]int{
		StatusGroup0: 2,
		StatusGroup6: 1,
		StatusGroup7: 5,
		StatusGroup8: 3,
	}, result.GroupCounts)
	require.ElementsMatch(t, []int64{1, 2}, result.OrdersForSignal[StatusGroup0])
	require.ElementsMatch(t, []int64{6}, result.OrdersForSignal[StatusGroup6])
	require.Len(t, result.OrdersForSignal, 2)
}

func TestGetOrdersForTabs_ReturnsGroupFetchError(t *testing.T) {
	ctx := context.Background()
	repo := stubRepository{
		fetchOrdersByStatusGroup: func(ctx context.Context, f BaseFilter) ([]int64, error) {
			return nil, errors.New("group failed")
		},
	}
	svc := &service{repo: repo}

	filter := WarningFilter{
		BaseFilter: BaseFilter{TenantID: 68},
	}

	result, err := svc.GetOrdersForTabs(ctx, filter)

	require.Error(t, err)
	require.Equal(t, GroupOrdersResult{}, result)
}

func TestShowCodeOrID(t *testing.T) {
	require.Equal(t, "q4ccf", ShowCodeOrID(true, "q4ccf", 123456))
	require.Equal(t, int64(123456), ShowCodeOrID(true, "", 123456))
	require.Equal(t, int64(123456), ShowCodeOrID(false, "q4ccf", 123456))
}

func TestGetCategoryAndDeviceName(t *testing.T) {
	require.Equal(t, "warning", GetCategory(45))
	require.Equal(t, "warning", GetCategory(38))
	require.Equal(t, "", GetCategory(999999))

	require.Equal(t, "Android", GetDeviceName(DeviceAndroid))
	require.Equal(t, "Диспетчер", GetDeviceName(DeviceDispatcher))
	require.Equal(t, "", GetDeviceName("UNKNOWN"))
}

func requireStatusSet(got, expected []int64) bool {
	if len(got) != len(expected) {
		return false
	}

	left := append([]int64(nil), got...)
	right := append([]int64(nil), expected...)
	sort.Slice(left, func(i, j int) bool { return left[i] < left[j] })
	sort.Slice(right, func(i, j int) bool { return right[i] < right[j] })
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func strPtr(v string) *string {
	return &v
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
