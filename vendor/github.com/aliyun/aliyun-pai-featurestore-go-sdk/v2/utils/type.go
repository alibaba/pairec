package utils

import (
	"encoding/json"
	"strconv"
)

func ToInt(i interface{}, defaultVal int) int {
	switch value := i.(type) {
	case int:
		return value
	case float64:
		return int(value)
	case uint:
		return int(value)
	case int32:
		return int(value)
	case int64:
		return int(value)
	case string:
		if val, err := strconv.Atoi(value); err == nil {
			return val
		} else {
			return defaultVal
		}
	case json.Number:
		if val, err := value.Int64(); err == nil {
			return int(val)
		} else {
			return defaultVal
		}
	default:
		return defaultVal
	}
}
func ToFloat(i interface{}, defaultVal float64) float64 {
	switch value := i.(type) {
	case float64:
		return value
	case int:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case uint32:
		return float64(value)
	case uint:
		return float64(value)
	case string:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		} else {
			return defaultVal
		}
	case json.Number:
		if f, err := value.Float64(); err == nil {
			return f
		} else {
			return defaultVal
		}
	default:
		return defaultVal
	}
}
func ToInt64(i interface{}, defaultVal int64) int64 {
	switch value := i.(type) {
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case uint:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case string:
		if val, err := strconv.ParseInt(value, 10, 64); err == nil {
			return val
		} else {
			return defaultVal
		}
	case json.Number:
		if val, err := value.Int64(); err == nil {
			return val
		} else {
			return defaultVal
		}
	default:
		return defaultVal
	}
}

func ToString(i interface{}, defaultVal string) string {
	switch value := i.(type) {
	case int:
		return strconv.Itoa(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 64)
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(value, 10)
	case string:
		return value
	case json.Number:
		return value.String()
	default:
		return defaultVal
	}
}
