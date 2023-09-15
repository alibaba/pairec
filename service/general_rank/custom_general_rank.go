package general_rank

import (
	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
)

type IGeneralRank interface {
	// Filter the custom rank of item
	Filter(User *module.User, item *module.Item, context *context.RecommendContext) bool

	Rank(User *module.User, items []*module.Item, context *context.RecommendContext)
}
