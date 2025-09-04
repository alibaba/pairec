package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/log/feature_log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/debug"
	"github.com/alibaba/pairec/v2/service/fallback"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/general_rank"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/service/metrics"
	"github.com/alibaba/pairec/v2/service/pipeline"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
)

type UserRecommendService struct {
	RecommendService
	recallService                *RecallService
	generalRankService           *general_rank.GeneralRankService
	rankService                  *rank.RankService
	userFeatureService           *feature.UserFeatureService
	featureService               *feature.FeatureService
	featureConsistencyJobService *rank.FeatureConsistencyJobService
}

func NewUserRecommendService() *UserRecommendService {
	service := UserRecommendService{
		recallService:                &RecallService{},
		rankService:                  rank.DefaultRankService(),
		userFeatureService:           feature.DefaultUserFeatureService(),
		featureService:               feature.DefaultFeatureService(),
		featureConsistencyJobService: new(rank.FeatureConsistencyJobService),
		generalRankService:           general_rank.DefaultGeneralRankService(),
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
		metrics.RecallCountTotal.WithLabelValues(scene).Add(float64(len(items)))

		recallCountMap := map[string]int{}
		for _, item := range items {
			recallCountMap[item.RetrieveId]++
		}

		for src, count := range recallCountMap {
			//metrics.RecallItemsPercentage.WithLabelValues(src).Set(float64(count) / float64(len(items)))
			metrics.RecallCount.WithLabelValues(scene, src).Add(float64(count))
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
	go r.featureConsistencyJobService.LogSampleResult(user, items, context)
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
	itemMap := make(map[module.ItemId]*module.Item, len(items))
	for _, item := range items {
		itemMap[item.Id] = item
	}
	// need to merge item properties of different pipelines
	for _, item := range pipelineItems {
		if exist, ok := itemMap[item.Id]; ok {
			exist.AddProperties(item.Properties)
		} else {
			itemMap[item.Id] = item
			items = append(items, item)
		}
	}
	return items
}

func (r *UserRecommendService) TryRecommendWithFallback(context *context.RecommendContext) []*module.Item {
	start := time.Now()
	scene, _ := context.Param.GetParameter("scene").(string)

	f := fallback.DefaultFallbackService().GetFallback(scene)
	if f == nil {
		return r.Recommend(context)
	}

	fallbackTimer := f.GetTimer()

	tryResult := make(chan []*module.Item, 1)

	go func() {
		tryResult <- r.Recommend(context)
	}()

	select {
	case <-fallbackTimer.C:
		f.PutTimer(fallbackTimer)
		fallbackResult := f.Recommend(context)
		log.Warning(fmt.Sprintf("requestId=%s\tmodule=recommend\tevent=fallback\tcause=timeout\tcost=%d", context.RecommendId, utils.CostTime(start)))
		return fallbackResult
	case ret := <-tryResult:
		f.PutTimer(fallbackTimer)

		if f.CompleteItemsIfNeed() && len(ret) < context.Size {
			originRetMap := make(map[module.ItemId]bool)
			for _, item := range ret {
				originRetMap[item.Id] = true
			}

			fallbackResult := f.Recommend(context)
			for i := 0; i < len(fallbackResult); i++ {
				fallbackItem := fallbackResult[i]

				if !originRetMap[fallbackItem.Id] { // must not appear in origin result
					ret = append(ret, fallbackItem)

					if len(ret) == context.Size {
						break
					}
				}
			}

			log.Warning(fmt.Sprintf("requestId=%s\tmodule=recommend\tevent=fallback\tcause=itemNotEnough\tcost=%d", context.RecommendId, utils.CostTime(start)))
			return ret
		} else {
			return ret
		}
	}
}
