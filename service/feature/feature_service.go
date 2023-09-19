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
	"github.com/alibaba/pairec/v2/utils"
)

var featureService *FeatureService

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
	s.FeatureSceneAsyncMap[sceneName] = async
}

func (s *FeatureService) SetFeatures(sceneName string, features []*Feature) {
	s.FeatureSceneMap[sceneName] = features
}

func (s *FeatureService) AddFeatureFunc(sceneName string, f FeatureFunc) {
	s.FeatureFuncMap[sceneName] = f
}

func (s *FeatureService) AddLoadFeatureFunc(sceneName string, f LoadFeatureFunc) {
	s.LoadFeatureFuncMap[sceneName] = f
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

	if features, ok := s.FeatureSceneMap[sceneName]; ok {
		async := s.FeatureSceneAsyncMap[sceneName]
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

	log.Info(fmt.Sprintf("requestId=%s\tmodule=LoadFeatures\tcost=%d", context.RecommendId, utils.CostTime(start)))

	return items
}

func (s *FeatureService) LoadFeaturesForGeneralRank(user *module.User, items []*module.Item, context *context.RecommendContext, pipeline string) {
	start := time.Now()
	sceneName := context.GetParameter("scene").(string)
	if features, ok := s.FeatureSceneMap[sceneName]; ok {
		async := s.FeatureSceneAsyncMap[sceneName]
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
	if features, ok := s.FeatureSceneMap[sceneName]; ok {
		async := s.FeatureSceneAsyncMap[sceneName]
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
	for name, sceneConf := range config.FeatureConfs {
		if _, ok := featureService.FeatureSceneMap[name]; ok {
			if signOfFeatureLoadConfs(sceneConf.FeatureLoadConfs) == featureService.FeatureSceneSigns[name] {
				continue
			}
		}

		var features []*Feature
		for _, conf := range sceneConf.FeatureLoadConfs {
			f := LoadWithConfig(conf)
			features = append(features, f)
		}

		featureService.FeatureSceneAsyncMap[name] = false
		if sceneConf.AsynLoadFeature {
			featureService.FeatureSceneAsyncMap[name] = true
		}
		featureService.FeatureSceneMap[name] = features
		featureService.FeatureSceneSigns[name] = signOfFeatureLoadConfs(sceneConf.FeatureLoadConfs)
	}
}

func signOfFeatureLoadConfs(confs []recconf.FeatureLoadConfig) string {
	var signs string
	for _, conf := range confs {
		sign, _ := json.Marshal(conf)

		signs += string(sign)
	}

	return utils.Md5(signs)

}
