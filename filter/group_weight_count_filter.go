package filter

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
	"github.com/alibaba/pairec/utils"
)

type GroupWeightStrategy interface {
	SetDimensionScoreMap(dimensionScoreMap map[string]float64)
	TotalScore()
	DimensionCount(dimension string, retainNum int) (count int, exceedNum bool)
}

type SoftmaxGroupWeightStrategy struct {
	dimensionScoreMap map[string]float64
	totalScore        float64
	scoreWeight       float64
	groupMinNum       int
	groupMaxNum       int
}

func (s *SoftmaxGroupWeightStrategy) SetDimensionScoreMap(dimensionScoreMap map[string]float64) {
	s.dimensionScoreMap = dimensionScoreMap
}

func (s *SoftmaxGroupWeightStrategy) TotalScore() {
	s.totalScore = 0
	for _, score := range s.dimensionScoreMap {
		s.totalScore += math.Exp(s.scoreWeight * score)
	}

	if s.totalScore == float64(0) {
		s.totalScore = 1
	}
}

func (s *SoftmaxGroupWeightStrategy) DimensionCount(dimension string, retainNum int) (count int, exceedNum bool) {
	count = utils.MaxInt(int(math.Round(float64(retainNum)*(math.Exp(s.scoreWeight*s.dimensionScoreMap[dimension])/s.totalScore))), s.groupMinNum)
	if s.groupMaxNum > 0 && count >= s.groupMaxNum {
		count = s.groupMaxNum
		exceedNum = true
	}
	return
}

type AvgGroupWeightStrategy struct {
	dimensionScoreMap map[string]float64
	totalScore        float64
	scoreWeight       float64
	groupMinNum       int
	groupMaxNum       int
}

func (s *AvgGroupWeightStrategy) SetDimensionScoreMap(dimensionScoreMap map[string]float64) {
	s.dimensionScoreMap = dimensionScoreMap
}

func (s *AvgGroupWeightStrategy) TotalScore() {
	s.totalScore = 0
	for _, score := range s.dimensionScoreMap {
		s.totalScore += s.scoreWeight * score
	}

	if s.totalScore == float64(0) {
		s.totalScore = 1
	}
}

func (s *AvgGroupWeightStrategy) DimensionCount(dimension string, retainNum int) (count int, exceedNum bool) {
	count = utils.MaxInt(int(math.Round(float64(retainNum)*(s.scoreWeight*s.dimensionScoreMap[dimension]/s.totalScore))), s.groupMinNum)
	if s.groupMaxNum > 0 && count >= s.groupMaxNum {
		count = s.groupMaxNum
		exceedNum = true
	}
	return
}

type GroupWeightCountFilter struct {
	retainNum                 int
	dimension                 string
	groupMinNum               int
	groupMaxNum               int
	scoreWeight               float64
	groupWeightStrategy       string
	groupWeightDimensionLimit map[string]int
}

func NewGroupWeightCountFilter(config recconf.FilterConfig) *GroupWeightCountFilter {
	filter := GroupWeightCountFilter{
		retainNum:                 config.RetainNum,
		dimension:                 config.Dimension,
		groupMinNum:               config.GroupMinNum,
		groupMaxNum:               config.GroupMaxNum,
		scoreWeight:               config.ScoreWeight,
		groupWeightStrategy:       config.GroupWeightStrategy,
		groupWeightDimensionLimit: config.GroupWeightDimensionLimit,
	}

	return &filter
}
func (f *GroupWeightCountFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")

	}
	return f.doFilter(filterData)
}

func (f *GroupWeightCountFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	ctx := filterData.Context
	items := filterData.Data.([]*module.Item)
	if len(items) <= f.retainNum {
		return nil
	}

	var groupWeightStrategy GroupWeightStrategy
	switch f.groupWeightStrategy {
	case "avg":
		groupWeightStrategy = &AvgGroupWeightStrategy{scoreWeight: f.scoreWeight,
			groupMinNum: f.groupMinNum, groupMaxNum: f.groupMaxNum}
	default:
		groupWeightStrategy = &SoftmaxGroupWeightStrategy{scoreWeight: f.scoreWeight,
			groupMinNum: f.groupMinNum, groupMaxNum: f.groupMaxNum}
	}

	if len(f.groupWeightDimensionLimit) > 0 {
		items = f.groupWeightDimensionFilter(items, ctx)
	}

	newItems := make([]*module.Item, 0, f.retainNum)
	groupDimensionMap := make(map[string][]*module.Item)
	retainGroupDimensionMap := make(map[string][]*module.Item)
	dimensionScoreMap := make(map[string]float64)
	exceedNumDimensionMap := make(map[string]bool)

	for _, item := range items {
		val := item.StringProperty(f.dimension)
		if val == "" {
			val = "DEFAULT"
		}

		groupDimensionMap[val] = append(groupDimensionMap[val], item)
		dimensionScoreMap[val] += item.Score
	}

	if ctx.Debug {
		for dimension, itemList := range groupDimensionMap {
			log.Info(fmt.Sprintf("requestId=%s\tdimension=%s\tcount=%d", ctx.RecommendId, dimension, len(itemList)))
		}

		log.Info(fmt.Sprintf("requestId=%s\tevent=befor GroupWeightCountFilter", ctx.RecommendId))
	}

	var (
		retainNum = 0
	)

	groupWeightStrategy.SetDimensionScoreMap(dimensionScoreMap)
	groupWeightStrategy.TotalScore()
	for dimension := range groupDimensionMap {
		_, exceed := groupWeightStrategy.DimensionCount(dimension, f.retainNum)
		if exceed {
			exceedNumDimensionMap[dimension] = true
		}
	}
	// find exceed num dimension first
	if len(exceedNumDimensionMap) > 0 {
		for dimension, itemList := range groupDimensionMap {
			if _, ok := exceedNumDimensionMap[dimension]; ok {
				delete(dimensionScoreMap, dimension)
				count := f.groupMaxNum
				if len(itemList) <= count {
					retainGroupDimensionMap[dimension] = append(retainGroupDimensionMap[dimension], itemList...)
					groupDimensionMap[dimension] = groupDimensionMap[dimension][:0]
				} else {
					retainGroupDimensionMap[dimension] = append(retainGroupDimensionMap[dimension], itemList[:count]...)
					groupDimensionMap[dimension] = groupDimensionMap[dimension][count:]
				}

				retainNum += len(retainGroupDimensionMap[dimension])
			}
		}
	} else {
		for dimension, itemList := range groupDimensionMap {
			count, _ := groupWeightStrategy.DimensionCount(dimension, f.retainNum)

			if len(itemList) <= count {
				retainGroupDimensionMap[dimension] = append(retainGroupDimensionMap[dimension], itemList...)
				groupDimensionMap[dimension] = groupDimensionMap[dimension][:0]
			} else {
				retainGroupDimensionMap[dimension] = append(retainGroupDimensionMap[dimension], itemList[:count]...)
				groupDimensionMap[dimension] = groupDimensionMap[dimension][count:]
			}

			retainNum += len(retainGroupDimensionMap[dimension])
		}

	}

	if retainNum < f.retainNum && len(exceedNumDimensionMap) > 0 {
		num := f.retainNum - len(exceedNumDimensionMap)*f.groupMaxNum
		groupWeightStrategy.SetDimensionScoreMap(dimensionScoreMap)
		groupWeightStrategy.TotalScore()
		for dimension, itemList := range groupDimensionMap {
			if _, ok := exceedNumDimensionMap[dimension]; ok {
				continue
			}
			count, _ := groupWeightStrategy.DimensionCount(dimension, num)

			if len(itemList) <= count {
				retainGroupDimensionMap[dimension] = append(retainGroupDimensionMap[dimension], itemList...)
				groupDimensionMap[dimension] = groupDimensionMap[dimension][:0]
			} else {
				retainGroupDimensionMap[dimension] = append(retainGroupDimensionMap[dimension], itemList[:count]...)
				groupDimensionMap[dimension] = groupDimensionMap[dimension][count:]
			}

			retainNum += len(retainGroupDimensionMap[dimension])
		}
	}

	if retainNum > f.retainNum {
		for {
			for dimension, itemList := range retainGroupDimensionMap {
				if len(itemList) > f.groupMinNum {
					retainGroupDimensionMap[dimension] = itemList[:len(itemList)-1]
					retainNum--
					if retainNum == f.retainNum {
						break
					}
				}
			}

			if retainNum == f.retainNum {
				break
			}
		}
	} else if retainNum < f.retainNum {
		for {
			for dimension, itemList := range groupDimensionMap {
				if len(itemList) > 0 {
					retainGroupDimensionMap[dimension] = append(retainGroupDimensionMap[dimension], itemList[0])
					groupDimensionMap[dimension] = itemList[1:]
					retainNum++
				}
				if retainNum == f.retainNum {
					break
				}
			}
			if retainNum == f.retainNum {
				break
			}
		}
	}

	for dimension, itemList := range retainGroupDimensionMap {
		if ctx.Debug {
			log.Info(fmt.Sprintf("requestId=%s\tdimension=%s\tcount=%d", ctx.RecommendId, dimension, len(itemList)))
		}
		newItems = append(newItems, itemList...)
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "GroupWeightCountFilter", len(newItems), start)
	return nil
}

func (f *GroupWeightCountFilter) groupWeightDimensionFilter(items []*module.Item, context *context.RecommendContext) []*module.Item {
	/**
	// log item position
	for i, item := range items {
		item.Extra = i
	}
	**/
	newItems := make([]*module.Item, 0, len(items))
	//defaultItems := make([]*module.Item, 0, f.retainNum)
	for dimension, sizeLimit := range f.groupWeightDimensionLimit {
		groupDimensionMap := make(map[string][]*module.Item)
		for i, item := range items {
			if item != nil {
				dimensionValue := item.StringProperty(dimension)
				if dimensionValue == "" {
					dimensionValue = "__DEFAULT__"
				}
				if (len(groupDimensionMap[dimensionValue]) + 1) > sizeLimit {
					items[i] = nil
				} else {
					groupDimensionMap[dimensionValue] = append(groupDimensionMap[dimensionValue], item)
				}
			}
		}
		if context.Debug {
			for dimensionValue, itemList := range groupDimensionMap {
				log.Info(fmt.Sprintf("requestId=%s\tevent=groupWeightDimensionFilter\tdimension_name=%s\tdimension_value=%s\tcount=%d", context.RecommendId, dimension, dimensionValue, len(itemList)))
			}
		}
	}

	for _, item := range items {
		if item != nil {
			newItems = append(newItems, item)
		}
	}

	return newItems
}
