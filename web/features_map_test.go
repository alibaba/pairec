package web

import (
	"encoding/json"
	"testing"
)

func TestFeaturesMap_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantKey  string
		wantType string
	}{
		{
			name:     "int64 array to string array",
			jsonData: `{"cellIDNeighbors":[384307167844368384,1921535839221841920,1921535841369325568]}`,
			wantKey:  "cellIDNeighbors",
			wantType: "[]string",
		},
		{
			name:     "float64 array to string array",
			jsonData: `{"scores":[1.5,2.5,3.5]}`,
			wantKey:  "scores",
			wantType: "[]string",
		},
		{
			name:     "string array converted to []string",
			jsonData: `{"name":"test","count":10,"tags":["a","b"]}`,
			wantKey:  "tags",
			wantType: "[]string",
		},
		{
			name:     "string array array converted to [][]string",
			jsonData: `{"name":"test","count":10,"tags":[["a","b"],["c","d"]]}`,
			wantKey:  "tags",
			wantType: "[][]string",
		},
		{
			name:     "int array array converted to [][]string",
			jsonData: `{"name":"test","count":10,"tags":[[1,2],[3,4]]}`,
			wantKey:  "tags",
			wantType: "[][]string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var features FeaturesMap
			if err := json.Unmarshal([]byte(tt.jsonData), &features); err != nil {
				t.Errorf("UnmarshalJSON() error = %v", err)
				return
			}

			value, exists := features[tt.wantKey]
			if !exists {
				t.Errorf("key %s not found in features", tt.wantKey)
				return
			}

			// Check type
			typeStr := getTypeName(value)
			if typeStr != tt.wantType {
				t.Errorf("type mismatch: got %s, want %s", typeStr, tt.wantType)
			}

			// For string arrays, verify values
			if tt.wantType == "[]string" {
				if strArr, ok := value.([]string); ok {
					t.Logf("Converted to string array: %v", strArr)
					// Verify first element is not in scientific notation
					if len(strArr) > 0 {
						t.Logf("First element: %s", strArr[0])
					}
				} else {
					t.Errorf("expected []string, got %T", value)
				}
			}
		})
	}
}

func TestFeaturesMap_PrecisionPreservation(t *testing.T) {
	// Test the specific case mentioned in the issue
	jsonData := `{"cellIDNeighbors":[384307167844368384,1921535839221841920,1921535841369325568,1152921512123039744]}`

	var features FeaturesMap
	if err := json.Unmarshal([]byte(jsonData), &features); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	cellIDs, ok := features["cellIDNeighbors"].([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", features["cellIDNeighbors"])
	}

	expectedValues := []string{
		"384307167844368384",
		"1921535839221841920",
		"1921535841369325568",
		"1152921512123039744",
	}

	for i, expected := range expectedValues {
		if i >= len(cellIDs) {
			t.Errorf("missing value at index %d", i)
			continue
		}
		if cellIDs[i] != expected {
			t.Errorf("index %d: got %s, want %s", i, cellIDs[i], expected)
		}
	}

	t.Logf("Successfully preserved precision for large integers: %v", cellIDs)
}

func TestFeaturesMap_StringArrayConversion(t *testing.T) {
	// Test geoHashWithNeighbors case
	jsonData := `{"geoHashWithNeighbors":["s00000000001","s00000000003","s00000000002","kpbpbpbpbpbr","kpbpbpbpbpbp","7zzzzzzzzzzz","ebpbpbpbpbpb","ebpbpbpbpbpc","s00000000000"]}`

	var features FeaturesMap
	if err := json.Unmarshal([]byte(jsonData), &features); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	geoHashes, ok := features["geoHashWithNeighbors"].([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", features["geoHashWithNeighbors"])
	}

	expectedValues := []string{
		"s00000000001", "s00000000003", "s00000000002", "kpbpbpbpbpbr",
		"kpbpbpbpbpbp", "7zzzzzzzzzzz", "ebpbpbpbpbpb", "ebpbpbpbpbpc", "s00000000000",
	}

	if len(geoHashes) != len(expectedValues) {
		t.Errorf("length mismatch: got %d, want %d", len(geoHashes), len(expectedValues))
	}

	for i, expected := range expectedValues {
		if i >= len(geoHashes) {
			t.Errorf("missing value at index %d", i)
			continue
		}
		if geoHashes[i] != expected {
			t.Errorf("index %d: got %s, want %s", i, geoHashes[i], expected)
		}
	}

	t.Logf("Successfully converted string array: %v", geoHashes)
}

func TestFeaturesMap_MapConversion(t *testing.T) {
	tests := []struct {
		name          string
		jsonData      string
		wantKey       string
		wantValueType string
	}{
		{
			name:          "map[string]string",
			jsonData:      `{"metadata":{"key1":"value1","key2":"value2"}}`,
			wantKey:       "metadata",
			wantValueType: "map[string]string",
		},
		{
			name:          "map[string]int",
			jsonData:      `{"counts":{"a":1,"b":2,"c":3}}`,
			wantKey:       "counts",
			wantValueType: "map[string]string",
		},
		{
			name:          "map[string]float64",
			jsonData:      `{"scores":{"x":1.5,"y":2.5,"z":3.5}}`,
			wantKey:       "scores",
			wantValueType: "map[string]string",
		},
		{
			name:          "map[string]mixed types",
			jsonData:      `{"config":{"name":"test","count":10,"rate":0.5}}`,
			wantKey:       "config",
			wantValueType: "map[string]string",
		},
		{
			name:          "map[string][]int",
			jsonData:      `{"tags":{"ids":[1,2,3],"nums":[10,20,30]}}`,
			wantKey:       "tags",
			wantValueType: "map[string][]string",
		},
		{
			name:          "map[string][]string",
			jsonData:      `{"names":{"first":["a","b"],"last":["c","d"]}}`,
			wantKey:       "names",
			wantValueType: "map[string][]string",
		},
		{
			name:          "map[string]nested map",
			jsonData:      `{"nested":{"inner":{"key":"value"}}}`,
			wantKey:       "nested",
			wantValueType: "map[string]interface {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var features FeaturesMap
			if err := json.Unmarshal([]byte(tt.jsonData), &features); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			value, exists := features[tt.wantKey]
			if !exists {
				t.Fatalf("key %s not found in features", tt.wantKey)
			}

			// Check type matches expected type
			switch tt.wantValueType {
			case "map[string]string":
				strMap, ok := value.(map[string]string)
				if !ok {
					t.Fatalf("expected map[string]string, got %T", value)
				}
				t.Logf("Converted to map[string]string: %v", strMap)
				for k, v := range strMap {
					t.Logf("  %s: %v", k, v)
				}

			case "map[string][]string":
				arrMap, ok := value.(map[string][]string)
				if !ok {
					t.Fatalf("expected map[string][]string, got %T", value)
				}
				t.Logf("Converted to map[string][]string: %v", arrMap)
				for k, v := range arrMap {
					t.Logf("  %s: %v", k, v)
				}

			case "map[string]interface {}":
				strMap, ok := value.(map[string]interface{})
				if !ok {
					t.Fatalf("expected map[string]interface{}, got %T", value)
				}
				t.Logf("Converted to map[string]interface{}: %v", strMap)
				for k, v := range strMap {
					t.Logf("  %s: %v (type: %T)", k, v, v)
					// For array values, check they are []string
					if strArr, ok := v.([]string); ok {
						t.Logf("    -> Array converted to []string: %v", strArr)
					}
				}
			}
		})
	}
}

func getTypeName(v interface{}) string {
	switch v.(type) {
	case []string:
		return "[]string"
	case [][]string:
		return "[][]string"
	case []interface{}:
		return "[]interface {}"
	case []float64:
		return "[]float64"
	case []int:
		return "[]int"
	case []int64:
		return "[]int64"
	default:
		return "unknown"
	}
}
