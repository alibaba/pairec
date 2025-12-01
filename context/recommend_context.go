package context

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pairec-config-go-sdk/v2/model"
	"github.com/tidwall/gjson"
)

type IParam interface {
	GetParameter(name string) interface{}
}
type RecommendContext struct {
	Debug            bool
	Size             int
	Param            IParam
	Config           *recconf.RecommendConfig
	ExperimentResult *model.ExperimentResult
	RecommendId      string
	ExpId            string
	Log              []string
	mu               sync.RWMutex
	contexParams     map[string]interface{}
}

func NewRecommendContext() *RecommendContext {
	context := RecommendContext{Size: -1, Debug: false, Log: make([]string, 0, 16), contexParams: make(map[string]interface{})}
	return &context
}
func (r *RecommendContext) GetParameter(name string) interface{} {
	return r.Param.GetParameter(name)
}

func (r *RecommendContext) GetParameterByPath(path string) interface{} {
	if strings.Contains(path, ".") {
		pos := strings.Index(path, ".")
		root := path[:pos]
		features := r.Param.GetParameter(root)
		if features == nil {
			return nil
		}
		newPath := path[pos+1:]
		if newPath == "" {
			return features
		}
		if dict, ok := features.(map[string]interface{}); ok {
			b, err := json.Marshal(dict)
			if err != nil {
				r.LogError(fmt.Sprintf("GetParameterByPath Marshal fail: %v", err))
				return nil
			}
			value := gjson.Get(string(b), newPath)
			return value.String()
		} else if feature, okay := features.(string); okay {
			value := gjson.Get(feature, newPath)
			return value.String()
		} else {
			r.LogError("GetParameterByPath fail: " + path)
			return nil
		}
	}
	return r.Param.GetParameter(path)
}

func (r *RecommendContext) LogDebug(msg string) {
	if r.Debug {
		r.Log = append(r.Log, fmt.Sprintf("[DEBUG] %s", strings.Replace(msg, "\t", "  ", -1)))
	}
}
func (r *RecommendContext) LogInfo(msg string) {
	r.Log = append(r.Log, fmt.Sprintf("[INFO] %s", strings.Replace(msg, "\t", "  ", -1)))
	log.Info(fmt.Sprintf("requestId=%s\t%s", r.RecommendId, msg))
}
func (r *RecommendContext) LogWarning(msg string) {
	r.Log = append(r.Log, fmt.Sprintf("[WARN] %s", strings.Replace(msg, "\t", "  ", -1)))
	log.Warning(fmt.Sprintf("requestId=%s\t%s", r.RecommendId, msg))
}
func (r *RecommendContext) LogError(msg string) {
	r.Log = append(r.Log, fmt.Sprintf("[ERROR] %s", strings.Replace(msg, "\t", "  ", -1)))
	log.Error(fmt.Sprintf("requestId=%s\t%s", r.RecommendId, msg))
}

func (r *RecommendContext) AddContextParam(name string, value interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.contexParams[name] = value
}

func (r *RecommendContext) GetContextParam(name string) interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.contexParams[name]
}
func (r *RecommendContext) DeleteContextParam(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.contexParams, name)
}
