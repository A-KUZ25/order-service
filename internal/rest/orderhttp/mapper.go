package orderhttp

import "orders-service/internal/app/order"

func buildOrdersResponse(
	totalCount int64,
	pageSize int,
	tabs order.GroupOrdersResult,
	orders []order.OrderView,
) ordersResponse {
	return ordersResponse{
		OrderTotalCount: totalCount,
		OrdersForSignal: mapStatusGroupIDs(tabs.OrdersForSignal),
		OrderCounts:     mapStatusGroupCounts(tabs.GroupCounts),
		CountPerPage:    pageSize,
		Orders:          mapOrderViews(orders),
	}
}

func buildAllOrdersResponse(result order.GetAllOrdersResult) allOrdersResponse {
	return allOrdersResponse{
		OrderTotalCount: result.OrderTotalCount,
		CountPerPage:    result.CountPerPage,
		Orders:          mapOrderViews(result.Orders),
	}
}

func mapStatusGroupIDs(values map[order.StatusGroup][]int64) map[string][]int64 {
	result := make(map[string][]int64, len(values))
	for key, ids := range values {
		result[string(key)] = ids
	}
	return result
}

func mapStatusGroupCounts(values map[order.StatusGroup]int) map[string]int {
	result := make(map[string]int, len(values))
	for key, count := range values {
		result[string(key)] = count
	}
	return result
}

func mapOrderViews(values []order.OrderView) []orderViewResponse {
	result := make([]orderViewResponse, 0, len(values))
	for _, value := range values {
		addresses := make([]addressResponse, 0, len(value.Address))
		for _, address := range value.Address {
			addresses = append(addresses, addressResponse{
				ID:      address.ID,
				City:    address.City,
				Street:  address.Street,
				Label:   address.Label,
				House:   address.House,
				Apt:     address.Apt,
				Parking: address.Parking,
				Type:    address.Type,
			})
		}

		options := make([]optionResponse, 0, len(value.Options))
		for _, option := range value.Options {
			options = append(options, optionResponse{
				OptionID: option.OptionID,
				Name:     option.Name,
				Quantity: option.Quantity,
			})
		}

		var worker *workerResponse
		if value.Worker != nil {
			worker = &workerResponse{
				WorkerID: value.Worker.WorkerID,
				Callsign: value.Worker.Callsign,
				Name:     value.Worker.Name,
				Phone:    value.Worker.Phone,
			}
		}

		var car *carResponse
		if value.Car != nil {
			car = &carResponse{
				CarID:  value.Car.CarID,
				Name:   value.Car.Name,
				Color:  value.Car.Color,
				Number: value.Car.Number,
			}
		}

		result = append(result, orderViewResponse{
			ID:             value.ID,
			OrderNumber:    value.OrderNumber,
			OrderIDForSort: value.OrderIDForSort,
			Status: statusResponse{
				StatusID: value.Status.StatusID,
				Name:     value.Status.Name,
				Category: value.Status.Category,
				Color:    value.Status.Color,
			},
			DateForSort: value.DateForSort,
			Date:        value.Date,
			Address:     addresses,
			CityID:      value.CityID,
			Phone:       value.Phone,
			Device:      value.Device,
			DeviceName:  value.DeviceName,
			Client: clientResponse{
				ClientID: value.Client.ClientID,
				Phone:    value.Client.Phone,
				Name:     value.Client.Name,
				LastName: value.Client.LastName,
			},
			Dispatcher: value.Dispatcher,
			Worker:     worker,
			Car:        car,
			Tariff: tariffResponse{
				TariffID:          value.Tariff.TariffID,
				Name:              value.Tariff.Name,
				QuantitativeTitle: value.Tariff.QuantitativeTitle,
				PriceForUnit:      value.Tariff.PriceForUnit,
				UnitName:          value.Tariff.UnitName,
			},
			Options:      options,
			Comment:      value.Comment,
			SummaryCost:  value.SummaryCost,
			StatusTime:   value.StatusTime,
			TimeToClient: value.TimeToClient,
			WaitTime:     value.WaitTime,
			CreateTime:   value.CreateTime,
			OrderTime:    value.OrderTime,
			PositionID:   value.PositionID,
			UnitQuantity: value.UnitQuantity,
		})
	}

	return result
}
