package model

import (
	"strconv"
)

// LayerParams offers Get* function to get value by the key
// If not found the key, defaultValue will return
type LayerParams interface {
	AddParam(key string, value interface{})

	AddParams(params map[string]interface{})

	Get(key string, defaultValue interface{}) interface{}

	GetString(key, defaultValue string) string

	GetInt(key string, defaultValue int) int

	GetFloat(key string, defaultValue float64) float64

	GetInt64(key string, defaultValue int64) int64
}

type layerParams struct {
	Parameters map[string]interface{}
}

func NewLayerParams() *layerParams {
	return &layerParams{
		Parameters: make(map[string]interface{}, 0),
	}
}

func (r *layerParams) AddParam(key string, value interface{}) {
	r.Parameters[key] = value
}

func (r *layerParams) AddParams(params map[string]interface{}) {
	for k, v := range params {
		r.Parameters[k] = v
	}
}

func (r *layerParams) Get(key string, defaultValue interface{}) interface{} {
	if val, ok := r.Parameters[key]; ok {
		return val
	}
	return defaultValue
}

func (r *layerParams) GetString(key, defaultValue string) string {
	val, ok := r.Parameters[key]
	if !ok {
		return defaultValue
	}

	switch value := val.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case float64:
		return strconv.Itoa(int(value))
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.Itoa(int(value))
	}
	return defaultValue
}

func (r *layerParams) GetInt(key string, defaultValue int) int {
	val, ok := r.Parameters[key]
	if !ok {
		return defaultValue
	}
	switch value := val.(type) {
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
			return defaultValue
		}
	default:
		return defaultValue
	}
}
func (r *layerParams) GetFloat(key string, defaultValue float64) float64 {
	val, ok := r.Parameters[key]
	if !ok {
		return defaultValue
	}

	switch value := val.(type) {
	case float64:
		return value
	case int:
		return float64(value)
	case string:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		} else {
			return defaultValue
		}
	default:
		return defaultValue
	}
}
func (r *layerParams) GetInt64(key string, defaultValue int64) int64 {
	val, ok := r.Parameters[key]
	if !ok {
		return defaultValue
	}

	switch value := val.(type) {
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
			return defaultValue
		}
	default:
		return defaultValue
	}
}

func MergeLayerParams(layersParamsMap map[string]LayerParams) LayerParams {
	mergedParams := NewLayerParams()
	for _, unmergedParams := range layersParamsMap {
		switch v := unmergedParams.(type) {
		case *layerParams:
			for k, p := range v.Parameters {
				mergedParams.Parameters[k] = p
			}
		}
	}
	return LayerParams(mergedParams)
}
