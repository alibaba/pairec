package datahub

import (
	"reflect"
	"testing"
)

func TestSyncLogDatahubItem(t *testing.T) {
	tests := []map[string]any{
		{
			"int_k":     int(123),
			"int64_k":   int64(890765),
			"float32_k": float32(0.4567),
			"float64_k": float64(12.129876),
			"bool_k":    true,
			"string_k":  "hello world",
		},
		{
			"int_k":     int(1234),
			"int64_k":   int64(890765),
			"float32_k": float32(0.67),
			"float64_k": float64(12.9876),
			"bool_k":    false,
			"string_k":  "h",
		},
	}

	for _, test := range tests {
		item := NewSyncLogDatahubItem(test)
		buf := item.Format()

		item.data = nil
		if err := item.Parse(buf); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(test, item.data) {
			t.Fatalf("test failed, expect: %v, got: %v", test, item.data)
		}
		t.Log(item.data)
	}
}
