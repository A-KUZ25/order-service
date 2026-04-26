package redisactive

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"orders-service/internal/app/order"
	legacyaddress "orders-service/internal/legacy/address"
	"orders-service/internal/legacy/phpdata"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type ActiveOrdersRepository struct {
	client *redis.Client
	parser *legacyaddress.Parser
}

func NewActiveOrdersRepository(client *redis.Client) *ActiveOrdersRepository {
	return &ActiveOrdersRepository{
		client: client,
		parser: legacyaddress.NewParser(),
	}
}

func (r *ActiveOrdersRepository) GetWorkerWaitingTime(
	ctx context.Context,
	tenantID, orderID int64,
) (int64, error) {
	raw, err := r.client.HGet(
		ctx,
		strconv.FormatInt(tenantID, 10),
		strconv.FormatInt(orderID, 10),
	).Bytes()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	payload, err := maybeGunzip(raw)
	if err != nil {
		return 0, err
	}

	value, err := phpdata.Unmarshal(payload)
	if err != nil {
		return 0, err
	}

	orderData, ok := value.(map[string]any)
	if !ok {
		return 0, nil
	}

	waitTime, ok := phpdata.CoerceInt64(orderData["wait_time"])
	if !ok {
		return 0, nil
	}

	return waitTime * 60, nil
}

func (r *ActiveOrdersRepository) GetFormattedActiveOrders(
	ctx context.Context,
	tenantID int64,
) ([]order.FormattedOrder, error) {
	values, err := r.client.HVals(ctx, strconv.FormatInt(tenantID, 10)).Result()
	if err != nil {
		return nil, err
	}

	result := make([]order.FormattedOrder, 0, len(values))
	for _, raw := range values {
		payload, err := maybeGunzip([]byte(raw))
		if err != nil {
			return nil, err
		}

		value, err := phpdata.Unmarshal(payload)
		if err != nil {
			return nil, err
		}

		orderData, ok := value.(map[string]any)
		if !ok {
			continue
		}

		formatted, ok := r.mapActiveOrder(orderData)
		if !ok {
			continue
		}
		result = append(result, formatted)
	}

	return result, nil
}

func (r *ActiveOrdersRepository) mapActiveOrder(value map[string]any) (order.FormattedOrder, bool) {
	orderID, ok := phpdata.CoerceInt64(value["order_id"])
	if !ok {
		return order.FormattedOrder{}, false
	}

	orderTime, _ := phpdata.CoerceInt64(value["order_time"])
	statusID, _ := phpdata.CoerceInt64(value["status_id"])
	tenantID, _ := phpdata.CoerceInt64(value["tenant_id"])
	cityID, _ := phpdata.CoerceInt64(value["city_id"])
	tariffID, _ := phpdata.CoerceInt64(value["tariff_id"])
	positionID, _ := phpdata.CoerceInt64(value["position_id"])
	clientID, _ := phpdata.CoerceInt64(value["client_id"])
	statusTime, _ := phpdata.CoerceInt64(value["status_time"])
	createTime, _ := phpdata.CoerceInt64(value["create_time"])
	orderNumber, _ := phpdata.CoerceInt64(value["order_number"])
	timeOffset, _ := phpdata.CoerceInt64(value["time_offset"])
	unitQuantity := coerceOptionalFloat64(value["unit_quantity"])
	timeToClient := coerceOptionalInt64(value["time_to_client"])

	addresses := r.parseAddressValue(value["address"])

	statusStatusID, _ := phpdata.CoerceInt64(value["status_status_id"])
	if statusStatusID == 0 {
		statusStatusID = statusID
	}
	statusName := phpdata.CoerceString(value["status_name"])

	formatted := order.FormattedOrder{
		OrderID:               orderID,
		TenantID:              tenantID,
		CityID:                cityID,
		TariffID:              tariffID,
		StatusID:              statusID,
		Address:               addresses,
		Comment:               coerceOptionalString(value["comment"]),
		PredvPrice:            coerceFloat64(value["predv_price"]),
		PredvPriceNoDiscount:  coerceFloat64(value["predv_price_no_discount"]),
		Device:                phpdata.CoerceString(value["device"]),
		OrderNumber:           orderNumber,
		CreateTime:            createTime,
		StatusTime:            statusTime,
		TimeToClient:          timeToClient,
		OrderTime:             orderTime,
		Phone:                 phpdata.CoerceString(value["phone"]),
		ClientID:              clientID,
		TimeOffset:            timeOffset,
		PositionID:            positionID,
		UnitQuantity:          unitQuantity,
		ShopID:                coerceOptionalInt64(value["shop_id"]),
		OrderCode:             phpdata.CoerceString(value["order_code"]),
		SummaryCost:           coerceOptionalString(value["summary_cost"]),
		SummaryCostNoDiscount: coerceOptionalString(value["summary_cost_no_discount"]),
		StatusStatusID:        statusStatusID,
		StatusName:            statusName,
		Callsign:              coerceOptionalInt64(value["callsign"]),
		WName:                 coerceOptionalString(value["wName"]),
		WLastName:             coerceOptionalString(value["wLastName"]),
		WSecondName:           coerceOptionalString(value["wSecondName"]),
		WPhone:                coerceOptionalString(value["wPhone"]),
		CPhone:                coerceOptionalString(value["cPhone"]),
		CName:                 coerceOptionalString(value["cName"]),
		CLastName:             coerceOptionalString(value["cLastName"]),
		CSecondName:           coerceOptionalString(value["cSecondName"]),
		CarName:               coerceOptionalString(value["car_name"]),
		CarColor:              coerceOptionalString(value["car_color"]),
		CarGosNumber:          coerceOptionalString(value["car_gos_number"]),
		TariffType:            phpdata.CoerceString(value["tariff_type"]),
		TName:                 phpdata.CoerceString(value["tName"]),
		QuantitativeTitle:     phpdata.CoerceString(value["quantitative_title"]),
		PriceForUnit:          coerceFloat64(value["price_for_unit"]),
		UnitName:              phpdata.CoerceString(value["unit_name"]),
		UserID:                coerceInt64(value["user_id"]),
		UName:                 phpdata.CoerceString(value["uName"]),
		ULastName:             phpdata.CoerceString(value["uLastName"]),
		USecondName:           coerceOptionalString(value["uSecondName"]),
		CurrencyName:          phpdata.CoerceString(value["currency_name"]),
		CurrencyCode:          phpdata.CoerceString(value["currency_code"]),
		Symbol:                phpdata.CoerceString(value["symbol"]),
		Status: order.StatusDTO{
			StatusID: statusStatusID,
			Name:     statusName,
		},
		Client: order.ClientDTO{
			ClientID:   clientID,
			Phone:      coerceOptionalString(value["cPhone"]),
			Name:       coerceOptionalString(value["cName"]),
			LastName:   coerceOptionalString(value["cLastName"]),
			SecondName: coerceOptionalString(value["cSecondName"]),
		},
		UserCreated: order.UserDTO{
			UserID:     coerceInt64(value["user_id"]),
			Name:       phpdata.CoerceString(value["uName"]),
			LastName:   phpdata.CoerceString(value["uLastName"]),
			SecondName: coerceOptionalString(value["uSecondName"]),
		},
		Worker: order.WorkerDTO{
			WorkerID:   coerceInt64(value["worker_id"]),
			Callsign:   coerceOptionalInt64(value["callsign"]),
			Name:       coerceOptionalString(value["wName"]),
			LastName:   coerceOptionalString(value["wLastName"]),
			SecondName: coerceOptionalString(value["wSecondName"]),
			Phone:      coerceOptionalString(value["wPhone"]),
		},
		Car: order.CarDTO{
			CarID:     coerceInt64(value["car_id"]),
			Name:      coerceOptionalString(value["car_name"]),
			Color:     coerceOptionalInt64(value["car_color"]),
			GosNumber: coerceOptionalString(value["car_gos_number"]),
		},
		Tariff: order.TariffDTO{
			TariffID:          tariffID,
			TariffType:        phpdata.CoerceString(value["tariff_type"]),
			Name:              phpdata.CoerceString(value["tName"]),
			QuantitativeTitle: phpdata.CoerceString(value["quantitative_title"]),
			PriceForUnit:      coerceFloat64(value["price_for_unit"]),
			UnitName:          phpdata.CoerceString(value["unit_name"]),
		},
		Currency: order.CurrencyDTO{
			Name:   phpdata.CoerceString(value["currency_name"]),
			Code:   phpdata.CoerceString(value["currency_code"]),
			Symbol: phpdata.CoerceString(value["symbol"]),
		},
	}

	return formatted, true
}

func (r *ActiveOrdersRepository) parseAddressValue(value any) []order.AddressView {
	switch raw := value.(type) {
	case string:
		addresses, err := r.parser.ParseAddress(raw)
		if err != nil {
			return nil
		}
		return addresses
	case map[string]any:
		var builder strings.Builder
		_ = builder
		return mapAddressMap(raw)
	default:
		return nil
	}
}

func mapAddressMap(value map[string]any) []order.AddressView {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}

	result := make([]order.AddressView, 0, len(keys))
	for _, key := range keys {
		item, ok := value[key].(map[string]any)
		if !ok {
			continue
		}
		address := order.AddressView{
			ID:      coerceOptionalString(item["city_id"]),
			City:    coerceOptionalString(item["city"]),
			Street:  coerceOptionalString(item["street"]),
			Label:   coerceOptionalString(item["label"]),
			House:   coerceOptionalString(item["house"]),
			Apt:     coerceOptionalString(item["apt"]),
			Parking: coerceOptionalString(item["parking"]),
			Type:    "house",
		}
		if phpdata.CoerceString(item["place_id"]) != "" && phpdata.CoerceString(item["place_id"]) != "0" {
			address.Type = "place"
		}
		result = append(result, address)
	}
	return result
}

func coerceOptionalString(value any) *string {
	s := phpdata.CoerceString(value)
	if s == "" {
		return nil
	}
	return &s
}

func coerceInt64(value any) int64 {
	v, _ := phpdata.CoerceInt64(value)
	return v
}

func coerceOptionalInt64(value any) *int64 {
	v, ok := phpdata.CoerceInt64(value)
	if !ok {
		return nil
	}
	return &v
}

func coerceFloat64(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case string:
		parsed, _ := strconv.ParseFloat(v, 64)
		return parsed
	default:
		parsed, _ := strconv.ParseFloat(phpdata.CoerceString(value), 64)
		return parsed
	}
}

func coerceOptionalFloat64(value any) *float64 {
	switch value.(type) {
	case nil:
		return nil
	}
	v := coerceFloat64(value)
	return &v
}

func maybeGunzip(raw []byte) ([]byte, error) {
	if len(raw) < 2 || raw[0] != 0x1f || raw[1] != 0x8b {
		return raw, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("gunzip active order payload: %w", err)
	}
	defer reader.Close()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read active order payload: %w", err)
	}

	return payload, nil
}
