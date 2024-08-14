package service

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/service/rank"
)

type UserRecallService struct {
	RecommendService
	recallService *RecallService
	rankService   *rank.RankService
}

func NewUserRecallService() *UserRecallService {
	service := UserRecallService{
		recallService: &RecallService{},
	}
	return &service
}
func (r *UserRecallService) Recommend(context *context.RecommendContext) []*module.Item {
	userId := r.GetUID(context)
	user := module.NewUserWithContext(userId, context)
	items := r.recallService.GetItems(user, context)

	// filter
	items = r.Filter(user, items, context)

	size := context.Size

	if len(items) < size {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=recommend\tevent=filter\tmsg=length of items less than size\tcount=%d", context.RecommendId, len(items)))
		return items
	}

	items = items[:size]

	// asynchronous clean hook func
	for _, hf := range hook.RecommendCleanHooks {
		go hf(context, user, items)
	}
	return items
}
