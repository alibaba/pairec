package filter

import (
	"math"
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func newRRFTestItem(id, recallName string, score float64) *module.Item {
	item := module.NewItem(id)
	item.RetrieveId = recallName
	item.Score = score
	return item
}

func newRRFFilterData(items []*module.Item) *FilterData {
	ctx := context.NewRecommendContext()
	ctx.RecommendId = "rrf-test"
	return &FilterData{Data: items, Context: ctx}
}

func assertFloatEqual(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-12 {
		t.Fatalf("got score %.15f, want %.15f", got, want)
	}
}

func TestRRFusionFilter(t *testing.T) {
	itemA := newRRFTestItem("a", "r1", 100)
	itemA.RecallScores = map[string]float64{"r1": 100, "r2": 80}
	itemB := newRRFTestItem("b", "r1", 90)
	itemC := newRRFTestItem("c", "r2", 100)

	filter := NewRRFusionFilter(recconf.FilterConfig{
		Name:      "rrf",
		RetainNum: 3,
		RRFConf: recconf.RRFConfig{
			K: 60,
			Rules: []recconf.RRFRule{
				{RecallName: "r1", Weight: 1},
				{RecallName: "r2", Weight: 2},
			},
		},
	})
	data := newRRFFilterData([]*module.Item{itemA, itemB, itemC})

	if err := filter.Filter(data); err != nil {
		t.Fatal(err)
	}

	got := data.Data.([]*module.Item)
	if len(got) != 3 {
		t.Fatalf("got %d items, want 3", len(got))
	}
	wantOrder := []module.ItemId{"a", "c", "b"}
	for i, id := range wantOrder {
		if got[i].Id != id {
			t.Fatalf("item[%d] = %s, want %s", i, got[i].Id, id)
		}
	}
	assertFloatEqual(t, itemA.Score, 1.0/61.0+2.0/62.0)
	assertFloatEqual(t, itemB.Score, 1.0/62.0)
	assertFloatEqual(t, itemC.Score, 2.0/61.0)
	assertFloatEqual(t, itemA.GetAlgoScore(rrfScoreName), itemA.Score)
}

func TestRRFusionFilterRanksScoresOnlyWithinEachRecall(t *testing.T) {
	itemAR1 := newRRFTestItem("a", "r1", 0.9)
	itemAR2 := newRRFTestItem("a", "r2", 100)
	itemB := newRRFTestItem("b", "r1", 0.8)
	itemC := newRRFTestItem("c", "r2", 200)

	filter := NewRRFusionFilter(recconf.FilterConfig{
		RRFConf: recconf.RRFConfig{
			K: 60,
			Rules: []recconf.RRFRule{
				{RecallName: "r1", Weight: 1},
				{RecallName: "r2", Weight: 1},
			},
		},
	})
	data := newRRFFilterData([]*module.Item{itemAR1, itemAR2, itemB, itemC})

	if err := filter.Filter(data); err != nil {
		t.Fatal(err)
	}

	got := data.Data.([]*module.Item)
	if len(got) != 3 {
		t.Fatalf("got %d items, want 3", len(got))
	}
	wantOrder := []module.ItemId{"a", "c", "b"}
	for i, id := range wantOrder {
		if got[i].Id != id {
			t.Fatalf("item[%d] = %s, want %s", i, got[i].Id, id)
		}
	}
	assertFloatEqual(t, got[0].Score, 1.0/61.0+1.0/62.0)
	assertFloatEqual(t, got[1].Score, 1.0/61.0)
	assertFloatEqual(t, got[2].Score, 1.0/62.0)
}

func TestRRFusionFilterDefaults(t *testing.T) {
	filter := NewRRFusionFilter(recconf.FilterConfig{
		RRFConf: recconf.RRFConfig{
			DefaultWeight: 2,
			Rules:         []recconf.RRFRule{{RecallName: "r1"}},
		},
	})

	if filter.k != rrfDefaultK {
		t.Fatalf("got k %f, want %f", filter.k, rrfDefaultK)
	}
	if len(filter.configs) != 1 || filter.configs[0].Weight != 2 {
		t.Fatalf("unexpected resolved config: %+v", filter.configs)
	}

	item := newRRFTestItem("a", "r1", 10)
	data := newRRFFilterData([]*module.Item{item})
	if err := filter.Filter(data); err != nil {
		t.Fatal(err)
	}
	assertFloatEqual(t, item.Score, 2.0/61.0)

	fallback := NewRRFusionFilter(recconf.FilterConfig{
		RRFConf: recconf.RRFConfig{Rules: []recconf.RRFRule{{RecallName: "r1"}}},
	})
	if fallback.configs[0].Weight != rrfDefaultWeight {
		t.Fatalf("got default weight %f, want %f", fallback.configs[0].Weight, rrfDefaultWeight)
	}
}

func TestRRFusionFilterIgnoresUnconfiguredRecallAndTruncates(t *testing.T) {
	items := []*module.Item{
		newRRFTestItem("a", "r1", 3),
		newRRFTestItem("b", "r1", 2),
		newRRFTestItem("c", "r1", 1),
		newRRFTestItem("ignored", "r2", 100),
	}
	filter := NewRRFusionFilter(recconf.FilterConfig{
		RetainNum: 2,
		RRFConf: recconf.RRFConfig{
			Rules: []recconf.RRFRule{{RecallName: "r1", Weight: 1}},
		},
	})
	data := newRRFFilterData(items)

	if err := filter.Filter(data); err != nil {
		t.Fatal(err)
	}

	got := data.Data.([]*module.Item)
	if len(got) != 2 || got[0].Id != "a" || got[1].Id != "b" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestRRFusionFilterDeterministicTieBreak(t *testing.T) {
	filter := NewRRFusionFilter(recconf.FilterConfig{
		RRFConf: recconf.RRFConfig{
			Rules: []recconf.RRFRule{
				{RecallName: "r1", Weight: 1},
				{RecallName: "r2", Weight: 1},
			},
		},
	})
	data := newRRFFilterData([]*module.Item{
		newRRFTestItem("b", "r2", 10),
		newRRFTestItem("a", "r1", 10),
	})

	if err := filter.Filter(data); err != nil {
		t.Fatal(err)
	}

	got := data.Data.([]*module.Item)
	if len(got) != 2 || got[0].Id != "a" || got[1].Id != "b" {
		t.Fatalf("unexpected tie order: %+v", got)
	}
}

func TestRRFusionFilterDeduplicatesSameRecallItem(t *testing.T) {
	low := newRRFTestItem("a", "r1", 1)
	high := newRRFTestItem("a", "r1", 10)
	filter := NewRRFusionFilter(recconf.FilterConfig{
		RRFConf: recconf.RRFConfig{
			K:     1,
			Rules: []recconf.RRFRule{{RecallName: "r1", Weight: 1}},
		},
	})
	data := newRRFFilterData([]*module.Item{low, high})

	if err := filter.Filter(data); err != nil {
		t.Fatal(err)
	}

	got := data.Data.([]*module.Item)
	if len(got) != 1 || got[0] != high {
		t.Fatalf("unexpected deduplicated result: %+v", got)
	}
	assertFloatEqual(t, got[0].Score, 0.5)
}

func TestRRFusionFilterNoRules(t *testing.T) {
	filter := NewRRFusionFilter(recconf.FilterConfig{Name: "rrf"})
	data := newRRFFilterData([]*module.Item{newRRFTestItem("a", "r1", 1)})

	if err := filter.Filter(data); err != nil {
		t.Fatal(err)
	}
	if got := data.Data.([]*module.Item); len(got) != 0 {
		t.Fatalf("got %d items, want empty result", len(got))
	}
}

func TestRRFusionFilterRejectsInvalidData(t *testing.T) {
	filter := NewRRFusionFilter(recconf.FilterConfig{})
	if err := filter.Filter(&FilterData{Data: "invalid"}); err == nil {
		t.Fatal("expected data type error")
	}
}

func TestRRFusionFilterClone(t *testing.T) {
	filter := NewRRFusionFilter(recconf.FilterConfig{
		Name: "rrf",
		RRFConf: recconf.RRFConfig{
			K:     60,
			Rules: []recconf.RRFRule{{RecallName: "r1", Weight: 1}},
		},
	})

	cloned := filter.CloneWithConfig(map[string]interface{}{
		"RetainNum": 1,
		"RRFConf": map[string]interface{}{
			"K":             10,
			"DefaultWeight": 3,
			"Rules": []interface{}{
				map[string]interface{}{"RecallName": "r2"},
			},
		},
	})
	clone, ok := cloned.(*RRFusionFilter)
	if !ok {
		t.Fatalf("unexpected clone type %T", cloned)
	}
	if clone == filter || clone.name != "rrf" || clone.k != 10 || clone.retainNum != 1 {
		t.Fatalf("unexpected clone: %+v", clone)
	}
	if len(clone.configs) != 1 || clone.configs[0].RecallName != "r2" || clone.configs[0].Weight != 3 {
		t.Fatalf("unexpected clone configs: %+v", clone.configs)
	}
	if filter.k != 60 || filter.configs[0].RecallName != "r1" {
		t.Fatalf("original filter was modified: %+v", filter)
	}
}
