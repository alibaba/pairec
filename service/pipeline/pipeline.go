package pipeline

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service/debug"
)

var pipelineService *PipelineService

func init() {
	pipelineService = NewPipelineService()
}

type PipelineService struct {
	userRecommendSceneMap map[string][]*UserRecommendService
}

func NewPipelineService() *PipelineService {
	service := PipelineService{
		userRecommendSceneMap: make(map[string][]*UserRecommendService),
	}

	return &service
}

func (p *PipelineService) clear() {
	p.userRecommendSceneMap = make(map[string][]*UserRecommendService, 0)
}

func LoadPipelineConfigs(conf *recconf.RecommendConfig) {
	userRecommendSceneMap := make(map[string][]*UserRecommendService)

	for sceneName, configs := range conf.PipelineConfs {
		for _, config := range configs {
			userRecommendService := NewUserRecommendService(&config)

			userRecommendSceneMap[sceneName] = append(userRecommendSceneMap[sceneName], userRecommendService)
		}
	}

	pipelineService.userRecommendSceneMap = userRecommendSceneMap
}
func Recommend(user *module.User, context *context.RecommendContext, debugService *debug.DebugService) (ret []*module.Item) {
	scene := context.GetParameter("scene").(string)

	services, ok := pipelineService.userRecommendSceneMap[scene]
	if !ok {
		return
	}

	ch := make(chan []*module.Item, len(services))

	for _, service := range services {
		go func(userRecommendService *UserRecommendService) {
			ch <- userRecommendService.Recommend(user, context, debugService)
		}(service)
	}

	for i := 0; i < len(services); i++ {
		items := <-ch
		ret = append(ret, items...)
	}
	close(ch)

	return
}
