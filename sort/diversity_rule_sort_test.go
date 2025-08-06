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
	fmt.Println("====sort before====")
	for _, item := range items {
		t.Log(item)
	}
	context := context.NewRecommendContext()
	context.Size = 10
	sortData := SortData{Data: items, Context: context}

	NewDiversityRuleSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	assert.Equal(t, 10, len(result))
	fmt.Println("====sort after====")
	for i, item := range result {
		t.Log(item)
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
	t.Run("diversitry_rule_sort_with_exclusion", func(t *testing.T) {
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
