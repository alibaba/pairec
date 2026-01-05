package recall

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/recallengine"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/recall/recallenginerecall"
	"github.com/alibaba/pairec/v2/utils"
)

const (
	reScoreFieldName  = "score"
	reItemIdFieldName = "item_id"
	//beMatchTypeFieldName = "match_type"
	reRecallName   = "recall_name"
	reRecallNameV2 = "__recall_name__"
)

type RecallEngineRecall struct {
	*BaseRecall
	bizRecall recallenginerecall.RecallEngineBaseRecall
}

func NewRecallEngineRecall(config recconf.RecallConfig) *RecallEngineRecall {
	client, err := recallengine.GetRecallEngineClient(config.RecallEngineConf.RecallEngineName)
	if err != nil {
		panic(err)
	}
	bizRecall := recallenginerecall.NewRecallEngineServiceRecall(client, config.RecallEngineConf, config.Name)

	if bizRecall == nil {
		panic("recall engine recall empty")
	}

	recall := &RecallEngineRecall{
		BaseRecall: NewBaseRecall(config),
		bizRecall:  bizRecall,
	}

	return recall
}

func (r *RecallEngineRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	items, err := r.bizRecall.GetItems(user, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=RecallEngineRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err.Error()))
		return
	}

	for _, item := range items {
		if item.RetrieveId == "" {
			item.RetrieveId = r.modelName
		}
	}

	ret = items

	log.Info(fmt.Sprintf("requestId=%s\tmodule=RecallEngineRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return

}
