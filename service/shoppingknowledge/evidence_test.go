package shoppingknowledge

import (
	"strings"
	"testing"

	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

func TestEvidenceContainsNonStringEntries(t *testing.T) {
	result := &recallsvc.KnowledgeSearchResult{
		Hits: []recallsvc.KnowledgeHit{
			{
				KnowledgeID:   "1",
				KnowledgeType: "category",
				Value:         "phone",
				ModelFields: map[string]interface{}{
					"category": "phone",
					"brands":   []string{"Apple", "HUAWEI"},
					"prices":   []interface{}{float64(100), float64(200)},
					"mixed":    []interface{}{"budget", float64(100)},
				},
			},
		},
	}
	evidence := NewEvidence(result)

	// The model selects terms from the prompt JSON, where numbers appear as
	// plain JSON text (100, not 100.0), so numeric entries must match that text.
	if !strings.Contains(evidence.PromptJSON(), `"prices":[100,200]`) {
		t.Errorf("prompt JSON %q does not show numeric entries as JSON text", evidence.PromptJSON())
	}

	cases := []struct {
		field string
		value string
		want  bool
	}{
		{"category", "phone", true},
		{"brands", "Apple", true},
		{"prices", "100", true},
		{"prices", "200", true},
		{"prices", "300", false},
		{"mixed", "budget", true},
		{"mixed", "100", true},
		{"mixed", "premium", false},
		{"missing", "100", false},
	}
	for _, tc := range cases {
		if got := evidence.Contains(tc.field, tc.value); got != tc.want {
			t.Errorf("Contains(%q, %q) = %v, want %v", tc.field, tc.value, got, tc.want)
		}
	}

	if !evidence.ContainsPair("mixed", "100", "category", "phone") {
		t.Error("ContainsPair should match a numeric entry and a string entry in the same record")
	}
	if evidence.ContainsPair("mixed", "100", "category", "tv") {
		t.Error("ContainsPair should fail when the other field value is absent")
	}
}
