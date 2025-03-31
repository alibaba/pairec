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
	/**
	case map[string]string:
		return "map<string,string>"
	case map[string]int:
		return "map<string,int>"
	case map[string]int64:
		return "map<string,int64>"
	**/
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
