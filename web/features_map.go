package web

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/alibaba/pairec/v2/utils"
)

// FeaturesMap is a custom type for Features field that converts numeric arrays to string arrays
// to avoid precision loss during JSON unmarshaling for large integers
type FeaturesMap map[string]interface{}

// UnmarshalJSON implements custom JSON unmarshaling for FeaturesMap
// It converts []int, []int64, []float32, []float64 to []string to preserve precision
func (f *FeaturesMap) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary map with interface{} values
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	*f = make(FeaturesMap, len(temp))
	for k, v := range temp {
		(*f)[k] = convertArrayToStringArray(v)
	}

	return nil
}

// convertArrayToStringArray converts numeric arrays to string arrays
// Preserves precision for large integers that would be lost in float64 conversion
// Also handles nested structures in maps
func convertArrayToStringArray(value interface{}) interface{} {
	if value == nil {
		return value
	}

	// Array element classification
	const (
		elemTypeScalar = iota // numeric, string, or other scalar
		elemTypeNested        // nested array
		elemTypeMixed         // mixed types
	)

	switch v := value.(type) {
	case []interface{}:
		if len(v) == 0 {
			return v
		}

		// Classify the array by its elements
		elemType := elemTypeScalar
		for _, elem := range v {
			switch elem.(type) {
			case []interface{}:
				if elemType == elemTypeScalar {
					elemType = elemTypeNested
				}
			case float64, float32, int, int32, int64, string:
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
			result := make([][]string, 0, len(v))
			for _, elem := range v {
				if arr, ok := elem.([]interface{}); ok {
					inner := make([]string, 0, len(arr))
					for _, innerElem := range arr {
						inner = append(inner, convertToString(innerElem))
					}
					result = append(result, inner)
				}
			}
			return result

		case elemTypeScalar:
			// All elements are scalar: convert to []string
			result := make([]string, 0, len(v))
			for _, elem := range v {
				result = append(result, convertToString(elem))
			}
			return result

		default:
			// Mixed types or unknown, return as-is
			return v
		}

	case map[string]interface{}:
		// Recursively convert map values
		// Determine the result type based on value types
		hasArray := false
		hasMap := false
		for _, val := range v {
			switch val.(type) {
			case []interface{}:
				hasArray = true
			case map[string]interface{}:
				hasMap = true
			}
		}

		// If has nested map, keep as map[string]interface{}
		if hasMap {
			result := make(map[string]interface{}, len(v))
			for k, val := range v {
				result[k] = convertArrayToStringArray(val)
			}
			return result
		}

		// If has array, return map[string][]string
		if hasArray {
			result := make(map[string][]string, len(v))
			for k, val := range v {
				converted := convertArrayToStringArray(val)
				if arr, ok := converted.([]string); ok {
					result[k] = arr
				} else if arr, ok := converted.([]interface{}); ok {
					// Convert []interface{} to []string
					strArr := make([]string, 0, len(arr))
					for _, elem := range arr {
						strArr = append(strArr, convertToString(elem))
					}
					result[k] = strArr
				} else {
					// Single value, wrap in slice
					result[k] = []string{convertToString(val)}
				}
			}
			return result
		}

		// All values are scalar types, return map[string]string
		result := make(map[string]string, len(v))
		for k, val := range v {
			result[k] = convertToString(val)
		}
		return result

	case float64:
		if v == float64(int64(v)) {
			return int(v)
		} else if v == float64(int32(v)) {
			return int(v)
		} else {
			return v
		}

	default:
		return value
	}
}

// convertToString converts a numeric value to string preserving precision
func convertToString(value interface{}) string {
	switch v := value.(type) {
	case float64:
		// Check if the float64 represents an integer
		if v == float64(int64(v)) {
			// It's an integer, use integer formatting to preserve precision
			return strconv.FormatInt(int64(v), 10)
		}
		// It's a float, use default formatting
		return fmt.Sprintf("%v", v)
	case float32:
		if v == float32(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return fmt.Sprintf("%v", v)
	case int:
		return strconv.Itoa(v)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case string:
		return v
	default:
		return utils.ToString(value, "")
	}
}
