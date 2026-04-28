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
	fetchUnpaidFunc             func(ctx context.Context, f UnpaidFilter) ([]int64, error)
	fetchBadReviewFunc          func(ctx context.Context, f BadReviewFilter) ([]int64, error)
	fetchExceededPriceFunc      func(ctx context.Context, f ExceededPriceFilter) ([]int64, error)
	fetchAllOrdersForGetAllFunc func(ctx context.Context, f GetAllOrdersFilter) ([]FullOrder, error)
	countOrdersWithWarningFunc  func(ctx context.Context, f BaseFilter, warningIDs []int64) (int64, error)
	fetchOrdersWithWarningFunc  func(ctx context.Context, f BaseFilter, warningIDs []int64, page, pageSize int) ([]FullOrder, error)
	fetchOrdersByStatusGroup    func(ctx context.Context, f BaseFilter) ([]int64, error)
	getOptionsForOrdersFunc     func(ctx context.Context, orderIDs []int64) (map[int64][]OptionDTO, error)
	getStatusChangeTimesFunc    func(ctx context.Context, keys []StatusKey) (map[StatusKey]int64, error)
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

func (s stubRepository) FetchAllOrdersForGetAll(
	ctx context.Context,
	f GetAllOrdersFilter,
) ([]FullOrder, error) {
	if s.fetchAllOrdersForGetAllFunc == nil {
		return nil, nil
	}
	return s.fetchAllOrdersForGetAllFunc(ctx, f)
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

func (m *MockRepository) FetchAllOrdersForGetAll(
	ctx context.Context,
	f GetAllOrdersFilter,
) ([]FullOrder, error) {
	args := m.Called(ctx, f)
	return args.Get(0).([]FullOrder), args.Error(1)
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

type stubAddressResolver struct {
	resolveFunc func(orders []FullOrder) (map[int64][]AddressView, error)
}

func (s stubAddressResolver) ResolveAddresses(orders []FullOrder) (map[int64][]AddressView, error) {
	return s.resolveFunc(orders)
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

type stubActiveOrdersReader struct {
	getFunc func(ctx context.Context, tenantID int64) ([]FormattedOrder, error)
}

func (s stubActiveOrdersReader) GetFormattedActiveOrders(
	ctx context.Context,
	tenantID int64,
) ([]FormattedOrder, error) {
	if s.getFunc == nil {
		return nil, nil
	}
	return s.getFunc(ctx, tenantID)
}

type testOrderViewAssembler struct {
	waitingTimeProvider WaitingTimeProvider
	statusTranslator    StatusTranslator
	showOrderCode       ShowOrderCodeProvider
}

func newTestOrderViewAssembler(
	waitingTimeProvider WaitingTimeProvider,
	statusTranslator StatusTranslator,
	showOrderCode ShowOrderCodeProvider,
) OrderViewAssembler {
	return &testOrderViewAssembler{
		waitingTimeProvider: waitingTimeProvider,
		statusTranslator:    statusTranslator,
		showOrderCode:       showOrderCode,
	}
}

func (a *testOrderViewAssembler) BuildOrderView(
	ctx context.Context,
	o FormattedOrder,
	f WarningFilter,
	statusChangeTimes map[StatusKey]int64,
) (OrderView, error) {
	statusName := o.Status.Name
	if a.statusTranslator != nil && statusName != "" {
		translated, err := a.statusTranslator.TranslateStatus(ctx, f.BaseFilter.Language, statusName)
		if err != nil {
			return OrderView{}, err
		}
		if translated != "" {
			statusName = translated
		}
	}

	waitTime := int64(0)
	if a.waitingTimeProvider != nil {
		value, err := a.waitingTimeProvider.GetWorkerWaitingTime(ctx, o.TenantID, o.OrderID)
		if err != nil {
			return OrderView{}, err
		}
		waitTime = value
	}

	showCode := false
	if a.showOrderCode != nil {
		value, err := a.showOrderCode.ShouldShowOrderCode(ctx, o.TenantID, o.CityID, o.PositionID)
		if err != nil {
			return OrderView{}, err
		}
		showCode = value
	}

	return OrderView{
		ID:             o.OrderID,
		OrderNumber:    ShowCodeOrID(showCode, o.OrderCode, o.OrderNumber),
		OrderIDForSort: o.OrderNumber,
		Status: OrderStatusView{
			StatusID: o.Status.StatusID,
			Name:     statusName,
			Category: GetCategory(o.StatusID),
			Color:    GetColor(o.StatusID),
		},
		DateForSort:  formatOrderTimeForSort(o.OrderTime),
		Date:         formatOrderTime(o.OrderTime),
		Address:      o.Address,
		CityID:       o.CityID,
		Phone:        o.Phone,
		Device:       o.Device,
		DeviceName:   GetDeviceName(o.Device),
		Client:       ClientView{ClientID: o.Client.ClientID, Phone: o.Client.Phone, Name: o.Client.Name, LastName: o.Client.LastName},
		Dispatcher:   BuildDispatcher(o),
		Worker:       BuildWorker(o),
		Car:          BuildCar(o),
		Tariff:       BuildTariff(o),
		Options:      o.Options,
		Comment:      o.Comment,
		SummaryCost:  resolveSummaryCost(o, f.BaseFilter.Group),
		StatusTime:   getTimeOrderStatusChanged(o.OrderID, o.StatusID, o.StatusTime, statusChangeTimes),
		TimeToClient: o.TimeToClient,
		WaitTime:     waitTime,
		CreateTime:   o.CreateTime,
		OrderTime:    o.OrderTime - o.TimeOffset,
		PositionID:   o.PositionID,
		UnitQuantity: o.UnitQuantity,
	}, nil
}

func newServiceWithRepo(repo Repository) *service {
	return &service{
		warningReader:      repo,
		orderListReader:    repo,
		groupOrderReader:   repo,
		optionsReader:      repo,
		statusChangeReader: repo,
		addressResolver:    stubAddressResolver{resolveFunc: func(orders []FullOrder) (map[int64][]AddressView, error) { return map[int64][]AddressView{}, nil }},
		assembler:          newTestOrderViewAssembler(nil, nil, nil),
	}
}

func TestGetOrdersByGroup_NotWarning(t *testing.T) {
	ctx := context.Background()

	repo := new(MockRepository)
	svc := newServiceWithRepo(repo)

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
	svc := newServiceWithRepo(repo)

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
	svc := newServiceWithRepo(repo)

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
	svc := newServiceWithRepo(repo)
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
		warningReader:      repo,
		orderListReader:    repo,
		groupOrderReader:   repo,
		optionsReader:      repo,
		statusChangeReader: repo,
		addressResolver: stubAddressResolver{
			resolveFunc: func(orders []FullOrder) (map[int64][]AddressView, error) {
				return map[int64][]AddressView{
					1: {{Street: strPtr("parsed a:1"), Type: "house"}},
				}, nil
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

func TestGetFormattedOrdersByGroup_ReturnsAddressResolverError(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	svc := &service{
		warningReader:      repo,
		orderListReader:    repo,
		groupOrderReader:   repo,
		optionsReader:      repo,
		statusChangeReader: repo,
		addressResolver: stubAddressResolver{
			resolveFunc: func(orders []FullOrder) (map[int64][]AddressView, error) {
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
		warningReader:      repo,
		orderListReader:    repo,
		groupOrderReader:   repo,
		optionsReader:      repo,
		statusChangeReader: repo,
		assembler: newTestOrderViewAssembler(
			stubWaitingTimeProvider{
				getFunc: func(ctx context.Context, tenantID, orderID int64) (int64, error) {
					require.Equal(t, int64(68), tenantID)
					require.Equal(t, int64(11), orderID)
					return 180, nil
				},
			},
			stubStatusTranslator{
				translateFunc: func(ctx context.Context, language, name string) (string, error) {
					require.Equal(t, "ru", language)
					require.Equal(t, "New order", name)
					return "Новый заказ", nil
				},
			},
			stubShowOrderCodeProvider{
				shouldFunc: func(ctx context.Context, tenantID, cityID, positionID int64) (bool, error) {
					require.Equal(t, int64(68), tenantID)
					require.Equal(t, int64(26068), cityID)
					require.Equal(t, int64(1), positionID)
					return true, nil
				},
			},
		),
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
			Address:              []AddressView{{Street: strPtr("Lenina"), Type: "house"}},
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
		warningReader:      repo,
		orderListReader:    repo,
		groupOrderReader:   repo,
		optionsReader:      repo,
		statusChangeReader: repo,
		assembler: newTestOrderViewAssembler(
			nil,
			stubStatusTranslator{
				translateFunc: func(ctx context.Context, language, name string) (string, error) {
					return "", nil
				},
			},
			stubShowOrderCodeProvider{
				shouldFunc: func(ctx context.Context, tenantID, cityID, positionID int64) (bool, error) {
					return false, nil
				},
			},
		),
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
			warningReader:      repo,
			orderListReader:    repo,
			groupOrderReader:   repo,
			optionsReader:      repo,
			statusChangeReader: repo,
			assembler: newTestOrderViewAssembler(
				nil,
				stubStatusTranslator{
					translateFunc: func(ctx context.Context, language, name string) (string, error) {
						return "", errors.New("translate failed")
					},
				},
				nil,
			),
		}

		prepared, err := svc.PrepareOrdersData(ctx, []FormattedOrder{baseOrder}, WarningFilter{
			BaseFilter: BaseFilter{Language: "ru"},
		})
		require.Error(t, err)
		require.Nil(t, prepared)
	})

	t.Run("wait time error", func(t *testing.T) {
		svc := &service{
			warningReader:      repo,
			orderListReader:    repo,
			groupOrderReader:   repo,
			optionsReader:      repo,
			statusChangeReader: repo,
			assembler: newTestOrderViewAssembler(
				stubWaitingTimeProvider{
					getFunc: func(ctx context.Context, tenantID, orderID int64) (int64, error) {
						return 0, errors.New("wait failed")
					},
				},
				nil,
				nil,
			),
		}

		prepared, err := svc.PrepareOrdersData(ctx, []FormattedOrder{baseOrder}, WarningFilter{
			BaseFilter: BaseFilter{Language: "ru"},
		})
		require.Error(t, err)
		require.Nil(t, prepared)
	})

	t.Run("show code error", func(t *testing.T) {
		svc := &service{
			warningReader:      repo,
			orderListReader:    repo,
			groupOrderReader:   repo,
			optionsReader:      repo,
			statusChangeReader: repo,
			assembler: newTestOrderViewAssembler(
				nil,
				nil,
				stubShowOrderCodeProvider{
					shouldFunc: func(ctx context.Context, tenantID, cityID, positionID int64) (bool, error) {
						return false, errors.New("show code failed")
					},
				},
			),
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
			case f.SelectForDate == false && requireStatusSet(f.Status, orderGroupIDs[StatusGroup0]):
				return []int64{1, 2}, nil
			case f.SelectForDate == false && requireStatusSet(f.Status, orderGroupIDs[StatusGroup6]):
				return []int64{6}, nil
			case f.SelectForDate == true && requireStatusSet(f.Status, orderGroupIDs[StatusGroup7]):
				return []int64{7, 8}, nil
			case f.SelectForDate == false && requireStatusSet(f.Status, orderGroupIDs[StatusGroup8]):
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
	svc := newServiceWithRepo(repo)

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
	svc := newServiceWithRepo(repo)

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

func TestMatchesSearchStatus_IncludesPreOrdersForWorksAndActive(t *testing.T) {
	require.True(t, matchesSearchStatus(6, "works"))
	require.True(t, matchesSearchStatus(6, "active"))
	require.True(t, matchesSearchStatus(6, "pre_order"))
	require.True(t, matchesSearchStatus(29, "works"))
	require.True(t, matchesSearchStatus(1, "active"))
}

func TestMergeGetAllOrders_KeepsPreOrdersAndDuplicatesForCountParity(t *testing.T) {
	filter := GetAllOrdersFilter{
		SearchStatus: "all",
	}

	mysqlFormatted := []FormattedOrder{
		{OrderID: 1, StatusID: 29},
		{OrderID: 2, StatusID: 37},
	}
	redisFormatted := []FormattedOrder{
		{OrderID: 1, StatusID: 29},
		{OrderID: 3, StatusID: 6},
	}

	merged := mergeGetAllOrders(mysqlFormatted, redisFormatted, filter)

	require.Len(t, merged, 4)
	require.Equal(t, int64(1), merged[0].OrderID)
	require.Equal(t, int64(2), merged[1].OrderID)
	require.Equal(t, int64(1), merged[2].OrderID)
	require.Equal(t, int64(3), merged[3].OrderID)
}

func TestMatchesAttribute_ClientUsesClientFields(t *testing.T) {
	clientPhone := "79990009999"
	clientName := "Тест"
	clientLastName := "Тест"

	o := FormattedOrder{
		Client: ClientDTO{
			Phone:    &clientPhone,
			Name:     &clientName,
			LastName: &clientLastName,
		},
	}

	require.True(t, matchesAttribute(o, "client", "7999"))
	require.True(t, matchesAttribute(o, "client", "тест"))
	require.False(t, matchesAttribute(o, "client", "другой"))
}

func TestGetAllOrders_SkipsMySQLForPreOrder(t *testing.T) {
	ctx := context.Background()
	mysqlCalled := false
	repo := stubRepository{
		fetchAllOrdersForGetAllFunc: func(ctx context.Context, f GetAllOrdersFilter) ([]FullOrder, error) {
			mysqlCalled = true
			return []FullOrder{{OrderID: 10}}, nil
		},
		getOptionsForOrdersFunc: func(ctx context.Context, orderIDs []int64) (map[int64][]OptionDTO, error) {
			return map[int64][]OptionDTO{}, nil
		},
	}
	svc := &service{
		allOrdersReader:    repo,
		optionsReader:      repo,
		statusChangeReader: repo,
		activeOrdersReader: stubActiveOrdersReader{getFunc: func(ctx context.Context, tenantID int64) ([]FormattedOrder, error) {
			return []FormattedOrder{{
				OrderID:     11,
				TenantID:    tenantID,
				CityID:      26068,
				StatusID:    6,
				OrderTime:   1777409100,
				OrderNumber: 1258182,
				Status: StatusDTO{
					StatusID: 6,
					Name:     "New pre-order",
				},
				Client: ClientDTO{
					ClientID: 1,
				},
			}}, nil
		}},
		assembler: newTestOrderViewAssembler(nil, nil, nil),
	}

	date := "2026-04-28"
	result, err := svc.GetAllOrders(ctx, GetAllOrdersFilter{
		BaseFilter: BaseFilter{
			TenantID: 68,
			CityIDs:  []int64{26068},
			Language: "ru",
			Date:     &date,
		},
		Page:         0,
		PageSize:     50,
		SearchStatus: "pre_order",
	})

	require.NoError(t, err)
	require.False(t, mysqlCalled)
	require.Equal(t, int64(1), result.OrderTotalCount)
	require.Len(t, result.Orders, 1)
	require.Equal(t, int64(11), result.Orders[0].ID)
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
