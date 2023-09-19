package general_rank

import (
	"fmt"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/filter"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/sort"
)

type ActionType int

const (
	UnknownAction ActionType = 0
	SortAction    ActionType = 1
	FilterAction  ActionType = 2
)

func CreateActionType(str string) ActionType {
	switch str {
	case "sort":
		return SortAction
	case "filter":
		return FilterAction
	default:
		return UnknownAction
	}
}

type Action struct {
	sort   sort.ISort
	filter filter.IFilter

	actionType ActionType
}

func NewAction(config *recconf.ActionConfig) (*Action, error) {
	actionType := CreateActionType(config.ActionType)
	if actionType == UnknownAction {
		return nil, fmt.Errorf("error to find actionType:%s", config.ActionType)
	}
	action := &Action{actionType: actionType}
	if actionType == SortAction {
		sort, err := sort.GetSort(config.ActionName)
		if err != nil {
			return nil, err
		}
		action.sort = sort
	} else if actionType == FilterAction {
		filter, err := filter.GetFilter(config.ActionName)
		if err != nil {
			return nil, err
		}
		action.filter = filter
	}

	return action, nil
}
func (a *Action) Do(user *module.User, items []*module.Item, context *context.RecommendContext) (ret []*module.Item) {
	switch a.actionType {
	case SortAction:
		if a.sort != nil {
			sortData := sort.SortData{Data: items, Context: context, User: user}

			a.sort.Sort(&sortData)
			return sortData.Data.([]*module.Item)

		}

	case FilterAction:
		if a.filter != nil {
			filterData := filter.FilterData{Data: items, Uid: user.Id, Context: context}

			a.filter.Filter(&filterData)
			return filterData.Data.([]*module.Item)
		}

	}

	return items
}
