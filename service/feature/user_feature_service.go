package feature

import (
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

var userFeatureService *UserFeatureService

// UserFeatureFunc is use for feature engineering, after all the features loaded
type UserFeatureFunc func(user *module.User, context *context.RecommendContext) []*module.Item

// LoadFeatureFunc is the function you can custom define to load features.
type UserLoadFeatureFunc func(user *module.User, context *context.RecommendContext)

func init() {
	userFeatureService = NewUserFeatureService()
}

func DefaultUserFeatureService() *UserFeatureService {
	return userFeatureService
}

func RegisterUserFeatureFunc(sceneName string, f UserFeatureFunc) {
	DefaultUserFeatureService().AddFeatureFunc(sceneName, func() FeatureFunc {
		return func(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {
			f(user, context)
			return items
		}
	}())
}

func RegisterUserLoadFeatureFunc(sceneName string, f UserLoadFeatureFunc) {
	DefaultUserFeatureService().AddLoadFeatureFunc(sceneName, func() LoadFeatureFunc {
		return func(user *module.User, items []*module.Item, context *context.RecommendContext) {
			f(user, context)
		}
	}())
}

type UserFeatureService struct {
	*FeatureService
	FeatureAsyncLoadMap map[*Feature]bool
}

func NewUserFeatureService() *UserFeatureService {
	return &UserFeatureService{
		FeatureService:      NewFeatureService(),
		FeatureAsyncLoadMap: make(map[*Feature]bool),
	}
}

func (s *UserFeatureService) GetLoadFeaturesStageNames(context *context.RecommendContext) (stageNames []string) {
	// if not have experiment, direct return scene name
	if context.ExperimentResult == nil {
		sceneName := context.GetParameter("scene").(string)
		stageNames = append(stageNames, sceneName)
		return
	}

	useStage := false
	stageFlage := context.ExperimentResult.GetExperimentParams().Get("user_features.multistage.on", false)
	useStage = utils.ToBool(stageFlage, false)

	// use mutlti stage to load features
	if useStage {
		if name := context.ExperimentResult.GetExperimentParams().GetString("user_features.stage.base.scene.name", ""); name != "" {
			stageNames = append(stageNames, name)
		}

		if name := context.ExperimentResult.GetExperimentParams().GetString("user_features.stage.recall.scene.name", ""); name != "" {
			stageNames = append(stageNames, name)
		}

		if name := context.ExperimentResult.GetExperimentParams().GetString("user_features.stage.generalrank.scene.name", ""); name != "" {
			stageNames = append(stageNames, name)
		}

		if name := context.ExperimentResult.GetExperimentParams().GetString("user_features.stage.rank.scene.name", ""); name != "" {
			stageNames = append(stageNames, name)
		}
		if len(stageNames) > 0 {
			return
		}

	}

	sceneName := context.ExperimentResult.GetExperimentParams().GetString("user_features.scene.name", "")
	if sceneName == "" {
		sceneName = context.GetParameter("scene").(string)
	}

	stageNames = append(stageNames, sceneName)

	return
}

// LoadUserFeatures load user  feature use feature.Feature
func (s *UserFeatureService) LoadUserFeatures(user *module.User, context *context.RecommendContext) {
	start := time.Now()
	stageNames := s.GetLoadFeaturesStageNames(context)

	for _, stageName := range stageNames {
		if features, ok := s.FeatureSceneMap[stageName]; ok {
			// for close featureAsyncLoadCh
			user.IncrementFeatureAsyncLoadCount(1)
			defer user.DescFeatureAsyncLoadCount(1)

			var wg sync.WaitGroup
			featureFunc, featureFuncExist := s.FeatureFuncMap[stageName]
			if featureFuncExist {
				var wg2 sync.WaitGroup
				for _, fea := range features {
					if s.FeatureAsyncLoadMap[fea] {
						user.IncrementFeatureAsyncLoadCount(1)
						wg2.Add(1)
						go func(fea *Feature) {
							defer wg2.Done()
							defer user.DescFeatureAsyncLoadCount(1)
							fea.LoadFeatures(user, nil, context)
						}(fea)

					} else {
						wg.Add(1)
						wg2.Add(1)
						go func(fea *Feature) {
							defer wg2.Done()
							defer wg.Done()
							fea.LoadFeatures(user, nil, context)
						}(fea)
					}
				}
				// custom user feature load
				if loadFeatureFunc, exist := s.LoadFeatureFuncMap[stageName]; exist {
					wg.Add(1)
					wg2.Add(1)
					go func() {
						defer wg2.Done()
						defer wg.Done()
						loadFeatureFunc(user, nil, context)
					}()
				}

				go func() {
					wg2.Wait()
					featureFunc(user, nil, context)
				}()
			} else {
				for _, fea := range features {
					if s.FeatureAsyncLoadMap[fea] {
						user.IncrementFeatureAsyncLoadCount(1)
						go func(fea *Feature) {
							defer user.DescFeatureAsyncLoadCount(1)
							fea.LoadFeatures(user, nil, context)
						}(fea)

					} else {
						wg.Add(1)
						go func(fea *Feature) {
							defer wg.Done()
							fea.LoadFeatures(user, nil, context)
						}(fea)
					}
				}
				if loadFeatureFunc, exist := s.LoadFeatureFuncMap[stageName]; exist {
					wg.Add(1)
					go func() {
						defer wg.Done()
						loadFeatureFunc(user, nil, context)
					}()
				}

			}

			wg.Wait()
			log.Info(fmt.Sprintf("requestId=%s\tmodule=UserLoadFeatures\tname=%s\tcost=%d", context.RecommendId, stageName, utils.CostTime(start)))
		}

	}

}

func (s *UserFeatureService) LoadUserFeaturesForCallback(user *module.User, context *context.RecommendContext) {
	start := time.Now()

	stageNames := s.GetLoadFeaturesStageNames(context)
	for _, stageName := range stageNames {
		if features, ok := s.FeatureSceneMap[stageName]; ok {
			// for close featureAsyncLoadCh
			user.IncrementFeatureAsyncLoadCount(1)
			defer user.DescFeatureAsyncLoadCount(1)

			var syncFeatures []*Feature
			for _, feature := range features {
				if s.FeatureAsyncLoadMap[feature] {
					user.IncrementFeatureAsyncLoadCount(1)
					go func(fea *Feature) {
						defer user.DescFeatureAsyncLoadCount(1)
						fea.LoadFeatures(user, nil, context)
					}(feature)
				} else {
					syncFeatures = append(syncFeatures, feature)
				}
			}

			go func() {
				if len(syncFeatures) > 0 {
					user.IncrementFeatureAsyncLoadCount(int32(len(syncFeatures)))
					defer user.DescFeatureAsyncLoadCount(int32(len(syncFeatures)))
					for _, feature := range syncFeatures {
						feature.LoadFeatures(user, nil, context)
					}
				}

			}()

			log.Info(fmt.Sprintf("requestId=%s\tmodule=UserLoadFeatures\tname=%s\tcost=%d", context.RecommendId, stageName, utils.CostTime(start)))
		}

	}

}

func UserLoadFeatureConfig(config *recconf.RecommendConfig) {
	for name, sceneConf := range config.UserFeatureConfs {
		if _, ok := userFeatureService.FeatureSceneMap[name]; ok {
			if signOfFeatureLoadConfs(sceneConf.FeatureLoadConfs) == userFeatureService.FeatureSceneSigns[name] {
				continue
			}
		}

		var features []*Feature
		for _, conf := range sceneConf.FeatureLoadConfs {
			if conf.FeatureDaoConf.FeatureStore != SOURCE_USER {
				panic("user load features config, FeatureStore value must is user")
			}
			f := LoadWithConfig(conf)
			userFeatureService.FeatureAsyncLoadMap[f] = false
			features = append(features, f)
			if conf.FeatureDaoConf.FeatureAsyncLoad {
				userFeatureService.FeatureAsyncLoadMap[f] = true
			}
		}

		/**
		userFeatureService.FeatureSceneAsyncMap[name] = true
		userFeatureService.FeatureSceneAsyncMap[name] = false
		if sceneConf.AsynLoadFeature {
			userFeatureService.FeatureSceneAsyncMap[name] = true
		}
		**/
		userFeatureService.FeatureSceneMap[name] = features
		userFeatureService.FeatureSceneSigns[name] = signOfFeatureLoadConfs(sceneConf.FeatureLoadConfs)
	}
}
