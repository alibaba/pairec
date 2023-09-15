package filter

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

// user exposure history filter
type User2ItemExposureFilter struct {
	user2ItemExposureDao module.User2ItemExposureDao
}

func NewUser2ItemExposureFilter(config recconf.FilterConfig) *User2ItemExposureFilter {
	filter := User2ItemExposureFilter{
		user2ItemExposureDao: module.NewUser2ItemExposureDao(config),
	}

	return &filter
}
func (f *User2ItemExposureFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *User2ItemExposureFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)

	newItems := f.user2ItemExposureDao.FilterByHistory(filterData.Uid, items)

	filterData.Data = newItems
	filterInfoLog(filterData, "User2ItemExposureFilter", len(newItems), start)
	return nil
}
func (f *User2ItemExposureFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
