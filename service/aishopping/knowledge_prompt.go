package aishopping

import (
	"encoding/json"
	"fmt"
	"strings"

	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

const (
	knowledgePromptMaxBytes          = 6000
	knowledgeCandidateSelectionLimit = 4
)

type knowledgeCandidate struct {
	KnowledgeID   string `json:"-"`
	CandidateID   string `json:"candidate_id"`
	KnowledgeType string `json:"knowledge_type"`
	Value         string `json:"value"`
	Category      string `json:"category"`
	SearchTerm    string `json:"search_term"`
}

type knowledgeEvidence struct {
	candidates []knowledgeCandidate
	byID       map[string]knowledgeCandidate
	promptJSON string
}

func newKnowledgeEvidence(result *recallsvc.KnowledgeSearchResult) *knowledgeEvidence {
	if result == nil || len(result.Hits) == 0 {
		return nil
	}
	evidence := &knowledgeEvidence{byID: make(map[string]knowledgeCandidate)}
	seen := make(map[string]struct{}, len(result.Hits))
	seenKnowledgeIDs := make(map[string]struct{}, len(result.Hits))
	for _, hit := range result.Hits {
		if _, ok := seenKnowledgeIDs[hit.KnowledgeID]; ok {
			continue
		}
		searchTerm := knowledgeSearchTerm(hit.KnowledgeType, hit.Value)
		if searchTerm == "" || hit.Category == "" {
			continue
		}
		key := strings.ToLower(hit.KnowledgeType + "\x00" + searchTerm + "\x00" + hit.Category)
		if _, ok := seen[key]; ok {
			continue
		}
		candidate := knowledgeCandidate{
			KnowledgeID:   hit.KnowledgeID,
			CandidateID:   fmt.Sprintf("K%d", len(evidence.candidates)+1),
			KnowledgeType: hit.KnowledgeType,
			Value:         hit.Value,
			Category:      hit.Category,
			SearchTerm:    searchTerm,
		}
		prospective := append(append([]knowledgeCandidate(nil), evidence.candidates...), candidate)
		payload, err := json.Marshal(prospective)
		if err != nil || len(payload) > knowledgePromptMaxBytes {
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

func (e *knowledgeEvidence) logSummary() []map[string]string {
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

func knowledgeSearchTerm(knowledgeType, value string) string {
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

func (e *knowledgeEvidence) candidateIDs() []string {
	if e == nil {
		return nil
	}
	ids := make([]string, 0, len(e.candidates))
	for _, candidate := range e.candidates {
		ids = append(ids, candidate.CandidateID)
	}
	return ids
}

func (e *knowledgeEvidence) apply(req *recallsvc.SearchGoodsRequest) error {
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
	selectedIDs := make([]string, 0, knowledgeCandidateSelectionLimit)
	searchTerms := make([]string, 0, knowledgeCandidateSelectionLimit)
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
		if len(searchTerms) >= knowledgeCandidateSelectionLimit {
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
