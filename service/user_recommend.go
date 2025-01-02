package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/log/feature_log"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/debug"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/general_rank"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/service/metrics"
	"github.com/alibaba/pairec/v2/service/pipeline"
	"github.com/alibaba/pairec/v2/service/rank"
)

type UserRecommendService struct {
	RecommendService
	recallService      *RecallService
	generalRankService *general_rank.GeneralRankService
	rankService        *rank.RankService
	userFeatureService *feature.UserFeatureService
	featureService     *feature.FeatureService
}

func NewUserRecommendService() *UserRecommendService {
	service := UserRecommendService{
		recallService:      &RecallService{},
		rankService:        rank.DefaultRankService(),
		userFeatureService: feature.DefaultUserFeatureService(),
		featureService:     feature.DefaultFeatureService(),
		generalRankService: general_rank.DefaultGeneralRankService(),
	}
	return &service
}

func (r *UserRecommendService) Recommend(context *context.RecommendContext) []*module.Item {
	start := time.Now()

	var scene, expId string
	if context.ExperimentResult != nil {
		scene = context.ExperimentResult.SceneName
		//expId = context.ExperimentResult.ExpId
	} else {
		scene, _ = context.Param.GetParameter("scene").(string)
	}

	userId := r.GetUID(context)
	user := module.NewUserWithContext(userId, context)

	//loadFeatureStart := time.Now()

	// load user features
	r.userFeatureService.LoadUserFeatures(user, context)

	//if metrics.Enabled() {
	//metrics.LoadFeatureDurSecs.WithLabelValues(scene, expId, "before_recall").Observe(time.Since(loadFeatureStart).Seconds())
	//}

	debugService := debug.NewDebugService(user, context)

	var pipelineItems []*module.Item

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipelineItems = pipeline.Recommend(user, context, debugService)
	}()

	recallStart := time.Now()

	items := r.recallService.GetItems(user, context)

	if metrics.Enabled() {
		metrics.RecallDurSecs.WithLabelValues(scene, expId).Observe(time.Since(recallStart).Seconds())

		recallCountMap := map[string]int{}
		for _, item := range items {
			recallCountMap[item.RetrieveId]++
		}

		for src, count := range recallCountMap {
			metrics.RecallItemsPercentage.WithLabelValues(src).Set(float64(count) / float64(len(items)))
		}
	}

	debugService.WriteRecallLog(user, items, context)

	filterStart := time.Now()

	// filter
	items = r.Filter(user, items, context)

	if metrics.Enabled() {
		metrics.FilterDurSecs.WithLabelValues(scene, expId).Observe(time.Since(filterStart).Seconds())
	}

	debugService.WriteFilterLog(user, items, context)

	generalRankStart := time.Now()

	// general rank
	items = r.generalRankService.Rank(user, items, context)

	if metrics.Enabled() {
		metrics.GeneralRankDurSecs.WithLabelValues(scene, expId).Observe(time.Since(generalRankStart).Seconds())
	}

	//debugService.WriteGeneralLog(user, items, context)

	//loadFeatureStart = time.Now()

	// load user or item features
	// can load data from datasource(holo, ots, redis)
	// after load data, use feature engine to create or modify features
	items = r.featureService.LoadFeatures(user, items, context)

	//if metrics.Enabled() {
	//metrics.LoadFeatureDurSecs.WithLabelValues(scene, expId, "before_rank").Observe(time.Since(loadFeatureStart).Seconds())
	//}

	rankStart := time.Now()

	r.rankService.Rank(user, items, context)

	if metrics.Enabled() {
		metrics.RankDurSecs.WithLabelValues(scene, expId).Observe(time.Since(rankStart).Seconds())
	}

	wg.Wait()
	items = r.mergePipelineItems(items, pipelineItems)

	debugService.WriteRankLog(user, items, context)

	sortStart := time.Now()

	// sort items
	items = r.Sort(user, items, context)

	if metrics.Enabled() {
		metrics.SortDurSecs.WithLabelValues(scene, expId).Observe(time.Since(sortStart).Seconds())
	}
	debugService.WriteSortLog(user, items, context)

	size := context.Size
	if size > len(items) {
		size = len(items)
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=recommend\tevent=filter\tuid=%s\tmsg=length of items less than size\tcount=%d", context.RecommendId, userId, len(items)))

		if metrics.Enabled() {
			metrics.SizeNotEnoughTotal.WithLabelValues(scene, expId).Inc()
		}
	}

	items = items[:size]
	go feature_log.FeatureLog(user, items, context)
	debugService.WriteRecommendLog(user, items, context)
	// asynchronous clean hook func
	for _, hf := range hook.RecommendCleanHooks {
		go hf(context, user, items)
	}

	if metrics.Enabled() {
		metrics.RecTotal.WithLabelValues(scene, expId).Inc()
		metrics.RecDurSecs.WithLabelValues(scene, expId).Observe(time.Since(start).Seconds())
	}

	return items
}

func (r *UserRecommendService) mergePipelineItems(items []*module.Item, pipelineItems []*module.Item) []*module.Item {
	itemMap := make(map[module.ItemId]bool, len(items))

	for _, item := range items {
		itemMap[item.Id] = true
	}

	for _, item := range pipelineItems {
		if _, ok := itemMap[item.Id]; !ok {
			itemMap[item.Id] = true
			items = append(items, item)
		}
	}

	return items
}
