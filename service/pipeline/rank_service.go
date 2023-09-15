package pipeline

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alibaba/pairec/algorithm"
	"github.com/alibaba/pairec/algorithm/eas"
	"github.com/alibaba/pairec/algorithm/response"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/service/rank"
	"github.com/alibaba/pairec/utils"
	"github.com/alibaba/pairec/utils/ast"
)

type RankService struct {
	pipelineName string
	rankConfig   recconf.RankConfig
}

func NewRankService(config *recconf.PipelineConfig) *RankService {
	rank := RankService{
		pipelineName: config.Name,
		rankConfig:   config.RankConf,
	}

	return &rank
}

func (r *RankService) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) {
	start := time.Now()
	if context.Debug {
		fmt.Println(r.pipelineName, user)
	}

	rankItems := items
	algoDataList := make([]rank.IAlgoData, 0)

	i := 0

	// find rank config
	var rankConfig recconf.RankConfig
	found := false
	if context.ExperimentResult != nil {
		rankconf := context.ExperimentResult.GetExperimentParams().Get("pipelines."+r.pipelineName+".RankConf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}
	if !found {
		rankConfig = r.rankConfig
	}

	if len(rankConfig.RankAlgoList) == 0 {
		return
	}

	batchCount := 100
	if rankConfig.BatchCount > 0 {
		batchCount = rankConfig.BatchCount
	}
	algoGenerator := rank.CreateAlgoDataGenerator(rankConfig.Processor, rankConfig.ContextFeatures)

	var userFeatures map[string]interface{}

	if rankConfig.Processor == eas.Eas_Processor_EASYREC {
		userFeatures = user.MakeUserFeatures2()
	} else {
		userFeatures = user.MakeUserFeatures()
	}

	for _, item := range rankItems {
		features := item.GetFeatures()
		algoGenerator.AddFeatures(item, features, userFeatures)
		i++
		if i%batchCount == 0 {
			algoData := algoGenerator.GeneratorAlgoData()
			algoDataList = append(algoDataList, algoData)
		}
	}

	if algoGenerator.HasFeatures() {
		algoData := algoGenerator.GeneratorAlgoData()
		algoDataList = append(algoDataList, algoData)
	}

	if len(algoDataList) == 0 {
		return
	}

	requestCh := make(chan rank.IAlgoData, len(algoDataList))
	responseCh := make(chan rank.IAlgoData, len(algoDataList))
	defer close(requestCh)
	defer close(responseCh)

	for _, data := range algoDataList {
		requestCh <- data
	}

	gCount := len(algoDataList)
	for i := 0; i < gCount; i++ {
		go func() {
			algoData := <-requestCh

			var wg sync.WaitGroup
			for _, algoName := range rankConfig.RankAlgoList {
				wg.Add(1)
				go func(algo string) {
					defer wg.Done()
					ret, err := algorithm.Run(algo, algoData.GetFeatures())
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\terror=run algorithm error(%v)", context.RecommendId, err))
						algoData.SetError(err)
					} else {
						if result, ok := ret.([]response.AlgoResponse); ok {
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
		log.Error(fmt.Sprintf("requestId=%s\tmodule=rank\tpipeline=%s\trankscore=%s\terror=%v", context.RecommendId, r.pipelineName, rankConfig.RankScore, err))
	}
	for i := 0; i < gCount; i++ {
		algoData := <-responseCh
		if algoData.Error() == nil && algoData.GetAlgoResult() != nil {
			for name, algoResult := range algoData.GetAlgoResult() {
				itemList := algoData.GetItems()
				for j := 0; j < len(algoResult) && j < len(itemList); j++ {
					if algoResult[j].GetModuleType() {
						arr_score := algoResult[j].GetScoreMap()
						for k, v := range arr_score {
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

	log.Info(fmt.Sprintf("requestId=%s\tmodule=rank\tpipeline=%s\tcost=%d", context.RecommendId, r.pipelineName, utils.CostTime(start)))
}
