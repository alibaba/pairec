package filter

import (
	"errors"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// user exposure history filter with condition
type User2ItemExposureWithConditionFilter struct {
	name                 string
	user2ItemExposureDao module.User2ItemExposureDao
	filterParam          *module.FilterParam
}

func NewUser2ItemExposureWithConditionFilter(config recconf.FilterConfig) *User2ItemExposureWithConditionFilter {
	filter := User2ItemExposureWithConditionFilter{
		name:                 config.Name,
		user2ItemExposureDao: module.NewUser2ItemExposureDao(config),
	}
	if len(config.Conditions) > 0 {
		filter.filterParam = module.NewFilterParamWithConfig(config.Conditions)
	}

	return &filter
}
func (f *User2ItemExposureWithConditionFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	userProperties := filterData.User.MakeUserFeatures2()
	if f.filterParam != nil {
		if flag, err := f.filterParam.EvaluateByDomain(userProperties, nil); err == nil {
			// if flag == true, should clear history, and no need to filter
			if flag {
				go f.user2ItemExposureDao.ClearHistory(filterData.User, filterData.Context)
				filterInfoLog(filterData, "User2ItemExposureWithConditionFilter", f.name, len(filterData.Data.([]*module.Item)), time.Now())
				return nil
			}
		} else {
			log.Error(fmt.Sprintf("requestId=%s\tmodule=%s\terror=%v", filterData.Context.RecommendId, "User2ItemExposureWithConditionFilter", err))
		}
	}
	return f.doFilter(filterData)
}

func (f *User2ItemExposureWithConditionFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)

	newItems := f.user2ItemExposureDao.FilterByHistory(filterData.Uid, items, filterData.Context)

	filterData.Data = newItems
	filterInfoLog(filterData, "User2ItemExposureWithConditionFilter", f.name, len(newItems), start)
	return nil
}
func (f *User2ItemExposureWithConditionFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
