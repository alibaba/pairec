package sort

import (
	"strconv"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// mockSortParam mimics the standard RecommendParam: only the "features" name
// is exposed, custom fields must be looked up inside the features map.
type mockSortParam struct {
	features map[string]interface{}
}

func (p *mockSortParam) GetParameter(name string) interface{} {
	if name == "features" {
		return p.features
	}
	return nil
}

// stubDiversityDao returns preset dimension values without any datasource.
type stubDiversityDao struct {
	values map[string]map[string]interface{}
	fields []string
}

func (d *stubDiversityDao) GetDistinctValue(items []*module.Item, ctx *context.RecommendContext) error {
	for _, item := range items {
		if vals, ok := d.values[string(item.Id)]; ok {
			item.AddProperties(vals)
		}
	}
	return nil
}

func (d *stubDiversityDao) GetDistinctFields() []string {
	return d.fields
}

func newTagItem(id, tag string) *module.Item {
	item := module.NewItem(id)
	item.RetrieveId = "r1"
	item.AddProperty("tag", tag)
	return item
}

func TestDiversityRuleV2MatchWithHistory(t *testing.T) {
	t.Run("interval_size_cross_history", func(t *testing.T) {
		rule := NewDiversityRuleV2(recconf.DiversityRuleConfig{
			Dimensions:   []string{"tag"},
			IntervalSize: 2,
		}, 10)
		rule.SetHistory([]*module.Item{newTagItem("p1", "t1"), newTagItem("p2", "t1")})
		assert.Equal(t, 2, rule.HistoryLen())

		// result is empty, but the history tail has 2 consecutive "t1"
		assert.Equal(t, false, rule.Match(newTagItem("c1", "t1"), nil))
		assert.Equal(t, true, rule.Match(newTagItem("c2", "t2"), nil))
	})

	t.Run("window_size_cross_history", func(t *testing.T) {
		rule := NewDiversityRuleV2(recconf.DiversityRuleConfig{
			Dimensions:    []string{"tag"},
			WindowSize:    6,
			FrequencySize: 1,
		}, 10)
		rule.SetHistory([]*module.Item{
			newTagItem("p1", "t9"), newTagItem("p2", "t8"), newTagItem("p3", "t1"),
			newTagItem("p4", "t7"), newTagItem("p5", "t6"),
		})
		assert.Equal(t, 5, rule.HistoryLen())

		// "t1" appears in the history window -> frequency exceeded
		assert.Equal(t, false, rule.Match(newTagItem("c1", "t1"), nil))
		assert.Equal(t, true, rule.Match(newTagItem("c2", "t2"), nil))

		// window slides as result grows: with 3 items in result, only the
		// last 2 history values (t7, t6) stay in the window of size 6
		result := []*module.Item{newTagItem("r1", "a"), newTagItem("r2", "b"), newTagItem("r3", "c")}
		assert.Equal(t, true, rule.Match(newTagItem("c3", "t1"), result))
		assert.Equal(t, false, rule.Match(newTagItem("c4", "t7"), result))
	})

	t.Run("history_truncated_to_max_len", func(t *testing.T) {
		rule := NewDiversityRuleV2(recconf.DiversityRuleConfig{
			Dimensions:    []string{"tag"},
			WindowSize:    3,
			FrequencySize: 1,
		}, 10)
		var prevItems []*module.Item
		for i := 0; i < 8; i++ {
			prevItems = append(prevItems, newTagItem("p"+strconv.Itoa(i), "t"+strconv.Itoa(i)))
		}
		rule.SetHistory(prevItems)
		// maxLen = WindowSize - 1 = 2, only the last 2 are kept
		assert.Equal(t, 2, rule.HistoryLen())
		assert.Equal(t, false, rule.Match(newTagItem("c1", "t7"), nil))
		assert.Equal(t, true, rule.Match(newTagItem("c2", "t0"), nil))
	})

	t.Run("empty_history_equals_v1", func(t *testing.T) {
		config := recconf.DiversityRuleConfig{
			Dimensions:    []string{"tag"},
			WindowSize:    4,
			FrequencySize: 2,
			IntervalSize:  1,
		}
		v1 := NewDiversityRule(config, 10)
		v2 := NewDiversityRuleV2(config, 10)

		var result []*module.Item
		for _, tag := range []string{"t1", "t2", "t1", "t1", "t2", "t3"} {
			item := newTagItem("i"+strconv.Itoa(len(result)), tag)
			assert.Equal(t, v1.Match(item, result), v2.Match(item, result))
			result = append(result, item)
		}
	})
}

func TestDiversityRuleMultiDimensionV2MatchWithHistory(t *testing.T) {
	multiDimensionMap := map[int]recconf.MultiValueDimensionConfig{
		0: {DimensionName: "tag", Delimiter: "/"},
	}
	rule := NewDiversityRuleMultiDimensionV2(recconf.DiversityRuleConfig{
		Dimensions:   []string{"tag"},
		IntervalSize: 1,
	}, 10, multiDimensionMap)
	rule.SetHistory([]*module.Item{newTagItem("p1", "b/c")})
	assert.Equal(t, 1, rule.HistoryLen())

	// "a/b" intersects with history tail "b/c" -> rejected
	assert.Equal(t, false, rule.Match(newTagItem("c1", "a/b"), nil))
	assert.Equal(t, true, rule.Match(newTagItem("c2", "d/e"), nil))
}

func TestParseLastPageItemIds(t *testing.T) {
	assert.Equal(t, []string{"1", "2", "3"}, parseLastPageItemIds([]interface{}{"1", "2", "3"}))
	assert.Equal(t, []string{"1", "2"}, parseLastPageItemIds([]interface{}{"1", 2}))
	assert.Equal(t, []string{"1", "2"}, parseLastPageItemIds(`["1","2"]`))
	assert.Equal(t, []string{"1", "2"}, parseLastPageItemIds("1, 2"))
	assert.Equal(t, 0, len(parseLastPageItemIds("")))
	assert.Equal(t, 0, len(parseLastPageItemIds(nil)))
	assert.Equal(t, 0, len(parseLastPageItemIds(123)))
	// invalid json degrades to comma split
	assert.Equal(t, []string{"[1", "2"}, parseLastPageItemIds("[1,2"))
}

func buildCrossPageSortV2(config recconf.SortConfig, dao module.DiversityDao) *DiversityRuleSortV2 {
	s := NewDiversityRuleSortV2(config)
	s.crossPageEnable = true
	s.lastItemIdsParam = defaultLastItemIdsParam
	s.dimLoader = dao
	return s
}

func TestDiversityRuleSortV2CrossPage(t *testing.T) {
	config := recconf.SortConfig{
		DiversityRules: []recconf.DiversityRuleConfig{
			{
				Dimensions:   []string{"tag"},
				IntervalSize: 2,
			},
		},
	}
	dao := &stubDiversityDao{
		values: map[string]map[string]interface{}{
			"p1": {"tag": "t1"},
			"p2": {"tag": "t1"},
		},
		fields: []string{"tag"},
	}

	newItems := func() []*module.Item {
		var items []*module.Item
		for i := 0; i < 6; i++ {
			tag := "t1"
			if i >= 3 {
				tag = "t2"
			}
			items = append(items, newTagItem(strconv.Itoa(i), tag))
		}
		return items
	}

	t.Run("warmup_constrains_first_item", func(t *testing.T) {
		ctx := context.NewRecommendContext()
		ctx.Size = 6
		ctx.Param = &mockSortParam{features: map[string]interface{}{
			"last_page_item_ids": []interface{}{"p1", "p2"},
		}}
		sortData := SortData{Data: newItems(), Context: ctx}

		err := buildCrossPageSortV2(config, dao).Sort(&sortData)
		assert.Equal(t, nil, err)

		result := sortData.Data.([]*module.Item)
		// history tail has 2 consecutive "t1" and IntervalSize=2, so the
		// first item of this page must not be "t1"
		assert.Equal(t, "t2", result[0].StringProperty("tag"))
	})

	t.Run("no_param_behaves_like_v1", func(t *testing.T) {
		ctx := context.NewRecommendContext()
		ctx.Size = 6
		ctx.Param = &mockSortParam{features: map[string]interface{}{}}
		sortData := SortData{Data: newItems(), Context: ctx}

		err := buildCrossPageSortV2(config, dao).Sort(&sortData)
		assert.Equal(t, nil, err)

		result := sortData.Data.([]*module.Item)
		assert.Equal(t, "0", string(result[0].Id))
	})
}

// TestDiversityRuleSortV2Alignment verifies that V2 without cross page config
// produces exactly the same order as V1 for the same input.
func TestDiversityRuleSortV2Alignment(t *testing.T) {
	configs := []recconf.SortConfig{
		{
			DiversityRules: []recconf.DiversityRuleConfig{
				{Dimensions: []string{"tag"}, IntervalSize: 1},
			},
		},
		{
			DiversityRules: []recconf.DiversityRuleConfig{
				{Dimensions: []string{"tag"}, WindowSize: 4, FrequencySize: 2},
			},
		},
		{
			DiversityRules: []recconf.DiversityRuleConfig{
				{Dimensions: []string{"tag"}, IntervalSize: 2, Weight: 10},
				{Dimensions: []string{"category"}, WindowSize: 3, FrequencySize: 1, Weight: 5},
			},
		},
		{
			DiversityRules: []recconf.DiversityRuleConfig{
				{Dimensions: []string{"tag"}, WindowSize: 5, FrequencySize: 1},
			},
			MultiValueDimensionConf: []recconf.MultiValueDimensionConfig{
				{DimensionName: "tag", Delimiter: "/"},
			},
		},
	}

	newItems := func() []*module.Item {
		var items []*module.Item
		tags := []string{"t1/a", "t1/b", "t2/a", "t1/a", "t3/c", "t2/b", "t1/c", "t3/a", "t2/c", "t1/b"}
		for i, tag := range tags {
			item := module.NewItem(strconv.Itoa(i))
			item.RetrieveId = "r1"
			item.AddProperty("tag", tag)
			item.AddProperty("category", "c"+strconv.Itoa(i%3))
			items = append(items, item)
		}
		return items
	}

	for i, config := range configs {
		ctxV1 := context.NewRecommendContext()
		ctxV1.Size = 10
		sortDataV1 := SortData{Data: newItems(), Context: ctxV1}
		if err := NewDiversityRuleSort(config).Sort(&sortDataV1); err != nil {
			t.Fatalf("config %d v1 sort error: %v", i, err)
		}

		ctxV2 := context.NewRecommendContext()
		ctxV2.Size = 10
		sortDataV2 := SortData{Data: newItems(), Context: ctxV2}
		if err := NewDiversityRuleSortV2(config).Sort(&sortDataV2); err != nil {
			t.Fatalf("config %d v2 sort error: %v", i, err)
		}

		resultV1 := sortDataV1.Data.([]*module.Item)
		resultV2 := sortDataV2.Data.([]*module.Item)
		assert.Equal(t, len(resultV1), len(resultV2))
		for j := range resultV1 {
			if resultV1[j].Id != resultV2[j].Id {
				t.Fatalf("config %d result mismatch at %d: v1=%s v2=%s", i, j, resultV1[j].Id, resultV2[j].Id)
			}
		}
	}
}

// TestNewDiversityDaoByConfUnknownType ensures the factory degrades to nil
// instead of panicking for unsupported adapter types.
func TestNewDiversityDaoByConfUnknownType(t *testing.T) {
	dao := module.NewDiversityDaoByConf(recconf.DiversityDaoConfig{
		DaoConfig: recconf.DaoConfig{AdapterType: "unknown"},
	})
	assert.Equal(t, nil, dao)
}
