package sort

import (
	"strconv"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestDistinctIdSort(t *testing.T) {
	t.Run("default distinctid ", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 10; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r1"

			item.AddProperty("recall_name", "r1")
			items = append(items, item)
		}

		for i := 10; i < 20; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r2"
			item.AddProperty("recall_name", "r2")

			items = append(items, item)
		}

		config := recconf.SortConfig{
			Debug: true,
			DistinctIdConditions: []recconf.DistinctIdCondition{
				{
					DistinctId: -1,
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "recall_name",
							Operator: "equal",
							Domain:   "item",
							Type:     "string",
							Value:    "r1",
						},
					},
				},
			},
		}
		user := module.NewUser("u1")
		sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{RecommendId: "test_req"}}
		NewDistinctIdSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i := 0; i < 10; i++ {
			t.Log(result[i])
			assert.Equal(t, result[i].GetProperty("__distinct_id__"), -1)
		}
		for i := 10; i < 20; i++ {
			t.Log(result[i])
			assert.NotEqual(t, result[i].GetProperty("__distinct_id__"), -1)
			assert.Equal(t, result[i].GetProperty("__distinct_id__"), i+1)
		}

	})
	t.Run("use distinctid name", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 10; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r1"

			item.AddProperty("recall_name", "r1")
			items = append(items, item)
		}

		for i := 10; i < 20; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r2"
			item.AddProperty("recall_name", "r2")

			items = append(items, item)
		}

		disinctIdName := "__recall_name__"
		config := recconf.SortConfig{
			Debug: true,
			DistinctIdConditions: []recconf.DistinctIdCondition{
				{
					DistinctId: -1,
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "recall_name",
							Operator: "equal",
							Domain:   "item",
							Type:     "string",
							Value:    "r1",
						},
					},
					DistinctIdName: disinctIdName,
				},
			},
		}
		user := module.NewUser("u1")
		sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{RecommendId: "test_req"}}
		NewDistinctIdSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i := 0; i < 10; i++ {
			t.Log(result[i])
			assert.Equal(t, result[i].GetProperty(disinctIdName), -1)
		}
		for i := 10; i < 20; i++ {
			t.Log(result[i])
			assert.NotEqual(t, result[i].GetProperty(disinctIdName), -1)
			assert.Equal(t, result[i].GetProperty(disinctIdName), i+1)
		}

	})
	t.Run("multi distinctid name", func(t *testing.T) {
		var items []*module.Item
		for i := 0; i < 10; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r1"

			item.AddProperty("recall_name", "r1")
			items = append(items, item)
		}

		for i := 10; i < 20; i++ {
			item := module.NewItem(strconv.Itoa(i))
			item.Score = float64(i)
			item.RetrieveId = "r2"
			item.AddProperty("recall_name", "r2")

			items = append(items, item)
		}

		disinctIdName1 := "__recall_name1__"
		disinctIdName2 := "__recall_name2__"
		config1 := recconf.SortConfig{
			Debug: true,
			DistinctIdConditions: []recconf.DistinctIdCondition{
				{
					DistinctId: -1,
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "recall_name",
							Operator: "equal",
							Domain:   "item",
							Type:     "string",
							Value:    "r1",
						},
					},
					DistinctIdName: disinctIdName1,
				},
			},
		}
		config2 := recconf.SortConfig{
			Debug: true,
			DistinctIdConditions: []recconf.DistinctIdCondition{
				{
					DistinctId: -1,
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "recall_name",
							Operator: "equal",
							Domain:   "item",
							Type:     "string",
							Value:    "r2",
						},
					},
					DistinctIdName: disinctIdName2,
				},
			},
		}
		user := module.NewUser("u1")
		sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{RecommendId: "test_req"}}
		NewDistinctIdSort(config1).Sort(&sortData)
		NewDistinctIdSort(config2).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		for i := 0; i < 10; i++ {
			t.Log(result[i])
			assert.Equal(t, result[i].GetProperty(disinctIdName1), -1)
			assert.Equal(t, result[i].GetProperty(disinctIdName2), i+1)
		}
		for i := 10; i < 20; i++ {
			t.Log(result[i])
			assert.Equal(t, result[i].GetProperty(disinctIdName2), -1)
			assert.Equal(t, result[i].GetProperty(disinctIdName1), i+1)
		}

	})

}
