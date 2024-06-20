package recall

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

type ContextItemRecall struct {
	*BaseRecall
}

func NewContextItemRecall(config recconf.RecallConfig) *ContextItemRecall {
	recall := &ContextItemRecall{
		BaseRecall: NewBaseRecall(config),
	}
	return recall
}

func (r *ContextItemRecall) GetCandidateItems(user *module.User, context *context.RecommendContext) (ret []*module.Item) {
	start := time.Now()

	if context.GetParameter("item_list") == nil {
		return
	}
	if item_list, ok := context.GetParameter("item_list").([]map[string]any); ok {
		ret = make([]*module.Item, 0, len(item_list))
		for _, itemData := range item_list {
			itemId := itemData["item_id"]
			itemIdStr := utils.ToString(itemId, "")
			if itemIdStr == "" {
				continue
			}
			item := module.NewItem(itemIdStr)
			item.RetrieveId = r.modelName

			for k, v := range itemData {
				if k == "item_id" {
					continue
				} else if k == "score" {
					item.Score = utils.ToFloat(v, 0)
				} else {
					item.AddProperty(k, v)
				}
			}

			ret = append(ret, item)

		}
	}

	log.Info(fmt.Sprintf("requestId=%s\tmodule=ContextItemRecall\tname=%s\tcount=%d\tcost=%d", context.RecommendId, r.modelName, len(ret), utils.CostTime(start)))
	return
}
