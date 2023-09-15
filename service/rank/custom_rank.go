package rank

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
)

type IRank interface {
	// Filter the custom rank of item
	Filter(User *module.User, item *module.Item, context *context.RecommendContext) bool

	Rank(User *module.User, items []*module.Item, requestData []map[string]interface{}, context *context.RecommendContext)
}

// customRank wapper the IRank interface, contains the items and requestData for IRank
// 1. Invoke the IRank filter, append the filter item and features
// 2. Invoke the IRank Rank function
type customRank struct {
	rankInter   IRank
	items       []*module.Item
	requestData []map[string]interface{}
}

func newCustomRank(rank IRank) *customRank {
	r := customRank{
		rankInter: rank,
	}

	return &r
}
func (r *customRank) appendFeature(userFeatures map[string]interface{}, item *module.Item, context *context.RecommendContext) {
	features := item.GetFeatures()
	if userFeatures != nil {
		for k, v := range userFeatures {
			features[k] = v
		}
	}

	r.requestData = append(r.requestData, features)
	r.items = append(r.items, item)
}
