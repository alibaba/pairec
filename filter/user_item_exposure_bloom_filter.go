package filter

import (
	"errors"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/filter/bloomfilter"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/service/hook"
	"github.com/alibaba/pairec/v2/utils"
)

// function for generate bloom filter key
type GenerateFilterKey func(uid module.UID, context *context.RecommendContext) string

// function for generate bloom filter values
type GenerateFilterValue func(uid module.UID, items []*module.Item, context *context.RecommendContext) [][]byte

// type GenerateFilter

// user exposure history filter
type User2ItemExposureBloomFilter struct {
	filterActive            bool
	logHistoryActive        bool
	generateFilterKeyFunc   GenerateFilterKey
	generateFilterValueFunc GenerateFilterValue
	bloom                   bloomfilter.BloomFilterInterface
}

func NewUser2ItemExposureBloomFilter(bloom bloomfilter.BloomFilterInterface, fkey GenerateFilterKey, fvalue GenerateFilterValue) *User2ItemExposureBloomFilter {
	if fkey == nil {
		panic("User2ItemExposureBloomFilter GenerateFilterKey func not nil")
	}

	if fvalue == nil {
		panic("User2ItemExposureBloomFilter GenerateFilterValue func not nil")
	}

	filter := User2ItemExposureBloomFilter{
		bloom:                   bloom,
		filterActive:            true,
		logHistoryActive:        true,
		generateFilterKeyFunc:   fkey,
		generateFilterValueFunc: fvalue,
	}

	hook.AddRecommendCleanHook(func(filter *User2ItemExposureBloomFilter) hook.RecommendCleanHookFunc {

		return func(context *context.RecommendContext, params ...interface{}) {
			user := params[0].(*module.User)
			items := params[1].([]*module.Item)
			filter.logHistory(user, items, context)
		}
	}(&filter))

	return &filter
}

func (f *User2ItemExposureBloomFilter) SetFilterActive(flag bool) {
	f.filterActive = flag
}

func (f *User2ItemExposureBloomFilter) SetLogHistoryActive(flag bool) {
	f.logHistoryActive = flag
}

func (f *User2ItemExposureBloomFilter) Filter(filterData *FilterData) error {
	if !f.filterActive {
		return nil
	}

	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *User2ItemExposureBloomFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)

	key := f.generateFilterKeyFunc(filterData.Uid, filterData.Context)
	values := f.generateFilterValueFunc(filterData.Uid, items, filterData.Context)

	bloomRet, err := f.bloom.Exists(key, values)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureBloomFilter\tuid=%s\terr=%v", filterData.Context.RecommendId, filterData.Uid, err))
		return nil
	}
	newItems := make([]*module.Item, 0, len(items))

	for i, b := range bloomRet {
		if !b {
			newItems = append(newItems, items[i])
		}
	}

	filterData.Data = newItems
	log.Info(fmt.Sprintf("requestId=%s\tevent=User2ItemExposureBloomFilter\tcost=%d", filterData.Context.RecommendId, utils.CostTime(start)))
	return nil
}

func (f *User2ItemExposureBloomFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}

func (f *User2ItemExposureBloomFilter) logHistory(user *module.User, items []*module.Item, context *context.RecommendContext) {
	if !f.logHistoryActive {
		return
	}
	key := f.generateFilterKeyFunc(user.Id, context)
	values := f.generateFilterValueFunc(user.Id, items, context)
	err := f.bloom.Add(key, values)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureBloomFilter\tuid=%s\terr=%v", context.RecommendId, user.Id, err))
		return
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=User2ItemExposureBloomFilter\tuid=%s\tmsg=log history success", context.RecommendId, user.Id))

}
