package fallback

import (
	"encoding/json"
	"sync"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

var fallbackService *FallbackService
var fallbackSigns map[string]string

func init() {
	fallbackService = NewFallbackService()
	fallbackSigns = make(map[string]string)
}

type FallbackService struct {
	fallbacks map[string]IFallback

	sync.RWMutex
}

func NewFallbackService() *FallbackService {
	service := FallbackService{
		fallbacks: make(map[string]IFallback),
	}

	return &service
}

func DefaultFallbackService() *FallbackService {
	return fallbackService
}

func RegisterFallback(sceneName string, fallback IFallback) {
	DefaultFallbackService().AddFallback(sceneName, fallback)
}

func RemoveFallback(sceneName string) {
	DefaultFallbackService().RemoveFallback(sceneName)
}

func (r *FallbackService) AddFallback(sceneName string, fallback IFallback) {
	r.Lock()
	defer r.Unlock()

	r.fallbacks[sceneName] = fallback
}

func (r *FallbackService) RemoveFallback(sceneName string) {
	r.Lock()
	defer r.Unlock()

	delete(r.fallbacks, sceneName)
}

func (r *FallbackService) GetFallback(sceneName string) (ret IFallback) {
	r.RLock()
	defer r.RUnlock()

	if fallback, ok := r.fallbacks[sceneName]; ok {
		ret = fallback
	}

	return
}

func LoadFallbackConfig(config *recconf.RecommendConfig) {
	for scene, conf := range config.SceneConfs {
		if categoryConf, ok := conf["default"]; ok && categoryConf.FallbackConfig != nil {
			sign, _ := json.Marshal(categoryConf.FallbackConfig)
			if utils.Md5(string(sign)) == fallbackSigns[scene] {
				continue
			}

			f := NewFeatureStoreFallback(*categoryConf.FallbackConfig)
			if f != nil {
				RegisterFallback(scene, f)
			}
		} else if ok && categoryConf.FallbackConfig == nil {
			RemoveFallback(scene)
		}
	}
}
