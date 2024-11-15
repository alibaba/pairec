package module

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/alibaba/pairec/v2/utils"
)

type ItemId string

type Item struct {
	Id         ItemId `json:"id"`
	Name       string `json:"name,omitempty"`
	Score      float64
	RetrieveId string
	ItemType   string
	Embedding  []float64
	//Extra     interface{}

	mutex        sync.RWMutex
	Properties   map[string]interface{} `json:"properties"`
	algoScores   map[string]float64
	RecallScores map[string]float64
}

func NewItem(id string) *Item {
	item := Item{
		Id:         ItemId(id),
		Score:      0,
		Properties: make(map[string]interface{}, 32),
	}
	//item.algoScores = make(map[string]float64)
	return &item
}
func NewItemWithProperty(id string, properties map[string]interface{}) *Item {
	item := NewItem(id)
	for k, v := range properties {
		item.Properties[k] = v
	}

	return item
}
func (t *Item) GetRecallName() string {
	return t.RetrieveId
}
func (t *Item) GetAlgoScores() map[string]float64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.algoScores
}
func (t *Item) GetAlgoScoreWithNames(names []string) map[string]float64 {
	ret := make(map[string]float64, len(names))
	for _, n := range names {
		ret[n] = t.algoScores[n]
	}
	return ret
}
func (t *Item) GetAlgoScore(key string) float64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.algoScores[key]
}
func (t *Item) IncrAlgoScore(name string, score float64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.algoScores == nil {
		t.algoScores = make(map[string]float64)
	}
	t.algoScores[name] += score
}
func (t *Item) CloneAlgoScores() map[string]float64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	algoScores := make(map[string]float64, len(t.algoScores))
	for k, v := range t.algoScores {
		algoScores[k] = v
	}
	return algoScores
}
func (t *Item) AddProperty(key string, value interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.Properties[key] = value
}

func (t *Item) GetProperty(key string) interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	val, ok := t.Properties[key]
	if !ok {
		return nil
	}
	return val
}
func (t *Item) StringProperty(key string) string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	val, ok := t.Properties[key]
	if !ok {
		return ""
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
	return ""
}
func (t *Item) FloatProperty(key string) (float64, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	val, ok := t.Properties[key]
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
		return float64(0), errors.New("unspport type")
	}
}
func (t *Item) IntProperty(key string) (int, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	val, ok := t.Properties[key]
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
		return int(0), errors.New("unspport type")
	}
}
func (t *Item) AddAlgoScore(name string, score float64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.algoScores == nil {
		t.algoScores = make(map[string]float64)
	}
	t.algoScores[name] = score
}
func (t *Item) AddAlgoScores(scores map[string]float64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.algoScores == nil {
		t.algoScores = make(map[string]float64)
	}

	for name, score := range scores {
		t.algoScores[name] = score
	}
}
func (t *Item) FloatExprData(name string) (float64, error) {
	if name == "current_score" {
		t.mutex.Lock()
		defer t.mutex.Unlock()
		if t.algoScores == nil {
			t.algoScores = make(map[string]float64)
		}
		t.algoScores["recall_score"] = t.Score
		return t.Score, nil
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()
	val, ok := t.algoScores[name]
	if ok {
		return val, nil
	}

	if val, ok := t.Properties[name]; ok {
		return utils.ToFloat(val, 0), nil
	}

	return float64(0), fmt.Errorf("not found,name:%s", name)
}
func (t *Item) ExprData() map[string]any {
	ret := make(map[string]any, len(t.algoScores)+len(t.Properties))

	t.mutex.RLock()
	defer t.mutex.RUnlock()
	for k, v := range t.algoScores {
		ret[k] = v
	}

	for k, v := range t.Properties {
		ret[k] = v
	}

	return ret
}

func (t *Item) GetFeatures() map[string]interface{} {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.RetrieveId != "" {
		if _, ok := t.Properties[t.RetrieveId]; !ok {
			t.Properties[t.RetrieveId] = t.Score
			t.Properties["recall_name"] = t.RetrieveId
			t.Properties["recall_score"] = t.Score
		}
	}

	features := make(map[string]interface{}, len(t.Properties))

	for k, v := range t.Properties {
		features[k] = v
	}

	return features
}
func (t *Item) AddRecallNameFeature() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.RetrieveId != "" {
		if _, ok := t.Properties[t.RetrieveId]; !ok {
			t.Properties[t.RetrieveId] = t.Score
			t.Properties["recall_name"] = t.RetrieveId
			t.Properties["recall_score"] = t.Score
		}
	}

}
func (t *Item) GetProperties() map[string]interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	return t.Properties
}
func (t *Item) GetCloneFeatures() map[string]interface{} {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	features := make(map[string]interface{}, len(t.Properties))

	for k, v := range t.Properties {
		features[k] = v
	}

	return features
}
func (t *Item) DeleteProperty(key string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.Properties, key)
}
func (t *Item) DeleteProperties(features []string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, key := range features {
		delete(t.Properties, key)
	}
}
func (t *Item) AddProperties(properties map[string]interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for key, val := range properties {
		t.Properties[key] = val
	}
}
func (t *Item) DeepClone() *Item {
	item := NewItemWithProperty(string(t.Id), t.Properties)

	item.Score = t.Score
	item.RetrieveId = t.RetrieveId
	item.ItemType = t.ItemType

	return item
}
