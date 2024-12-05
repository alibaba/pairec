package filter

import (
	"errors"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// user exposure history filter
type User2ItemCustomFilter struct {
	name               string
	user2ItemCustomDao module.User2ItemCustomFilterDao
}

func NewUser2ItemCustomFilter(config recconf.FilterConfig) *User2ItemCustomFilter {
	filter := User2ItemCustomFilter{
		name:               config.Name,
		user2ItemCustomDao: module.NewUser2ItemCustomFilterDao(config),
	}

	return &filter
}
func (f *User2ItemCustomFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *User2ItemCustomFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)

	newItems := f.user2ItemCustomDao.Filter(filterData.Uid, items, filterData.Context)

	filterData.Data = newItems

	filterInfoLog(filterData, "User2ItemCustomFilter", f.name, len(newItems), start)
	return nil
}
func (f *User2ItemCustomFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
