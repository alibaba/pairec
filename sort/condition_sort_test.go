package sort

import (
	"testing"

	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
)

// mockSort is a simple sort for testing
type mockSort struct {
	name   string
	called bool
}

func (m *mockSort) Sort(sortData *SortData) error {
	m.called = true
	return nil
}

func TestConditionSort_NoConditionMatch_UsesDefault(t *testing.T) {
	// Register mock sorts for testing
	defaultSort := &mockSort{name: "default_sort"}
	RegisterSort("test_default_sort", defaultSort)
	defer func() {
		delete(sortMapping, "test_default_sort")
	}()

	config := recconf.SortConfig{
		Name:     "test_condition_sort",
		SortType: "ConditionSort",
	}
	config.ConditionSortConfs.SortConfs = []struct {
		Conditions []recconf.FilterParamConfig
		SortName   string
	}{
		{
			Conditions: []recconf.FilterParamConfig{
				{Name: "user_level", Domain: "user", Operator: "equal", Type: "string", Value: "vip"},
			},
			SortName: "vip_sort",
		},
	}
	config.ConditionSortConfs.DefaultSortName = "test_default_sort"

	condSort := NewConditionSort(config)

	// Create test data with non-VIP user
	user := module.NewUser("test_user")
	user.AddProperty("user_level", "normal")

	items := []*module.Item{
		{Id: "item1", Score: 1.0},
		{Id: "item2", Score: 2.0},
	}

	ctx := context.NewRecommendContext()
	ctx.RecommendId = "test_req_1"

	sortData := &SortData{
		Data:    items,
		Context: ctx,
		User:    user,
	}

	err := condSort.Sort(sortData)
	if err != nil {
		t.Errorf("Sort() error = %v", err)
	}

	if !defaultSort.called {
		t.Error("Expected default sort to be called")
	}
}

func TestConditionSort_ConditionMatch_UsesMatchedSort(t *testing.T) {
	// Register mock sorts for testing
	vipSort := &mockSort{name: "vip_sort"}
	defaultSort := &mockSort{name: "default_sort"}
	RegisterSort("test_vip_sort", vipSort)
	RegisterSort("test_default_sort_2", defaultSort)
	defer func() {
		delete(sortMapping, "test_vip_sort")
		delete(sortMapping, "test_default_sort_2")
	}()

	config := recconf.SortConfig{
		Name:     "test_condition_sort_2",
		SortType: "ConditionSort",
	}
	config.ConditionSortConfs.SortConfs = []struct {
		Conditions []recconf.FilterParamConfig
		SortName   string
	}{
		{
			Conditions: []recconf.FilterParamConfig{
				{Name: "user_level", Domain: "user", Operator: "equal", Type: "string", Value: "vip"},
			},
			SortName: "test_vip_sort",
		},
	}
	config.ConditionSortConfs.DefaultSortName = "test_default_sort_2"

	condSort := NewConditionSort(config)

	// Create test data with VIP user
	user := module.NewUser("test_user")
	user.AddProperty("user_level", "vip")

	items := []*module.Item{
		{Id: "item1", Score: 1.0},
		{Id: "item2", Score: 2.0},
	}

	ctx := context.NewRecommendContext()
	ctx.RecommendId = "test_req_2"

	sortData := &SortData{
		Data:    items,
		Context: ctx,
		User:    user,
	}

	err := condSort.Sort(sortData)
	if err != nil {
		t.Errorf("Sort() error = %v", err)
	}

	if !vipSort.called {
		t.Error("Expected VIP sort to be called")
	}
	if defaultSort.called {
		t.Error("Default sort should not be called")
	}
}

func TestConditionSort_NoConditions_NoDefault(t *testing.T) {
	config := recconf.SortConfig{
		Name:     "test_condition_sort_empty",
		SortType: "ConditionSort",
	}
	// No conditions and no default sort configured

	condSort := NewConditionSort(config)

	user := module.NewUser("test_user")
	items := []*module.Item{
		{Id: "item1", Score: 1.0},
	}

	ctx := context.NewRecommendContext()
	ctx.RecommendId = "test_req_empty"

	sortData := &SortData{
		Data:    items,
		Context: ctx,
		User:    user,
	}

	// Should return nil without error (no-op)
	err := condSort.Sort(sortData)
	if err != nil {
		t.Errorf("Sort() error = %v, expected nil", err)
	}
}

func TestConditionSort_MultipleConditions_FirstMatch(t *testing.T) {
	// Register mock sorts for testing
	vipSort := &mockSort{name: "vip_sort"}
	newUserSort := &mockSort{name: "new_user_sort"}
	defaultSort := &mockSort{name: "default_sort"}
	RegisterSort("test_vip_sort_3", vipSort)
	RegisterSort("test_new_user_sort", newUserSort)
	RegisterSort("test_default_sort_3", defaultSort)
	defer func() {
		delete(sortMapping, "test_vip_sort_3")
		delete(sortMapping, "test_new_user_sort")
		delete(sortMapping, "test_default_sort_3")
	}()

	config := recconf.SortConfig{
		Name:     "test_condition_sort_multi",
		SortType: "ConditionSort",
	}
	config.ConditionSortConfs.SortConfs = []struct {
		Conditions []recconf.FilterParamConfig
		SortName   string
	}{
		{
			Conditions: []recconf.FilterParamConfig{
				{Name: "user_level", Domain: "user", Operator: "equal", Type: "string", Value: "vip"},
			},
			SortName: "test_vip_sort_3",
		},
		{
			Conditions: []recconf.FilterParamConfig{
				{Name: "is_new_user", Domain: "user", Operator: "equal", Type: "int", Value: 1},
			},
			SortName: "test_new_user_sort",
		},
	}
	config.ConditionSortConfs.DefaultSortName = "test_default_sort_3"

	condSort := NewConditionSort(config)

	// Create test data with VIP user who is also a new user
	// First condition should match, second should be skipped
	user := module.NewUser("test_user")
	user.AddProperty("user_level", "vip")
	user.AddProperty("is_new_user", 1)

	items := []*module.Item{
		{Id: "item1", Score: 1.0},
	}

	ctx := context.NewRecommendContext()
	ctx.RecommendId = "test_req_multi"

	sortData := &SortData{
		Data:    items,
		Context: ctx,
		User:    user,
	}

	err := condSort.Sort(sortData)
	if err != nil {
		t.Errorf("Sort() error = %v", err)
	}

	// Only VIP sort should be called (first match)
	if !vipSort.called {
		t.Error("Expected VIP sort to be called (first match)")
	}
	if newUserSort.called {
		t.Error("New user sort should not be called (second condition)")
	}
	if defaultSort.called {
		t.Error("Default sort should not be called")
	}
}
