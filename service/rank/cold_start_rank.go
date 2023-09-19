package rank

import (
	"fmt"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/eas"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type ColdStartRank struct {
	RecallName           string
	AlgoName             string
	OnlyEmbeddingFeature bool
}

func NewColdStartRank(conf *recconf.ColdStartRankConfig) *ColdStartRank {
	return &ColdStartRank{
		RecallName:           conf.RecallName,
		AlgoName:             conf.AlgoName,
		OnlyEmbeddingFeature: conf.OnlyEmbeddingFeature,
	}
}

func (r *ColdStartRank) Filter(User *module.User, item *module.Item, context *context.RecommendContext) bool {
	return item.GetRecallName() == r.RecallName
}

func (r *ColdStartRank) Rank(user *module.User, items []*module.Item, requestData []map[string]interface{}, context *context.RecommendContext) {
	// if algo name not set, no need rank
	if r.AlgoName == "" {
		return
	}
	if len(items) == 0 {
		return
	}
	var itemIds []string
	for _, item := range items {
		itemIds = append(itemIds, string(item.Id))
	}

	limit := context.Size
	if limit > len(items) {
		limit = len(items)
	}

	var userFeatures map[string]interface{}
	if r.OnlyEmbeddingFeature {
		userFeatures = user.GetEmbeddingFeature()
	} else {
		userFeatures = user.MakeUserFeatures2()
	}
	request := eas.LincubRequestData{
		RequestId:    context.RecommendId,
		Scene:        context.GetParameter("scene").(string),
		AlgoName:     r.AlgoName,
		UserId:       string(user.Id),
		UserFeature:  userFeatures,
		Items:        itemIds,
		ItemFeatures: requestData,
		Limit:        limit,
	}

	ret, err := algorithm.Run(r.AlgoName, &request)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=ColdStartRank\terr=%v", context.RecommendId, err))
		return
	}

	results, err := eas.LinucbResponseFunc(ret)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tevent=ColdStartRank\terr=%v", context.RecommendId, err))
		return
	}

	for _, linucbItem := range results {
		for _, item := range items {
			if linucbItem.ItemId == string(item.Id) {
				item.Score = linucbItem.GetScore()
			}
		}
	}

}

func LoadColdStartRankConfig(config *recconf.RecommendConfig) {
	for scene, conf := range config.ColdStartRankConfs {
		rank := NewColdStartRank(&conf)
		RegisterRank(scene, rank)
	}
}
