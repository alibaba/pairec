package filter

import (
	"errors"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type ConditionFilterItem struct {
	filterParam *module.FilterParam
	filterName  string
}

type ConditionFilter struct {
	filterItems       []*ConditionFilterItem
	defaultFitlerName string
}

func NewConditionFilter(config recconf.FilterConfig) *ConditionFilter {
	var items []*ConditionFilterItem
	for _, item := range config.ConditionFilterConfs.FilterConfs {

		filterItem := ConditionFilterItem{
			filterName: item.FilterName,
		}
		if len(item.Conditions) > 0 {
			filterItem.filterParam = module.NewFilterParamWithConfig(item.Conditions)
		}

		items = append(items, &filterItem)
	}
	filter := ConditionFilter{
		defaultFitlerName: config.ConditionFilterConfs.DefaultFilterName,
		filterItems:       items,
	}

	return &filter
}
func (f *ConditionFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	userProperties := filterData.User.MakeUserFeatures2()
	for _, item := range f.filterItems {
		if item.filterParam != nil {
			if flag, err := item.filterParam.EvaluateByDomain(userProperties, nil); err == nil {
				// if match condition
				if flag {
					filter, err := GetFilter(item.filterName)
					if err != nil {
						log.Error(fmt.Sprintf("requestId=%s\tmodule=ConditionFilter\tfilterName=%s\terror=%v", filterData.Context.RecommendId, item.filterName, err))
						return err
					}
					log.Info(fmt.Sprintf("requestId=%s\tmodule=ConditionFilter\tfilterName=%s", filterData.Context.RecommendId, item.filterName))
					return filter.Filter(filterData)
				}
			} else {
				log.Error(fmt.Sprintf("requestId=%s\tmodule=ConditionFilter\tfilterName=%s\terror=%v", filterData.Context.RecommendId, item.filterName, err))
			}
		}
	}

	if len(f.defaultFitlerName) > 0 {
		filter, err := GetFilter(f.defaultFitlerName)
		if err != nil {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=ConditionFilter\tdefaultFilterName=%s\terror=%v", filterData.Context.RecommendId, f.defaultFitlerName, err))
			return err
		}
		log.Info(fmt.Sprintf("requestId=%s\tmodule=ConditionFilter\tfilterName=%s", filterData.Context.RecommendId, f.defaultFitlerName))
		return filter.Filter(filterData)

	}

	return nil
}

func (f *ConditionFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
