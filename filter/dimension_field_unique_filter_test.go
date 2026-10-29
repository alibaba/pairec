package filter

import (
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func newUniqueTestItem(id, brand string, props map[string]interface{}) *module.Item {
	item := module.NewItem(id)
	if brand != "" {
		item.AddProperty("brand_id", brand)
	}
	for k, v := range props {
		item.AddProperty(k, v)
	}
	return item
}

func runDimensionFieldUniqueFilter(t *testing.T, config recconf.FilterConfig, items []*module.Item) []*module.Item {
	t.Helper()
	filter := NewDimensionFieldUniqueFilter(config)
	filterData := FilterData{Context: &context.RecommendContext{}, Data: items}
	if err := filter.doFilter(&filterData); err != nil {
		t.Fatal(err)
	}
	return filterData.Data.([]*module.Item)
}

func TestDimensionFieldUniqueFilterKeepFirst(t *testing.T) {
	// without UniqueKeepConf, keep the first occurrence of each dimension value (legacy behavior)
	items := []*module.Item{
		newUniqueTestItem("item_1", "brand_a", map[string]interface{}{"distance": 3.0}),
		newUniqueTestItem("item_2", "brand_b", nil),
		newUniqueTestItem("item_3", "brand_a", map[string]interface{}{"distance": 1.0}),
		newUniqueTestItem("item_4", "", map[string]interface{}{"distance": 0.5}),
		newUniqueTestItem("item_5", "brand_b", map[string]interface{}{"distance": 2.0}),
	}

	newItems := runDimensionFieldUniqueFilter(t, recconf.FilterConfig{
		Name:      "brand_unique_filter",
		Dimension: "brand_id",
	}, items)

	// item_4 has empty dimension value, it is kept and never deduplicated
	assert.Equal(t, 3, len(newItems))
	assert.Equal(t, module.ItemId("item_1"), newItems[0].Id)
	assert.Equal(t, module.ItemId("item_2"), newItems[1].Id)
	assert.Equal(t, module.ItemId("item_4"), newItems[2].Id)
}

func TestDimensionFieldUniqueFilterKeepMin(t *testing.T) {
	items := []*module.Item{
		newUniqueTestItem("item_1", "brand_a", map[string]interface{}{"distance": 3.0}),
		newUniqueTestItem("item_2", "brand_b", map[string]interface{}{"distance": 9.0}),
		newUniqueTestItem("item_3", "brand_a", map[string]interface{}{"distance": 1.0}),
		newUniqueTestItem("item_4", "brand_a", map[string]interface{}{"distance": 2.0}),
	}

	newItems := runDimensionFieldUniqueFilter(t, recconf.FilterConfig{
		Name:      "brand_unique_filter",
		Dimension: "brand_id",
		UniqueKeepConf: recconf.UniqueKeepConfig{
			Property: "distance",
			Strategy: "min",
		},
	}, items)

	// the winner replaces the group in place, so output positions follow first occurrences
	assert.Equal(t, 2, len(newItems))
	assert.Equal(t, module.ItemId("item_3"), newItems[0].Id)
	assert.Equal(t, module.ItemId("item_2"), newItems[1].Id)
}

func TestDimensionFieldUniqueFilterKeepMax(t *testing.T) {
	items := []*module.Item{
		newUniqueTestItem("item_1", "brand_a", map[string]interface{}{"score": 0.3}),
		newUniqueTestItem("item_2", "brand_a", map[string]interface{}{"score": 0.9}),
		newUniqueTestItem("item_3", "brand_a", map[string]interface{}{"score": 0.6}),
	}

	newItems := runDimensionFieldUniqueFilter(t, recconf.FilterConfig{
		Name:      "brand_unique_filter",
		Dimension: "brand_id",
		UniqueKeepConf: recconf.UniqueKeepConfig{
			Property: "score",
			Strategy: "max",
		},
	}, items)

	assert.Equal(t, 1, len(newItems))
	assert.Equal(t, module.ItemId("item_2"), newItems[0].Id)
}

func TestDimensionFieldUniqueFilterKeepMissingProperty(t *testing.T) {
	items := []*module.Item{
		// item_1 has no distance property, it loses to any item with a comparable value
		newUniqueTestItem("item_1", "brand_a", nil),
		newUniqueTestItem("item_2", "brand_a", map[string]interface{}{"distance": 5.0}),
		// item_4 has no distance property either, both sides keep the first occurring one
		newUniqueTestItem("item_3", "brand_b", nil),
		newUniqueTestItem("item_4", "brand_b", nil),
	}

	newItems := runDimensionFieldUniqueFilter(t, recconf.FilterConfig{
		Name:      "brand_unique_filter",
		Dimension: "brand_id",
		UniqueKeepConf: recconf.UniqueKeepConfig{
			Property: "distance",
			Strategy: "min",
		},
	}, items)

	assert.Equal(t, 2, len(newItems))
	assert.Equal(t, module.ItemId("item_2"), newItems[0].Id)
	assert.Equal(t, module.ItemId("item_3"), newItems[1].Id)
}

func TestDimensionFieldUniqueFilterKeepTieAndStringNumber(t *testing.T) {
	items := []*module.Item{
		newUniqueTestItem("item_1", "brand_a", map[string]interface{}{"distance": "2.5"}),
		// tie keeps the first occurring item
		newUniqueTestItem("item_2", "brand_a", map[string]interface{}{"distance": 2.5}),
		// numeric string is comparable via FloatProperty
		newUniqueTestItem("item_3", "brand_a", map[string]interface{}{"distance": "1.5"}),
	}

	newItems := runDimensionFieldUniqueFilter(t, recconf.FilterConfig{
		Name:      "brand_unique_filter",
		Dimension: "brand_id",
		UniqueKeepConf: recconf.UniqueKeepConfig{
			Property: "distance",
			Strategy: "min",
		},
	}, items)

	assert.Equal(t, 1, len(newItems))
	assert.Equal(t, module.ItemId("item_3"), newItems[0].Id)
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic, got none", name)
		}
	}()
	fn()
}

func TestNewDimensionFieldUniqueFilterInvalidConfig(t *testing.T) {
	assertPanics(t, "invalid strategy", func() {
		NewDimensionFieldUniqueFilter(recconf.FilterConfig{
			Name:      "brand_unique_filter",
			Dimension: "brand_id",
			UniqueKeepConf: recconf.UniqueKeepConfig{
				Property: "distance",
				Strategy: "asc",
			},
		})
	})

	// strategy without property
	assertPanics(t, "strategy without property", func() {
		NewDimensionFieldUniqueFilter(recconf.FilterConfig{
			Name:      "brand_unique_filter",
			Dimension: "brand_id",
			UniqueKeepConf: recconf.UniqueKeepConfig{
				Strategy: "min",
			},
		})
	})

	// property without strategy
	assertPanics(t, "property without strategy", func() {
		NewDimensionFieldUniqueFilter(recconf.FilterConfig{
			Name:      "brand_unique_filter",
			Dimension: "brand_id",
			UniqueKeepConf: recconf.UniqueKeepConfig{
				Property: "distance",
			},
		})
	})
}
