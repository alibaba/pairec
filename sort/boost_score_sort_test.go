package sort

import (
	"strconv"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestBoostScoreSort(t *testing.T) {
	t.Run("boost_score_sort", func(t *testing.T) {
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
			BoostScoreConditions: []recconf.BoostScoreCondition{
				{
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "recall_name",
							Type:     "string",
							Operator: "equal",
							Value:    "r1",
						},
					},
					Expression: "score * 100",
				},
				{
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "recall_name",
							Type:     "string",
							Operator: "equal",
							Value:    "r2",
						},
					},
					Expression: "score * (-10)",
				},
			},
		}
		user := module.NewUser("u1")
		sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{RecommendId: "test_req"}}
		NewBoostScoreSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		assert.Equal(t, result[0].Score, float64(0))
		assert.Equal(t, result[1].Score, float64(100))
		assert.Equal(t, result[10].Score, float64(-100))

	})

	t.Run("boost_score_sort_With_round", func(t *testing.T) {
		var items []*module.Item
		item1 := module.NewItem("item_1")
		item1.Score = float64(0.311)

		items = append(items, item1)
		config := recconf.SortConfig{
			Debug: true,
			BoostScoreConditions: []recconf.BoostScoreCondition{
				{
					Conditions: []recconf.FilterParamConfig{},
					Expression: "round(score * 3, 2)",
				},
			},
		}
		user := module.NewUser("u1")
		sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{RecommendId: "test_req"}}
		NewBoostScoreSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		assert.Equal(t, 0.93, result[0].Score)
	})
	t.Run("boost_score_sort_With_round_v2", func(t *testing.T) {
		var items []*module.Item
		item1 := module.NewItem("item_1")
		item1.Score = float64(0.311)

		items = append(items, item1)
		config := recconf.SortConfig{
			Debug: true,
			BoostScoreConditions: []recconf.BoostScoreCondition{
				{
					Conditions: []recconf.FilterParamConfig{},
					Expression: "round(score * 3)",
				},
			},
		}
		user := module.NewUser("u1")
		sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{RecommendId: "test_req"}}
		NewBoostScoreSort(config).Sort(&sortData)

		result := sortData.Data.([]*module.Item)

		assert.Equal(t, float64(1), result[0].Score)
	})

}
