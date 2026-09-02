package shoppingknowledge

import (
	"encoding/json"
	"fmt"
	"strings"

	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

const (
	promptMaxBytes          = 6000
	candidateSelectionLimit = 4
)

type Candidate struct {
	KnowledgeID   string `json:"-"`
	CandidateID   string `json:"candidate_id"`
	KnowledgeType string `json:"knowledge_type"`
	Value         string `json:"value"`
	Category      string `json:"category"`
	SearchTerm    string `json:"search_term"`
}

type SuggestionKnowledge struct {
	KnowledgeType     string `json:"knowledge_type"`
	Value             string `json:"value"`
	Category          string `json:"category"`
	SearchTerm        string `json:"search_term"`
	SelectedByPlanner bool   `json:"selected_by_planner,omitempty"`
}

type Evidence struct {
	candidates []Candidate
	byID       map[string]Candidate
	promptJSON string
}

func NewEvidence(result *recallsvc.KnowledgeSearchResult) *Evidence {
	if result == nil || len(result.Hits) == 0 {
		return nil
	}
	evidence := &Evidence{byID: make(map[string]Candidate)}
	seen := make(map[string]struct{}, len(result.Hits))
	seenKnowledgeIDs := make(map[string]struct{}, len(result.Hits))
	for _, hit := range result.Hits {
		if _, ok := seenKnowledgeIDs[hit.KnowledgeID]; ok {
			continue
		}
		searchTerm := searchTerm(hit.KnowledgeType, hit.Value)
		if searchTerm == "" || strings.TrimSpace(hit.Category) == "" {
			continue
		}
		key := strings.ToLower(hit.KnowledgeType + "\x00" + searchTerm + "\x00" + hit.Category)
		if _, ok := seen[key]; ok {
			continue
		}
		candidate := Candidate{
			KnowledgeID:   hit.KnowledgeID,
			CandidateID:   fmt.Sprintf("K%d", len(evidence.candidates)+1),
			KnowledgeType: strings.TrimSpace(hit.KnowledgeType),
			Value:         strings.TrimSpace(hit.Value),
			Category:      strings.TrimSpace(hit.Category),
			SearchTerm:    searchTerm,
		}
		prospective := append(append([]Candidate(nil), evidence.candidates...), candidate)
		payload, err := json.Marshal(prospective)
		if err != nil || len(payload) > promptMaxBytes {
			break
		}
		seen[key] = struct{}{}
		seenKnowledgeIDs[hit.KnowledgeID] = struct{}{}
		evidence.candidates = prospective
		evidence.byID[candidate.CandidateID] = candidate
		evidence.promptJSON = string(payload)
	}
	if len(evidence.candidates) == 0 {
		return nil
	}
	return evidence
}

func (e *Evidence) Len() int {
	if e == nil {
		return 0
	}
	return len(e.candidates)
}

func (e *Evidence) PromptJSON() string {
	if e == nil {
		return ""
	}
	return e.promptJSON
}

func (e *Evidence) CandidateIDs() []string {
	if e == nil {
		return nil
	}
	ids := make([]string, 0, len(e.candidates))
	for _, candidate := range e.candidates {
		ids = append(ids, candidate.CandidateID)
	}
	return ids
}

func (e *Evidence) LogSummary() []map[string]string {
	if e == nil {
		return nil
	}
	result := make([]map[string]string, 0, len(e.candidates))
	for _, candidate := range e.candidates {
		result = append(result, map[string]string{
			"candidate_id":   candidate.CandidateID,
			"knowledge_id":   candidate.KnowledgeID,
			"knowledge_type": candidate.KnowledgeType,
			"search_term":    candidate.SearchTerm,
			"category":       candidate.Category,
		})
	}
	return result
}

func (e *Evidence) SuggestionKnowledge(selectedIDs []string) []SuggestionKnowledge {
	if e == nil {
		return nil
	}
	selected := make(map[string]struct{}, len(selectedIDs))
	for _, id := range selectedIDs {
		selected[id] = struct{}{}
	}
	result := make([]SuggestionKnowledge, 0, len(e.candidates))
	for _, candidate := range e.candidates {
		_, isSelected := selected[candidate.CandidateID]
		result = append(result, SuggestionKnowledge{
			KnowledgeType:     candidate.KnowledgeType,
			Value:             candidate.Value,
			Category:          candidate.Category,
			SearchTerm:        candidate.SearchTerm,
			SelectedByPlanner: isSelected,
		})
	}
	return result
}

func (e *Evidence) Apply(req *recallsvc.SearchGoodsRequest) error {
	if e == nil {
		if len(req.KnowledgeCandidateIDs) > 0 {
			return fmt.Errorf("knowledge_candidate_ids is unavailable for this request")
		}
		return nil
	}
	if req.KnowledgeCandidateIDs == nil {
		return fmt.Errorf("knowledge_candidate_ids is required")
	}
	if len(req.KnowledgeCandidateIDs) == 0 {
		return nil
	}
	seenIDs := make(map[string]struct{}, len(req.KnowledgeCandidateIDs))
	selectedIDs := make([]string, 0, candidateSelectionLimit)
	searchTerms := make([]string, 0, candidateSelectionLimit)
	seenTerms := make(map[string]struct{}, len(req.KnowledgeCandidateIDs))
	for _, id := range req.KnowledgeCandidateIDs {
		if _, exists := seenIDs[id]; exists {
			return fmt.Errorf("knowledge_candidate_ids contains duplicate value %q", id)
		}
		seenIDs[id] = struct{}{}
		candidate, exists := e.byID[id]
		if !exists {
			return fmt.Errorf("knowledge_candidate_ids contains unknown value %q", id)
		}
		key := strings.ToLower(candidate.SearchTerm)
		if _, exists := seenTerms[key]; exists {
			continue
		}
		seenTerms[key] = struct{}{}
		if len(searchTerms) >= candidateSelectionLimit {
			continue
		}
		selectedIDs = append(selectedIDs, id)
		searchTerms = append(searchTerms, candidate.SearchTerm)
	}
	if len(searchTerms) == 0 {
		return fmt.Errorf("knowledge_candidate_ids does not resolve to a search term")
	}
	req.KnowledgeCandidateIDs = selectedIDs
	req.ProductTypeKeywords = searchTerms
	return nil
}

func searchTerm(knowledgeType, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if knowledgeType != "categories" {
		return value
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == '/' || r == '>' })
	for index := len(parts) - 1; index >= 0; index-- {
		if part := strings.TrimSpace(parts[index]); part != "" {
			return part
		}
	}
	return ""
}
