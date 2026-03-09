package sort

import (
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestNewCustomFieldSort(t *testing.T) {
	// 测试默认配置
	config := recconf.SortConfig{
		Name:        "testSort",
		SortByField: "price",
	}
	sort := NewCustomFieldSort(config)
	if sort.sortByField != "price" {
		t.Errorf("expected sortByField to be 'price', got '%s'", sort.sortByField)
	}
	if sort.sortOrder != "desc" {
		t.Errorf("expected default sortOrder to be 'desc', got '%s'", sort.sortOrder)
	}

	// 测试升序配置
	config.SortOrder = "asc"
	sort = NewCustomFieldSort(config)
	if sort.sortOrder != "asc" {
		t.Errorf("expected sortOrder to be 'asc', got '%s'", sort.sortOrder)
	}

	// 测试大写配置
	config.SortOrder = "ASC"
	sort = NewCustomFieldSort(config)
	if sort.sortOrder != "asc" {
		t.Errorf("expected sortOrder to be 'asc' (lowercase), got '%s'", sort.sortOrder)
	}

	// 测试无效配置（应默认为降序）
	config.SortOrder = "invalid"
	sort = NewCustomFieldSort(config)
	if sort.sortOrder != "desc" {
		t.Errorf("expected invalid sortOrder to default to 'desc', got '%s'", sort.sortOrder)
	}

	// 测试空字段（应默认为 current_score）
	config.SortByField = ""
	config.SortOrder = "desc"
	sort = NewCustomFieldSort(config)
	if sort.sortByField != "current_score" {
		t.Errorf("expected empty sortByField to default to 'current_score', got '%s'", sort.sortByField)
	}
}

func TestCustomFieldSort_Sort_Desc(t *testing.T) {
	config := recconf.SortConfig{
		Name:        "testSort",
		SortByField: "price",
		SortOrder:   "desc",
	}
	s := NewCustomFieldSort(config)

	ctx := context.NewRecommendContext()
	items := []*module.Item{
		createItemWithPrice("1", 100),
		createItemWithPrice("2", 300),
		createItemWithPrice("3", 200),
		createItemWithPrice("4", 50),
	}

	sortData := &SortData{
		Data:    items,
		Context: ctx,
	}

	err := s.Sort(sortData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	result := sortData.Data.([]*module.Item)
	expected := []string{"2", "3", "1", "4"} // 300, 200, 100, 50
	for i, item := range result {
		if string(item.Id) != expected[i] {
			t.Errorf("position %d: expected id '%s', got '%s'", i, expected[i], item.Id)
		}
	}
}

func TestCustomFieldSort_Sort_Asc(t *testing.T) {
	config := recconf.SortConfig{
		Name:        "testSort",
		SortByField: "price",
		SortOrder:   "asc",
	}
	s := NewCustomFieldSort(config)

	ctx := context.NewRecommendContext()
	items := []*module.Item{
		createItemWithPrice("1", 100),
		createItemWithPrice("2", 300),
		createItemWithPrice("3", 200),
		createItemWithPrice("4", 50),
	}

	sortData := &SortData{
		Data:    items,
		Context: ctx,
	}

	err := s.Sort(sortData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	result := sortData.Data.([]*module.Item)
	expected := []string{"4", "1", "3", "2"} // 50, 100, 200, 300
	for i, item := range result {
		if string(item.Id) != expected[i] {
			t.Errorf("position %d: expected id '%s', got '%s'", i, expected[i], item.Id)
		}
	}
}

func TestCustomFieldSort_Sort_ByCurrentScore(t *testing.T) {
	config := recconf.SortConfig{
		Name:        "testSort",
		SortByField: "current_score",
		SortOrder:   "desc",
	}
	s := NewCustomFieldSort(config)

	ctx := context.NewRecommendContext()
	items := []*module.Item{
		createItemWithScore("1", 10),
		createItemWithScore("2", 30),
		createItemWithScore("3", 20),
	}

	sortData := &SortData{
		Data:    items,
		Context: ctx,
	}

	err := s.Sort(sortData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	result := sortData.Data.([]*module.Item)
	expected := []string{"2", "3", "1"} // 30, 20, 10
	for i, item := range result {
		if string(item.Id) != expected[i] {
			t.Errorf("position %d: expected id '%s', got '%s'", i, expected[i], item.Id)
		}
	}
}

func TestCustomFieldSort_Sort_InvalidDataType(t *testing.T) {
	config := recconf.SortConfig{
		Name:        "testSort",
		SortByField: "price",
	}
	s := NewCustomFieldSort(config)

	ctx := context.NewRecommendContext()
	sortData := &SortData{
		Data:    "invalid data type",
		Context: ctx,
	}

	err := s.Sort(sortData)
	if err == nil {
		t.Error("expected error for invalid data type, got nil")
	}
}

func createItemWithPrice(id string, price float64) *module.Item {
	item := module.NewItem(id)
	item.Properties["price"] = price
	return item
}

func createItemWithScore(id string, score float64) *module.Item {
	item := module.NewItem(id)
	item.Score = score
	return item
}
