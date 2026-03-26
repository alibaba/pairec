package web

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// FeaturesMap is a custom type for Features field that converts numeric arrays to string arrays
// to avoid precision loss during JSON unmarshaling for large integers
type FeaturesMap map[string]interface{}

// UnmarshalJSON implements custom JSON unmarshaling for FeaturesMap
// It uses json.Number to preserve precision for large integers
func (f *FeaturesMap) UnmarshalJSON(data []byte) error {
	// Use json.Decoder with UseNumber() to preserve numeric precision
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var temp map[string]interface{}
	if err := decoder.Decode(&temp); err != nil {
		return err
	}

	*f = make(FeaturesMap, len(temp))
	for k, v := range temp {
		(*f)[k] = convertValueToString(v)
	}

	return nil
}

// convertValueToString converts values to string representation
// Handles arrays, nested structures, and json.Number for precision preservation
func convertValueToString(value interface{}) interface{} {
	if value == nil {
		return value
	}

	switch v := value.(type) {
	case json.Number:
		// Try to restore original numeric type
		return parseNumber(v)

	case []interface{}:
		return convertArray(v)

	case map[string]interface{}:
		return convertMap(v)

	default:
		return value
	}
}

// convertArray handles array conversion with nested array support
func convertArray(arr []interface{}) interface{} {
	if len(arr) == 0 {
		return arr
	}

	// Classify the array by its elements
	const (
		elemTypeScalar = iota
		elemTypeNested
		elemTypeMixed
	)

	elemType := elemTypeScalar
	for _, elem := range arr {
		switch elem.(type) {
		case []interface{}:
			if elemType == elemTypeScalar {
				elemType = elemTypeNested
			}
		case json.Number, string:
			if elemType == elemTypeNested {
				elemType = elemTypeMixed
			}
		default:
			elemType = elemTypeMixed
		}
		if elemType == elemTypeMixed {
			break
		}
	}

	switch elemType {
	case elemTypeNested:
		// Convert nested arrays: []interface{} -> [][]string
		result := make([][]string, 0, len(arr))
		for _, elem := range arr {
			if innerArr, ok := elem.([]interface{}); ok {
				inner := make([]string, 0, len(innerArr))
				for _, innerElem := range innerArr {
					inner = append(inner, formatValue(innerElem))
				}
				result = append(result, inner)
			}
		}
		return result

	case elemTypeScalar:
		// All elements are scalar: convert to []string
		result := make([]string, 0, len(arr))
		for _, elem := range arr {
			result = append(result, formatValue(elem))
		}
		return result

	default:
		// Mixed types or unknown, return as-is
		return arr
	}
}

// convertMap handles map conversion with type-aware result types
func convertMap(m map[string]interface{}) interface{} {
	// Determine the result type based on value types
	hasArray := false
	hasMap := false
	for _, val := range m {
		switch val.(type) {
		case []interface{}:
			hasArray = true
		case map[string]interface{}:
			hasMap = true
		}
	}

	// If has nested map, keep as map[string]interface{}
	if hasMap {
		result := make(map[string]interface{}, len(m))
		for k, val := range m {
			result[k] = convertValueToString(val)
		}
		return result
	}

	// If has array, return map[string][]string
	if hasArray {
		result := make(map[string][]string, len(m))
		for k, val := range m {
			converted := convertValueToString(val)
			if arr, ok := converted.([]string); ok {
				result[k] = arr
			} else if arr, ok := converted.([]interface{}); ok {
				// Convert []interface{} to []string
				strArr := make([]string, 0, len(arr))
				for _, elem := range arr {
					strArr = append(strArr, formatValue(elem))
				}
				result[k] = strArr
			} else {
				// Single value, wrap in slice
				result[k] = []string{formatValue(val)}
			}
		}
		return result
	}

	// All values are scalar types, return map[string]string
	result := make(map[string]string, len(m))
	for k, val := range m {
		result[k] = formatValue(val)
	}
	return result
}

// parseNumber attempts to restore the original numeric type from json.Number
func parseNumber(n json.Number) interface{} {
	// Try int64 first (covers most integer cases)
	if i64, err := n.Int64(); err == nil {
		return int(i64)
	}

	// Try float64 for decimal numbers
	if f64, err := n.Float64(); err == nil {
		return f64
	}

	// Fallback to string if it can't be parsed as a number
	return n.String()
}

// formatValue converts a single value to string
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case json.Number:
		return v.String()
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
