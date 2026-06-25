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

func TestParseTriggerValues(t *testing.T) {
	t.Run("single value single dimension", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "category"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"category": "sports"}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "sports")
	})

	t.Run("array value single dimension", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "triggers"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"triggers": []string{"trigger_a", "trigger_b"}}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "trigger_a,trigger_b")
	})

	t.Run("array value with underscore in values", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "tags"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"tags": []any{"item_tag_1", "item_tag_2", "item_tag_3"}}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "item_tag_1,item_tag_2,item_tag_3")
	})

	t.Run("multi dimension single values", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "sex"},
			{TriggerKey: "os"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"sex": "Male", "os": "IOS"}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "Male_IOS")
	})

	t.Run("multi dimension with array", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "tags"},
			{TriggerKey: "os"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"tags": []string{"tag_a", "tag_b"}, "os": "Android"}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "tag_a_Android,tag_b_Android")
	})

	t.Run("multi dimension both arrays", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "tags"},
			{TriggerKey: "levels"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"tags": []string{"tag_a", "tag_b"}, "levels": []any{"L1", "L2"}}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "tag_a_L1,tag_a_L2,tag_b_L1,tag_b_L2")
	})

	t.Run("with boundaries", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "sex"},
			{TriggerKey: "age", Boundaries: []int{20, 30, 40}},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"sex": "Male", "age": 25}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "Male_20-30")
	})

	t.Run("missing feature uses default", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "tags"},
			{TriggerKey: "city", DefaultValue: "unknown"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"tags": []string{"trigger_a", "trigger_b"}}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "trigger_a_unknown,trigger_b_unknown")
	})

	t.Run("empty string value preserves dimension", func(t *testing.T) {
		// When first dimension is empty string, second dimension should still be "_Android"
		result := ParseTriggerValues([]string{"", "Android"})
		assert.Equal(t, result, "_Android")
	})

	t.Run("empty array produces empty string", func(t *testing.T) {
		config := []recconf.TriggerConfig{
			{TriggerKey: "tags"},
			{TriggerKey: "os"},
		}
		trigger := NewTrigger(config)
		features := map[string]interface{}{"tags": []string{}, "os": "Android"}
		result := ParseTriggerValues(trigger.GetTriggerValues(features))
		assert.Equal(t, result, "_Android")
	})
}
