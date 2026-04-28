package order

import (
	"context"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *service) GetAllOrders(ctx context.Context, f GetAllOrdersFilter) (GetAllOrdersResult, error) {
	mysqlOrders := []FullOrder{}
	var err error
	if shouldFetchMySQLForGetAll(f.SearchStatus) {
		log.Printf("getAll: mysql fetch start tenant=%d searchStatus=%q page=%d pageSize=%d", f.TenantID, f.SearchStatus, f.Page, f.PageSize)
		mysqlOrders, err = s.allOrdersReader.FetchAllOrdersForGetAll(ctx, f)
		if err != nil {
			return GetAllOrdersResult{}, err
		}
		log.Printf("getAll: mysql fetch done count=%d", len(mysqlOrders))
	} else {
		log.Printf("getAll: mysql fetch skipped searchStatus=%q", f.SearchStatus)
	}

	addressMap := make(map[int64][]AddressView, len(mysqlOrders))
	if s.addressResolver != nil {
		log.Printf("getAll: mysql address resolve start count=%d", len(mysqlOrders))
		addressMap, err = s.addressResolver.ResolveAddresses(mysqlOrders)
		if err != nil {
			return GetAllOrdersResult{}, err
		}
		log.Printf("getAll: mysql address resolve done count=%d", len(addressMap))
	}

	mysqlFormatted := s.MapOrders(mysqlOrders, map[int64][]OptionDTO{}, addressMap)
	log.Printf("getAll: mysql formatted done count=%d", len(mysqlFormatted))

	redisFormatted := []FormattedOrder{}
	if s.activeOrdersReader != nil && shouldFetchRedisForGetAll(f.SearchStatus) {
		log.Printf("getAll: redis fetch start tenant=%d", f.TenantID)
		redisFormatted, err = s.activeOrdersReader.GetFormattedActiveOrders(ctx, f.TenantID)
		if err != nil {
			return GetAllOrdersResult{}, err
		}
		log.Printf("getAll: redis fetch done count=%d", len(redisFormatted))
	} else {
		log.Printf("getAll: redis fetch skipped searchStatus=%q", f.SearchStatus)
	}

	allOrders := mergeGetAllOrders(mysqlFormatted, redisFormatted, f)
	log.Printf("getAll: merged filtered count=%d", len(allOrders))

	sortFormattedOrders(allOrders, f.SortField, f.SortOrder)
	log.Printf("getAll: sort done field=%q dir=%q", f.SortField, f.SortOrder)

	totalCount := int64(len(allOrders))
	pagedOrders := paginateFormattedOrders(allOrders, f.Page, f.PageSize)
	log.Printf("getAll: pagination done total=%d pageCount=%d", totalCount, len(pagedOrders))

	orderIDs := make([]int64, 0, len(pagedOrders))
	for _, value := range pagedOrders {
		orderIDs = append(orderIDs, value.OrderID)
	}

	log.Printf("getAll: options fetch start count=%d", len(orderIDs))
	optionsMap, err := s.optionsReader.GetOptionsForOrders(ctx, orderIDs)
	if err != nil {
		return GetAllOrdersResult{}, err
	}
	log.Printf("getAll: options fetch done count=%d", len(optionsMap))

	for i := range pagedOrders {
		pagedOrders[i].Options = optionsMap[pagedOrders[i].OrderID]
	}

	prepared, err := s.PrepareOrdersData(ctx, pagedOrders, WarningFilter{
		BaseFilter: BaseFilter{
			Language: f.Language,
			Group:    normalizeGetAllGroup(f.SearchStatus),
		},
	})
	if err != nil {
		return GetAllOrdersResult{}, err
	}
	log.Printf("getAll: prepare orders done count=%d", len(prepared))

	return GetAllOrdersResult{
		OrderTotalCount: totalCount,
		CountPerPage:    f.PageSize,
		Orders:          prepared,
	}, nil
}

func paginateFormattedOrders(orders []FormattedOrder, page, pageSize int) []FormattedOrder {
	if page < 0 {
		page = 0
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	offset := page * pageSize
	if offset >= len(orders) {
		return []FormattedOrder{}
	}

	end := offset + pageSize
	if end > len(orders) {
		end = len(orders)
	}

	return orders[offset:end]
}

func sortFormattedOrders(orders []FormattedOrder, field, direction string) {
	desc := !strings.EqualFold(direction, "asc")
	switch field {
	case "o.order_time", "order_time":
		sort.SliceStable(orders, func(i, j int) bool {
			if desc {
				return orders[i].OrderTime > orders[j].OrderTime
			}
			return orders[i].OrderTime < orders[j].OrderTime
		})
	default:
		sort.SliceStable(orders, func(i, j int) bool {
			if desc {
				return orders[i].OrderID > orders[j].OrderID
			}
			return orders[i].OrderID < orders[j].OrderID
		})
	}
}

func matchesGetAllFilter(o FormattedOrder, f GetAllOrdersFilter) bool {
	if !matchesSearchStatus(o.StatusID, f.SearchStatus) {
		return false
	}
	if len(f.CityIDs) > 0 && !containsInt64(f.CityIDs, o.CityID) {
		return false
	}
	if len(f.Tariffs) > 0 && !containsInt64(f.Tariffs, o.TariffID) {
		return false
	}
	if len(f.ShopIDs) > 0 && (o.ShopID == nil || !containsInt64(f.ShopIDs, *o.ShopID)) {
		return false
	}
	if !matchesDate(o.OrderTime, f.Date) {
		return false
	}
	if !matchesSearchAttributes(o, f.Attributes) {
		return false
	}
	if !matchesSearchString(o, "client", f.SearchString["client"]) {
		return false
	}
	if !matchesSearchString(o, "worker", f.SearchString["worker"]) {
		return false
	}

	return true
}

func matchesSearchStatus(statusID int64, searchStatus string) bool {
	switch searchStatus {
	case "", "all":
		return true
	case "new":
		return GetCategory(statusID) == "new"
	case "works":
		return GetCategory(statusID) == "works"
	case "active":
		category := GetCategory(statusID)
		return category == "new" || category == "works"
	case "completed":
		return GetCategory(statusID) == "completed"
	case "rejected":
		return GetCategory(statusID) == "rejected"
	case "warning":
		return GetCategory(statusID) == "warning"
	case "pre_order":
		return GetCategory(statusID) == "pre_order"
	default:
		return GetCategory(statusID) == searchStatus
	}
}

func shouldFetchMySQLForGetAll(searchStatus string) bool {
	return normalizeGetAllSearchStatus(searchStatus) != "pre_order"
}

func shouldFetchRedisForGetAll(searchStatus string) bool {
	switch normalizeGetAllSearchStatus(searchStatus) {
	case "completed", "rejected":
		return false
	default:
		return true
	}
}

func shouldIncludeRedisOrderForGetAll(o FormattedOrder, searchStatus string) bool {
	switch normalizeGetAllSearchStatus(searchStatus) {
	case "all":
		return GetCategory(o.StatusID) != "pre_order"
	default:
		return matchesSearchStatus(o.StatusID, searchStatus)
	}
}

func mergeGetAllOrders(mysqlFormatted, redisFormatted []FormattedOrder, f GetAllOrdersFilter) []FormattedOrder {
	allOrders := make([]FormattedOrder, 0, len(mysqlFormatted)+len(redisFormatted))
	seen := make(map[int64]struct{}, len(mysqlFormatted)+len(redisFormatted))

	appendUnique := func(value FormattedOrder) {
		if _, ok := seen[value.OrderID]; ok {
			return
		}
		seen[value.OrderID] = struct{}{}
		allOrders = append(allOrders, value)
	}

	for _, value := range mysqlFormatted {
		if matchesGetAllFilter(value, f) {
			appendUnique(value)
		}
	}
	for _, value := range redisFormatted {
		if shouldIncludeRedisOrderForGetAll(value, f.SearchStatus) && matchesGetAllFilter(value, f) {
			appendUnique(value)
		}
	}

	return allOrders
}

func matchesDate(orderTime int64, date *string) bool {
	if date == nil || *date == "" {
		return true
	}

	for _, layout := range []string{"2006-01-02", "02.01.2006", "02.01.2006 15:04:05"} {
		parsed, err := time.Parse(layout, *date)
		if err == nil {
			return time.Unix(orderTime, 0).UTC().Format("2006-01-02") == parsed.Format("2006-01-02")
		}
	}

	return time.Unix(orderTime, 0).UTC().Format("2006-01-02") == *date
}

func matchesSearchAttributes(o FormattedOrder, attributes []SearchAttribute) bool {
	for _, attribute := range attributes {
		search := strings.TrimSpace(attribute.SearchString)
		if search == "" {
			continue
		}

		matched := true
		for _, part := range strings.Fields(search) {
			if !matchesAttribute(o, attribute.Attribute, part) {
				matched = false
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func matchesAttribute(o FormattedOrder, attribute, search string) bool {
	needle := strings.ToLower(search)

	switch attribute {
	case "number":
		return strings.Contains(strings.ToLower(strconv.FormatInt(o.OrderNumber, 10)), needle) ||
			strings.Contains(strings.ToLower(o.OrderCode), needle)
	case "address":
		for _, address := range o.Address {
			if strings.Contains(strings.ToLower(joinAddress(address)), needle) {
				return true
			}
		}
		return false
	case "comment":
		return strings.Contains(strings.ToLower(stringValue(o.Comment)), needle)
	case "client":
		return matchesSearchString(o, "client", search)
	case "worker":
		return matchesSearchString(o, "worker", search)
	default:
		return true
	}
}

func matchesSearchString(o FormattedOrder, field, search string) bool {
	needle := strings.ToLower(strings.TrimSpace(search))
	if needle == "" {
		return true
	}

	switch field {
	case "client":
		phone := normalizePhone(stringValue(o.Client.Phone))
		return strings.Contains(strings.ToLower(stringValue(o.Client.LastName)), needle) ||
			strings.Contains(strings.ToLower(stringValue(o.Client.Name)), needle) ||
			strings.Contains(strings.ToLower(stringValue(o.Client.SecondName)), needle) ||
			strings.Contains(phone, normalizePhone(needle))
	case "worker":
		return strings.Contains(strings.ToLower(stringValue(o.Worker.LastName)), needle) ||
			strings.Contains(strings.ToLower(stringValue(o.Worker.Name)), needle) ||
			strings.Contains(strings.ToLower(stringValue(o.Worker.SecondName)), needle) ||
			strings.Contains(strings.ToLower(derefInt64(o.Worker.Callsign)), needle) ||
			strings.Contains(strings.ToLower(stringValue(o.Car.GosNumber)), needle)
	default:
		return true
	}
}

func joinAddress(address AddressView) string {
	parts := []string{
		stringValue(address.City),
		stringValue(address.Street),
		stringValue(address.Label),
		stringValue(address.House),
		stringValue(address.Apt),
		stringValue(address.Parking),
	}
	return strings.ToLower(strings.Join(parts, " "))
}

func normalizePhone(value string) string {
	value = strings.ReplaceAll(value, "+", "")
	value = strings.ReplaceAll(value, "(", "")
	value = strings.ReplaceAll(value, ")", "")
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefInt64(value *int64) string {
	if value == nil {
		return ""
	}
	return strconv.FormatInt(*value, 10)
}

func normalizeGetAllGroup(searchStatus string) string {
	switch normalizeGetAllSearchStatus(searchStatus) {
	case "completed":
		return "completed"
	default:
		return searchStatus
	}
}

func normalizeGetAllSearchStatus(searchStatus string) string {
	return strings.TrimSpace(strings.ToLower(searchStatus))
}

func containsInt64(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
