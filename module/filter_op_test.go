package module

import (
	"testing"

	"github.com/alibaba/pairec/v2/recconf"
)

func TestContainsFilterOp(t *testing.T) {

	testcases := []struct {
		Config         recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []any{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{},
			Expect:         false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "contains",
				Type:     "[]string",
				Value:    []string{"40", "41", "41.5", "42"},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "contains",
				Type:     "[]string",
				Value:    []any{"40", "41", "41.5", "42"},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "contains",
				Type:     "[]int",
				Value:    []any{40, 41, 42},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []int{42, 43},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		containsOp := NewContainsFilterOp(case1.Config)

		result, err := containsOp.DomainEvaluate(case1.ItemProperties, case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error")
		}

	}

}

func TestNotContainsFilterOp(t *testing.T) {

	testcases := []struct {
		Config         recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "not_contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "not_contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "not_contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "not_contains",
				Type:     "[]string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{},
			Expect:         false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "not_contains",
				Type:     "[]string",
				Value:    []string{"40", "41", "41.5", "42"},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "valid_list",
				Domain:   "item",
				Operator: "not_contains",
				Type:     "[]string",
				Value:    []string{"40", "41", "41.5"},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		containsOp := NewNotContainsFilterOp(case1.Config)

		result, err := containsOp.DomainEvaluate(case1.ItemProperties, case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error")
		}

	}

}

func TestFilterParamByFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "valid_list",
					Domain:   "item",
					Operator: "not_contains",
					Type:     "[]string",
					Value:    "user.list",
				},
				{
					Name:     "val1",
					Domain:   "item",
					Operator: "equal",
					Type:     "string",
					Value:    "test",
				},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
				"val1":       "test",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "valid_list",
					Domain:   "item",
					Operator: "contains",
					Type:     "[]string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []interface{}{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "valid_list",
					Domain:   "item",
					Operator: "not_contains",
					Type:     "[]string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{},
			Expect:         true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"40", "41"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"40", "41", "42"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []any{"40", "41", "42"},
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error")
		}

	}

}

func TestNotInFilterOp(t *testing.T) {

	testcases := []struct {
		Config         recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: recconf.FilterParamConfig{
				Name:     "foo",
				Domain:   "item",
				Operator: "not_in",
				Type:     "string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "foo",
				Domain:   "item",
				Operator: "not_in",
				Type:     "string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"foo": "42.5",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "foo",
				Domain:   "item",
				Operator: "not_in",
				Type:     "string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "foo",
				Domain:   "item",
				Operator: "not_in",
				Type:     "string",
				Value:    "user.list",
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{},
			Expect:         true,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "foo",
				Domain:   "item",
				Operator: "not_in",
				Type:     "string",
				Value:    []any{"40", "41", "41.5", "42"},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list2": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: recconf.FilterParamConfig{
				Name:     "foo",
				Domain:   "item",
				Operator: "not_in",
				Type:     "string",
				Value:    []string{"40", "41", "41.5"},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		containsOp := NewNotInFilterOp(case1.Config)

		result, err := containsOp.DomainEvaluate(case1.ItemProperties, case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error")
		}

	}

}
func TestInFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    []any{"41", "42", "44"},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    []string{"41", "42", "44"},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "int",
					Value:    []any{"41", "42", "44"},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "int",
					Value:    []int{41, 42, 44},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "int",
					Value:    []any{41, 42, 44},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "int",
					Value:    []any{41, 44},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    []any{"41", "44"},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "42", "43.5", "44"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "int",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "42", "43.5", "44"},
				"foo":  42,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "44"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				//"list": []string{"41", "43.5", "44"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "in",
					Type:     "string",
					Value:    "user.list",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{},
			},
			Expect: false,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error")
		}

	}

}
func TestEqualFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "string",
					Value:    "",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "string",
					Value:    "",
				},
			},
			ItemProperties: map[string]interface{}{},
			Expect:         false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "string",
					Value:    "42",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "48"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "string",
					Value:    "43",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "string",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
				"bar":  "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Operator: "equal",
					Type:     "string",
					Value:    "item.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
				"bar": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "int",
					Value:    "item.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
				"bar": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "int",
					Value:    "42",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
				"bar": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "int",
					Value:    42,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
				"bar": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "int",
					Value:    43,
				},
			},
			ItemProperties: map[string]interface{}{
				"bar": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "int64",
					Value:    43,
				},
			},
			ItemProperties: map[string]interface{}{
				"bar": 42,
				"foo": 43,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "equal",
					Type:     "int64",
					Value:    "item.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"bar": 42,
				"foo": 43,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestNotEqualFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "string",
					Value:    "",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "string",
					Value:    "",
				},
			},
			ItemProperties: map[string]interface{}{},
			Expect:         true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "string",
					Value:    "42",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "string",
					Value:    "43",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "int",
					Value:    43,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "int",
					Value:    42,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": 42,
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "user",
					Operator: "not_equal",
					Type:     "string",
					Value:    "item.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"bar": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
				"foo":  "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "user",
					Operator: "not_equal",
					Type:     "string",
					Value:    "item.bar",
				},
			},
			ItemProperties: map[string]interface{}{},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
				"foo":  "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "string",
					Value:    "item.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
				"bar": "43",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "not_equal",
					Type:     "string",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
				"bar":  "43",
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestIsNullFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "is_null",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "is_null",
				},
			},
			ItemProperties: map[string]interface{}{},
			Expect:         true,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestIsNotNullFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "is_not_null",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "is_not_null",
				},
			},
			ItemProperties: map[string]interface{}{},
			Expect:         false,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestGreaterFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "int",
					Value:    45,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "int",
					Value:    40,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "int",
					Value:    "45",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 45,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "40",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2024-02-04 20:16:59",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2025-01-05 20:17:00",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greater",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2014-01-02",
			},
			UserProperties: map[string]interface{}{
				"bar": "2014-01-01 23:59:59",
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestGreaterThanFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "int",
					Value:    45,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "int",
					Value:    40,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "int",
					Value:    "45",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 45,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 42,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "40",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2024-02-04 20:16:59",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "greaterThan",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2025-01-05 20:17:00",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestLessFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "int",
					Value:    45,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "int",
					Value:    40,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "int",
					Value:    "45",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 45,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "40",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 45,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2024-02-04 20:16:59",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "less",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2025-01-05 20:17:00",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: false,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestLessThanFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "int",
					Value:    45,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "int",
					Value:    40,
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "int",
					Value:    "42",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 45,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "int",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "40",
			},
			UserProperties: map[string]interface{}{
				"bar": 40,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "float",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"bar": 45,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2025-01-04 20:16:59",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2024-02-04 20:16:59",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Name:     "foo",
					Domain:   "item",
					Operator: "lessThan",
					Type:     "time",
					Value:    "user.bar",
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "2025-01-05 20:17:00",
			},
			UserProperties: map[string]interface{}{
				"bar": "2025-01-05 20:16:59",
			},
			Expect: false,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestBoolFilterOp(t *testing.T) {

	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "or",
					Configs: []recconf.FilterParamConfig{
						{
							Operator: "equal",
							Type:     "string",
							Name:     "foo",
							Value:    "42",
						},
						{
							Operator: "equal",
							Type:     "string",
							Name:     "foo",
							Value:    "43",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "or",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "lessThan",
							Type:     "string",
							Name:     "foo",
							Value:    "40",
						},
						{
							Domain:   "item",
							Operator: "lessThan",
							Type:     "string",
							Name:     "foo",
							Value:    "41",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "and",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "lessThan",
							Type:     "string",
							Name:     "foo",
							Value:    "40",
						},
						{
							Domain:   "item",
							Operator: "lessThan",
							Type:     "string",
							Name:     "foo",
							Value:    "41",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "or",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "lessThan",
							Type:     "string",
							Name:     "foo",
							Value:    "40",
						},
						{
							Domain:   "item",
							Operator: "in",
							Type:     "string",
							Name:     "foo",
							Value:    "user.list",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "or",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "lessThan",
							Type:     "string",
							Name:     "foo",
							Value:    "40",
						},
						{
							Domain:   "item",
							Operator: "in",
							Type:     "string",
							Name:     "foo",
							Value:    "user.list",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"foo": "42",
			},
			UserProperties: map[string]interface{}{
				"list": []string{"41", "43.5", "42", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "or",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "is_null",
							Type:     "string",
							Name:     "valid_list",
						},
						{
							Name:     "valid_list",
							Domain:   "item",
							Operator: "contains",
							Type:     "[]string",
							Value:    "user.list",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "42.5", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "or",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "is_null",
							Type:     "string",
							Name:     "valid_list",
						},
						{
							Name:     "valid_list",
							Domain:   "item",
							Operator: "contains",
							Type:     "[]string",
							Value:    "user.list",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "and",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "is_null",
							Type:     "string",
							Name:     "valid_list",
						},
						{
							Name:     "valid_list",
							Domain:   "item",
							Operator: "contains",
							Type:     "[]string",
							Value:    "user.list",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "and",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "is_null",
							Type:     "string",
							Name:     "valid_list",
						},
						{
							Name:     "valid_list",
							Domain:   "item",
							Operator: "contains",
							Type:     "[]string",
							Value:    "user.list",
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"valid_list": []string{"42", "43"},
			},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "42.5"},
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{
				{
					Operator: "bool",
					Type:     "or",
					Configs: []recconf.FilterParamConfig{
						{
							Domain:   "item",
							Operator: "not_in",
							Type:     "string",
							Name:     "cate_id",
							Value:    []any{"75", "765", "74"},
						},
						{
							Operator: "bool",
							Type:     "and",
							Configs: []recconf.FilterParamConfig{
								{
									Domain:   "item",
									Operator: "equal",
									Type:     "string",
									Name:     "cate_id",
									Value:    "74",
								},
								{
									Domain:   "item",
									Operator: "not_in",
									Type:     "string",
									Name:     "length",
									Value:    []any{"Cropped", "Hip"},
								},
							},
						},
					},
				},
			},
			ItemProperties: map[string]interface{}{
				"cate_id": "74",
				"length":  "Hip2",
			},
			UserProperties: map[string]interface{}{
				"list": []any{"41", "43.5", "42.5"},
			},
			Expect: true,
		},
	}

	for _, case1 := range testcases {
		filterParam := NewFilterParamWithConfig(case1.Config)

		result, err := filterParam.EvaluateByDomain(case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error", result, case1.Expect)
		}

	}

}

func TestExpressionFilterOp(t *testing.T) {
	testcases := []struct {
		Config         []recconf.FilterParamConfig
		UserProperties map[string]interface{}
		ItemProperties map[string]interface{}
		Expect         bool
	}{
		{
			Config: []recconf.FilterParamConfig{ // test false result
				{
					Operator: "expression",
					Value:    "item.size == 43",
				},
			},
			ItemProperties: map[string]interface{}{
				"size": 42,
			},
			Expect: false,
		},
		{
			Config: []recconf.FilterParamConfig{ // test true result
				{
					Operator: "expression",
					Value:    "item.size in [42, 43]",
				},
			},
			ItemProperties: map[string]interface{}{
				"size": 42,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{ // test properties ref
				{
					Operator: "expression",
					Value:    "properties.size in [42, 43]",
				},
			},
			ItemProperties: map[string]interface{}{
				"size": 42,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{ // test set domain
				{
					Domain:   USER,
					Operator: "expression",
					Value:    "item.size in properties.list ",
				},
			},
			UserProperties: map[string]interface{}{
				"list": []int{42, 43},
			},
			ItemProperties: map[string]interface{}{
				"size": 42,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{ // test item ref and user ref
				{
					Operator: "expression",
					Value:    "item.size in user.list ",
				},
			},
			UserProperties: map[string]interface{}{
				"list": []int{42, 43},
			},
			ItemProperties: map[string]interface{}{
				"size": 42,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{ // test complex expr
				{
					Operator: "expression",
					Value:    "!item.sold_out and item.size in user.list ",
				},
			},
			UserProperties: map[string]interface{}{
				"list": []int{42, 43},
			},
			ItemProperties: map[string]interface{}{
				"size":     42,
				"sold_out": false,
			},
			Expect: true,
		},
		{
			Config: []recconf.FilterParamConfig{ // test nil
				{
					Operator: "expression",
					Value:    "!item.sold_out and user.list != nil ? item.size in user.list : true",
				},
			},
			UserProperties: map[string]interface{}{},
			ItemProperties: map[string]interface{}{
				"size":     42,
				"sold_out": false,
			},
			Expect: true,
		},
	}

	for _, testcase := range testcases {
		filterParam := NewFilterParamWithConfig(testcase.Config)

		result, err := filterParam.EvaluateByDomain(testcase.UserProperties, testcase.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != testcase.Expect {
			t.Error(testcase, "result error", result, testcase.Expect)
		}
	}
}
