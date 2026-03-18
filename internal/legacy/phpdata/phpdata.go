package phpdata

import (
	"fmt"
	"strconv"

	"github.com/elliotchance/phpserialize"
)

func Unmarshal(raw []byte) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	switch raw[0] {
	case 'a', 'O':
		var value map[interface{}]interface{}
		if err := phpserialize.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return normalize(value), nil
	case 's':
		var value string
		if err := phpserialize.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return value, nil
	case 'i':
		var value int64
		if err := phpserialize.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return value, nil
	case 'd':
		var value float64
		if err := phpserialize.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return value, nil
	case 'b':
		var value bool
		if err := phpserialize.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return value, nil
	case 'N':
		return nil, nil
	default:
		var value map[interface{}]interface{}
		if err := phpserialize.Unmarshal(raw, &value); err != nil {
			return nil, err
		}
		return normalize(value), nil
	}
}

func normalize(v any) any {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]any, len(val))
		for key, item := range val {
			m[CoerceString(key)] = normalize(item)
		}
		return m
	case []interface{}:
		items := make([]any, 0, len(val))
		for _, item := range val {
			items = append(items, normalize(item))
		}
		return items
	default:
		return val
	}
}

func CoerceString(v any) string {
	switch value := v.(type) {
	case nil:
		return ""
	case string:
		return value
	case []byte:
		return string(value)
	case fmt.Stringer:
		return value.String()
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	default:
		return fmt.Sprint(value)
	}
}

func CoerceInt64(v any) (int64, bool) {
	switch value := v.(type) {
	case int:
		return int64(value), true
	case int8:
		return int64(value), true
	case int16:
		return int64(value), true
	case int32:
		return int64(value), true
	case int64:
		return value, true
	case uint:
		return int64(value), true
	case uint8:
		return int64(value), true
	case uint16:
		return int64(value), true
	case uint32:
		return int64(value), true
	case uint64:
		return int64(value), true
	case float32:
		return int64(value), true
	case float64:
		return int64(value), true
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	case []byte:
		parsed, err := strconv.ParseInt(string(value), 10, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
