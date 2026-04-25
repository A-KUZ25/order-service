package mysql

import (
	"context"
	"log"
	"orders-service/internal/app/order"
	"strings"
	"time"
)

func (r *OrdersRepository) FetchOrdersWithWarning(
	ctx context.Context,
	f order.BaseFilter,
	warningIDs []int64,
	page, pageSize int,
) ([]order.FullOrder, error) {
	var sb strings.Builder
	var args []any

	sb.WriteString(`
SELECT
    o.order_id,
    o.tenant_id,
    o.worker_id,
    o.car_id,
    o.city_id,
    o.tariff_id,
    o.user_create,
    o.status_id,
    o.user_modifed,
    o.company_id,
    o.parking_id,
    o.address,
    o.comment,
    o.predv_price,
    o.predv_price_no_discount,
    o.device,
    o.order_number,
    o.payment,
    o.show_phone,
    o.create_time,
    o.status_time,
    o.time_to_client,
    o.client_device_token,
    o.app_id,
    o.order_time,
    o.predv_distance,
    o.predv_time,
    o.call_warning_id,
    o.phone,
    o.client_id,
    o.bonus_payment,
    o.currency_id,
    o.time_offset,
    o.is_fix,
    o.update_time,
    o.deny_refuse_order,
    o.position_id,
    o.promo_code_id,
    o.tenant_company_id,
    o.mark,
    o.processed_exchange_program_id,
    o.client_passenger_id,
    o.client_passenger_phone,
    o.active,
    o.is_pre_order,
    o.app_version,
    o.agent_commission,
    o.is_fix_by_dispatcher,
    o.finish_time,
    o.comment_for_dispatcher,
    o.worker_manual_surcharge,
    o.realtime_price,
    o.unit_quantity,
    o.shop_id,
    o.require_prepayment,
    o.order_code,
    o.client_offered_price,
    o.idempotent_key,
    o.additional_tariff_id,
    o.initial_price,
    o.time_to_order,
    o.sort,
    d.summary_cost,
    d.summary_cost_no_discount,
    s.status_id AS status_status_id,
    s.name AS status_name,
    w.worker_id,
    w.callsign,
    w.name,
    w.last_name,
    w.second_name,
    w.phone,
    cl.client_id,
    cl.phone,
    cl.name,
    cl.last_name,
    cl.second_name,
    car.car_id,
    car.name,
    car.color,
    car.gos_number,
    t.tariff_id,
    t.tariff_type,
    t.name,
    t.quantitative_title,
    t.price_for_unit,
    t.unit_name,
    u.user_id,
    u.name,
    u.last_name,
    u.second_name,
    curr.name,
    curr.code,
    curr.symbol
FROM tbl_order o
LEFT JOIN tbl_client cl ON o.client_id = cl.client_id
LEFT JOIN tbl_order_status s ON o.status_id = s.status_id
LEFT JOIN tbl_worker w ON o.worker_id = w.worker_id
LEFT JOIN tbl_car car ON o.car_id = car.car_id
LEFT JOIN tbl_taxi_tariff t ON o.tariff_id = t.tariff_id
LEFT JOIN tbl_order_detail_cost d ON o.order_id = d.order_id
LEFT JOIN tbl_user u ON o.user_create = u.user_id
LEFT JOIN tbl_currency curr ON o.currency_id = curr.currency_id
WHERE ( 1=1
`)

	r.buildBaseQuery(&sb, &args, f)
	sb.WriteString(") ")
	if len(warningIDs) > 0 {
		sb.WriteString(" OR (o.order_id IN (")
		for i, id := range warningIDs {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			args = append(args, id)
		}
		sb.WriteString("))\n")
	}

	r.appendOrderBy(&sb, f)

	sb.WriteString(" LIMIT ? OFFSET ?")
	args = append(args, pageSize, page*pageSize)

	start := time.Now()
	rows, err := r.db.QueryContext(ctx, sb.String(), args...)
	log.Println("BASE REQUEST TIME:", time.Since(start))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []order.FullOrder
	for rows.Next() {
		var o order.FullOrder

		err = rows.Scan(
			&o.OrderID,
			&o.TenantID,
			&o.WorkerID,
			&o.CarID,
			&o.CityID,
			&o.TariffID,
			&o.UserCreate,
			&o.StatusID,
			&o.UserModified,
			&o.CompanyID,
			&o.ParkingID,
			&o.Address,
			&o.Comment,
			&o.PredvPrice,
			&o.PredvPriceNoDiscount,
			&o.Device,
			&o.OrderNumber,
			&o.Payment,
			&o.ShowPhone,
			&o.CreateTime,
			&o.StatusTime,
			&o.TimeToClient,
			&o.ClientDeviceToken,
			&o.AppID,
			&o.OrderTime,
			&o.PredvDistance,
			&o.PredvTime,
			&o.CallWarningID,
			&o.Phone,
			&o.ClientID,
			&o.BonusPayment,
			&o.CurrencyID,
			&o.TimeOffset,
			&o.IsFix,
			&o.UpdateTime,
			&o.DenyRefuseOrder,
			&o.PositionID,
			&o.PromoCodeID,
			&o.TenantCompanyID,
			&o.Mark,
			&o.ProcessedExchangeProgramID,
			&o.ClientPassengerID,
			&o.ClientPassengerPhone,
			&o.Active,
			&o.IsPreOrder,
			&o.AppVersion,
			&o.AgentCommission,
			&o.IsFixByDispatcher,
			&o.FinishTime,
			&o.CommentForDispatcher,
			&o.WorkerManualSurcharge,
			&o.RealtimePrice,
			&o.UnitQuantity,
			&o.ShopID,
			&o.RequirePrepayment,
			&o.OrderCode,
			&o.ClientOfferedPrice,
			&o.IdempotentKey,
			&o.AdditionalTariffID,
			&o.InitialPrice,
			&o.TimeToOrder,
			&o.Sort,
			&o.SummaryCost,
			&o.SummaryCostNoDiscount,
			&o.StatusStatusID,
			&o.StatusName,
			&o.WorkerWorkerID,
			&o.WorkerCallsign,
			&o.WorkerName,
			&o.WorkerLastName,
			&o.WorkerSecondName,
			&o.WorkerPhone,
			&o.ClientClientID,
			&o.ClientPhone,
			&o.ClientName,
			&o.ClientLastName,
			&o.ClientSecondName,
			&o.CarCarID,
			&o.CarName,
			&o.CarColor,
			&o.CarGosNumber,
			&o.TariffTariffID,
			&o.TariffType,
			&o.TariffName,
			&o.TariffQuantitativeTitle,
			&o.TariffPriceForUnit,
			&o.TariffUnitName,
			&o.UserUserID,
			&o.UserName,
			&o.UserLastName,
			&o.UserSecondName,
			&o.CurrencyName,
			&o.CurrencyCode,
			&o.CurrencySymbol,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, o)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *OrdersRepository) FetchOrdersByStatusGroup(
	ctx context.Context,
	f order.BaseFilter,
) ([]int64, error) {
	var sb strings.Builder
	args := []any{}

	sb.WriteString(`
SELECT o.order_id
FROM tbl_order o
WHERE (1=1
`)

	r.buildBaseQuery(&sb, &args, f, true)
	sb.WriteString(") ")

	if len(f.Status) > 0 {
		sb.WriteString(" AND o.status_id IN (")
		for i, st := range f.Status {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
			args = append(args, st)
		}
		sb.WriteString(")\n")
	}

	r.appendOrderBy(&sb, f)

	return r.executeQuery(ctx, sb.String(), args)
}
