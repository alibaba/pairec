package filter

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

const (
	// rrfDefaultK is the default rank smoothing constant, aligned with common
	// implementations (e.g. Elasticsearch).
	rrfDefaultK = 60.0
	// rrfDefaultWeight is the fallback per-recall weight when not configured.
	rrfDefaultWeight = 1.0
	// rrfScoreName is the algo score name used to record the fusion score.
	rrfScoreName = "rrf_score"
)

// rrfRecallConfig holds the resolved weight for a single configured recall path.
type rrfRecallConfig struct {
	RecallName string
	Weight     float64
}

// RRFusionFilter fuses multiple recall paths using Reciprocal Rank Fusion (RRF).
// For each configured recall path, candidates are ranked by their in-path score
// in descending order (rank starts from 1). The final score of a document d is
// score(d) = sum_i w_i / (k + rank_i(d)), where a path that does not contain d
// contributes nothing. Only explicitly configured recall paths participate; items
// hit by no configured path are dropped.
type RRFusionFilter struct {
	name           string
	configs        []*rrfRecallConfig
	k              float64
	retainNum      int
	cloneInstances sync.Map
}

func newRRFRecallConfig(config recconf.RRFRule, defaultWeight float64) *rrfRecallConfig {
	weight := config.Weight
	if weight <= 0 {
		weight = defaultWeight
	}
	return &rrfRecallConfig{
		RecallName: config.RecallName,
		Weight:     weight,
	}
}

func NewRRFusionFilter(config recconf.FilterConfig) *RRFusionFilter {
	defaultWeight := config.RRFConf.DefaultWeight
	if defaultWeight <= 0 {
		defaultWeight = rrfDefaultWeight
	}

	k := config.RRFConf.K
	if k <= 0 {
		k = rrfDefaultK
	}

	filter := RRFusionFilter{
		name:      config.Name,
		k:         k,
		retainNum: config.RetainNum,
	}

	for _, conf := range config.RRFConf.Rules {
		if conf.RecallName == "" {
			continue
		}
		filter.configs = append(filter.configs, newRRFRecallConfig(conf, defaultWeight))
	}

	return &filter
}

func (f *RRFusionFilter) Filter(filterData *FilterData) error {
	if _, ok := filterData.Data.([]*module.Item); !ok {
		return errors.New("filter data type error")
	}

	return f.doFilter(filterData)
}

// scoreInRecall returns the item's score in the given recall path and whether
// the item is present in that path. It supports both single-recall items and
// items merged by UniqueFilter (whose per-path scores live in RecallScores).
func scoreInRecall(item *module.Item, recallName string) (float64, bool) {
	if item.RetrieveId == recallName {
		return item.Score, true
	}
	if item.RecallScores != nil {
		if score, ok := item.RecallScores[recallName]; ok {
			return score, true
		}
	}
	return 0, false
}

func (f *RRFusionFilter) doFilter(filterData *FilterData) error {
	start := time.Now()
	items := filterData.Data.([]*module.Item)

	fusionScores := make(map[module.ItemId]float64, len(items))
	itemMap := make(map[module.ItemId]*module.Item, len(items))

	for _, config := range f.configs {
		// Collect candidates that appear in this recall path.
		type scoredItem struct {
			item  *module.Item
			score float64
		}
		recallCandidateMap := make(map[module.ItemId]scoredItem, len(items))
		for _, item := range items {
			if score, ok := scoreInRecall(item, config.RecallName); ok {
				// Deduplicate only within the current recall path. Scores from
				// different recall paths are never compared.
				current, exists := recallCandidateMap[item.Id]
				if !exists || score > current.score {
					recallCandidateMap[item.Id] = scoredItem{item: item, score: score}
				}
			}
		}
		candidates := make([]scoredItem, 0, len(recallCandidateMap))
		for _, candidate := range recallCandidateMap {
			candidates = append(candidates, candidate)
		}

		// Rank by in-path score desc; break ties by item id asc for determinism.
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].score != candidates[j].score {
				return candidates[i].score > candidates[j].score
			}
			return candidates[i].item.Id < candidates[j].item.Id
		})

		for idx, c := range candidates {
			rank := idx + 1 // rank starts from 1
			fusionScores[c.item.Id] += config.Weight / (f.k + float64(rank))
			if _, ok := itemMap[c.item.Id]; !ok {
				itemMap[c.item.Id] = c.item
			}
		}
	}

	newItems := make([]*module.Item, 0, len(itemMap))
	for id, item := range itemMap {
		score := fusionScores[id]
		item.Score = score
		item.AddAlgoScore(rrfScoreName, score)
		newItems = append(newItems, item)
	}

	// Order by fusion score desc; break ties by item id asc for determinism.
	sort.SliceStable(newItems, func(i, j int) bool {
		if newItems[i].Score != newItems[j].Score {
			return newItems[i].Score > newItems[j].Score
		}
		return newItems[i].Id < newItems[j].Id
	})

	if f.retainNum > 0 && len(newItems) > f.retainNum {
		newItems = newItems[:f.retainNum]
	}

	if len(f.configs) == 0 {
		log.Warning("module=RRFusionFilter\tname=" + f.name + "\terror=no recall configured, result is empty")
	}

	filterData.Data = newItems
	filterInfoLog(filterData, "RRFusionFilter", f.name, len(newItems), start)
	return nil
}

func (f *RRFusionFilter) MatchTag(tag string) bool {
	// default filter, so filter all tag
	return true
}
