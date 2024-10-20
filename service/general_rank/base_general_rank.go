package general_rank

import (
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/debug"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/alibaba/pairec/v2/utils/ast"
)

type BaseGeneralRank struct {
	scene                        string
	featureService               *feature.FeatureService
	rankConfig                   *recconf.RankConfig
	actions                      []*Action
	featureConsistencyJobService *FeatureConsistencyJobService
}

func NewBaseGeneralRankWithConfig(scene string, config recconf.GeneralRankConfig) *BaseGeneralRank {
	preRank := BaseGeneralRank{
		featureService:               feature.NewFeatureService(),
		rankConfig:                   &config.RankConf,
		scene:                        scene,
		featureConsistencyJobService: new(FeatureConsistencyJobService),
	}

	var features []*feature.Feature
	for _, conf := range config.FeatureLoadConfs {
		f := feature.LoadWithConfig(conf)
		features = append(features, f)
	}

	preRank.featureService.SetFeatureSceneAsync(scene, true)
	preRank.featureService.SetFeatures(scene, features)

	for _, conf := range config.ActionConfs {
		if action, err := NewAction(&conf); err == nil {
			preRank.actions = append(preRank.actions, action)
		} else {
			log.Error(fmt.Sprintf("create action error:%v", err))
		}
	}

	if preRank.rankConfig.BatchCount <= 0 {
		preRank.rankConfig.BatchCount = 100
	}
	return &preRank
}

// Rank GeneralRank for items , return the items for Rank
// 1. first load user features
// 2. construct data use AlgoDataGenerator, if processor is easyrec, create easyrecRequest
// 3. use goroutines to invoke eas module
// 4. iterator actions invoke Do function to apply items
func (r *BaseGeneralRank) DoRank(user *module.User, items []*module.Item, context *context.RecommendContext, pipeline string) (ret []*module.Item) {
	debugService := debug.NewDebugService(user, context)
	rankItems := items

	// get user feature by the FeatureService
	r.featureService.LoadFeaturesForGeneralRank(user, rankItems, context, pipeline)
	if context.Debug {
		data, _ := json.Marshal(user)
		size := len(data)
		for i := 0; i < size; {
			end := i + 4096
			if end >= size {
				end = size
			} else {
				for end > i {
					if data[end] == ',' {
						end++
						break
					}
					end--
				}

				if end == i {
					end = i + 4096
				}
			}
			log.Info(fmt.Sprintf("requestId=%s\tmodule=general_rank\tuser=%s", context.RecommendId, string(data[i:end])))
			i = end
		}
	}

	if len(r.rankConfig.RankAlgoList) > 0 {
		rankItems = r.doRankWithAlgo(user, rankItems, context)
	}

	debugService.WriteGeneralLog(user, rankItems, context)

	for _, action := range r.actions {
		rankItems = action.Do(user, rankItems, context)
	}

	ret = append(ret, rankItems...)
	return
}

func (r *BaseGeneralRank) doRankWithAlgo(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {
	start := time.Now()
	algoDataList := make([]rank.IAlgoData, 0)
	i := 0
	algoGenerator := rank.CreateAlgoDataGenerator(r.rankConfig.Processor, r.rankConfig.ContextFeatures)

	userFeatures := user.MakeUserFeatures2()

	//emptyFeatures := make(map[string]interface{}, 0)
	/**
	go func() {
		for _, item := range items {
			item.AddRecallNameFeature()
		}
	}()
	**/
	for _, item := range items {
		features := item.GetFeatures()
		if r.rankConfig.Processor == eas.Eas_Processor_EASYREC {
			algoGenerator.AddFeatures(item, features, userFeatures)
		} else {
			algoGenerator.AddFeatures(item, features, userFeatures)
		}

		i++
		if i%r.rankConfig.BatchCount == 0 {
			algoData := algoGenerator.GeneratorAlgoData()
			algoDataList = append(algoDataList, algoData)
		}
	}

	if algoGenerator.HasFeatures() {
		algoData := algoGenerator.GeneratorAlgoData()
		algoDataList = append(algoDataList, algoData)
	}

	if len(algoDataList) == 0 {
		return items
	}

	requestCh := make(chan rank.IAlgoData, len(algoDataList))
	responseCh := make(chan rank.IAlgoData, len(algoDataList))
	defer close(requestCh)
	defer close(responseCh)

	for _, data := range algoDataList {
		requestCh <- data
	}

	rankConfig := r.rankConfig

	gCount := len(algoDataList)
	for i := 0; i < gCount; i++ {
		go func() {
			algoData := <-requestCh

			var wg sync.WaitGroup
			for _, algoName := range rankConfig.RankAlgoList {
				wg.Add(1)
				go func(algo string) {
					defer wg.Done()
					algoResponses, err := algorithm.Run(algo, algoData.GetFeatures())
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\terror=run algorithm error(%v)", context.RecommendId, err))
						algoData.SetError(err)
					} else {
						if result, ok := algoResponses.([]response.AlgoResponse); ok {
							algoData.SetAlgoResult(algo, result)
						}
					}
				}(algoName)
			}
			wg.Wait()
			responseCh <- algoData
		}()
	}

	exprAst, err := ast.GetExpAST(rankConfig.RankScore)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\trankscore=%s\terror=%v", context.RecommendId, rankConfig.RankScore, err))
	}
	for i := 0; i < gCount; i++ {
		algoData := <-responseCh
		if algoData.Error() == nil && algoData.GetAlgoResult() != nil {
			for name, algoResult := range algoData.GetAlgoResult() {
				itemList := algoData.GetItems()
				for j := 0; j < len(algoResult) && j < len(itemList); j++ {
					if algoResult[j].GetModuleType() {
						scoreMap := algoResult[j].GetScoreMap()
						for k, v := range scoreMap {
							itemList[j].AddAlgoScore(name+"_"+k, v)
						}
					} else {
						itemList[j].AddAlgoScore(name, algoResult[j].GetScore())
					}
				}
			}
		}

		if rankConfig.RankScore != "" {
			itemList := algoData.GetItems()
			for k := range itemList {
				if exprAst != nil {
					itemList[k].Score = ast.ExprASTResult(exprAst, itemList[k])
				}
			}
		}
	}

	if context.Debug && len(items) > 0 {
		fmt.Println("general rank", items[0])
	}

	go r.featureConsistencyJobService.LogRankResult(user, items, context)
	log.Info(fmt.Sprintf("requestId=%s\tmodule=GeneralRankWithAlgo\tcount=%d\tcost=%d", context.RecommendId, len(items), utils.CostTime(start)))
	return items
}
