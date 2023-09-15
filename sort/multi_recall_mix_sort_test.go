package sort

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

func TestFixPositionStrategyByRecallName(t *testing.T) {
	positions := []int{1, 4, 6, 8, 10}
	config := recconf.SortConfig{
		MixSortRules: []recconf.MixSortConfig{
			{
				MixStrategy: "fix_position",
				Positions:   positions,
				RecallNames: []string{"r1"},
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r2"

		items = append(items, item)
	}

	for size := 10; size < 20; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		size := 10
		for _, pos := range positions {
			if result[pos-1].GetRecallName() != "r1" {
				t.Error("item position error")
			}

			if result[pos-1].Score != float64(size-1) {
				t.Error("item score error")
			}
			size--
		}
		for _, item := range result {
			if item == nil {
				t.Error("item has nil item")
			}
		}
	}

}

func TestRandomPositionStrategyByRecallName(t *testing.T) {
	config := recconf.SortConfig{
		MixSortRules: []recconf.MixSortConfig{
			{
				MixStrategy: "random_position",
				NumberRate:  0.3,
				RecallNames: []string{"r1"},
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r2"

		items = append(items, item)
	}

	for size := 10; size <= 20; size++ {
		fmt.Println("start")
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		if len(result) != context.Size {
			t.Error("items len error", len(items), context.Size)
		}
		for _, item := range result {
			if item == nil {
				t.Error("item has nil item")
			}
		}

	}
}

func TestRemainItemByRecallName(t *testing.T) {
	config := recconf.SortConfig{
		RemainItem: true,
		MixSortRules: []recconf.MixSortConfig{
			{
				MixStrategy: "random_position",
				NumberRate:  0.1,
				RecallNames: []string{"r1", "r4"},
			},
			{
				MixStrategy: "random_position",
				NumberRate:  0.1,
				RecallNames: []string{"r2"},
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r2"

		items = append(items, item)
	}

	for i := 20; i < 30; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r3"

		items = append(items, item)
	}
	for i := 30; i < 40; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r4"

		items = append(items, item)
	}

	for size := 10; size <= 30; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		if len(result) != len(items) {
			t.Error("items len error")
		}
		for _, item := range result {
			if item == nil {
				t.Error("item has nil item")
			}
		}

	}
}

func TestFixPositionStrategyByCondition(t *testing.T) {
	positions := []int{1, 4, 6, 8, 10}
	config := recconf.SortConfig{
		MixSortRules: []recconf.MixSortConfig{
			{
				MixStrategy: "fix_position",
				Positions:   positions,
				Conditions: []recconf.FilterParamConfig{
					{
						Name:     "sex",
						Type:     "string",
						Value:    "man",
						Operator: "equal",
					},
				},
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "man")
		item.RetrieveId = "r1"

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "woman")
		item.RetrieveId = "r2"

		items = append(items, item)
	}

	for size := 10; size < 20; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		size := 10
		for _, pos := range positions {
			if result[pos-1].StringProperty("sex") != "man" {
				t.Error("item position error")
			}
			size--
		}
		for _, item := range result {
			if item == nil {
				t.Error("item has nil item")
			}
		}
	}

}

func TestRandomPositionStrategyByCondition(t *testing.T) {
	config := recconf.SortConfig{
		MixSortRules: []recconf.MixSortConfig{
			{
				MixStrategy: "random_position",
				NumberRate:  0.3,
				Conditions: []recconf.FilterParamConfig{
					{
						Name:     "sex",
						Type:     "string",
						Value:    "man",
						Operator: "equal",
					},
				},
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "woman")
		item.RetrieveId = "r1"

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "woman")
		item.RetrieveId = "r2"

		items = append(items, item)
	}

	for size := 10; size <= 20; size++ {
		fmt.Println("start")
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		if len(result) != context.Size {
			t.Error("items len error", len(items), context.Size)
		}
		for _, item := range result {
			if item == nil {
				t.Error("item has nil item")
			}
		}

	}
}

func TestRemainItemByCondition(t *testing.T) {
	config := recconf.SortConfig{
		RemainItem: true,
		MixSortRules: []recconf.MixSortConfig{
			{
				MixStrategy: "random_position",
				NumberRate:  0.1,
				Conditions: []recconf.FilterParamConfig{
					{
						Name:     "sex",
						Type:     "string",
						Value:    "man",
						Operator: "equal",
					},
				},
			},
			{
				MixStrategy: "random_position",
				NumberRate:  0.1,
				Conditions: []recconf.FilterParamConfig{
					{
						Name:     "sex",
						Type:     "string",
						Value:    "woman",
						Operator: "equal",
					},
				},
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "man")
		item.RetrieveId = "r1"

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "woman")
		item.RetrieveId = "r2"

		items = append(items, item)
	}

	for size := 10; size <= 30; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		if len(result) != len(items) {
			t.Error("items len error")
		}
		for _, item := range result {
			if item == nil {
				t.Error("item has nil item")
			}
		}

	}
}
