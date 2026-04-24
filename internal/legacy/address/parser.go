package address

import (
	"orders-service/internal/legacy/phpdata"
	"orders-service/order"
	"sort"
	"strconv"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseAddress(raw string) ([]order.AddressView, error) {
	if raw == "" {
		return nil, nil
	}

	value, err := phpdata.Unmarshal([]byte(raw))
	if err != nil {
		return nil, err
	}

	addressMap, ok := value.(map[string]any)
	if !ok {
		return []order.AddressView{}, nil
	}

	result := make([]order.AddressView, 0, len(addressMap))
	keys := make([]string, 0, len(addressMap))
	for key := range addressMap {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		leftInt, leftErr := strconv.Atoi(keys[i])
		rightInt, rightErr := strconv.Atoi(keys[j])

		switch {
		case leftErr == nil && rightErr == nil:
			return leftInt < rightInt
		case leftErr == nil:
			return true
		case rightErr == nil:
			return false
		default:
			return keys[i] < keys[j]
		}
	})

	for _, key := range keys {
		itemValue := addressMap[key]
		item, ok := itemValue.(map[string]any)
		if !ok {
			continue
		}

		address := order.AddressView{
			ID:      nullablePHPString(item["city_id"]),
			City:    nullablePHPString(item["city"]),
			Street:  nullablePHPString(item["street"]),
			Label:   nullablePHPString(item["label"]),
			House:   nullablePHPString(item["house"]),
			Apt:     nullablePHPString(item["apt"]),
			Parking: nullablePHPString(item["parking"]),
			Type:    "house",
		}

		if !isEmptyPHPValue(item["place_id"]) {
			address.Type = "place"
		}

		result = append(result, address)
	}

	return result, nil
}

func isEmptyPHPValue(v any) bool {
	switch value := v.(type) {
	case nil:
		return true
	case bool:
		return !value
	case string:
		return value == "" || value == "0"
	case []byte:
		return len(value) == 0 || string(value) == "0"
	case int:
		return value == 0
	case int8:
		return value == 0
	case int16:
		return value == 0
	case int32:
		return value == 0
	case int64:
		return value == 0
	case uint:
		return value == 0
	case uint8:
		return value == 0
	case uint16:
		return value == 0
	case uint32:
		return value == 0
	case uint64:
		return value == 0
	case float32:
		return value == 0
	case float64:
		return value == 0
	default:
		return false
	}
}

func nullablePHPString(v any) *string {
	if isEmptyPHPValue(v) {
		return nil
	}

	value := phpdata.CoerceString(v)
	return &value
}
