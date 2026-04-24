package order

import (
	"context"
	"strconv"
)

func (s *service) MapFullOrderToFormatted(
	o FullOrder,
	options []OptionDTO,
	address []AddressView,
) FormattedOrder {
	predvPrice := 0.0
	if o.SummaryCost.Valid && o.SummaryCost.String != "" {
		predvPrice = parseFloat(o.SummaryCost.String)
	} else if o.PredvPrice.Valid {
		predvPrice = o.PredvPrice.Float64
	}

	return FormattedOrder{
		OrderID:      o.OrderID,
		TenantID:     o.TenantID,
		WorkerID:     nullableInt64(o.WorkerID),
		CarID:        nullableInt64(o.CarID),
		CityID:       o.CityID.Int64,
		TariffID:     o.TariffID,
		UserCreate:   o.UserCreate.Int64,
		StatusID:     o.StatusID,
		UserModified: o.UserModified.Int64,
		CompanyID:    nullableInt64(o.CompanyID),
		ParkingID:    nullableInt64(o.ParkingID),
		Address:      address,
		Comment:      nullableString(o.Comment),

		PredvPrice:           predvPrice,
		PredvPriceNoDiscount: o.PredvPriceNoDiscount.Float64,
		Device:               o.Device.String,
		OrderNumber:          o.OrderNumber,
		Payment:              o.Payment.String,
		ShowPhone:            o.ShowPhone.Int64,
		CreateTime:           o.CreateTime.Int64,
		StatusTime:           o.StatusTime,
		TimeToClient:         nullableInt64(o.TimeToClient),
		ClientDeviceToken:    nullableString(o.ClientDeviceToken),
		AppID:                nullableInt64(o.AppID),
		OrderTime:            o.OrderTime.Int64,
		PredvDistance:        o.PredvDistance.Float64,
		PredvTime:            o.PredvTime.Int64,
		CallWarningID:        nullableInt64(o.CallWarningID),
		Phone:                o.Phone.String,
		ClientID:             o.ClientID,
		BonusPayment:         o.BonusPayment.Int64,
		CurrencyID:           o.CurrencyID,
		TimeOffset:           o.TimeOffset.Int64,
		IsFix:                o.IsFix,
		UpdateTime:           o.UpdateTime.Int64,
		DenyRefuseOrder:      o.DenyRefuseOrder.Int64,
		PositionID:           o.PositionID,
		PromoCodeID:          nullableInt64(o.PromoCodeID),
		TenantCompanyID:      nullableInt64(o.TenantCompanyID),
		Mark:                 o.Mark.Int64,

		ProcessedExchangeProgramID: nullableInt64(o.ProcessedExchangeProgramID),
		ClientPassengerID:          nullableInt64(o.ClientPassengerID),
		ClientPassengerPhone:       nullableString(o.ClientPassengerPhone),
		Active:                     o.Active.Int64,
		IsPreOrder:                 o.IsPreOrder.Int64,
		AppVersion:                 nullableString(o.AppVersion),
		AgentCommission:            o.AgentCommission.Float64,
		IsFixByDispatcher:          o.IsFixByDispatcher.Int64,
		FinishTime:                 nullableInt64(o.FinishTime),
		CommentForDispatcher:       nullableString(o.CommentForDispatcher),
		WorkerManualSurcharge:      o.WorkerManualSurcharge.Float64,
		RealtimePrice:              nullableFloat64(o.RealtimePrice),
		UnitQuantity:               nullableFloat64(o.UnitQuantity),
		ShopID:                     nullableInt64(o.ShopID),
		RequirePrepayment:          o.RequirePrepayment.Int64,
		OrderCode:                  o.OrderCode.String,
		ClientOfferedPrice:         nullableFloat64(o.ClientOfferedPrice),
		IdempotentKey:              o.IdempotentKey.String,
		AdditionalTariffID:         nullableInt64(o.AdditionalTariffID),
		InitialPrice:               nullableFloat64(o.InitialPrice),
		TimeToOrder:                nullableInt64(o.TimeToOrder),
		Sort:                       nullableInt64(o.Sort),
		SummaryCost:                nullableString(o.SummaryCost),
		SummaryCostNoDiscount:      nullableString(o.SummaryCostNoDiscount),
		StatusStatusID:             o.StatusStatusID,
		StatusName:                 o.StatusName,
		Status:                     StatusDTO{StatusID: o.StatusStatusID, Name: o.StatusName},
		Client: ClientDTO{
			ClientID:   o.ClientClientID.Int64,
			Phone:      nullableString(o.ClientPhone),
			Name:       nullableString(o.ClientName),
			LastName:   nullableString(o.ClientLastName),
			SecondName: nullableString(o.ClientSecondName),
		},
		UserCreated: UserDTO{
			UserID:     o.UserUserID.Int64,
			Name:       o.UserName.String,
			LastName:   o.UserLastName.String,
			SecondName: nullableString(o.UserSecondName),
		},
		Worker: WorkerDTO{
			WorkerID:   o.WorkerWorkerID.Int64,
			Callsign:   nullableInt64(o.WorkerCallsign),
			Name:       nullableString(o.WorkerName),
			LastName:   nullableString(o.WorkerLastName),
			SecondName: nullableString(o.WorkerSecondName),
			Phone:      nullableString(o.WorkerPhone),
		},
		Car: CarDTO{
			CarID:     o.CarCarID.Int64,
			Name:      nullableString(o.CarName),
			Color:     nullableInt64(o.CarColor),
			GosNumber: nullableString(o.CarGosNumber),
		},
		Tariff: TariffDTO{
			TariffID:          o.TariffTariffID.Int64,
			TariffType:        o.TariffType.String,
			Name:              o.TariffName.String,
			QuantitativeTitle: o.TariffQuantitativeTitle.String,
			PriceForUnit:      o.TariffPriceForUnit.Float64,
			UnitName:          o.TariffUnitName.String,
		},
		Currency: CurrencyDTO{
			Name:   o.CurrencyName.String,
			Code:   o.CurrencyCode.String,
			Symbol: o.CurrencySymbol.String,
		},
		Options: options,
	}
}

func parseFloat(v string) float64 {
	if v == "" {
		return 0
	}

	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}

	return f
}

func (s *service) MapOrders(
	orders []FullOrder,
	optionsMap map[int64][]OptionDTO,
	addressMap map[int64][]AddressView,
) []FormattedOrder {
	result := make([]FormattedOrder, 0, len(orders))

	for _, o := range orders {
		result = append(result, s.MapFullOrderToFormatted(
			o,
			optionsMap[o.OrderID],
			addressMap[o.OrderID],
		))
	}

	return result
}

func (s *service) GetFormattedOrdersByGroup(
	ctx context.Context,
	f WarningFilter,
	page, pageSize int,
) (int64, []FormattedOrder, error) {
	count, orders, err := s.GetOrdersByGroup(ctx, f, page, pageSize)
	if err != nil {
		return 0, nil, err
	}

	if len(orders) == 0 {
		return count, []FormattedOrder{}, nil
	}

	orderIDs := make([]int64, len(orders))
	for i := range orders {
		orderIDs[i] = orders[i].OrderID
	}

	addressMap := make(map[int64][]AddressView, len(orders))
	if s.addressResolver != nil {
		resolved, err := s.addressResolver.ResolveAddresses(orders)
		if err != nil {
			return 0, nil, err
		}
		addressMap = resolved
	}

	optionsMap, err := s.optionsReader.GetOptionsForOrders(ctx, orderIDs)
	if err != nil {
		return 0, nil, err
	}

	return count, s.MapOrders(orders, optionsMap, addressMap), nil
}
