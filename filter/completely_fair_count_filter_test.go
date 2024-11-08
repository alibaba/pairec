package filter

import (
	"fmt"
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"testing"
)

func TestCompletelyFairCountFilter(t *testing.T) {
	fairCountFilter := NewCompletelyFairCountFilter(recconf.FilterConfig{
		RetainNum: 10,
	})

	user := module.NewUser("user_1")

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

	err := fairCountFilter.Filter(filterData)
	if err != nil {
		t.Error(err)
	}
	filterItems := filterData.Data.([]*module.Item)

	if len(filterItems) != 10 {
		t.Error("len(items) != 10", len(filterItems))
	}
	recallNames := []string{"recall1", "recall2"}
	for i, item := range filterItems {
		if item.RetrieveId != recallNames[i%len(recallNames)] {
			t.Errorf("item.RetrieveId != %s, %v", recallNames[i%len(recallNames)], item)
		}
	}

	filterData = &FilterData{
		Context: &context.RecommendContext{
			RecommendId: "1",
		},
		Data: items,
		User: user,
	}
	fairCountFilter2 := NewCompletelyFairCountFilter(recconf.FilterConfig{
		RetainNum: 100,
	})
	err = fairCountFilter2.Filter(filterData)
	if err != nil {
		t.Error(err)
	}
	filterItems = filterData.Data.([]*module.Item)

	if len(filterItems) != 20 {
		t.Error("len(items) != 20", len(filterItems))
	}

	for i, item := range filterItems {
		if item.RetrieveId != recallNames[i%len(recallNames)] {
			t.Errorf("item.RetrieveId != %s, %v", recallNames[i%len(recallNames)], item)
		}
	}
}
