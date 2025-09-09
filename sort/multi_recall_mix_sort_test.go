package sort

import (
	"strconv"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestFixPositionStrategyByRecallName(t *testing.T) {
	t.Run("fix position", func(t *testing.T) {
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
		user := module.NewUser("user_1")

		for size := 10; size < 20; size++ {
			context := context.NewRecommendContext()
			context.Size = size
			sortData := SortData{Data: items, Context: context, User: user}

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

	})
	t.Run("fix position by order", func(t *testing.T) {
		positions := []int{1, 3}
		config := recconf.SortConfig{
			MixSortRules: []recconf.MixSortConfig{
				{
					MixStrategy: "fix_position",
					Positions:   positions,
					RecallNames: []string{"r1"},
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "tag",
							Operator: "equal",
							Value:    "tag1",
							Domain:   "item",
						},
					},
				},
			},
		}

		var items []*module.Item
		for i := 0; i < 10; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i) * 10
			item.RetrieveId = "r1"

			item.AddProperty("tag", "tag1")
			items = append(items, item)
		}
		for i := 10; i < 20; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r2"

			item.AddProperty("tag", "tag2")
			items = append(items, item)
		}
		user := module.NewUser("user_1")

		for size := 10; size < 11; size++ {
			context := context.NewRecommendContext()
			context.Size = size
			sortData := SortData{Data: items, Context: context, User: user}

			NewItemRankScoreSort().Sort(&sortData)
			result := sortData.Data.([]*module.Item)
			for _, item := range result {
				t.Log(item)
			}

			NewMultiRecallMixSort(config).Sort(&sortData)

			result = sortData.Data.([]*module.Item)

			if result[0].GetRecallName() != "r1" {
				t.Error("item position error")
			}

			assert.Equal(t, result[0].Score, float64(90))
			assert.Equal(t, result[1].Score, float64(70))
			assert.Equal(t, result[2].Score, float64(80))
			for _, item := range result {
				if item == nil {
					t.Error("item has nil item")
				}
			}
		}

	})

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

	user := module.NewUser("user_1")
	for size := 10; size <= 20; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context, User: user}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		if len(result) != context.Size {
			t.Error("items len error", len(items), context.Size)
		}
		for _, item := range result {
			if size == 10 {
				t.Log(item)
			}
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

	user := module.NewUser("user_1")
	for size := 10; size <= 30; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context, User: user}

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
						Domain:   "item",
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
		item.Score = float64(i)

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "woman")
		item.RetrieveId = "r2"
		item.Score = float64(i)

		items = append(items, item)
	}

	user := module.NewUser("user_1")
	for size := 10; size < 20; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context, User: user}

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
		item.AddProperty("sex", "man")
		item.RetrieveId = "r1"
		item.Score = float64(i)

		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.AddProperty("sex", "woman")
		item.RetrieveId = "r2"
		item.Score = float64(i)

		items = append(items, item)
	}

	user := module.NewUser("user_1")
	for size := 10; size <= 20; size++ {
		context := context.NewRecommendContext()
		context.Size = size
		sortData := SortData{Data: items, Context: context, User: user}

		NewItemRankScoreSort().Sort(&sortData)

		NewMultiRecallMixSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		if len(result) != context.Size {
			t.Error("items len error", len(items), context.Size)
		}
		for _, item := range result {
			if size == 10 {
				t.Log(item)
			}
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
