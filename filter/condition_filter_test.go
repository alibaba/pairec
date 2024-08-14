package filter

import (
	"fmt"
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestConditionFilterInit(t *testing.T) {

	config := recconf.FilterConfig{
		ConditionFilterConfs: struct {
			FilterConfs []struct {
				Conditions []recconf.FilterParamConfig
				FilterName string
			}
			DefaultFilterName string
		}{
			FilterConfs: []struct {
				Conditions []recconf.FilterParamConfig
				FilterName string
			}{
				{
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "query",
							Type:     "string",
							Domain:   "user",
							Operator: "equal",
						},
					},
					FilterName: "filter1",
				},
			},
			DefaultFilterName: "filter2",
		},
	}
	filter := NewConditionFilter(config)
	if filter == nil {
		t.Error("filter is nil")
	}
}

func TestConditionFilterMatch(t *testing.T) {
	adjustCountFilter := NewAdjustCountFilter(recconf.FilterConfig{
		ShuffleItem: false,
		RetainNum:   10,
	})

	RegisterFilter("adjustCountFilter", adjustCountFilter)

	fairCountFilter := NewCompletelyFairCountFilter(recconf.FilterConfig{
		RetainNum: 10,
	})

	RegisterFilter("fairCountFilter", fairCountFilter)

	config := recconf.FilterConfig{
		ConditionFilterConfs: struct {
			FilterConfs []struct {
				Conditions []recconf.FilterParamConfig
				FilterName string
			}
			DefaultFilterName string
		}{
			FilterConfs: []struct {
				Conditions []recconf.FilterParamConfig
				FilterName string
			}{
				{
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "query",
							Type:     "string",
							Domain:   "user",
							Operator: "equal",
							Value:    "1",
						},
					},
					FilterName: "adjustCountFilter",
				},
			},
			DefaultFilterName: "fairCountFilter",
		},
	}
	filter := NewConditionFilter(config)
	if filter == nil {
		t.Error("filter is nil")
	}

	user := module.NewUser("user_1")
	user.AddProperty("query", "1")

	var items []*module.Item

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_a_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "recall1"
		items = append(items, item)
	}

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_b_%d", i))
		item.RetrieveId = "recall2"
		items = append(items, item)
	}

	filterData := &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}
	err := filter.Filter(filterData)
	if err != nil {
		t.Error(err)
	}
	filterItems := filterData.Data.([]*module.Item)
	if len(filterItems) != 10 {
		t.Error("len(items) != 10", len(filterItems))
	}
	for _, item := range filterItems {
		if item.RetrieveId != "recall1" {
			t.Error("item.RetrieveId != recall1", item)
		}
	}

	user.AddProperty("query", "2")

	filterData = &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}
	err = filter.Filter(filterData)
	if err != nil {
		t.Error(err)
	}
	filterItems = filterData.Data.([]*module.Item)
	if len(filterItems) != 10 {
		t.Error("len(items) != 10", len(filterItems))
	}
	recallNames := []string{"recall1", "recall2"}
	for i, item := range filterItems {
		if item.RetrieveId != recallNames[i%len(recallNames)] {
			t.Errorf("item.RetrieveId != %s, %v", recallNames[i%len(recallNames)], item)
		}
	}
}

func TestConditionFilterMatch2(t *testing.T) {
	adjustCountFilter := NewAdjustCountFilter(recconf.FilterConfig{
		ShuffleItem: false,
		RetainNum:   10,
	})

	RegisterFilter("adjustCountFilter", adjustCountFilter)

	fairCountFilter := NewCompletelyFairCountFilter(recconf.FilterConfig{
		RetainNum: 10,
	})

	RegisterFilter("fairCountFilter", fairCountFilter)

	config := recconf.FilterConfig{
		ConditionFilterConfs: struct {
			FilterConfs []struct {
				Conditions []recconf.FilterParamConfig
				FilterName string
			}
			DefaultFilterName string
		}{
			FilterConfs: []struct {
				Conditions []recconf.FilterParamConfig
				FilterName string
			}{
				{
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "query",
							Type:     "string",
							Domain:   "user",
							Operator: "equal",
							Value:    "1",
						},
					},
					FilterName: "adjustCountFilter",
				},
				{
					Conditions: []recconf.FilterParamConfig{
						{
							Name:     "query",
							Type:     "string",
							Domain:   "user",
							Operator: "equal",
							Value:    "2",
						},
					},
					FilterName: "fairCountFilter",
				},
			},
		},
	}
	filter := NewConditionFilter(config)
	if filter == nil {
		t.Error("filter is nil")
	}

	user := module.NewUser("user_1")
	user.AddProperty("query", "1")

	var items []*module.Item

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_a_%d", i))
		item.Score = float64(i)
		item.RetrieveId = "recall1"
		items = append(items, item)
	}

	for i := 0; i < 10; i++ {
		item := module.NewItem(fmt.Sprintf("item_b_%d", i))
		item.RetrieveId = "recall2"
		items = append(items, item)
	}

	filterData := &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}
	err := filter.Filter(filterData)
	if err != nil {
		t.Error(err)
	}
	filterItems := filterData.Data.([]*module.Item)
	if len(filterItems) != 10 {
		t.Error("len(items) != 10", len(filterItems))
	}
	for _, item := range filterItems {
		if item.RetrieveId != "recall1" {
			t.Error("item.RetrieveId != recall1", item)
		}
	}

	user.AddProperty("query", "2")

	filterData = &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}
	err = filter.Filter(filterData)
	if err != nil {
		t.Error(err)
	}
	filterItems = filterData.Data.([]*module.Item)
	if len(filterItems) != 10 {
		t.Error("len(items) != 10", len(filterItems))
	}
	recallNames := []string{"recall1", "recall2"}
	for i, item := range filterItems {
		if item.RetrieveId != recallNames[i%len(recallNames)] {
			t.Errorf("item.RetrieveId != %s, %v", recallNames[i%len(recallNames)], item)
		}
	}
}
