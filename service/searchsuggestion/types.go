package searchsuggestion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/alibaba/pairec/v2/service/shoppingknowledge"
)

const (
	EmbeddedCount   = 3
	StandaloneCount = 1
)

type ConversationTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type SearchIntent struct {
	Keywords            []string `json:"keywords"`
	ProductTypeKeywords []string `json:"product_type_keywords"`
	AttributeKeywords   []string `json:"attribute_keywords"`
	Operator            string   `json:"operator"`
	ExcludeKeywords     []string `json:"exclude_keywords,omitempty"`
	MinPrice            *float64 `json:"min_price,omitempty"`
	MaxPrice            *float64 `json:"max_price,omitempty"`

	SelectedKnowledgeCandidateIDs []string `json:"-"`
}

type ProductSummary struct {
	Total int                 `json:"total"`
	Hits  []ProductSummaryHit `json:"hits"`
}

type ProductSummaryHit struct {
	Index    int    `json:"index"`
	Title    string `json:"title,omitempty"`
	Category string `json:"category,omitempty"`
	Brand    string `json:"brand,omitempty"`
	Price    string `json:"price,omitempty"`
}

type GenerationInput struct {
	Language              string                                  `json:"language"`
	SuggestionCount       int                                     `json:"suggestion_count"`
	CurrentQuery          string                                  `json:"current_query,omitempty"`
	Conversation          []ConversationTurn                      `json:"conversation,omitempty"`
	FinalSearchIntent     *SearchIntent                           `json:"final_search_intent,omitempty"`
	Knowledge             []shoppingknowledge.SuggestionKnowledge `json:"knowledge"`
	CurrentProductSummary *ProductSummary                         `json:"current_product_summary,omitempty"`
}

func BuildStandaloneInput(language, query string, knowledge []shoppingknowledge.SuggestionKnowledge) (*GenerationInput, *Error) {
	query = strings.TrimSpace(query)
	if query == "" || len(knowledge) == 0 {
		return nil, NewError(CodeKnowledgeEmpty, false, fmt.Errorf("query and knowledge are required"))
	}
	return &GenerationInput{
		Language:        language,
		SuggestionCount: StandaloneCount,
		CurrentQuery:    query,
		Knowledge:       append([]shoppingknowledge.SuggestionKnowledge(nil), knowledge...),
	}, nil
}

func (in *GenerationInput) validate() *Error {
	if in == nil || in.SuggestionCount <= 0 || len(in.Knowledge) == 0 {
		return NewError(CodeInvalidStructure, false, fmt.Errorf("invalid suggestion context"))
	}
	if in.FinalSearchIntent == nil && strings.TrimSpace(in.CurrentQuery) == "" {
		return NewError(CodeInvalidStructure, false, fmt.Errorf("search baseline is empty"))
	}
	if in.FinalSearchIntent != nil && in.CurrentProductSummary == nil {
		return NewError(CodeProductContextInvalid, true, fmt.Errorf("product summary is missing"))
	}
	return nil
}

func (in *GenerationInput) JSON() ([]byte, *Error) {
	if err := in.validate(); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(in)
	if err != nil {
		return nil, NewError(CodeInvalidStructure, false, err)
	}
	return payload, nil
}

func BuildProductSummary(payload string, expectedIndexes []int, expectedTotal int) (*ProductSummary, *Error) {
	type rawHit struct {
		Index int                        `json:"index"`
		Title string                     `json:"title"`
		Raw   map[string]json.RawMessage `json:"raw"`
	}
	type rawPayload struct {
		Total int      `json:"total"`
		Hits  []rawHit `json:"hits"`
	}
	var decoded rawPayload
	decoder := json.NewDecoder(bytes.NewBufferString(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return nil, NewError(CodeProductContextInvalid, true, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return nil, NewError(CodeProductContextInvalid, true, err)
	}
	if decoded.Total != expectedTotal || len(decoded.Hits) != len(expectedIndexes) {
		return nil, NewError(CodeProductContextInvalid, true, fmt.Errorf("product snapshot mismatch"))
	}
	summary := &ProductSummary{Total: decoded.Total, Hits: make([]ProductSummaryHit, 0, len(decoded.Hits))}
	seen := make(map[int]struct{}, len(decoded.Hits))
	for i, hit := range decoded.Hits {
		if hit.Index <= 0 || hit.Index != expectedIndexes[i] {
			return nil, NewError(CodeProductContextInvalid, true, fmt.Errorf("product index mismatch"))
		}
		if _, exists := seen[hit.Index]; exists {
			return nil, NewError(CodeProductContextInvalid, true, fmt.Errorf("duplicate product index"))
		}
		seen[hit.Index] = struct{}{}
		summary.Hits = append(summary.Hits, ProductSummaryHit{
			Index:    hit.Index,
			Title:    truncateRunes(strings.TrimSpace(hit.Title), 96),
			Category: rawScalar(hit.Raw, 64, "category", "category_name"),
			Brand:    rawScalar(hit.Raw, 48, "brand", "brand_name"),
			Price:    rawScalar(hit.Raw, 32, "price"),
		})
	}
	return summary, nil
}

func rawScalar(values map[string]json.RawMessage, limit int, names ...string) string {
	for _, name := range names {
		for key, raw := range values {
			if !strings.EqualFold(key, name) {
				continue
			}
			var value interface{}
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				continue
			}
			switch typed := value.(type) {
			case string:
				return truncateRunes(strings.TrimSpace(typed), limit)
			case json.Number:
				return truncateRunes(typed.String(), limit)
			case bool:
				return fmt.Sprint(typed)
			}
		}
	}
	return ""
}

func truncateRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}
