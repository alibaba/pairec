package shoppingknowledge

import (
	"encoding/json"
	"strings"

	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

const promptMaxBytes = 6000

// SuggestionKnowledge is the knowledge view shared by all model inputs.
// Retrieval metadata stays server-side and must not enter this view.
type SuggestionKnowledge struct {
	Value string `json:"value"`
}

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
	for _, hit := range result.Hits {
		value := strings.TrimSpace(hit.Value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		prospective := append(append([]SuggestionKnowledge(nil), evidence.values...), SuggestionKnowledge{Value: value})
		payload, err := json.Marshal(prospective)
		if err != nil || len(payload) > promptMaxBytes {
			break
		}
		seen[value] = struct{}{}
		evidence.hits = append(evidence.hits, hit)
		evidence.values = prospective
		evidence.promptJSON = string(payload)
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
