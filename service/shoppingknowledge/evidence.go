package shoppingknowledge

import (
	"encoding/json"
	"strings"

	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

// SuggestionKnowledge contains only explicitly configured model-visible fields.
type SuggestionKnowledge map[string]interface{}

type Evidence struct {
	values     []SuggestionKnowledge
	promptJSON string
}

func NewEvidence(result *recallsvc.KnowledgeSearchResult) *Evidence {
	if result == nil {
		return nil
	}
	evidence := &Evidence{}
	seen := make(map[string]struct{}, len(result.Hits))
	var payload strings.Builder
	payload.WriteByte('[')
	for _, hit := range result.Hits {
		if len(hit.ModelFields) == 0 {
			continue
		}
		record, err := json.Marshal(hit.ModelFields)
		if err != nil {
			continue
		}
		key := string(record)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		if len(evidence.values) > 0 {
			payload.WriteByte(',')
		}
		payload.Write(record)
		evidence.values = append(evidence.values, SuggestionKnowledge(hit.ModelFields))
	}
	if len(evidence.values) == 0 {
		return nil
	}
	payload.WriteByte(']')
	evidence.promptJSON = payload.String()
	return evidence
}

func (e *Evidence) Len() int {
	if e == nil {
		return 0
	}
	return len(e.values)
}

func (e *Evidence) PromptJSON() string {
	if e == nil {
		return ""
	}
	return e.promptJSON
}

func (e *Evidence) SuggestionKnowledge() []SuggestionKnowledge {
	if e == nil {
		return nil
	}
	return append([]SuggestionKnowledge(nil), e.values...)
}

// Records returns the model-visible records in retrieval order for configured
// missing-parameter defaults. Callers must not mutate the record maps.
func (e *Evidence) Records() []map[string]interface{} {
	if e == nil {
		return nil
	}
	records := make([]map[string]interface{}, len(e.values))
	for i, value := range e.values {
		records[i] = value
	}
	return records
}

func (e *Evidence) Contains(field, value string) bool {
	if e == nil {
		return false
	}
	for _, record := range e.values {
		if knowledgeContains(record[field], value) {
			return true
		}
	}
	return false
}

func (e *Evidence) ContainsPair(field, value, otherField, otherValue string) bool {
	if e == nil {
		return false
	}
	for _, record := range e.values {
		if knowledgeContains(record[field], value) && knowledgeContains(record[otherField], otherValue) {
			return true
		}
	}
	return false
}

func knowledgeContains(raw interface{}, value string) bool {
	switch raw := raw.(type) {
	case string:
		return raw == value
	case []string:
		for _, entry := range raw {
			if entry == value {
				return true
			}
		}
	case []interface{}:
		found := false
		for _, entry := range raw {
			text, ok := entry.(string)
			if !ok {
				return false
			}
			found = found || text == value
		}
		return found
	}
	return false
}
