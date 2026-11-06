package feature

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/metrics"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/goburrow/cache"
)

var featureService *FeatureService
var featureCacheAfterFunc = time.AfterFunc

// FeatureFunc is use for feature engineering, after all the features loaded
type FeatureFunc func(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item

// LoadFeatureFunc is the function you can custom define to load features.
type LoadFeatureFunc func(user *module.User, items []*module.Item, context *context.RecommendContext)

func init() {
	featureService = NewFeatureService()
}

func DefaultFeatureService() *FeatureService {
	return featureService
}
func RegisterFeatureFunc(sceneName string, f FeatureFunc) {
	DefaultFeatureService().AddFeatureFunc(sceneName, f)
}

func RegisterLoadFeatureFunc(sceneName string, f LoadFeatureFunc) {
	DefaultFeatureService().AddLoadFeatureFunc(sceneName, f)
}

type FeatureService struct {
	mu                   sync.RWMutex
	loadMu               sync.Mutex
	FeatureSceneMap      map[string][]*Feature
	FeatureSceneSigns    map[string]string
	FeatureSceneAsyncMap map[string]bool
	FeatureFuncMap       map[string]FeatureFunc
	LoadFeatureFuncMap   map[string]LoadFeatureFunc
}

func NewFeatureService() *FeatureService {
	fs := &FeatureService{
		FeatureSceneMap:      make(map[string][]*Feature),
		FeatureSceneSigns:    make(map[string]string),
		FeatureSceneAsyncMap: make(map[string]bool),
		FeatureFuncMap:       make(map[string]FeatureFunc),
		LoadFeatureFuncMap:   make(map[string]LoadFeatureFunc),
	}

	return fs
}

func (s *FeatureService) SetFeatureSceneAsync(sceneName string, async bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FeatureSceneAsyncMap[sceneName] = async
}

func (s *FeatureService) SetFeatures(sceneName string, features []*Feature) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FeatureSceneMap[sceneName] = features
}

func (s *FeatureService) AddFeatureFunc(sceneName string, f FeatureFunc) {
	s.FeatureFuncMap[sceneName] = f
}

func (s *FeatureService) AddLoadFeatureFunc(sceneName string, f LoadFeatureFunc) {
	s.LoadFeatureFuncMap[sceneName] = f
}

func (s *FeatureService) getFeatures(scene string) ([]*Feature, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	features, ok := s.FeatureSceneMap[scene]
	return features, s.FeatureSceneAsyncMap[scene], ok
}

// LoadFeatures load user or item feature use feature.Feature
func (s *FeatureService) LoadFeatures(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {
	start := time.Now()
	var sceneName string
	if context.ExperimentResult != nil {
		sceneName = context.ExperimentResult.GetExperimentParams().GetString("features.scene.name", "")
	}
	if sceneName == "" {
		sceneName = context.GetParameter("scene").(string)
	}

	if features, async, ok := s.getFeatures(sceneName); ok {
		if async {
			var wg sync.WaitGroup
			for _, fea := range features {
				wg.Add(1)
				go func(fea *Feature) {
					defer wg.Done()
					fea.LoadFeatures(user, items, context)
				}(fea)
			}
			if loadFeatureFunc, exist := s.LoadFeatureFuncMap[sceneName]; exist {
				wg.Add(1)
				go func() {
					defer wg.Done()
					loadFeatureFunc(user, items, context)
				}()
			}

			wg.Wait()
		} else {
			for _, feature := range features {
				feature.LoadFeatures(user, items, context)
			}

			if loadFeatureFunc, exist := s.LoadFeatureFuncMap[sceneName]; exist {
				loadFeatureFunc(user, items, context)
			}

		}

		featureFunc, ok := s.FeatureFuncMap[sceneName]
		if ok {
			items = featureFunc(user, items, context)
		}
	}

	if metrics.Enabled() {
		metrics.LoadFeatureDurSecs.WithLabelValues(sceneName, "", "before_rank").Observe(time.Since(start).Seconds())
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=LoadFeatures\tscene=%s\tcost=%d", context.RecommendId, sceneName, utils.CostTime(start)))

	return items
}

func (s *FeatureService) LoadFeaturesForGeneralRank(user *module.User, items []*module.Item, context *context.RecommendContext, pipeline string) {
	start := time.Now()
	sceneName := context.GetParameter("scene").(string)
	if features, async, ok := s.getFeatures(sceneName); ok {
		if async {
			var wg sync.WaitGroup
			for _, fea := range features {
				wg.Add(1)
				go func(fea *Feature) {
					defer wg.Done()
					fea.LoadFeatures(user, items, context)
				}(fea)
			}

			wg.Wait()
		} else {
			for _, feature := range features {
				feature.LoadFeatures(user, items, context)
			}

		}

	}

	if pipeline != "" {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=LoadFeaturesForGeneralRank\tpipeline=%s\tcost=%d", context.RecommendId, pipeline, utils.CostTime(start)))

	} else {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=LoadFeaturesForGeneralRank\tcost=%d", context.RecommendId, utils.CostTime(start)))

	}
}
func (s *FeatureService) LoadFeaturesForPipelineRank(user *module.User, items []*module.Item, context *context.RecommendContext, pipeline string) {
	start := time.Now()
	sceneName := context.GetParameter("scene").(string)
	if features, async, ok := s.getFeatures(sceneName); ok {
		if async {
			var wg sync.WaitGroup
			for _, fea := range features {
				wg.Add(1)
				go func(fea *Feature) {
					defer wg.Done()
					fea.LoadFeatures(user, items, context)
				}(fea)
			}

			wg.Wait()
		} else {
			for _, feature := range features {
				feature.LoadFeatures(user, items, context)
			}

		}

	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=LoadFeatures\tpipeline=%s\tcost=%d", context.RecommendId, pipeline, utils.CostTime(start)))

}

func LoadFeatureConfig(config *recconf.RecommendConfig) {
	s := featureService
	s.loadMu.Lock()
	defer s.loadMu.Unlock()
	for name, sceneConf := range config.FeatureConfs {
		sign := signOfFeatureLoadConfs(sceneConf.FeatureLoadConfs)
		s.mu.RLock()
		old, ok := s.FeatureSceneMap[name]
		unchanged := ok && sign == s.FeatureSceneSigns[name]
		s.mu.RUnlock()
		if unchanged {
			continue
		}

		features, reused := loadFeatureScene(sceneConf.FeatureLoadConfs, old)
		s.mu.Lock()
		s.FeatureSceneAsyncMap[name] = sceneConf.AsynLoadFeature
		s.FeatureSceneMap[name] = features
		s.FeatureSceneSigns[name] = sign
		s.mu.Unlock()
		for _, f := range old {
			if f.daoSign == "" {
				continue
			}
			if c := f.featureCache(); c != nil && !reused[c] {
				featureCacheAfterFunc(30*time.Second, c.InvalidateAll)
				reused[c] = true
			}
		}
	}
}

func loadFeatureScene(confs []recconf.FeatureLoadConfig, old []*Feature) ([]*Feature, map[cache.Cache]bool) {
	candidates := make(map[string][]*Feature)
	for _, f := range old {
		if f.daoSign != "" && f.featureCache() != nil {
			candidates[f.daoSign] = append(candidates[f.daoSign], f)
		}
	}
	reused := make(map[cache.Cache]bool)
	var created []cache.Cache
	var features []*Feature
	completed := false
	defer func() {
		if !completed {
			for _, c := range created {
				_ = c.Close()
			}
		}
	}()
	for _, conf := range confs {
		var sign string
		if (conf.FeatureDaoConf.AdapterType == recconf.DataSource_Type_FeatureStore || conf.FeatureDaoConf.AdapterType == recconf.DaoConf_Adapter_Hologres) && conf.FeatureDaoConf.FeatureCacheName == "" {
			sign = featureDaoSign(conf.FeatureDaoConf)
		}
		var inherited cache.Cache
		if queue := candidates[sign]; len(queue) > 0 {
			inherited = queue[0].featureCache()
			candidates[sign] = queue[1:]
		}
		f := loadWithCache(conf, inherited)
		features = append(features, f)
		if inherited != nil {
			reused[inherited] = true
		} else if c := f.featureCache(); c != nil {
			created = append(created, c)
		}
	}
	completed = true
	return features, reused
}

func signOfFeatureLoadConfs(confs []recconf.FeatureLoadConfig) string {
	var signs string
	for _, conf := range confs {
		sign, _ := json.Marshal(conf)

		signs += string(sign)
	}

	return utils.Md5(signs)

}
