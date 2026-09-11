package aishopping

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/algorithm/response"
	paireccontext "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/feature"
	"github.com/alibaba/pairec/v2/service/rank"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/alibaba/pairec/v2/utils/ast"
)

type fineRankRuntime struct {
	prepared             bool
	prepareErr           error
	recommendContext     *paireccontext.RecommendContext
	user                 *module.User
	loadUserFeatures     bool
	loadRankUserFeatures bool
}

func newFineRankRuntime(req *Request) *fineRankRuntime {
	recommendContext := paireccontext.NewRecommendContext()
	recommendContext.Param = req
	recommendContext.Config = req.Config
	recommendContext.RecommendId = req.RequestId

	runtime := &fineRankRuntime{
		recommendContext: recommendContext,
		user:             module.NewUserWithContext(module.UID(req.Uid), recommendContext),
	}
	if req.Config == nil {
		return runtime
	}
	if scene, ok := req.Config.UserFeatureConfs[req.SceneId]; ok {
		runtime.loadUserFeatures = len(scene.FeatureLoadConfs) > 0
	}
	if scene, ok := req.Config.FeatureConfs[req.SceneId]; ok {
		for _, conf := range scene.FeatureLoadConfs {
			if conf.FeatureDaoConf.FeatureStore == module.Feature_Store_User {
				runtime.loadRankUserFeatures = true
				break
			}
		}
	}
	return runtime
}

func (r *fineRankRuntime) prepareUser(ctx context.Context) (*module.User, error) {
	if r == nil || r.user == nil || r.recommendContext == nil {
		return nil, fmt.Errorf("fine rank runtime is nil")
	}
	if r.prepared {
		return r.user, r.prepareErr
	}
	r.prepared = true
	if r.loadUserFeatures {
		feature.DefaultUserFeatureService().LoadUserFeatures(r.user, r.recommendContext)
		if r.user.FeatureAsyncLoadCount() > 0 {
			select {
			case <-r.user.FeatureAsyncLoadCh():
			case <-ctx.Done():
				r.prepareErr = ctx.Err()
			}
		}
	}
	if r.prepareErr == nil && r.loadRankUserFeatures {
		feature.DefaultFeatureService().LoadFeatures(r.user, nil, r.recommendContext)
	}
	return r.user, r.prepareErr
}

func fineRankGoods(ctx context.Context, hits []recallsvc.GoodsHit, runtime *fineRankRuntime, conf *recconf.AIShoppingFineRankConfig) ([]recallsvc.GoodsHit, error) {
	if conf == nil || len(hits) <= 1 {
		return append([]recallsvc.GoodsHit(nil), hits...), nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	rankConf := conf.RankConf
	algoName := rankConf.RankAlgoList[0]
	generator := rank.CreateAlgoDataGenerator(rankConf.Processor, rankConf.ContextFeatures)
	if rankConf.Processor == eas.Eas_Processor_EASYREC {
		generator.SetItemFeatures(rankConf.ItemFeatures)
	}

	user, err := runtime.prepareUser(ctx)
	if err != nil {
		return nil, err
	}
	userFeatures := user.MakeUserFeatures()
	if rankConf.Processor == eas.Eas_Processor_EASYREC {
		userFeatures = user.MakeUserFeatures2()
	}

	items := make([]*module.Item, 0, len(hits))
	for i, hit := range hits {
		item := fineRankItem(hit, i)
		items = append(items, item)
		generator.AddFeatures(item, item.GetFeatures(), userFeatures)
	}
	if !generator.HasFeatures() {
		return nil, fmt.Errorf("fine rank request has no features")
	}

	algoData := generator.GeneratorAlgoData()
	result, err := algorithm.Run(algoName, algoData.GetFeatures())
	if err != nil {
		return nil, fmt.Errorf("run fine rank algorithm %s: %w", algoName, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	responses, ok := result.([]response.AlgoResponse)
	if !ok {
		return nil, fmt.Errorf("fine rank algorithm %s returned %T", algoName, result)
	}
	if len(responses) != len(items) {
		return nil, fmt.Errorf("fine rank result count mismatch: got %d, want %d", len(responses), len(items))
	}

	expr, err := ast.GetExpASTWithType(rankConf.RankScore, rankConf.ASTType)
	if err != nil {
		return nil, fmt.Errorf("parse fine rank score: %w", err)
	}
	type scoredHit struct {
		hit   recallsvc.GoodsHit
		score float64
	}
	scored := make([]scoredHit, len(hits))
	for i, algoResponse := range responses {
		if err := addFineRankResponse(items[i], algoName, algoResponse); err != nil {
			return nil, err
		}
		score := ast.ExprASTResultWithType(expr, items[i], rankConf.ASTType)
		if math.IsNaN(score) || math.IsInf(score, 0) {
			return nil, fmt.Errorf("fine rank returned non-finite score for item %s", hits[i].ItemId)
		}
		scored[i] = scoredHit{hit: hits[i], score: score}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	ranked := make([]recallsvc.GoodsHit, len(scored))
	for i, item := range scored {
		ranked[i] = item.hit
	}
	return ranked, nil
}

func fineRankItem(hit recallsvc.GoodsHit, position int) *module.Item {
	properties := make(map[string]interface{}, len(hit.Properties)+5)
	for key, value := range hit.Properties {
		properties[key] = value
	}
	properties["item_id"] = hit.ItemId
	properties["title"] = hit.Title
	properties["content"] = hit.Content
	properties["aishopping_recall_position"] = position + 1
	if hit.Score != nil {
		properties["aishopping_recall_score"] = utils.ToFloat(hit.Score, 0)
	}
	item := module.NewItemWithProperty(hit.ItemId, properties)
	item.Score = utils.ToFloat(hit.Score, 0)
	return item
}

func addFineRankResponse(item *module.Item, algoName string, algoResponse response.AlgoResponse) error {
	if algoResponse == nil {
		return fmt.Errorf("fine rank algorithm %s returned nil response", algoName)
	}
	if algoResponse.GetModuleType() {
		scores := algoResponse.GetScoreMap()
		if len(scores) == 0 {
			return fmt.Errorf("fine rank algorithm %s returned empty score map", algoName)
		}
		for name, score := range scores {
			if math.IsNaN(score) || math.IsInf(score, 0) {
				return fmt.Errorf("fine rank algorithm %s returned non-finite score", algoName)
			}
			item.AddAlgoScore(algoName+"_"+name, score)
		}
		return nil
	}

	score := algoResponse.GetScore()
	if math.IsNaN(score) || math.IsInf(score, 0) {
		return fmt.Errorf("fine rank algorithm %s returned non-finite score", algoName)
	}
	item.AddAlgoScore(algoName, score)
	return nil
}
