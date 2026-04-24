package orderview

import (
	"context"
	"orders-service/order"
	"time"
)

type Assembler struct {
	waitingTimeProvider order.WaitingTimeProvider
	statusTranslator    order.StatusTranslator
	showOrderCode       order.ShowOrderCodeProvider
}

func NewAssembler(
	waitingTimeProvider order.WaitingTimeProvider,
	statusTranslator order.StatusTranslator,
	showOrderCode order.ShowOrderCodeProvider,
) *Assembler {
	return &Assembler{
		waitingTimeProvider: waitingTimeProvider,
		statusTranslator:    statusTranslator,
		showOrderCode:       showOrderCode,
	}
}

func (a *Assembler) BuildOrderView(
	ctx context.Context,
	o order.FormattedOrder,
	f order.WarningFilter,
	statusChangeTimes map[order.StatusKey]int64,
) (order.OrderView, error) {
	status, err := a.buildOrderStatusView(ctx, f.BaseFilter.Language, o)
	if err != nil {
		return order.OrderView{}, err
	}

	waitTime, err := a.getWorkerWaitingTime(ctx, o.TenantID, o.OrderID)
	if err != nil {
		return order.OrderView{}, err
	}

	orderNumber, err := a.resolveOrderNumber(ctx, o)
	if err != nil {
		return order.OrderView{}, err
	}

	return order.OrderView{
		ID:             o.OrderID,
		OrderNumber:    orderNumber,
		OrderIDForSort: o.OrderNumber,
		Status:         status,
		DateForSort:    formatOrderTimeForSort(o.OrderTime),
		Date:           formatOrderTime(o.OrderTime),
		Address:        o.Address,
		CityID:         o.CityID,
		Phone:          o.Phone,
		Device:         o.Device,
		DeviceName:     order.GetDeviceName(o.Device),
		Client: order.ClientView{
			ClientID: o.Client.ClientID,
			Phone:    o.Client.Phone,
			Name:     o.Client.Name,
			LastName: o.Client.LastName,
		},
		Dispatcher:   order.BuildDispatcher(o),
		Worker:       order.BuildWorker(o),
		Car:          order.BuildCar(o),
		Tariff:       order.BuildTariff(o),
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

func (a *Assembler) buildOrderStatusView(
	ctx context.Context,
	language string,
	o order.FormattedOrder,
) (order.OrderStatusView, error) {
	translatedStatusName, err := a.translateStatus(ctx, language, o.Status.Name)
	if err != nil {
		return order.OrderStatusView{}, err
	}

	return order.OrderStatusView{
		StatusID: o.Status.StatusID,
		Name:     translatedStatusName,
		Category: order.GetCategory(o.StatusID),
		Color:    order.GetColor(o.StatusID),
	}, nil
}

func (a *Assembler) resolveOrderNumber(
	ctx context.Context,
	o order.FormattedOrder,
) (any, error) {
	showOrderCode, err := a.shouldShowOrderCode(ctx, o.TenantID, o.CityID, o.PositionID)
	if err != nil {
		return nil, err
	}

	return order.ShowCodeOrID(showOrderCode, o.OrderCode, o.OrderNumber), nil
}

func (a *Assembler) getWorkerWaitingTime(
	ctx context.Context,
	tenantID, orderID int64,
) (int64, error) {
	if a.waitingTimeProvider == nil {
		return 0, nil
	}

	return a.waitingTimeProvider.GetWorkerWaitingTime(ctx, tenantID, orderID)
}

func (a *Assembler) translateStatus(
	ctx context.Context,
	language string,
	name string,
) (string, error) {
	if a.statusTranslator == nil || name == "" {
		return name, nil
	}

	translated, err := a.statusTranslator.TranslateStatus(ctx, language, name)
	if err != nil {
		return "", err
	}
	if translated == "" {
		return name, nil
	}

	return translated, nil
}

func (a *Assembler) shouldShowOrderCode(
	ctx context.Context,
	tenantID, cityID, positionID int64,
) (bool, error) {
	if a.showOrderCode == nil {
		return false, nil
	}

	return a.showOrderCode.ShouldShowOrderCode(ctx, tenantID, cityID, positionID)
}

func resolveSummaryCost(o order.FormattedOrder, group string) any {
	summaryCost := any(o.PredvPrice)
	if o.PredvPriceNoDiscount > 0 {
		summaryCost = o.PredvPriceNoDiscount
	}

	if group == "completed" {
		switch {
		case o.SummaryCostNoDiscount != nil:
			summaryCost = *o.SummaryCostNoDiscount
		case o.SummaryCost != nil:
			summaryCost = *o.SummaryCost
		}
	}

	return summaryCost
}

func formatOrderTimeForSort(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("2006-01-02 15:04:05")
}

func formatOrderTime(timestamp int64) string {
	return time.Unix(timestamp, 0).UTC().Format("02.01.06 15:04")
}

func getTimeOrderStatusChanged(
	orderID int64,
	statusID int64,
	statusTime int64,
	statusChangeTimes map[order.StatusKey]int64,
) int64 {
	key := order.StatusKey{
		OrderID:  orderID,
		StatusID: statusID,
	}

	if t, ok := statusChangeTimes[key]; ok {
		return t
	}

	return statusTime
}
