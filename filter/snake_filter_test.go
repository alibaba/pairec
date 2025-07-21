package filter

import (
	"fmt"
	"math/rand"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestSnakeItemIterator(t *testing.T) {
	iter := &snakeItemIterator{
		index:               0,
		recallName:          "test",
		alreadyExistItemMap: map[module.ItemId]bool{},
		config: &snakeAdjustCountConfig{
			Count:      10,
			RecallName: "test",
		},
		scoreMap:    map[module.ItemId]float64{},
		itemRankMap: map[module.ItemId][]string{},
	}

	t.Run("test sort", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				RetrieveId: "test",
				Score:      float64(i),
			}
			iter.AddItem(item)
		}
		iter.Sort()
		for i, item := range iter.items {
			assert.Equal(t, item.Id, module.ItemId(fmt.Sprintf("item_%d", 9-i)))
		}
	})
	t.Run("test next", func(t *testing.T) {
		index := 9
		for i := 0; i < 10; i++ {
			ret := iter.Next(3)
			for _, item := range ret {
				assert.Equal(t, item.Id, module.ItemId(fmt.Sprintf("item_%d", index)))
				index--
			}
		}
	})
}

func TestSnakeFilter(t *testing.T) {
	t.Run("test snake filter", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 100; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_A",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 100; i < 200; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_B",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 200; i < 300; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_C",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		t.Log(len(items))
		filter := NewSnakeFilter(recconf.FilterConfig{
			Name: "snake_filter",
			AdjustCountConfs: []recconf.AdjustCountConfig{
				{
					RecallName: "recall_A",
					Weight:     1,
				},
				{
					RecallName: "recall_B",
					Weight:     1,
				},
				{
					RecallName: "recall_C",
					Weight:     1,
				},
			},
			RetainNum: 30,
		})

		filterData := FilterData{
			Context: &context.RecommendContext{},
			Data:    items,
		}
		filter.doFilter(&filterData)
		newItems := filterData.Data.([]*module.Item)
		assert.Equal(t, 30, len(newItems))
		index := 0
		for i := 99; i >= 90; i-- {
			assert.Equal(t, fmt.Sprintf("item_%d", i), string(newItems[index].Id))
			assert.Equal(t, fmt.Sprintf("item_%d", i+100), string(newItems[index+1].Id))
			assert.Equal(t, fmt.Sprintf("item_%d", i+200), string(newItems[index+2].Id))
			index += 3
		}
	})
	t.Run("test snake filter with weight", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 100; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_A",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 100; i < 200; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_B",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 200; i < 300; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_C",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		t.Log(len(items))
		filter := NewSnakeFilter(recconf.FilterConfig{
			Name: "snake_filter",
			AdjustCountConfs: []recconf.AdjustCountConfig{
				{
					RecallName: "recall_A",
					Weight:     3,
				},
				{
					RecallName: "recall_B",
					Weight:     3,
				},
				{
					RecallName: "recall_C",
					Weight:     4,
				},
			},
			RetainNum: 20,
		})

		filterData := FilterData{
			Context: &context.RecommendContext{},
			Data:    items,
		}
		filter.doFilter(&filterData)
		newItems := filterData.Data.([]*module.Item)
		assert.Equal(t, 20, len(newItems))
		size := 0
		for size < len(newItems) {
			for _, config := range filter.configs {
				end := size + config.Weight
				items := newItems[size:end]
				for _, item := range items {
					assert.Equal(t, config.RecallName, item.RetrieveId)
				}
				size = end
			}
		}
		for _, item := range newItems {
			t.Log(item)
		}
	})
	t.Run("test snake filter with no recall", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 100; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_A",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 200; i < 300; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_C",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		t.Log(len(items))
		filter := NewSnakeFilter(recconf.FilterConfig{
			Name: "snake_filter",
			AdjustCountConfs: []recconf.AdjustCountConfig{
				{
					RecallName: "recall_A",
					Weight:     3,
				},
				{
					RecallName: "recall_B",
					Weight:     3,
				},
				{
					RecallName: "recall_C",
					Weight:     4,
				},
			},
			RetainNum: 20,
		})

		filterData := FilterData{
			Context: &context.RecommendContext{},
			Data:    items,
		}
		filter.doFilter(&filterData)
		newItems := filterData.Data.([]*module.Item)
		assert.Equal(t, 20, len(newItems))
		items = newItems[20-3 : 20]
		for _, item := range items {
			assert.Equal(t, "recall_C", item.RetrieveId)
		}
		items = newItems[20-6 : 20-3]
		for _, item := range items {
			assert.Equal(t, "recall_A", item.RetrieveId)
		}
	})
	t.Run("test snake filter with recall merge", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 100; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_A",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 0; i < 20; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_B",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 20; i < 40; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_C",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
		uniqFilter := NewUniqueFilter()
		filterData := FilterData{
			Context: &context.RecommendContext{},
			Data:    items,
		}
		uniqFilter.doFilter(&filterData)
		filter := NewSnakeFilter(recconf.FilterConfig{
			Name: "snake_filter",
			AdjustCountConfs: []recconf.AdjustCountConfig{
				{
					RecallName: "recall_A",
					Weight:     1,
				},
				{
					RecallName: "recall_B",
					Weight:     1,
				},
				{
					RecallName: "recall_C",
					Weight:     2,
				},
			},
			RetainNum: 20,
		})
		filterData = FilterData{
			Context: &context.RecommendContext{},
			Data:    filterData.Data,
		}

		filter.doFilter(&filterData)
		newItems := filterData.Data.([]*module.Item)
		assert.Equal(t, 20, len(newItems))
		size := 0
		for size < len(newItems) {
			for _, config := range filter.configs {
				end := size + config.Weight
				items := newItems[size:end]
				for _, item := range items {
					assert.Equal(t, config.RecallName, item.RetrieveId)
				}
				size = end
			}
		}
	})
	t.Run("test snake filter with recall merge v2", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 100; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_A",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 95; i < 105; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_B",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 98; i < 108; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_C",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
		uniqFilter := NewUniqueFilter()
		filterData := FilterData{
			Context: &context.RecommendContext{},
			Data:    items,
		}
		uniqFilter.doFilter(&filterData)
		filter := NewSnakeFilter(recconf.FilterConfig{
			Name: "snake_filter",
			AdjustCountConfs: []recconf.AdjustCountConfig{
				{
					RecallName: "recall_A",
					Weight:     1,
				},
				{
					RecallName: "recall_B",
					Weight:     2,
				},
				{
					RecallName: "recall_C",
					Weight:     3,
				},
			},
			RetainNum: 60,
		})
		filterData = FilterData{
			Context: &context.RecommendContext{},
			Data:    filterData.Data,
		}

		filter.doFilter(&filterData)
		newItems := filterData.Data.([]*module.Item)
		assert.Equal(t, 60, len(newItems))
		for _, item := range newItems {
			t.Log(item)
		}
	})
	t.Run("test snake filter with recall not config", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 100; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_A",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 100; i < 200; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_B",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 200; i < 300; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_C",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		for i := 300; i < 400; i++ {
			item := &module.Item{
				Id:         module.ItemId(fmt.Sprintf("item_%d", i)),
				Score:      float64(i),
				RetrieveId: "recall_D",
				Properties: map[string]interface{}{},
			}
			items = append(items, item)
		}
		filter := NewSnakeFilter(recconf.FilterConfig{
			Name: "snake_filter",
			AdjustCountConfs: []recconf.AdjustCountConfig{
				{
					RecallName: "recall_A",
					Weight:     1,
				},
				{
					RecallName: "recall_B",
					Weight:     2,
				},
				{
					RecallName: "recall_C",
					Weight:     3,
				},
			},
			RetainNum: 20,
		})

		filterData := FilterData{
			Context: &context.RecommendContext{},
			Data:    items,
		}
		filter.doFilter(&filterData)
		newItems := filterData.Data.([]*module.Item)
		assert.Equal(t, 20, len(newItems))
		size := 0
		for size < len(newItems) {
			for _, config := range filter.configs {
				end := size + config.Weight
				items := newItems[size:end]
				for _, item := range items {
					assert.Equal(t, config.RecallName, item.RetrieveId)
				}
				size = end
			}
		}
	})
}
