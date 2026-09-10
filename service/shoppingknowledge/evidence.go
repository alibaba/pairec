package shoppingknowledge

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/log"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

const promptMaxBytes = 6000

// SuggestionKnowledge contains only explicitly configured model-visible fields.
type SuggestionKnowledge map[string]interface{}

type Evidence struct {
	hits       []recallsvc.KnowledgeHit
	values     []SuggestionKnowledge
	promptJSON string
}

func NewEvidence(result *recallsvc.KnowledgeSearchResult) *Evidence {
	if result == nil {
		return nil
	}
	evidence := &Evidence{}
	seen := make(map[string]struct{}, len(result.Hits))
	skipped := 0
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
		prospective := append(append([]SuggestionKnowledge(nil), evidence.values...), SuggestionKnowledge(hit.ModelFields))
		payload, err := json.Marshal(prospective)
		if err != nil || len(payload) > promptMaxBytes {
			skipped++
			continue
		}
		evidence.hits = append(evidence.hits, hit)
		evidence.values = prospective
		evidence.promptJSON = string(payload)
	}
	if skipped > 0 {
		log.Info(fmt.Sprintf("module=ShoppingKnowledge\tevent=model_view_budget\tkept=%d\tskipped=%d\tmaxBytes=%d", len(evidence.values), skipped, promptMaxBytes))
	}
	if len(evidence.values) == 0 {
		return nil
	}
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

func (e *Evidence) LogSummary() []map[string]string {
	if e == nil {
		return nil
	}
	result := make([]map[string]string, 0, len(e.hits))
	for _, hit := range e.hits {
		result = append(result, map[string]string{
			"knowledge_id":   hit.KnowledgeID,
			"knowledge_type": hit.KnowledgeType,
			"value":          hit.Value,
			"category":       hit.Category,
		})
	}
	return result
}

func (e *Evidence) SuggestionKnowledge() []SuggestionKnowledge {
	if e == nil {
		return nil
	}
	return append([]SuggestionKnowledge(nil), e.values...)
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
