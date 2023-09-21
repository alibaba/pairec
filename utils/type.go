package utils

import (
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
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
	case uint32:
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
func ToBool(i interface{}, defaultVal bool) bool {
	switch value := i.(type) {
	case bool:
		return value
	case int:
		return value > 0
	case int32:
		return value > 0
	case int64:
		return value > 0
	case string:
		v := strings.ToLower(value)
		if v == "false" || v == "" || v == "f" || v == "off" {
			return false
		} else if v == "true" || v == "t" || v == "on" {
			return true
		} else {
			return false
		}
	default:
		return defaultVal
	}
}

func Equal(l interface{}, r interface{}) bool {
	switch value := l.(type) {
	case int:
		return value == ToInt(r, math.MinInt32)
	case float32:
		return float64(value) == ToFloat(r, math.MaxFloat64)
	case float64:
		return value == ToFloat(r, math.MaxFloat64)
	case uint:
		return int(value) == ToInt(r, math.MinInt32)
	case int32:
		return int(value) == ToInt(r, math.MinInt32)
	case int64:
		return value == ToInt64(r, math.MinInt64)
	case string:
		return value == ToString(r, "")
	case json.Number:
		return value.String() == ToString(r, "")
	default:
		return false
	}
}

func NotEqual(l interface{}, r interface{}) bool {
	return !Equal(l, r)
}

func Greater(l interface{}, r interface{}) bool {
	switch value := l.(type) {
	case int:
		return value > ToInt(r, math.MinInt32)
	case float32:
		return float64(value) > ToFloat(r, math.MaxFloat64)
	case float64:
		return value > ToFloat(r, math.MaxFloat64)
	case uint:
		return int(value) > ToInt(r, math.MinInt32)
	case int32:
		return int(value) > ToInt(r, math.MinInt32)
	case int64:
		return value > ToInt64(r, math.MinInt64)
	case string:
		return value > ToString(r, "")
	case json.Number:
		return value.String() > ToString(r, "")
	default:
		return false
	}
}

func GreaterEqual(l interface{}, r interface{}) bool {
	switch value := l.(type) {
	case int:
		return value >= ToInt(r, math.MinInt32)
	case float32:
		return float64(value) >= ToFloat(r, math.MaxFloat64)
	case float64:
		return value >= ToFloat(r, math.MaxFloat64)
	case uint:
		return int(value) >= ToInt(r, math.MinInt32)
	case int32:
		return int(value) >= ToInt(r, math.MinInt32)
	case int64:
		return value >= ToInt64(r, math.MinInt64)
	case string:
		return value >= ToString(r, "")
	case json.Number:
		return value.String() >= ToString(r, "")
	default:
		return false
	}
}

func Less(l interface{}, r interface{}) bool {
	return !GreaterEqual(l, r)
}

func LessEqual(l interface{}, r interface{}) bool {
	return !Greater(l, r)
}

func In(l interface{}, r interface{}) bool {
	values := ToString(r, "")
	values = strings.Trim(values, "()")
	elements := strings.Split(values, ",")
	for _, element := range elements {
		if Equal(l, element) {
			return true
		}
		element = strings.Trim(element, " '\"")
		if Equal(l, element) {
			return true
		}
	}
	return false
}
func Contains(leftList []interface{}, rightList []interface{}) bool {
	for _, left := range leftList {
		for _, right := range rightList {
			if Equal(left, right) {
				return true
			}
		}
	}
	return false
}

func StringContains(leftList []string, rightList []string) bool {
	for _, left := range leftList {
		for _, right := range rightList {
			if Equal(left, right) {
				return true
			}
		}
	}
	return false
}

func IntContains(leftList []int, rightList []int) bool {
	for _, left := range leftList {
		for _, right := range rightList {
			if Equal(left, right) {
				return true
			}
		}
	}
	return false
}
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
