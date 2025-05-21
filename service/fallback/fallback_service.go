package fallback

import (
	"sync"

	"github.com/alibaba/pairec/v2/recconf"
)

var fallbackService *FallbackService

func init() {
	fallbackService = NewFallbackService()
}

type FallbackService struct {
	fallbacks map[string]IFallback

	sync.RWMutex
}

func NewFallbackService() *FallbackService {
	service := FallbackService{
		fallbacks: make(map[string]IFallback, 0),
	}

	return &service
}

func DefaultFallbackService() *FallbackService {
	return fallbackService
}

func RegisterFallback(sceneName string, fallback IFallback) {
	DefaultFallbackService().AddFallback(sceneName, fallback)
}

func (r *FallbackService) AddFallback(sceneName string, fallback IFallback) {
	r.Lock()
	defer r.Unlock()

	r.fallbacks[sceneName] = fallback
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
			f := NewFeatureStoreFallback(*categoryConf.FallbackConfig)
			if f != nil {
				RegisterFallback(scene, f)
			}
		}
	}
}
