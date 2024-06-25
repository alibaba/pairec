package module

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/alibaba/pairec/v2/context"
)

type UID string

type User struct {
	Id UID `json:"uid"`
	//Vector     string
	mutex                    sync.RWMutex
	Properties               map[string]interface{}    `json:"properties"`
	cacheFeatures            map[string]map[string]any `json:"-"`
	featureAsyncLoadCount    int32
	featureAsyncLoadCh       chan struct{}
	featureAsyncLoadChClosed bool
}

func NewUser(id string) *User {
	user := User{
		Id:                    UID(id),
		cacheFeatures:         make(map[string]map[string]any),
		featureAsyncLoadCount: 0,
		featureAsyncLoadCh:    make(chan struct{}, 1),
	}
	user.Properties = make(map[string]interface{})
	return &user
}
func NewUserWithContext(id UID, context *context.RecommendContext) *User {
	user := NewUser(string(id))

	features := context.GetParameter("features")

	properties := make(map[string]any, 64)
	if featuresMap, ok := features.(map[string]interface{}); ok {
		properties = make(map[string]any, len(featuresMap))
		for k, v := range featuresMap {
			if strValue, ok := v.(string); ok {
				if strValue != "" {
					properties[k] = v
				}
			} else {
				properties[k] = v
			}
		}
	}

	properties["uid"] = string(id)
	user.Properties = properties
	return user
}
func (u *User) Clone() *User {
	user := User{
		Id:                       u.Id,
		Properties:               make(map[string]interface{}),
		cacheFeatures:            make(map[string]map[string]any),
		featureAsyncLoadCh:       make(chan struct{}, 1),
		featureAsyncLoadChClosed: true,
	}
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	close(user.featureAsyncLoadCh)
	for k, v := range u.Properties {
		user.Properties[k] = v
	}

	for name, m := range u.cacheFeatures {
		cloneMap := make(map[string]any, len(m))
		for k, v := range m {
			cloneMap[k] = v
		}

		user.cacheFeatures[name] = cloneMap

	}

	return &user
}

func (u *User) AddProperty(key string, value interface{}) {
	u.mutex.Lock()
	u.Properties[key] = value
	u.mutex.Unlock()
}
func (u *User) AddProperties(properties map[string]interface{}) {
	u.mutex.Lock()
	for key, val := range properties {
		u.Properties[key] = val
	}
	u.mutex.Unlock()
}
func (u *User) FloatProperty(key string) (float64, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	val, ok := u.Properties[key]
	if !ok {
		return float64(0), errors.New("property key not exist")
	}

	switch value := val.(type) {
	case float64:
		return value, nil
	case int:
		return float64(value), nil
	case string:
		f, err := strconv.ParseFloat(value, 64)
		return f, err
	default:
		return float64(0), errors.New("unsupported type")
	}
}

func (u *User) GetEmbeddingFeature() (features map[string]interface{}) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	features = make(map[string]interface{})
	for k, v := range u.Properties {
		if strings.HasSuffix(k, "embedding") {
			if emb, ok := v.(string); ok {
				features[k] = strings.Trim(emb, "{}")
			} else {
				features[k] = v
			}
		}
	}
	return
}

func (u *User) MakeUserFeatures() (features map[string]interface{}) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	features = make(map[string]interface{})
	for k, v := range u.Properties {
		if k == "type" {
			continue
		}
		if s, ok := v.(float64); ok {
			features[k] = s
			continue
		}
		if str, ok := v.(string); ok {
			if s, err := strconv.ParseFloat(str, 64); err == nil {
				features[k] = s
				continue
			}
		}

		features[k] = v
	}
	return
}

// MakeUserFeatures2 for easyrec processor
func (u *User) MakeUserFeatures2() (features map[string]interface{}) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	features = make(map[string]interface{}, len(u.Properties))
	for k, v := range u.Properties {
		features[k] = v
	}
	return
}
func (u *User) StringProperty(key string) string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	val, ok := u.Properties[key]
	if !ok {
		return ""
	}

	switch value := val.(type) {
	case string:
		return value
	case int:
		return strconv.Itoa(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case int32:
		return strconv.Itoa(int(value))
	case int64:
		return strconv.Itoa(int(value))
	}
	return ""
}

func (u *User) SetProperties(p map[string]interface{}) {
	u.mutex.Lock()
	u.Properties = p
	u.mutex.Unlock()
}

func (u *User) DeleteProperty(key string) {
	u.mutex.Lock()
	delete(u.Properties, key)
	u.mutex.Unlock()
}

func (u *User) DeleteProperties(features []string) {
	u.mutex.Lock()
	for _, key := range features {
		delete(u.Properties, key)
	}
	u.mutex.Unlock()
}
func (u *User) IntProperty(key string) (int, error) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	val, ok := u.Properties[key]
	if !ok {
		return int(0), errors.New("property key not exist")
	}

	switch value := val.(type) {
	case float64:
		return int(value), nil
	case int:
		return value, nil
	case uint:
		return int(value), nil
	case int32:
		return int(value), nil
	case int64:
		return int(value), nil
	case string:
		return strconv.Atoi(value)
	default:
		return int(0), errors.New("unsupported type")
	}
}
func (u *User) GetProperty(key string) interface{} {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	val, ok := u.Properties[key]
	if !ok {
		return nil
	}
	return val
}
func (u *User) AddCacheFeatures(key string, features map[string]any) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	m, ok := u.cacheFeatures[key]
	if !ok {
		m = make(map[string]any, len(features))

	}

	for k, v := range features {
		m[k] = v
	}
	u.cacheFeatures[key] = m
}

func (u *User) LoadCacheFeatures(key string) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if m, ok := u.cacheFeatures[key]; ok {
		for k, v := range m {
			u.Properties[k] = v
		}
	}
}

func (u *User) GetCacheFeatures(key string) (result map[string]interface{}) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	result = make(map[string]interface{})
	if m, ok := u.cacheFeatures[key]; ok {
		for k, v := range m {
			result[k] = v
		}
	}
	return
}

func (u *User) IncrementFeatureAsyncLoadCount(count int32) {
	atomic.AddInt32(&u.featureAsyncLoadCount, count)
}

func (u *User) DescFeatureAsyncLoadCount(count int32) {
	u.mutex.Lock()
	defer u.mutex.Unlock()
	if atomic.LoadInt32(&u.featureAsyncLoadCount) < 1 {
		panic("featureAsyncLoadCount not less than 0")
	}
	curr := atomic.AddInt32(&u.featureAsyncLoadCount, -1*count)
	if curr == 0 {
		if !u.featureAsyncLoadChClosed {
			close(u.featureAsyncLoadCh)
			u.featureAsyncLoadChClosed = true
		}
	}
}

func (u *User) FeatureAsyncLoadCount() int32 {
	return atomic.LoadInt32(&u.featureAsyncLoadCount)
}

func (u *User) GetCacheFeaturesNames() (ret []string) {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	for k := range u.cacheFeatures {
		ret = append(ret, k)
	}
	return
}

func (u *User) FeatureAsyncLoadCh() <-chan struct{} {
	return u.featureAsyncLoadCh
}
