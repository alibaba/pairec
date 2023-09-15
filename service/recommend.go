package service

import (
	"fmt"
	"time"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/filter"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/sort"
	"github.com/alibaba/pairec/utils"
)

type RecommendService struct {
}

func (s *RecommendService) GetUID(context *context.RecommendContext) module.UID {
	uid := context.GetParameter("uid")
	if uid == nil {
		uid = ""
	}

	userId := module.UID(uid.(string))
	return userId
}

func (s *RecommendService) Filter(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {
	start := time.Now()
	filterData := filter.FilterData{Data: items, Uid: user.Id, Context: context, User: user}

	filter.Filter(&filterData, "")
	log.Info(fmt.Sprintf("requestId=%s\tmodule=Filter\tcost=%d", context.RecommendId, utils.CostTime(start)))
	return filterData.Data.([]*module.Item)
}

func (s *RecommendService) Sort(user *module.User, items []*module.Item, context *context.RecommendContext) []*module.Item {
	sortData := sort.SortData{Data: items, Context: context, User: user}

	sort.Sort(&sortData, "")
	return sortData.Data.([]*module.Item)
}

func (s *RecommendService) PreSort(items []*module.Item, context *context.RecommendContext) []*module.Item {
	sortData := sort.SortData{Data: items, Context: context}

	sort.Sort(&sortData, "_PreSort")
	return sortData.Data.([]*module.Item)
}
