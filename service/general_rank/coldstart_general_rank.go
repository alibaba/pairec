package general_rank

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type ColdStartGeneralRank struct {
	*BaseGeneralRank
	recallNames map[string]bool
}

func NewColdStartGeneralRankWithConfig(scene string, config recconf.ColdStartGeneralRankConfig) (*ColdStartGeneralRank, error) {

	preRank := ColdStartGeneralRank{
		BaseGeneralRank: NewBaseGeneralRankWithConfig(scene, config.GeneralRankConfig),
		recallNames:     make(map[string]bool, 0),
	}

	for _, name := range config.RecallNames {
		preRank.recallNames[name] = true
	}
	return &preRank, nil
}

func (r *ColdStartGeneralRank) Rank(user *module.User, items []*module.Item, context *context.RecommendContext) (ret []*module.Item) {

	ret = r.DoRank(user, items, context, "")
	log.Info(fmt.Sprintf("requestId=%s\tmodule=ColdStartGeneralRank\tcount=%d", context.RecommendId, len(ret)))
	return
}

func (r *ColdStartGeneralRank) Filter(User *module.User, item *module.Item, context *context.RecommendContext) bool {
	if _, ok := r.recallNames[item.GetRecallName()]; ok {
		return true
	}

	return false

}
