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
				"valid_list": []string{"42", "42.5", "43"},
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
	}

	for _, case1 := range testcases {
		containsOp := NewContainsFilterOp(case1.Config)

		result, err := containsOp.ContainsEvaluate(case1.ItemProperties, case1.UserProperties, case1.ItemProperties)

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

		result, err := containsOp.ContainsEvaluate(case1.ItemProperties, case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error")
		}

	}

}

func TestFilterParam(t *testing.T) {

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
				"list": []string{"41", "43.5", "44"},
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
				Value:    []string{"40", "41", "41.5", "42"},
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

		result, err := containsOp.NotInEvaluate(case1.ItemProperties, case1.UserProperties, case1.ItemProperties)

		if err != nil {
			t.Error(err)
		}

		if result != case1.Expect {
			t.Error(case1, "result error")
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
