package sort

import (
	"fmt"
	"strconv"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestDiversitryRuleSortByIntervalSize(t *testing.T) {
	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		if i < 5 {
			item.AddProperty("tag", "t1")
		} else {
			item.AddProperty("tag", "t2")
		}

		items = append(items, item)
	}
	t.Run("interval_size", func(t *testing.T) {
		config := recconf.SortConfig{
			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:   []string{"tag"},
					IntervalSize: 1,
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i, item := range result {
			if i%2 == 0 && item.StringProperty("tag") != "t1" {
				t.Error("item error")
			}
		}
	})
	t.Run("interval_size_more", func(t *testing.T) {
		config := recconf.SortConfig{
			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:   []string{"tag"},
					IntervalSize: 2,
				},
			},
		}

		fmt.Println("====sort before====")
		for _, item := range items {
			t.Log(item)
		}
		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		fmt.Println("====sort after====")
		for i, item := range result {
			t.Log(item)
			if i == 2 || i == 5 {
				assert.Equal(t, "t2", item.StringProperty("tag"))

			}
		}
	})
	t.Run("interval_size_with_exclusion_rule", func(t *testing.T) {
		config := recconf.SortConfig{
			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:   []string{"tag"},
					IntervalSize: 1,
				},
			},
			ExclusionRules: []recconf.ExclusionRuleConfig{
				{
					Positions: []int{1},
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "tag",
							Type:     "string",
							Operator: "equal",
							Value:    "t1",
						},
					},
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i, item := range result {
			if i%2 == 1 && item.StringProperty("tag") != "t1" {
				t.Error("item error")
			}
		}
	})
}

func TestDiversitryRuleSortByIntervalSize2(t *testing.T) {
	config := recconf.SortConfig{

		DiversityRules: []recconf.DiversityRuleConfig{
			{
				Dimensions:   []string{"tag"},
				IntervalSize: 1,
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		item.AddProperty("tag", "t1")

		items = append(items, item)
	}
	context := context.NewRecommendContext()
	context.Size = 10
	sortData := SortData{Data: items, Context: context}

	NewDiversityRuleSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	for i, item := range result {
		assert.Equal(t, string(item.Id), strconv.Itoa(i))
	}
}
func TestDiversitryRuleSortByExclusionRule(t *testing.T) {
	config := recconf.SortConfig{

		DiversityRules: []recconf.DiversityRuleConfig{
			{
				Dimensions:   []string{"tag"},
				IntervalSize: 1,
			},
		},
		ExclusionRules: []recconf.ExclusionRuleConfig{
			{
				Positions: []int{1, 2, 3, 4, 5},
				Conditions: []recconf.FilterParamConfig{
					{
						Name:     "tag",
						Type:     "string",
						Operator: "equal",
						Value:    "t1",
					},
				},
			},
		},
	}

	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		if i < 5 {
			item.AddProperty("tag", "t1")
		} else {
			item.AddProperty("tag", "t2")
		}

		items = append(items, item)
	}
	context := context.NewRecommendContext()
	context.Size = 10
	sortData := SortData{Data: items, Context: context}

	NewDiversityRuleSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	assert.Equal(t, 10, len(result))
	for i, item := range result {
		if i < 5 {
			assert.Equal(t, "t2", item.StringProperty("tag"))
		} else {
			assert.Equal(t, "t1", item.StringProperty("tag"))
		}
	}
}

func TestDiversitryRuleSort(t *testing.T) {
	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		if i < 3 {
			item.AddProperty("tag", "t1")
		} else if i < 6 {
			item.AddProperty("tag", "t2")
		} else {
			item.AddProperty("tag", "t3")
		}

		items = append(items, item)
	}
	t.Run("diversitry_rule_sort", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:    []string{"tag"},
					WindowSize:    5,
					FrequencySize: 2,
					IntervalSize:  1,
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i, item := range result {
			fmt.Println(i, item)
		}
	})
	t.Run("diversitry_rule_sort_with_exclusion", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:    []string{"tag"},
					WindowSize:    5,
					FrequencySize: 2,
					IntervalSize:  1,
				},
			},
			ExclusionRules: []recconf.ExclusionRuleConfig{
				{
					Positions: []int{1},
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "tag",
							Type:     "string",
							Operator: "equal",
							Value:    "t1",
						},
					},
				},
				{
					Positions: []int{1},
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "tag",
							Type:     "string",
							Operator: "equal",
							Value:    "t2",
						},
					},
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		assert.Equal(t, 10, len(result))
		for i, item := range result {
			if i == 0 {
				assert.Equal(t, "t3", item.GetProperty("tag"))
			}
			fmt.Println(i, item)
		}
	})
}

func TestDiversitryRuleExploreItemSize(t *testing.T) {
	var items []*module.Item
	for i := 0; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"
		if i < 10 {
			item.AddProperty("tag", "t1")
		} else {
			item.AddProperty("tag", "t2")
		}

		items = append(items, item)
	}
	t.Run("diversitry_rule_sort", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:   []string{"tag"},
					IntervalSize: 1,
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i, item := range result {
			if i == 10 {
				break
			}
			if i%2 == 0 {
				assert.Equal(t, "t1", item.StringProperty("tag"))
			} else {
				assert.Equal(t, "t2", item.StringProperty("tag"))
			}
		}
	})
	t.Run("diversitry_rule_sort_with_explore_item_size", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:   []string{"tag"},
					IntervalSize: 1,
				},
			},
			ExploreItemSize: 10,
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		assert.Equal(t, 20, len(result))
		for i, item := range result {
			if i < 10 {
				if i%2 == 0 {
					assert.Equal(t, "t1", item.StringProperty("tag"))
				} else {
					assert.Equal(t, "t2", item.StringProperty("tag"))
				}
			}
			t.Log(i, item)
		}
	})
}

func TestDiversitryRuleWeight(t *testing.T) {
	var items []*module.Item
	for i := 0; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"

		item.AddProperty("is_new_item", i%2)
		item.AddProperty("tag", fmt.Sprintf("t%d", i%3))
		item.AddProperty("category", fmt.Sprintf("c%d", i%4))

		if i > 10 {
			item.AddProperty("is_new_item", 1)
			item.AddProperty("tag", "t2")
			item.AddProperty("category", "c1")
		}
		if i == 19 {
			item.AddProperty("tag", "t3")
		}
		if i == 18 {
			item.AddProperty("category", "c4")
			item.AddProperty("is_new_item", 0)
		}

		items = append(items, item)
	}
	t.Run("diversitry_rule_sort", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:   []string{"is_new_item"},
					IntervalSize: 1,
				},
				{
					Dimensions:    []string{"tag"},
					WindowSize:    5,
					FrequencySize: 1,
				},
				{
					Dimensions:    []string{"category"},
					WindowSize:    5,
					FrequencySize: 1,
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i, item := range result {
			assert.Equal(t, strconv.Itoa(i), string(item.Id))
		}

	})
	t.Run("diversitry_rule_sort_with_weight", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:   []string{"is_new_item"},
					IntervalSize: 1,
					Weight:       1,
				},
				{
					Dimensions:    []string{"category"},
					WindowSize:    5,
					FrequencySize: 1,
					Weight:        3,
				},
				{
					Dimensions:    []string{"tag"},
					WindowSize:    5,
					FrequencySize: 1,
					Weight:        5,
				},
			},
		}
		fmt.Println("====sort before====")
		for _, item := range items {
			t.Log(item)
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		fmt.Println("====sort after====")
		for i, item := range result {
			if i == 3 {
				assert.Equal(t, "19", string(item.Id))
			}
			if i == 4 {
				assert.Equal(t, "18", string(item.Id))
			}
			t.Log(item)
		}

	})
}

func TestDiversitryRuleMulitValue(t *testing.T) {
	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"

		item.AddProperty("tag", "A")
		items = append(items, item)
	}
	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"

		item.AddProperty("tag", "A/B")
		items = append(items, item)
	}
	t.Run("diversitry_rule_sort", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions: []string{"tag"},
					WindowSize: 5,
					//FrequencySize: 1,
					IntervalSize: 1,
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 20
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		baseIndex := 0
		for i, item := range result {
			if i%2 == 0 {
				assert.Equal(t, strconv.Itoa(baseIndex), string(item.Id))
				baseIndex++
			} else {
				assert.Equal(t, strconv.Itoa(baseIndex+9), string(item.Id))
			}
		}

	})
	t.Run("diversitry_rule_sort_multi_value", func(t *testing.T) {
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions: []string{"tag"},
					WindowSize: 5,
					//FrequencySize: 1,
					IntervalSize: 1,
				},
			},
			MultiValueDimensionConf: []recconf.MultiValueDimensionConfig{
				{
					DimensionName: "tag",
					Delimiter:     "/",
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 20
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i, item := range result {
			assert.Equal(t, strconv.Itoa(i), string(item.Id))
		}

	})
	t.Run("diversitry_rule_sort_multi_value_v2", func(t *testing.T) {
		items = items[:0]
		for i := 0; i < 10; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r1"

			if i%3 == 0 {
				item.AddProperty("tag", "A")
			} else if i%3 == 1 {
				//item.AddProperty("tag", "A/B")
			} else {
				item.AddProperty("tag", "B/C")
			}
			items = append(items, item)
		}
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions:    []string{"tag"},
					WindowSize:    3,
					FrequencySize: 1,
					IntervalSize:  1,
				},
			},
			MultiValueDimensionConf: []recconf.MultiValueDimensionConfig{
				{
					DimensionName: "tag",
					Delimiter:     "/",
				},
			},
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i, item := range result {
			assert.Equal(t, strconv.Itoa(i), string(item.Id))
		}

	})
	t.Run("diversitry_rule_sort_multi_value_v3", func(t *testing.T) {
		items = items[:0]
		for i := 0; i < 10; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r1"

			if i%3 == 0 {
				item.AddProperty("tag", "A")
			} else if i%3 == 1 {
				item.AddProperty("tag", "A/B")
			} else {
				item.AddProperty("tag", "B/C")
			}
			if i < 5 {
				item.AddProperty("category", strconv.Itoa(0))
			} else {
				item.AddProperty("category", strconv.Itoa(1))
			}
			items = append(items, item)
		}
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions: []string{"category", "tag"},
					WindowSize: 3,
					//FrequencySize: 1,
					IntervalSize: 1,
				},
			},
			MultiValueDimensionConf: []recconf.MultiValueDimensionConfig{
				{
					DimensionName: "tag",
					Delimiter:     "/",
				},
			},
		}
		fmt.Println("====sort before====")
		for _, item := range items {
			t.Log(item)
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		fmt.Println("====sort after====")
		for _, item := range result {
			//assert.Equal(t, strconv.Itoa(i), string(item.Id))
			t.Log(item)
		}

	})
	t.Run("diversitry_rule_sort_multi_value_v4", func(t *testing.T) {
		items = items[:0]
		for i := 0; i < 10; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r1"

			if i%3 == 0 {
				item.AddProperty("tag", "A")
			} else if i%3 == 1 {
				item.AddProperty("tag", "A/B")
			} else {
				item.AddProperty("tag", "B/C")
			}
			if i < 3 {
				item.AddProperty("category", "c1")
			} else if i < 6 {
				item.AddProperty("category", "c1#c2")
			} else {
				item.AddProperty("category", "c3#c4")
			}
			items = append(items, item)
		}
		config := recconf.SortConfig{

			DiversityRules: []recconf.DiversityRuleConfig{
				{
					Dimensions: []string{"tag"},
					WindowSize: 3,
					//FrequencySize: 1,
					IntervalSize: 1,
				},
				{
					Dimensions: []string{"category"},
					WindowSize: 3,
					//FrequencySize: 1,
					IntervalSize: 1,
				},
			},
			MultiValueDimensionConf: []recconf.MultiValueDimensionConfig{
				{
					DimensionName: "tag",
					Delimiter:     "/",
				},
				{
					DimensionName: "category",
					Delimiter:     "#",
				},
			},
		}
		fmt.Println("====sort before====")
		for _, item := range items {
			t.Log(item)
		}

		context := context.NewRecommendContext()
		context.Size = 10
		sortData := SortData{Data: items, Context: context}

		NewDiversityRuleSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		assert.Equal(t, string(result[1].Id), strconv.Itoa(8))
		assert.Equal(t, string(result[2].Id), strconv.Itoa(3))
		assert.Equal(t, string(result[3].Id), strconv.Itoa(1))

	})
}
