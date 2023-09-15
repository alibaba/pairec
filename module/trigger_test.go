package module

import (
	"fmt"
	"testing"

	"github.com/alibaba/pairec/recconf"
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

	trigger := NewTrigger(config)

	features := []map[string]interface{}{
		{"sex": "Male",
			"os":  "IOS",
			"age": 23,
		},

		{"sex": "Male",
			"os": "Android",
		},
		{"sex": "Male",
			"os":  "Android",
			"age": 60,
		},
		{"sex": "Male",
			"os":  "Android",
			"age": 50,
		},
		{"sex": "Male",
			"os":  "Android",
			"age": 40,
		},
		{"sex": "Male",
			"os":  "Android",
			"age": 20,
		},
	}

	for _, f := range features {
		fmt.Println(trigger.GetValue(f))
	}
}
