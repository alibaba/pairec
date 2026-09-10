package aichat

import (
	"encoding/json"
	"fmt"
)

// SearchGoodsParams is shared by tool calls, session snapshots and suggestions.
// Configured extension parameters remain at the top level of its JSON object.
type SearchGoodsParams struct {
	Keywords          []string                   `json:"keywords"`
	PreferredKeywords []string                   `json:"preferred_keywords,omitempty"`
	Constraints       map[string]json.RawMessage `json:"constraints,omitempty"`
	ExcludeKeywords   []string                   `json:"exclude_keywords,omitempty"`
	MinPrice          *float64                   `json:"min_price,omitempty"`
	MaxPrice          *float64                   `json:"max_price,omitempty"`
	ToolParams        map[string]json.RawMessage `json:"-"`
}

func IsBuiltinSearchParam(name string) bool {
	switch name {
	case "keywords", "preferred_keywords", "constraints", "exclude_keywords", "min_price", "max_price":
		return true
	}
	return false
}

func (p SearchGoodsParams) MarshalJSON() ([]byte, error) {
	type plain SearchGoodsParams
	data, err := json.Marshal(plain(p))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return nil, err
	}
	for name, value := range p.ToolParams {
		if IsBuiltinSearchParam(name) {
			return nil, fmt.Errorf("extension parameter conflicts with %q", name)
		}
		fields[name] = value
	}
	return json.Marshal(fields)
}

func (p *SearchGoodsParams) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if fields == nil {
		return fmt.Errorf("search parameters must be an object")
	}
	type plain SearchGoodsParams
	var parsed plain
	builtin := make(map[string]json.RawMessage)
	for name, value := range fields {
		if IsBuiltinSearchParam(name) {
			builtin[name] = value
			delete(fields, name)
		}
	}
	payload, err := json.Marshal(builtin)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return err
	}
	parsed.ToolParams = fields
	*p = SearchGoodsParams(parsed)
	return nil
}
