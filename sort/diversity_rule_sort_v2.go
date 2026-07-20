package sort

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

const defaultLastItemIdsParam = "last_page_item_ids"

// DiversityRuleSortV2 is a standalone variant of DiversityRuleSort that
// supports cross page diversity: the tail item ids of the previous page are
// passed by the client, their dimension values are loaded from the configured
// datasource and used to warm up the diversity window, so that the first
// items of the current page are also constrained by the previous page.
type DiversityRuleSortV2 struct {
	diversitySize           int
	diversityRules          []recconf.DiversityRuleConfig
	exclusionRules          []recconf.ExclusionRuleConfig
	excludeRecallMap        map[string]bool
	filterParam             *module.FilterParam
	cloneMutex              sync.RWMutex
	cloneInstances          map[string]*DiversityRuleSortV2
	name                    string
	exploreItemSize         int
	multiValueDimensionConf []recconf.MultiValueDimensionConfig
	multiDimensionMaps      []map[int]recconf.MultiValueDimensionConfig // size is equal config.DiversityRules, per entry is map of multi dimension config for each diversity rule of dimensions
	lastItemIdsParam        string
	dimLoader               module.DiversityDao
	crossPageEnable         bool
}

func NewDiversityRuleSortV2(config recconf.SortConfig) *DiversityRuleSortV2 {
	sort := DiversityRuleSortV2{
		diversitySize:           config.DiversitySize,
		diversityRules:          config.DiversityRules,
		exclusionRules:          config.ExclusionRules,
		excludeRecallMap:        make(map[string]bool, len(config.ExcludeRecalls)),
		filterParam:             nil,
		name:                    config.Name,
		cloneInstances:          make(map[string]*DiversityRuleSortV2),
		exploreItemSize:         -1,
		multiValueDimensionConf: config.MultiValueDimensionConf,
	}

	for _, recallName := range config.ExcludeRecalls {
		sort.excludeRecallMap[recallName] = true
	}

	if len(config.Conditions) > 0 {
		filterParam := module.NewFilterParamWithConfig(config.Conditions)
		sort.filterParam = filterParam
	}
	if config.ExploreItemSize > 0 {
		sort.exploreItemSize = config.ExploreItemSize
	}
	if len(config.MultiValueDimensionConf) > 0 {
		for _, diversityRuleConfig := range sort.diversityRules {
			multiDimensionMap := make(map[int]recconf.MultiValueDimensionConfig)
			for i, dimension := range diversityRuleConfig.Dimensions {
				for _, multiDimension := range config.MultiValueDimensionConf {
					if multiDimension.DimensionName == dimension {
						multiDimensionMap[i] = multiDimension
						break
					}
				}
			}
			sort.multiDimensionMaps = append(sort.multiDimensionMaps, multiDimensionMap)

		}

	}

	if config.CrossPageDiversity.Enable {
		sort.lastItemIdsParam = config.CrossPageDiversity.LastItemIdsParam
		if sort.lastItemIdsParam == "" {
			sort.lastItemIdsParam = defaultLastItemIdsParam
		}
		sort.dimLoader = module.NewDiversityDaoByConf(config.CrossPageDiversity.DiversityDaoConf)
		if sort.dimLoader != nil {
			sort.crossPageEnable = true
			checkDistinctFieldsCoverage(config, sort.dimLoader.GetDistinctFields())
		} else {
			log.Error("module=DiversityRuleSortV2\terror=create diversity dao failed, cross page diversity disabled")
		}
	}

	return &sort
}

// checkDistinctFieldsCoverage warns when some rule dimensions are not covered
// by the DistinctFields of the dao conf, in which case the history dimension
// values would be empty strings and the warmup is ineffective for that rule.
func checkDistinctFieldsCoverage(config recconf.SortConfig, distinctFields []string) {
	fieldMap := make(map[string]bool, len(distinctFields))
	for _, field := range distinctFields {
		fieldMap[field] = true
	}
	for _, rule := range config.DiversityRules {
		for _, dimension := range rule.Dimensions {
			if !fieldMap[dimension] {
				log.Warning("module=DiversityRuleSortV2\twarn=dimension " + dimension + " not covered by DistinctFields, cross page warmup is ineffective for it")
			}
		}
	}
}

func (s *DiversityRuleSortV2) Sort(sortData *SortData) error {
	if _, ok := sortData.Data.([]*module.Item); !ok {
		return errors.New("sort data type error")
	}

	// if condition is empty
	if s.filterParam == nil {
		return s.doSort(sortData)
	} else {
		userProperties := sortData.User.MakeUserFeatures2()
		flag, err := s.filterParam.EvaluateByDomain(userProperties, nil)
		if err != nil {
			return err
		}
		if flag {
			return s.doSort(sortData)
		}
	}

	return nil
}

func (s *DiversityRuleSortV2) createDiversityRules(size int) (ret []DiversityRuleV2Interface) {
	for i, config := range s.diversityRules {
		var rule DiversityRuleV2Interface
		if len(s.multiValueDimensionConf) > 0 {
			rule = NewDiversityRuleMultiDimensionV2(config, size, s.multiDimensionMaps[i])
		} else {
			rule = NewDiversityRuleV2(config, size)
		}

		ret = append(ret, rule)
	}

	return
}

func (s *DiversityRuleSortV2) createExclusionRules(user *module.User, size int) (ret []*DiversityExclusionRule) {
	features := make(map[string]any)
	if user != nil {
		features = user.MakeUserFeatures2()
	}
	for _, config := range s.exclusionRules {
		rule := NewDiversityExclusionRule(config, features, size)
		ret = append(ret, rule)
	}

	return
}

// getLastPageParam looks up the raw parameter value: first from the top-level
// request param (custom IParam impls), then from the "features" map which is
// the standard way to pass custom fields in the recommend API.
func getLastPageParam(ctx *context.RecommendContext, name string) interface{} {
	if ctx == nil || ctx.Param == nil {
		return nil
	}
	if value := ctx.GetParameter(name); value != nil {
		return value
	}
	if features, ok := ctx.GetParameter("features").(map[string]interface{}); ok {
		return features[name]
	}
	return nil
}

// parseLastPageItemIds parses item ids from the request parameter value.
// It supports []interface{}, []string, JSON array string and comma separated
// string. Invalid JSON array payload returns nil so that the warmup is skipped
// instead of feeding garbage ids.
func parseLastPageItemIds(value interface{}) []string {
	switch v := value.(type) {
	case []interface{}:
		return normalizeItemIds(v)
	case []string:
		ids := make([]string, 0, len(v))
		for _, part := range v {
			if id := strings.TrimSpace(part); id != "" {
				ids = append(ids, id)
			}
		}
		return ids
	case string:
		str := strings.TrimSpace(v)
		if str == "" {
			return nil
		}
		if strings.HasPrefix(str, "[") {
			// treat as JSON array, use []interface{} to accept both string and
			// number elements, no comma split fallback for malformed payload
			var raw []interface{}
			if err := json.Unmarshal([]byte(str), &raw); err != nil {
				return nil
			}
			return normalizeItemIds(raw)
		}
		parts := strings.Split(str, ",")
		ids := make([]string, 0, len(parts))
		for _, part := range parts {
			if id := strings.TrimSpace(part); id != "" {
				ids = append(ids, id)
			}
		}
		return ids
	default:
		return nil
	}
}

// normalizeItemIds converts raw values to trimmed non empty id strings.
func normalizeItemIds(raw []interface{}) []string {
	ids := make([]string, 0, len(raw))
	for _, item := range raw {
		if id := strings.TrimSpace(utils.ToString(item, "")); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

// warmupCrossPage loads dimension values of the previous page tail items and
// sets them as the history of each diversity rule. It returns true when at
// least one rule gets a non empty history.
func (s *DiversityRuleSortV2) warmupCrossPage(sortData *SortData, diversityRules []DiversityRuleV2Interface) bool {
	ctx := sortData.Context
	ids := parseLastPageItemIds(getLastPageParam(ctx, s.lastItemIdsParam))
	if len(ids) == 0 {
		return false
	}

	maxNeed := 0
	for _, config := range s.diversityRules {
		if need := historyMaxLen(config); need > maxNeed {
			maxNeed = need
		}
	}
	if maxNeed <= 0 {
		return false
	}
	if len(ids) > maxNeed {
		ids = ids[len(ids)-maxNeed:]
	}

	prevItems := make([]*module.Item, 0, len(ids))
	for _, id := range ids {
		prevItems = append(prevItems, module.NewItem(id))
	}

	if err := s.dimLoader.GetDistinctValue(prevItems, ctx); err != nil {
		ctx.LogError("module=DiversityRuleSortV2\terror=load prev page dimension failed, skip cross page warmup")
		return false
	}

	warmed := false
	for _, rule := range diversityRules {
		rule.SetHistory(prevItems)
		if rule.HistoryLen() > 0 {
			warmed = true
		}
	}
	return warmed
}

func (s *DiversityRuleSortV2) doSort(sortData *SortData) error {
	start := time.Now()
	items := sortData.Data.([]*module.Item)

	diversityRules := s.createDiversityRules(len(items))
	if len(diversityRules) == 0 {
		return nil
	}

	warmed := false
	if s.crossPageEnable && s.dimLoader != nil {
		warmed = s.warmupCrossPage(sortData, diversityRules)
	}

	exclusionRules := s.createExclusionRules(sortData.User, len(items))
	var excludeItems []*module.Item
	if len(s.excludeRecallMap) > 0 {
		newItems := make([]*module.Item, 0, len(items))
		for _, item := range items {
			if _, ok := s.excludeRecallMap[item.GetRecallName()]; ok {
				excludeItems = append(excludeItems, item)
			} else {
				newItems = append(newItems, item)
			}
		}

		items = newItems
	}

	itemLength := len(items)
	//if items empty
	if itemLength == 0 {
		return nil
	}

	diversitySize := sortData.Context.Size

	if s.diversitySize > 0 {
		diversitySize = s.diversitySize
		if diversitySize > itemLength {
			diversitySize = itemLength
		}
	}

	result := make([]*module.Item, 0, diversitySize)
	alreadyMatchItems := make(map[module.ItemId]bool, diversitySize)
	exFlag := false
	index := 0
	// when warmed, the seed item is not appended unconditionally: the main
	// loop below matches every item (including the first one) against the
	// history prefix, and it has its own weight/first-item fallback.
	if !warmed {
		if len(exclusionRules) > 0 {
			for _, item := range items {
				exFlag = false
				for _, rule := range exclusionRules {
					if rule.Match(1, item) {
						exFlag = true
						break
					}
				}
				if !exFlag { // item not match any exclusion rule
					alreadyMatchItems[item.Id] = true
					result = append(result, item)
					break
				}
			}

			if len(result) == 0 {
				alreadyMatchItems[items[0].Id] = true
				result = append(result, items[0])
				items = items[1:]
			}

		} else {
			alreadyMatchItems[items[0].Id] = true
			result = append(result, items[0])
			items = items[1:]
		}
		index = 1
	}

	weight := 0
	itemWeightContainer := itemWeightContainerPool.Get().(*itemWeightContainer)
	defer itemWeightContainerPool.Put(itemWeightContainer)
	hasWeight := false
	for _, rule := range diversityRules {
		if rule.GetWeight() > 0 {
			hasWeight = true
			break
		}
	}
	for len(result) <= diversitySize {
		if index == itemLength {
			break
		}

		flag := true
		// if all the rest items not match diversity rule, use the first item append to the result
		firstItemIndex := -1
		itemWeightContainer.reset()
		for i, item := range items {
			if _, ok := alreadyMatchItems[item.Id]; ok {
				continue
			}
			if len(exclusionRules) > 0 {
				exFlag = false
				for _, rule := range exclusionRules {
					if rule.Match(len(result)+1, item) { // next position check item is match the exclusion rule
						exFlag = true
						break
					}
				}
				if exFlag { // if item match the exclusion rule, so skip it, search next item
					continue
				}
			}

			if firstItemIndex == -1 {
				firstItemIndex = i
			}
			if s.exploreItemSize > 0 && i-firstItemIndex >= s.exploreItemSize {
				break
			}
			flag = true
			weight = 0
			for _, rule := range diversityRules {
				if hasWeight { // if has weight, so need to iterate all the rules find the max weight
					if rule.Match(item, result) {
						weight += rule.GetWeight()
					} else {
						flag = false
					}
				} else {
					if flag = rule.Match(item, result); !flag {
						break
					}
				}
			}

			// if the item match all the diversity rule, so add it to the result
			if flag {
				alreadyMatchItems[item.Id] = true
				result = append(result, item)
				index++
				break
			} else {
				itemWeightContainer.addItemWeight(item, weight)
			}
		}

		if !flag {
			if item := itemWeightContainer.getItem(); item != nil {
				alreadyMatchItems[item.Id] = true
				result = append(result, item)
				index++
			} else {
				alreadyMatchItems[items[firstItemIndex].Id] = true
				result = append(result, items[firstItemIndex])
				index++
			}
		} else if firstItemIndex == -1 { // all items are in alreadyMatchItems map
			break
		}
	}

	for _, item := range items {
		if _, ok := alreadyMatchItems[item.Id]; ok {
			continue
		}
		result = append(result, item)
	}

	result = append(result, excludeItems...)

	sortData.Data = result
	sortInfoLogWithName(sortData, "DiversityRuleSortV2", s.name, len(result), start)
	return nil
}
