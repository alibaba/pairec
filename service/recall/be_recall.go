package recall

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/datasource/beengine"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

const (
	beScoreFieldName     = "__score__"
	beMatchTypeFieldName = "match_type"
	beRecallName         = "recall_name"
	beRecallNameV2       = "__recall_name__"
)

type BeBaseRecall interface {
	GetItems(user *module.User, context *context.RecommendContext) ([]*module.Item, error)
	//BuildRecallParam(user *module.User, context *context.RecommendContext) *be.RecallParam
	BuildQueryParams(user *module.User, context *context.RecommendContext) (ret map[string]string)
	CloneWithConfig(params map[string]interface{}) BeBaseRecall
}

type BeRecall struct {
	*BaseRecall
	bizRecall BeBaseRecall
}

func NewBeRecall(config recconf.RecallConfig) *BeRecall {
	client, err := beengine.GetBeClient(config.BeConf.BeName)
	if err != nil {
		panic(err)
	}
	var bizRecall BeBaseRecall
	switch config.BeConf.BeRecallType {
	case recconf.BE_RecallType_X2I:
		bizRecall = NewBeX2IRecall(client, config.BeConf)
	case recconf.BE_RecallType_Vector:
		bizRecall = NewBeVectorRecall(client, config.BeConf)
	case recconf.BE_RecallType_MultiMerge:
		bizRecall = NewBeMultiBizRecall(client, config.BeConf, config.Name)
	}

	if bizRecall == nil {
		panic("be biz recall empty")
	}

	recall := &BeRecall{
		BaseRecall: NewBaseRecall(config),
		bizRecall:  bizRecall,
	}

	return recall
}

func (r *BeRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()
	items, err := r.bizRecall.GetItems(user, context)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=BeRecall\tname=%s\terr=%v", context.RecommendId, r.modelName, err.Error()))
		return
	}

	for _, item := range items {
		if item.RetrieveId == "" {
			item.RetrieveId = r.modelName
		}
	}

	ret = items

	log.Info(fmt.Sprintf("requestId=%s\tmodule=BeRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return

}
