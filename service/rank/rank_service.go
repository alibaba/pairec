package rank

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/algorithm"
	"github.com/alibaba/pairec/algorithm/eas"
	"github.com/alibaba/pairec/algorithm/response"
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
	"github.com/alibaba/pairec/utils/ast"
)

var rankService *RankService

func init() {
	rankService = NewRankService()
}

type RankService struct {
	rankInters                   map[string][]IRank
	featureConsistencyJobService *FeatureConsistencyJobService
}

func NewRankService() *RankService {
	rank := RankService{
		rankInters:                   make(map[string][]IRank, 0),
		featureConsistencyJobService: new(FeatureConsistencyJobService),
	}

	return &rank
}

func DefaultRankService() *RankService {
	return rankService
}

func RegisterRank(sceneName string, ranks ...IRank) {
	DefaultRankService().AddRanks(sceneName, ranks...)
}

func (r *RankService) AddRanks(sceneName string, ranks ...IRank) {
	var rankInters []IRank
	if preRanks, ok := r.rankInters[sceneName]; ok {
		rankInters = preRanks
	}

	for _, rank := range ranks {
		rankInters = append(rankInters, rank)
	}

	r.rankInters[sceneName] = rankInters
}

func (r *RankService) GetRanks(sceneName string, context *context.RecommendContext) (ret []IRank) {
	// first find experiment info
	if context.ExperimentResult != nil {
		rankconf := context.ExperimentResult.GetExperimentParams().Get("coldStartRankConf", "")
		if rankconf != "" {
			var rankConfig recconf.ColdStartRankConfig
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				rank := NewColdStartRank(&rankConfig)
				ret = append(ret, rank)
				// when rank is not ColdStartRank, add it
				if ranks, exist := r.rankInters[sceneName]; exist {
					for _, rank := range ranks {
						if _, ok := rank.(*ColdStartRank); !ok {
							ret = append(ret, rank)
						}
					}
				}
				return
			} else {
				context.LogError(fmt.Sprintf("Unmarshal rank config error\terr=%v\tconfig=%s", err, rankconf))
			}
		}
	}

	if ranks, ok := r.rankInters[sceneName]; ok {
		ret = ranks
	}

	return
}

type boostFunc func(score float64, user *module.User, item *module.Item, context *context.RecommendContext) float64

var boostScoreFunc boostFunc

func SetBoostFunc(bf boostFunc) {
	boostScoreFunc = bf
}

func (r *RankService) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) {
	start := time.Now()
	if context.Debug {
		data, _ := json.Marshal(user)
		fmt.Println(fmt.Sprintf("requestId=%s\tuser=%s", context.RecommendId, string(data)))
	}

	rankItems := items
	algoDataList := make([]IAlgoData, 0)
	scene := context.GetParameter("scene").(string)

	i := 0
	var customRanks []*customRank
	for _, rank := range r.GetRanks(scene, context) {
		customRank := newCustomRank(rank)
		customRanks = append(customRanks, customRank)
	}

	// find rank config
	var rankConfig recconf.RankConfig
	var rankscore string
	found := false
	if context.ExperimentResult != nil {
		rankscore = context.ExperimentResult.GetExperimentParams().GetString("rankscore", "")
		rankconf := context.ExperimentResult.GetExperimentParams().Get("rankconf", "")
		if rankconf != "" {
			d, _ := json.Marshal(rankconf)
			if err := json.Unmarshal(d, &rankConfig); err == nil {
				found = true
			}
		}
	}
	if !found {
		if rankConfigs, ok := recconf.Config.RankConf[scene]; ok {
			rankConfig = rankConfigs
		}
	}

	if rankscore != "" {
		rankConfig.RankScore = rankscore
	}

	batchCount := 100
	if rankConfig.BatchCount > 0 {
		batchCount = rankConfig.BatchCount
	}
	algoGenerator := CreateAlgoDataGenerator(rankConfig.Processor, rankConfig.ContextFeatures)

	var userFeatures map[string]interface{}

	if rankConfig.Processor == eas.Eas_Processor_EASYREC {
		userFeatures = user.MakeUserFeatures2()
	} else {
		userFeatures = user.MakeUserFeatures()
	}

	var filter bool
	for _, item := range rankItems {
		filter = false
		if len(customRanks) > 0 {
			for _, rank := range customRanks {
				if rank.rankInter.Filter(user, item, context) {
					filter = true
					if _, ok := rank.rankInter.(*ColdStartRank); ok {
						rank.appendFeature(nil, item, context)
					} else {
						rank.appendFeature(userFeatures, item, context)
					}
					break
				}
			}
		}

		if filter {
			continue
		}

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

	var rankWG sync.WaitGroup
	// invoke custom rank
	if len(customRanks) > 0 {
		for _, rank := range customRanks {
			rankWG.Add(1)
			go func(customRank *customRank) {
				defer rankWG.Done()
				customRank.rankInter.Rank(user, customRank.items, customRank.requestData, context)
			}(rank)
		}
	}

	if len(algoDataList) == 0 {
		if len(customRanks) > 0 {
			rankWG.Wait()
		}
		return
	}
	requestCh := make(chan IAlgoData, len(algoDataList))
	responseCh := make(chan IAlgoData, len(algoDataList))
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

	exprAst, err := ast.GetExpASTWithType(rankConfig.RankScore, rankConfig.ASTType)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=rank\trankscore=%s\terror=%v", context.RecommendId, rankConfig.RankScore, err))
	}

	var scoreRewriteAst map[string]ast.ExprAST
	if len(rankConfig.ScoreRewrite) > 0 {
		scoreRewriteAst = make(map[string]ast.ExprAST, len(rankConfig.ScoreRewrite))
		for source, sourceExpr := range rankConfig.ScoreRewrite {
			if strings.Contains(sourceExpr, source) {
				ast, err := ast.GetExpASTByAntlrWithStatement(sourceExpr, source)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=rank\tscorerewrite=%s\terror=%v", context.RecommendId, rankConfig.RankScore, err))
					continue
				}
				scoreRewriteAst[source] = ast
			} else {
				ast, err := ast.GetExpASTWithType(sourceExpr, rankConfig.ASTType)
				if err != nil {
					log.Error(fmt.Sprintf("requestId=%s\tmodule=rank\tscorerewrite=%s\terror=%v", context.RecommendId, rankConfig.RankScore, err))
					continue
				}
				scoreRewriteAst[source] = ast
			}
		}
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
				// score rewrite 重写
				if len(rankConfig.ScoreRewrite) > 0 {
					scores := make(map[string]float64, len(rankConfig.ScoreRewrite))
					for source, sourceExpr := range rankConfig.ScoreRewrite {
						if strings.Contains(sourceExpr, source) {
							if exprAst, ok := scoreRewriteAst[source]; ok {
								scores[source] = ast.ExprASTResultByAntlrWithStatement(exprAst, itemList[k])
							} else {
								scores[source] = 0
							}

						} else {
							if exprAst, ok := scoreRewriteAst[source]; ok {
								scores[source] = ast.ExprASTResultWithType(exprAst, itemList[k], rankConfig.ASTType)
							} else {
								scores[source] = 0
							}
						}

					}
					itemList[k].AddAlgoScores(scores)
				}

				if exprAst != nil {
					itemList[k].Score = ast.ExprASTResultWithType(exprAst, itemList[k], rankConfig.ASTType)
				}

				if boostScoreFunc != nil {
					itemList[k].Score = boostScoreFunc(itemList[k].Score, user, itemList[k], context)
				}
			}
		}
	}

	if len(customRanks) > 0 {
		rankWG.Wait()
	}

	go r.featureConsistencyJobService.LogRankResult(user, items, context)
	log.Info(fmt.Sprintf("requestId=%s\tmodule=rank\tcost=%d", context.RecommendId, utils.CostTime(start)))
}
