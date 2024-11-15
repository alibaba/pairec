package filter

import (
	"fmt"
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestPriorityAdjustCountFilterV2_FixCount(t *testing.T) {
	PAFilter := NewPriorityAdjustCountFilterV2(recconf.FilterConfig{
		Name: "PAF_fix",
		AdjustCountConfs: []recconf.AdjustCountConfig{
			{
				RecallName: "u2i",
				Type:       Fix_Count_Type,
				Count:      5,
			},
			{
				RecallName: "hot",
				Type:       Fix_Count_Type,
				Count:      5,
			},
		}})

	user := module.NewUser("user_1")

	var items []*module.Item

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_a_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "hot"
		items = append(items, item)
	}

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_b_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "u2i"
		items = append(items, item)
	}

	filterData := &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}

	PAFilter.Filter(filterData)

	filterItems := filterData.Data.([]*module.Item)

	if len(filterItems) != 10 {
		t.Error("len(items) != 10", len(filterItems))
	}

	for i := 0; i < 5; i++ {
		item := filterItems[i]
		if item.RetrieveId != "u2i" {
			t.Error("item.RetrieveId != u2i")
		}
		id := fmt.Sprintf("item_b_%d", 9-i)
		if string(item.Id) != id {
			t.Error("wrong item.Id")
		}
	}

	for i := 5; i < 10; i++ {
		item := filterItems[i]
		if item.RetrieveId != "hot" {
			t.Error("item.RetrieveId != hot")
		}
		id := fmt.Sprintf("item_a_%d", 14-i)
		if string(item.Id) != id {
			t.Error("wrong item.Id")
		}
	}
}

func TestPriorityAdjustCountFilterV2_AccumulateCount(t *testing.T) {
	PAFilter := NewPriorityAdjustCountFilterV2(recconf.FilterConfig{
		Name: "PAF_fix",
		AdjustCountConfs: []recconf.AdjustCountConfig{
			{
				RecallName: "u2i",
				Type:       Accumulate_Count_Type,
				Count:      15,
			},
			{
				RecallName: "hot",
				Type:       Accumulate_Count_Type,
				Count:      15,
			},
		}})

	user := module.NewUser("user_1")

	var items []*module.Item

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_a_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "hot"
		items = append(items, item)
	}

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_b_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "u2i"
		items = append(items, item)
	}

	filterData := &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}

	PAFilter.Filter(filterData)

	filterItems := filterData.Data.([]*module.Item)

	if len(filterItems) != 15 {
		t.Error("len(items) != 15", len(filterItems))
	}

	for i := 0; i < 10; i++ {
		item := filterItems[i]
		if item.RetrieveId != "u2i" {
			t.Error("item.RetrieveId != u2i")
		}
		id := fmt.Sprintf("item_b_%d", 9-i)
		if string(item.Id) != id {
			t.Error("wrong item.Id")
		}
	}

	for i := 10; i < 15; i++ {
		item := filterItems[i]
		if item.RetrieveId != "hot" {
			t.Error("item.RetrieveId != hot")
		}
		id := fmt.Sprintf("item_a_%d", 19-i)
		if string(item.Id) != id {
			t.Error("wrong item.Id")
		}
	}
}

func TestPriorityAdjustCountFilterV2_AccumulateCount_Unique(t *testing.T) {
	PAFilter := NewPriorityAdjustCountFilterV2(recconf.FilterConfig{
		Name: "PAF_fix",
		AdjustCountConfs: []recconf.AdjustCountConfig{
			{
				RecallName: "u2i",
				Type:       Accumulate_Count_Type,
				Count:      15,
			},
			{
				RecallName: "hot",
				Type:       Accumulate_Count_Type,
				Count:      15,
			},
		}})

	user := module.NewUser("user_1")

	var items []*module.Item

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_a_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "hot"
		items = append(items, item)
	}

	duplicateRecallItem1 := module.NewItem("item_a&b")
	duplicateRecallItem1.Score = -1
	duplicateRecallItem1.RetrieveId = "hot"
	items = append(items, duplicateRecallItem1)

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_b_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "u2i"
		items = append(items, item)
	}

	duplicateRecallItem2 := module.NewItem("item_a&b")
	duplicateRecallItem2.Score = 11
	duplicateRecallItem2.RetrieveId = "u2i"
	items = append(items, duplicateRecallItem2)

	filterData := &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}

	NewUniqueFilter().Filter(filterData)
	PAFilter.Filter(filterData)

	filterItems := filterData.Data.([]*module.Item)

	if len(filterItems) != 15 {
		t.Error("len(items) != 15", len(filterItems))
	}

	if filterItems[0].Id != "item_a&b" || filterItems[0].RetrieveId != "u2i" {
		t.Error("wrong first item", filterItems[0])
	}

	for i := 1; i < 11; i++ {
		item := filterItems[i]
		if item.RetrieveId != "u2i" {
			t.Error("item.RetrieveId != u2i")
		}
		id := fmt.Sprintf("item_b_%d", 10-i)
		if string(item.Id) != id {
			t.Error("wrong item.Id")
		}
	}

	for i := 11; i < 15; i++ {
		item := filterItems[i]
		if item.RetrieveId != "hot" {
			t.Error("item.RetrieveId != hot")
		}
		id := fmt.Sprintf("item_a_%d", 20-i)
		if string(item.Id) != id {
			t.Error("wrong item.Id")
		}
	}
}
