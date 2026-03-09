package sort

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

type CustomFieldSort struct {
	sortByField string
	sortOrder   string // "asc" 表示升序，"desc" 表示降序
	name        string
}

func NewCustomFieldSort(config recconf.SortConfig) *CustomFieldSort {
	if config.SortByField == "" {
		log.Error(fmt.Sprintf("sort by field is not configured for sort %s", config.Name))
		config.SortByField = "current_score"
	}

	// 默认降序
	sortOrder := strings.ToLower(config.SortOrder)
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return &CustomFieldSort{
		sortByField: config.SortByField,
		sortOrder:   sortOrder,
		name:        config.Name,
	}
}

func (s *CustomFieldSort) Sort(sortData *SortData) error {
	start := time.Now()
	items, ok := sortData.Data.([]*module.Item)
	if !ok {
		return errors.New("sort data type error")
	}

	ctx := sortData.Context
	sortByField := s.sortByField
	isAsc := s.sortOrder == "asc"

	sort.Slice(items, func(i, j int) bool {
		iScore, err1 := items[i].FloatExprData(sortByField)
		if err1 != nil {
			iScore = items[i].Score
			ctx.LogWarning(fmt.Sprintf("get sort field %s from item %s failed", sortByField, items[i].Id))
		}
		jScore, err2 := items[j].FloatExprData(sortByField)
		if err2 != nil {
			jScore = items[j].Score
			ctx.LogWarning(fmt.Sprintf("get sort field %s from item %s failed", sortByField, items[j].Id))
		}
		if isAsc {
			return iScore < jScore
		}
		return iScore > jScore
	})
	sortData.Data = items
	sortInfoLogWithName(sortData, "CustomFieldSort", s.name, len(items), start)
	return nil
}
