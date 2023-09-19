package filter

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	psort "github.com/alibaba/pairec/v2/sort"
	"github.com/alibaba/pairec/v2/utils"
)

const (
	Fix_Count_Type        = "fix"
	Accumulate_Count_Type = "accumulator"
)

type PriorityAdjustCountFilter struct {
	ensureDiversity   bool
	diversityDao      module.DiversityDao
	configs           []recconf.AdjustCountConfig
	diversityMinCount int
}

func NewPriorityAdjustCountFilter(config recconf.FilterConfig) *PriorityAdjustCountFilter {
	filter := PriorityAdjustCountFilter{
		configs:           config.AdjustCountConfs,
		ensureDiversity:   config.EnsureDiversity,
		diversityMinCount: config.DiversityMinCount,
	}
	if filter.diversityMinCount <= 0 {
		filter.diversityMinCount = 10
	}

	if config.DiversityDaoConf.AdapterType != "" {
		filter.diversityDao = module.NewDiversityDao(config)
	}

	return &filter
}
func (f *PriorityAdjustCountFilter) Filter(filterData *FilterData) error {
	context := filterData.Context

	ensureDiversity := f.ensureDiversity
	diversityMinCount := f.diversityMinCount
	if context.ExperimentResult != nil {
		params := context.ExperimentResult.GetExperimentParams()
		/**
		rankconf := params.Get("generalRankConf", "")
		if rankconf != "" {
			return nil
		}
		**/
		ensure := params.Get("ensure_diversity_min_count", nil)
		if ensure != nil {
			e := utils.ToInt(ensure, 0)
			diversityMinCount = e
			ensureDiversity = e != 0
		}
	}

	if items, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")
	} else if ensureDiversity {
		context.LogDebug("module=PriorityAdjustCountFilter\tensure diversity")
		_ = f.diversityDao.GetDistinctValue(items, context)
	}

	return f.doFilter(filterData, ensureDiversity, diversityMinCount)
}

func (f *PriorityAdjustCountFilter) doFilter(filterData *FilterData, ensureDiversity bool, diversityMinCnt int) error {
	ctx := filterData.Context
	start := time.Now()
	items := filterData.Data.([]*module.Item)
	newItems := make([]*module.Item, 0, 200)
	recallToItemMap := make(map[string][]*module.Item)

	// first random
	rand.Shuffle(len(items)/2, func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	sort.Sort(sort.Reverse(psort.ItemScoreSlice(items)))

	var distinctFields []string
	if f.diversityDao != nil {
		distinctFields = f.diversityDao.GetDistinctFields()
		if len(distinctFields) == 0 {
			ensureDiversity = false
		}
	}

	fieldValueCnt := make([]map[interface{}]int, len(distinctFields))
	for j, item := range items {
		recallToItemMap[item.RetrieveId] = append(recallToItemMap[item.RetrieveId], item)
		if ensureDiversity {
			for i, field := range distinctFields {
				if j == 0 {
					fieldValueCnt[i] = make(map[interface{}]int)
				}
				value := item.GetProperty(field)
				if value == nil {
					value = "null"
				}
				if cnt, ok := fieldValueCnt[i][value]; ok {
					fieldValueCnt[i][value] = cnt + 1
				} else {
					fieldValueCnt[i][value] = 1
				}
			}
		}
	}

	quota := utils.MinInt(f.configs[len(f.configs)-1].Count, len(items))
	total := float32(len(items))
	ctx.LogDebug(fmt.Sprintf("model=priority_adjust_count_filter\tquota=%d\ttotal=%v", quota, total))
	fieldValueQuota := make([]map[interface{}]int, len(distinctFields))
	if ensureDiversity {
		for i, valueCnt := range fieldValueCnt {
			fieldValueQuota[i] = make(map[interface{}]int)
			for value, cnt := range valueCnt {
				ratio := float32(cnt) / total
				fieldValueQuota[i][value] = utils.MaxInt(int(ratio*float32(quota)), diversityMinCnt)
			}
		}
	}

	fieldValueAccum := make([]map[interface{}]int, len(distinctFields))
	for i := range distinctFields {
		fieldValueAccum[i] = make(map[interface{}]int)
	}
	recallCntMap := make(map[string]int)
	accumulator := 0
	for _, config := range f.configs {
		recallItems := recallToItemMap[config.RecallName]
		if config.Type == Fix_Count_Type {
			if len(recallItems) < config.Count {
				newItems = append(newItems, recallItems...)
			} else {
				newItems = append(newItems, recallItems[:config.Count]...)
			}
		} else if config.Type == Accumulate_Count_Type {
			if ensureDiversity {
				accumulator = len(newItems)
				for j, item := range recallItems {
					valid := true
					for i, field := range distinctFields {
						value := item.GetProperty(field)
						if value == nil {
							value = "null"
						}
						q := fieldValueQuota[i][value]
						if cnt, ok := fieldValueAccum[i][value]; ok && cnt > q {
							valid = false
							break
						}
					}
					if valid {
						newItems = append(newItems, item)
						recallItems[j] = nil // set a mark
						for i, field := range distinctFields {
							value := item.GetProperty(field)
							if value == nil {
								value = "null"
							}
							if cnt, ok := fieldValueAccum[i][value]; ok {
								fieldValueAccum[i][value] = cnt + 1
							} else {
								fieldValueAccum[i][value] = 1
							}
						}
						if len(newItems) >= config.Count {
							break
						}
					}
				}
				recallCnt := len(newItems) - accumulator
				if cnt, ok := recallCntMap[config.RecallName]; ok {
					recallCntMap[config.RecallName] = cnt + recallCnt
				} else {
					recallCntMap[config.RecallName] = recallCnt
				}
			} else {
				count := config.Count - accumulator
				if len(recallItems) >= count {
					newItems = append(newItems, recallItems[:count]...)
					accumulator += count
				} else {
					newItems = append(newItems, recallItems...)
					accumulator += len(recallItems)
				}
			}
		}
	}

	if ensureDiversity && len(newItems) < quota {
		for _, config := range f.configs {
			quo := config.Count
			if recallCnt, ok := recallCntMap[config.RecallName]; ok {
				quo -= recallCnt
			}
			if quo > 0 {
				recallItems := recallToItemMap[config.RecallName]
				for i, item := range recallItems {
					if item == nil {
						continue
					}
					newItems = append(newItems, item)
					for j, field := range distinctFields {
						value := item.GetProperty(field)
						if value == nil {
							value = "null"
						}
						if cnt, ok := fieldValueAccum[j][value]; ok {
							fieldValueAccum[j][value] = cnt + 1
						} else {
							fieldValueAccum[j][value] = 1
						}
					}
					if len(newItems) >= quota {
						break
					}
					quo -= 1
					if quo == 0 {
						break
					}
					recallItems[i] = nil
				}
			}
		}
	}

	if filterData.Context.Debug {
		for _, accumMap := range fieldValueAccum {
			filterData.Context.LogDebug(fmt.Sprintf("%v", accumMap))
		}
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "PriorityAdjustCountFilter", len(newItems), start)
	return nil
}
