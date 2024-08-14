package pipeline

import (
	"encoding/json"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/utils"
)

type FeatureService struct {
	pipelineName     string
	featureServices  map[string]*feature.FeatureService
	featureLoadConfs []recconf.FeatureLoadConfig
}

func NewFeatureService(config *recconf.PipelineConfig) *FeatureService {

	service := FeatureService{
		pipelineName:    config.Name,
		featureServices: make(map[string]*feature.FeatureService),
	}

	return &service
}

func (fs *FeatureService) LoadFeatures(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {

	featureService := fs.GetFeatureServiceByContext(context)
	if featureService == nil {
		return items
	}

	featureService.LoadFeaturesForPipelineRank(user, items, context, fs.pipelineName)

	return items
}

func (fs *FeatureService) GetFeatureServiceByContext(context *context.RecommendContext) *feature.FeatureService {
	scene := context.GetParameter("scene").(string)

	var featureLoadConfs []recconf.FeatureLoadConfig
	found := false
	if context.ExperimentResult != nil {
		featureconf := context.ExperimentResult.GetExperimentParams().Get("pipelines."+fs.pipelineName+".FeatureLoadConfs", "")
		if featureconf != "" {
			d, _ := json.Marshal(featureconf)
			if err := json.Unmarshal(d, &featureLoadConfs); err == nil {
				found = true
			}
		}
	}

	if !found {
		featureLoadConfs = fs.featureLoadConfs
		found = true
	}

	if !found {
		return nil
	}

	d, _ := json.Marshal(featureLoadConfs)
	id := scene + "#" + utils.Md5(string(d))
	if featureService, ok := fs.featureServices[id]; ok {
		return featureService
	}

	featureService := feature.NewFeatureService()
	var features []*feature.Feature
	for _, conf := range featureLoadConfs {
		f := feature.LoadWithConfig(conf)
		features = append(features, f)
	}

	featureService.SetFeatureSceneAsync(scene, true)
	featureService.SetFeatures(scene, features)

	fs.featureServices[id] = featureService

	return featureService
}
