package utils

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/model"
	"strings"
	"time"
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
	default:
		return value
	}
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
