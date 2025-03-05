package filter

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	psort "github.com/alibaba/pairec/v2/sort"
)

type diversityAdjustCountConfig struct {
	Id                  int
	Count               int
	Type                string
	evaluableExpression *govaluate.EvaluableExpression
	Expression          string
}

func newDiversityAdjustCountConfig(config recconf.AdjustCountConfig) (*diversityAdjustCountConfig, error) {
	expression, err := govaluate.NewEvaluableExpression(config.Expression)
	if err != nil {
		return nil, err
	}
	return &diversityAdjustCountConfig{
		Count:               config.Count,
		Type:                config.Type,
		evaluableExpression: expression,
		Expression:          config.Expression,
	}, nil
}

type DiversityAdjustCountFilter struct {
	name           string
	configs        []*diversityAdjustCountConfig
	cloneInstances map[string]*DiversityAdjustCountFilter
}

func NewDiversityAdjustCountFilter(config recconf.FilterConfig) *DiversityAdjustCountFilter {
	filter := DiversityAdjustCountFilter{
		name:           config.Name,
		cloneInstances: make(map[string]*DiversityAdjustCountFilter),
	}
	for i, conf := range config.AdjustCountConfs {
		diversityConfig, err := newDiversityAdjustCountConfig(conf)
		if err != nil {
			panic(err)
		}
		diversityConfig.Id = i
		filter.configs = append(filter.configs, diversityConfig)
		if i > 0 {
			if diversityConfig.Type == Accumulate_Count_Type && filter.configs[i-1].Type == Accumulate_Count_Type {
				if diversityConfig.Count < filter.configs[i-1].Count {
					panic(fmt.Sprintf("diversity adjust count config error, accumulator type must greater than pre config, %v", config.AdjustCountConfs))
				}
			}
		}
	}

	return &filter
}
func (f *DiversityAdjustCountFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")
	}

	return f.doFilter(filterData)
}

func (f *DiversityAdjustCountFilter) doFilter(filterData *FilterData) error {
	ctx := filterData.Context
	start := time.Now()
	items := filterData.Data.([]*module.Item)
	newItems := make([]*module.Item, 0, 200)
	recallToItemMap := make(map[int][]*module.Item)

	// first random
	rand.Shuffle(len(items)/2, func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})

	sort.Sort(sort.Reverse(psort.ItemScoreSlice(items)))
	itemFeatruesList := make([]map[string]any, len(items))
	for i, item := range items {
		itemFeatruesList[i] = item.GetFeatures()
	}
	for i, item := range items {
		for _, config := range f.configs {
			result, err := config.evaluableExpression.Evaluate(itemFeatruesList[i])
			if err != nil {
				ctx.LogWarning(fmt.Sprintf("requestId=%s\tmodel=DiversityAdjustCountFilter\tmsg=evaluate error\titem=%v\terror=%v", ctx.RecommendId, item, err))
				continue
			}
			if b, ok := result.(bool); ok && b {
				recallToItemMap[config.Id] = append(recallToItemMap[config.Id], item)
			}
		}
	}
	if ctx.Debug {
		for id, itemList := range recallToItemMap {
			for _, config := range f.configs {
				if config.Id == id {
					log.Info(fmt.Sprintf("requestId=%s\tmodel=DiversityAdjustCountFilter\texpression=%s\titemCnt=%d", ctx.RecommendId, config.Expression, len(itemList)))
					break
				}
			}
		}
	}

	duplicateItemMap := make(map[*module.Item]bool, f.configs[len(f.configs)-1].Count)
	accumulator := 0
	for _, config := range f.configs {
		recallItems := recallToItemMap[config.Id]
		if config.Type == Fix_Count_Type {
			for i := 0; i < len(recallItems) && i < config.Count; i++ {
				if _, ok := duplicateItemMap[recallItems[i]]; ok {
					continue
				}
				newItems = append(newItems, recallItems[i])
				duplicateItemMap[recallItems[i]] = true
			}
		} else if config.Type == Accumulate_Count_Type {
			count := config.Count - accumulator

			for i := 0; i < len(recallItems) && i < count; i++ {
				if _, ok := duplicateItemMap[recallItems[i]]; ok {
					continue
				}
				newItems = append(newItems, recallItems[i])
				duplicateItemMap[recallItems[i]] = true

				accumulator++
			}
		}
	}
	filterData.Data = newItems
	filterInfoLog(filterData, "DiversityAdjustCountFilter", f.name, len(newItems), start)
	return nil
}
