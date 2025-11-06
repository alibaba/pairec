package module

import (
	"fmt"
	"strings"
	"testing"

	"fortio.org/assert"
	"github.com/alibaba/pairec/v2/recconf"
)

func TestTrigger(t *testing.T) {
	config := []recconf.TriggerConfig{
		{
			TriggerKey: "sex",
		},
		{
			TriggerKey: "age",
			Boundaries: []int{20, 30, 40, 50},
		},
		{
			TriggerKey: "os",
		},
	}

	testcases := []struct {
		features  map[string]interface{}
		expectVal string
	}{
		{
			features: map[string]interface{}{"sex": "Male",
				"os":  "IOS",
				"age": 23,
			},
			expectVal: "Male_20-30_IOS",
		},
		{
			features: map[string]interface{}{"sex": "Male",
				"os": "Android",
			},
			expectVal: "Male_NULL_Android",
		},
		{
			features: map[string]interface{}{"sex": "Male",
				"os":  "Android",
				"age": 60,
			},
			expectVal: "Male_>50_Android",
		},
		{
			features: map[string]interface{}{"sex": "Male",
				"os":  "Android",
				"age": 50,
			},
			expectVal: "Male_40-50_Android",
		},
		{
			features: map[string]interface{}{"sex": "Female",
				"os":  "Android",
				"age": 40,
			},
			expectVal: "Female_30-40_Android",
		},
		{
			features: map[string]interface{}{"sex": "Female",
				"os":  "Android",
				"age": 20,
			},
			expectVal: "Female_<=20_Android",
		},
		{
			features: map[string]interface{}{"sex": "Female",
				"os":  "Android",
				"age": 19,
			},
			expectVal: "Female_<=20_Android",
		},
	}

	trigger := NewTrigger(config)

	for _, testcase := range testcases {
		fmt.Println(trigger.GetValue(testcase.features))
		assert.Equal(t, trigger.GetValue(testcase.features), testcase.expectVal)
	}
}

func TestMultiTrigger(t *testing.T) {
	config := []recconf.TriggerConfig{
		{
			TriggerKey: "tags",
		},
	}

	testcases := []struct {
		features  map[string]interface{}
		expectVal string
	}{
		{
			features: map[string]interface{}{"sex": "Male",
				"tags": []string{"tag1", "tag2", "tag3"},
				"age":  23,
			},
			expectVal: strings.Join([]string{"tag1", "tag2", "tag3"}, TIRRGER_SPLIT),
		},
		{
			features: map[string]interface{}{"sex": "Male",
				"tags": []any{"tag1", "tag2", "tag3"},
				"age":  23,
			},
			expectVal: strings.Join([]string{"tag1", "tag2", "tag3"}, TIRRGER_SPLIT),
		},
	}

	trigger := NewTrigger(config)

	for _, testcase := range testcases {
		fmt.Println(trigger.GetValue(testcase.features))
		assert.Equal(t, trigger.GetValue(testcase.features), testcase.expectVal)
	}
}

func TestParseTrigger(t *testing.T) {

	t.Run("parse sigle multi trigger", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{
				TriggerKey: "tags",
			},
		}

		testcases := []struct {
			features  map[string]interface{}
			expectVal string
		}{
			{
				features: map[string]interface{}{"sex": "Male",
					"tags": []string{"tag1", "tag2", "tag3"},
					"age":  23,
				},
				expectVal: strings.Join([]string{"tag1", "tag2", "tag3"}, TIRRGER_SPLIT),
			},
			{
				features: map[string]interface{}{"sex": "Male",
					"tags": []any{"tag1", "tag2", "tag3"},
					"age":  23,
				},
				expectVal: strings.Join([]string{"tag1", "tag2", "tag3"}, TIRRGER_SPLIT),
			},
		}

		trigger := NewTrigger(config)

		for _, testcase := range testcases {
			fmt.Println(trigger.GetValue(testcase.features))
			assert.Equal(t, trigger.GetValue(testcase.features), testcase.expectVal)
			triggers := ParseTriggerId(trigger.GetValue(testcase.features))
			var strs []string
			for _, trigger := range triggers {
				strs = append(strs, trigger.(string))
			}
			assert.Equal(t, strings.Join(strs, TIRRGER_SPLIT), testcase.expectVal)
		}

	})
	t.Run("parse  multi triggers", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{
				TriggerKey: "tags",
			},
			{
				TriggerKey: "age",
			},
		}

		testcases := []struct {
			features  map[string]interface{}
			expectVal []any
		}{
			{
				features: map[string]interface{}{"sex": "Male",
					"tags": []string{"tag1", "tag2", "tag3"},
					"age":  23,
				},
				expectVal: []any{"tag1_23", "tag2_23", "tag3_23"},
			},
			{
				features: map[string]interface{}{"sex": "Male",
					"tags": []any{"tag1", "tag2", "tag3"},
					"age":  34,
				},
				expectVal: []any{"tag1_34", "tag2_34", "tag3_34"},
			},
			{
				features: map[string]interface{}{"sex": "Male",
					"tags": "tag1",
					"age":  34,
				},
				expectVal: []any{"tag1_34"},
			},
			{
				features: map[string]interface{}{"sex": "Male",
					"tags": []any{"tag1", "tag2"},
					"age":  []any{34, 23},
				},
				expectVal: []any{"tag1_34", "tag1_23", "tag2_34", "tag2_23"},
			},
		}

		trigger := NewTrigger(config)

		for _, testcase := range testcases {
			fmt.Println(trigger.GetValue(testcase.features))
			triggers := ParseTriggerId(trigger.GetValue(testcase.features))
			t.Log(triggers...)
			assert.Equal(t, triggers, testcase.expectVal)
		}

	})
}
