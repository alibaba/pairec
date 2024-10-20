package sort

import (
	"strconv"
	"testing"

	"github.com/alibaba/pairec/context"
	"github.com/alibaba/pairec/module"
	"github.com/alibaba/pairec/recconf"
)

func TestBoostScoreSort(t *testing.T) {

	user := module.NewUser("user1")
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
	sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{}}
	NewBoostScoreSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	if result[0].Score != float64(0) {
		t.Error("test fail")
	}
	if result[1].Score != float64(100) {
		t.Error("test fail")
	}
	if result[10].Score != float64(-100) {
		t.Error("test fail")
	}

}

func TestBoostScoreSortV2(t *testing.T) {

	user := module.NewUser("user1")
	var items []*module.Item
	for i := 0; i < 10; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r1"

		item.AddAlgoScore("algo_score", 1)
		items = append(items, item)
	}

	for i := 10; i < 20; i++ {
		item := module.NewItem(strconv.Itoa(i))
		item.Score = float64(i)
		item.RetrieveId = "r2"

		item.AddAlgoScore("algo_score", 10)
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
				Expression: "score * 100 + algo_score",
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
				Expression: "score * (-10) + algo_score",
			},
		},
	}
	sortData := SortData{Data: items, User: user, Context: &context.RecommendContext{}}
	NewBoostScoreSort(config).Sort(&sortData)

	result := sortData.Data.([]*module.Item)

	if result[0].Score != float64(1) {
		t.Error("test fail")
	}
	if result[1].Score != float64(101) {
		t.Error("test fail")
	}
	if result[10].Score != float64(-90) {
		t.Error("test fail")
	}

}
