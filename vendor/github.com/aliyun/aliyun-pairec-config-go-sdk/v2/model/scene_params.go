package model

import (
	"encoding/json"
	"strconv"

	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/common"
)

// SceneParams offers Get* function to get value by the key
// If not found the key, defaultValue will return
type SceneParams interface {
	AddParam(key string, value interface{})

	AddParams(params map[string]interface{})

	Get(key string, defaultValue interface{}) interface{}

	GetString(key, defaultValue string) string

	GetInt(key string, defaultValue int) int

	GetFloat(key string, defaultValue float64) float64

	GetInt64(key string, defaultValue int64) int64

	GetFeatureConsistencyJobs() []*FeatureConsistencyJob
}

type sceneParams struct {
	Parameters             map[string]interface{}
	featureConsistencyJobs []*FeatureConsistencyJob
}

func NewSceneParams() *sceneParams {
	return &sceneParams{
		Parameters: make(map[string]interface{}, 0),
	}
}

func (r *sceneParams) AddParam(key string, value interface{}) {
	r.Parameters[key] = value
	if key == common.Feature_Consistency_Job_Param_Name {
		json.Unmarshal([]byte(value.(string)), &r.featureConsistencyJobs)
	}
}

func (r *sceneParams) AddParams(params map[string]interface{}) {
	for k, v := range params {
		r.AddParam(k, v)
	}
}

func (r *sceneParams) GetFeatureConsistencyJobs() []*FeatureConsistencyJob {
	return r.featureConsistencyJobs
}

func (r *sceneParams) Get(key string, defaultValue interface{}) interface{} {
	if val, ok := r.Parameters[key]; ok {
		return val
	}
	return defaultValue
}

func (r *sceneParams) GetString(key, defaultValue string) string {
	val, ok := r.Parameters[key]
	if !ok {
		return defaultValue
	}

	switch value := val.(type) {
	case string:
		if value == "" {
			return defaultValue
		}
		return value
	case int:
		return strconv.Itoa(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.FormatInt(value, 10)
	}
	return defaultValue
}

func (r *sceneParams) GetInt(key string, defaultValue int) int {
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
func (r *sceneParams) GetFloat(key string, defaultValue float64) float64 {
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
func (r *sceneParams) GetInt64(key string, defaultValue int64) int64 {
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

type emptySceneParams struct {
}

func NewEmptySceneParams() *emptySceneParams {
	return &emptySceneParams{}
}

func (r *emptySceneParams) AddParam(key string, value interface{}) {}

func (r *emptySceneParams) AddParams(params map[string]interface{}) {}

func (r *emptySceneParams) Get(key string, defaultValue interface{}) interface{} {
	return defaultValue
}

func (r *emptySceneParams) GetString(key, defaultValue string) string {
	return defaultValue
}

func (r *emptySceneParams) GetInt(key string, defaultValue int) int {
	return defaultValue
}
func (r *emptySceneParams) GetFloat(key string, defaultValue float64) float64 {
	return defaultValue
}
func (r *emptySceneParams) GetInt64(key string, defaultValue int64) int64 {
	return defaultValue
}

func (r *emptySceneParams) GetFeatureConsistencyJobs() (ret []*FeatureConsistencyJob) {
	return
}
