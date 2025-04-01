package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/tidwall/gjson"
)

func CostTime(start time.Time) int64 {
	duration := time.Now().UnixNano() - start.UnixNano()

	return duration / 1e6
}

type FeatureInfo struct {
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

func ConvertFeatures(features map[string]interface{}) map[string]*FeatureInfo {
	ret := make(map[string]*FeatureInfo, len(features))
	for name, value := range features {
		info := &FeatureInfo{
			Value: value,
			Type:  GetTypeOf(value),
		}

		ret[name] = info
	}
	return ret
}

func GetTypeOf(value interface{}) string {
	switch value.(type) {
	case int:
		return "int"
	case int32:
		return "int"
	case int64:
		return "bigint"
	case string:
		return "string"
	case float64:
		return "float64"
	case bool:
		return "bool"
	case []int:
		return "list<int>"
	case []int32:
		return "list<int>"
	case []int64:
		return "list<int64>"
	case []float32:
		return "list<float>"
	case []float64:
		return "list<double>"
	case []string:
		return "list<string>"
	case [][]float32:
		return "list<list<float>>"
	case [][]float64:
		return "list<list<double>>"
	case map[string]string:
		return "map<string,string>"
	case map[string]int:
		return "map<string,int>"
	case map[string]int32:
		return "map<string,int>"
	case map[string]int64:
		return "map<string,int64>"
	case map[string]float32:
		return "map<string,float>"
	case map[string]float64:
		return "map<string,double>"
	case map[int]int:
		return "map<int,int>"
	case map[int]int32:
		return "map<int,int>"
	case map[int32]int32:
		return "map<int,int>"
	case map[int]int64:
		return "map<int,int64>"
	case map[int]float32:
		return "map<int,float>"
	case map[int]float64:
		return "map<int,double>"
	case map[int]string:
		return "map<int,string>"
	case map[int64]int64:
		return "map<int64,int64>"
	case map[int64]int:
		return "map<int64,int>"
	case map[int64]int32:
		return "map<int64,int>"
	case map[int64]float32:
		return "map<int64,float>"
	case map[int64]float64:
		return "map<int64,double>"
	case map[int64]string:
		return "map<int64,string>"
	default:
		return "none"
	}
}
func GetValueByType(value interface{}, vtype string) interface{} {
	switch vtype {
	case "int":
		return ToInt(value, 0)
	case "string":
		return ToString(value, "")
	case "float64":
		return ToFloat(value, float64(0))
	case "int64":
		return ToInt64(value, 0)
	case "bigint":
		return ToInt64(value, 0)
	case "bool":
		return ToBool(value, false)
	case "list<string>":
		return ToStringArray(value)
	case "list<int>":
		return ToIntArray(value)
	case "list<int64>":
		if vals, ok := value.([]any); ok {
			values := make([]int64, 0, len(vals))
			for _, v := range vals {
				values = append(values, ToInt64(v, 0))
			}
			return values
		}
	case "list<double>":
		if vals, ok := value.([]any); ok {
			values := make([]float64, 0, len(vals))
			for _, v := range vals {
				values = append(values, ToFloat(v, 0))
			}
			return values
		}
	case "list<float>":
		if vals, ok := value.([]any); ok {
			values := make([]float32, 0, len(vals))
			for _, v := range vals {
				values = append(values, ToFloat32(v, 0))
			}
			return values
		}
	case "list<list<float>>":
		if vals, ok := value.([]any); ok {
			values := make([][]float32, 0, len(vals))
			for _, vlist := range vals {
				if lists, ok := vlist.([]any); ok {
					list := make([]float32, 0, len(lists))
					for _, v := range lists {
						list = append(list, ToFloat32(v, 0))
					}
					values = append(values, list)
				}
			}

			return values
		}
	case "list<list<double>>":
		if vals, ok := value.([]any); ok {
			values := make([][]float64, 0, len(vals))
			for _, vlist := range vals {
				if lists, ok := vlist.([]any); ok {
					list := make([]float64, 0, len(lists))
					for _, v := range lists {
						list = append(list, ToFloat(v, 0))
					}
					values = append(values, list)
				}
			}

			return values
		}
	case "map<string,string>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[string]string, len(vals))
			for k, v := range vals {
				values[k] = ToString(v, "")
			}
			return values
		}
	case "map<string,int>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[string]int, len(vals))
			for k, v := range vals {
				values[k] = ToInt(v, 0)
			}
			return values
		}
	case "map<string,int64>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[string]int64, len(vals))
			for k, v := range vals {
				values[k] = ToInt64(v, 0)
			}
			return values
		}
	case "map<string,float>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[string]float32, len(vals))
			for k, v := range vals {
				values[k] = ToFloat32(v, 0)
			}
			return values
		}
	case "map<string,double>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[string]float64, len(vals))
			for k, v := range vals {
				values[k] = ToFloat(v, 0)
			}
			return values
		}
	case "map<int,int>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int]int, len(vals))
			for k, v := range vals {
				values[ToInt(k, 0)] = ToInt(v, 0)
			}
			return values
		}
	case "map<int,int64>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int]int64, len(vals))
			for k, v := range vals {
				values[ToInt(k, 0)] = ToInt64(v, 0)
			}
			return values
		}
	case "map<int,double>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int]float64, len(vals))
			for k, v := range vals {
				values[ToInt(k, 0)] = ToFloat(v, 0)
			}
			return values
		}
	case "map<int,float>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int]float32, len(vals))
			for k, v := range vals {
				values[ToInt(k, 0)] = ToFloat32(v, 0)
			}
			return values
		}
	case "map<int,string>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int]string, len(vals))
			for k, v := range vals {
				values[ToInt(k, 0)] = ToString(v, "")
			}
			return values
		}
	case "map<int64,int>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int64]int, len(vals))
			for k, v := range vals {
				values[ToInt64(k, 0)] = ToInt(v, 0)
			}
			return values
		}
	case "map<int64,int64>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int64]int64, len(vals))
			for k, v := range vals {
				values[ToInt64(k, 0)] = ToInt64(v, 0)
			}
			return values
		}
	case "map<int64,double>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int64]float64, len(vals))
			for k, v := range vals {
				values[ToInt64(k, 0)] = ToFloat(v, 0)
			}
			return values
		}
	case "map<int64,float>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int64]float32, len(vals))
			for k, v := range vals {
				values[ToInt64(k, 0)] = ToFloat32(v, 0)
			}
			return values
		}
	case "map<int64,string>":
		if vals, ok := value.(map[string]any); ok {
			values := make(map[int64]string, len(vals))
			for k, v := range vals {
				values[ToInt64(k, 0)] = ToString(v, "")
			}
			return values
		}
	default:
		return value
	}

	return value
}

func IndexOf(a []string, e string) int {
	n := len(a)
	var i = 0
	for ; i < n; i++ {
		if e == a[i] {
			return i
		}
	}
	return -1
}

func UniqueStrings(strSlice []string) []string {
	keys := make(map[string]bool)
	list := make([]string, 0, len(strSlice))
	for _, entry := range strSlice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func GetExperimentParamByPath(params model.LayerParams, path string, defaultValue interface{}) interface{} {
	if strings.Contains(path, ".") {
		pos := strings.Index(path, ".")
		root := path[:pos]
		rootConf := params.Get(root, nil)
		if rootConf == nil {
			return defaultValue
		}
		newPath := path[pos+1:]
		if newPath == "" {
			return rootConf
		}
		if dict, ok := rootConf.(map[string]interface{}); ok {
			b, err := json.Marshal(dict)
			if err != nil {
				fmt.Sprintf("GetExpParamByPath Marshal fail: %v", err)
				return defaultValue
			}
			value := gjson.Get(string(b), newPath).Value()
			if value == nil {
				return defaultValue
			}
			return value
		} else if feature, okay := rootConf.(string); okay {
			value := gjson.Get(feature, newPath).Value()
			if value == nil {
				return defaultValue
			}
			return value
		} else {
			return defaultValue
		}
	}
	return params.Get(path, defaultValue)
}
