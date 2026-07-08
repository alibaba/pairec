package filter

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

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
	t.Run("test snake filter with recall SKIP_ON_DUPLICATE type", func(t *testing.T) {
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
			SnakeType: "SKIP_ON_DUPLICATE",
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
}

// TestSnakeFilterSkipNoEarlyBreak is a regression test for a bug where the
// SKIP_ON_DUPLICATE main loop terminates early (iterSize==0 -> break) and drops
// still-available items.
//
// Setup: X and Y are multi-recall items ranked in OPPOSITE order in the two
// iterators (X high in A / low in B, Y low in A / high in B). With weight=1 each
// Next(1) advances only one slot; on round 2 both iterators point at an
// already-taken duplicate at the same time, so iterSize==0 triggers break and the
// still-fresh items a1 and b1 are lost.
//
// RetainNum is 100 and there are only 4 unique items, so a correct implementation
// must keep all 4. This test FAILS before the fix (returns 2) and passes after.
func TestSnakeFilterSkipNoEarlyBreak(t *testing.T) {
	X := &module.Item{
		Id: "X", Score: 100, RetrieveId: "recall_A", Properties: map[string]interface{}{},
		RecallScores: map[string]float64{"recall_A": 100, "recall_B": 1},
	}
	Y := &module.Item{
		Id: "Y", Score: 100, RetrieveId: "recall_B", Properties: map[string]interface{}{},
		RecallScores: map[string]float64{"recall_A": 1, "recall_B": 100},
	}
	a1 := &module.Item{Id: "a1", Score: 0.5, RetrieveId: "recall_A", Properties: map[string]interface{}{}}
	b1 := &module.Item{Id: "b1", Score: 0.5, RetrieveId: "recall_B", Properties: map[string]interface{}{}}
	items := []*module.Item{X, Y, a1, b1}

	filter := NewSnakeFilter(recconf.FilterConfig{
		Name:      "snake_filter",
		RetainNum: 100,
		SnakeType: "SKIP_ON_DUPLICATE",
		AdjustCountConfs: []recconf.AdjustCountConfig{
			{RecallName: "recall_A", Weight: 1},
			{RecallName: "recall_B", Weight: 1},
		},
	})

	filterData := FilterData{Context: &context.RecommendContext{}, Data: items}
	filter.doFilter(&filterData)
	newItems := filterData.Data.([]*module.Item)

	got := make(map[string]bool, len(newItems))
	for _, item := range newItems {
		got[string(item.Id)] = true
	}
	assert.Equal(t, 4, len(newItems))
	for _, id := range []string{"X", "Y", "a1", "b1"} {
		assert.Equal(t, true, got[id])
	}
}

// TestSnakeFilterSkipStableCount is a regression test that the output count of
// SKIP_ON_DUPLICATE does not collapse in the presence of cross-recall duplicates.
//
// This mirrors the production shape (post-UniqueFilter): a dominant recall
// (recall_A) owns 100 unique items, 10 of which are multi-recall (also belong to
// recall_B). All scores are 0 (e.g. a context list carrying no score), so the
// per-iterator sort is unstable and the 10 shared items land at arbitrary
// positions. With weight=1 the main loop advances one slot per recall per round;
// once recall_B is exhausted, the first round on which recall_A happens to land
// on a shared item that recall_B already took makes iterSize==0 and the loop
// breaks -- dropping all of recall_A's remaining unique items.
//
// 100 unique items exist and RetainNum is 540, so a correct implementation keeps
// all 100. Before the fix the count collapses (far below 100); after the fix it
// is stable at 100.
func TestSnakeFilterSkipStableCount(t *testing.T) {
	var items []*module.Item
	for i := 0; i < 100; i++ {
		item := &module.Item{
			Id: module.ItemId(fmt.Sprintf("item_%d", i)), Score: 0,
			RetrieveId: "recall_A", Properties: map[string]interface{}{},
		}
		if i%10 == 0 { // item_0, item_10, ..., item_90 are also recalled by recall_B
			item.RecallScores = map[string]float64{"recall_A": 0, "recall_B": 0}
		}
		items = append(items, item)
	}

	filter := NewSnakeFilter(recconf.FilterConfig{
		Name:      "snake_filter",
		RetainNum: 540,
		SnakeType: "SKIP_ON_DUPLICATE",
		AdjustCountConfs: []recconf.AdjustCountConfig{
			{RecallName: "recall_A", Weight: 1},
			{RecallName: "recall_B", Weight: 1},
		},
	})

	filterData := FilterData{Context: &context.RecommendContext{}, Data: items}
	filter.doFilter(&filterData)
	newItems := filterData.Data.([]*module.Item)

	// 100 unique items exist; none should be dropped by dedup early-break.
	assert.Equal(t, 100, len(newItems))
}

// TestSnakeFilterZeroWeightTerminates guards the loop against a recall configured
// with weight <= 0. Next(0) never advances the iterator index, so a termination
// condition keyed purely on "all iterators exhausted" would spin forever. The
// loop must instead stop when a full round makes no progress. This test fails by
// timing out if that guarantee regresses.
func TestSnakeFilterZeroWeightTerminates(t *testing.T) {
	items := []*module.Item{
		{Id: "a1", Score: 1, RetrieveId: "recall_A", Properties: map[string]interface{}{}},
		{Id: "b1", Score: 1, RetrieveId: "recall_B", Properties: map[string]interface{}{}},
	}
	filter := NewSnakeFilter(recconf.FilterConfig{
		Name:      "snake_filter",
		RetainNum: 100,
		SnakeType: "SKIP_ON_DUPLICATE",
		AdjustCountConfs: []recconf.AdjustCountConfig{
			{RecallName: "recall_A", Weight: 1},
			{RecallName: "recall_B", Weight: 0}, // misconfigured: weight omitted -> 0
		},
	})

	done := make(chan int, 1)
	go func() {
		filterData := FilterData{Context: &context.RecommendContext{}, Data: items}
		filter.doFilter(&filterData)
		done <- len(filterData.Data.([]*module.Item))
	}()

	select {
	case n := <-done:
		// recall_A (weight 1) still yields its item; recall_B (weight 0) yields none.
		assert.Equal(t, 1, n)
	case <-time.After(3 * time.Second):
		t.Fatal("doFilter did not terminate: infinite loop with weight <= 0")
	}
}
