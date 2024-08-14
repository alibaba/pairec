package pipeline

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/filter"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/debug"
	"github.com/alibaba/pairec/v2/sort"
	"github.com/alibaba/pairec/v2/utils"
)

type UserRecommendService struct {
	pipelineName         string
	recallService        *RecallService
	filterService        *FilterService
	generalRankService   *GeneralRankService
	featureService       *FeatureService
	rankService          *RankService
	coldStartRankService *ColdStartRankService
	sortService          *SortService
}

func NewUserRecommendService(config *recconf.PipelineConfig) *UserRecommendService {
	service := UserRecommendService{
		pipelineName:         config.Name,
		recallService:        NewRecallService(config),
		filterService:        NewFilterService(config),
		featureService:       NewFeatureService(config),
		generalRankService:   NewGeneralRankService(config),
		rankService:          NewRankService(config),
		coldStartRankService: NewColdStartRankService(config),
		sortService:          NewSortService(config),
	}

	return &service
}
func (r *UserRecommendService) Recommend(user *module.User, context *context.RecommendContext, debugService *debug.DebugService) []*module.Item {

	start := time.Now()
	var items []*module.Item
	items = r.recallService.GetItems(user, context)
	if len(items) == 0 {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=pipeline\tpipeline=%s\tcount=0", context.RecommendId, r.pipelineName))
		return items
	}

	debugService.WriteRecallLog(user, items, context)

	// filter
	items = r.Filter(user, items, context)

	debugService.WriteFilterLog(user, items, context)

	// general rank
	items = r.generalRankService.Rank(user, items, context)

	debugService.WriteGeneralLog(user, items, context)

	// load user or item features
	// can load data from datasource(holo, ots, redis)
	// after load data, use feature engine to create or modify features
	items = r.featureService.LoadFeatures(user, items, context)

	r.rankService.Rank(user, items, context)

	r.coldStartRankService.Rank(user, items, context)

	items = r.Sort(user, items, context)
	log.Info(fmt.Sprintf("requestId=%s\tmodule=pipeline\tpipeline=%s\tcount=%d\tcost=%d", context.RecommendId, r.pipelineName, len(items), utils.CostTime(start)))
	return items
}
func (s *UserRecommendService) Filter(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {
	start := time.Now()
	filterData := filter.FilterData{Data: items, Uid: user.Id, Context: context, PipelineName: s.pipelineName, User: user}

	s.filterService.Filter(&filterData)
	log.Info(fmt.Sprintf("requestId=%s\tmodule=Filter\tpipeline=%s\tcost=%d", context.RecommendId, s.pipelineName, utils.CostTime(start)))
	return filterData.Data.([]*module.Item)
}

func (s *UserRecommendService) Sort(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {
	start := time.Now()
	sortData := sort.SortData{Data: items, Context: context, User: user, PipelineName: s.pipelineName}

	s.sortService.Sort(&sortData)
	log.Info(fmt.Sprintf("requestId=%s\tmodule=Sort\tpipeline=%s\tcost=%d", context.RecommendId, s.pipelineName, utils.CostTime(start)))
	return sortData.Data.([]*module.Item)
}
