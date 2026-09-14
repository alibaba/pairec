package recall

import (
	"encoding/json"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/recconf"
)

// FillMissingToolParams applies only explicitly enabled knowledge defaults.
// No usable record leaves the original parameters unchanged, without an error.
func (r *Ha3ChatRecall) FillMissingToolParams(current aichat.SearchGoodsParams, previous *aichat.SearchGoodsParams, evidence ToolParamEvidence) aichat.SearchGoodsParams {
	provider, ok := evidence.(interface {
		Records() []map[string]interface{}
	})
	if !ok {
		return current
	}
	params := make(map[string]recconf.SearchToolParamConfig)
	for _, param := range r.toolParams() {
		params[param.Name] = param
	}
	// Missing parameters from the same parent chain are selected together, so
	// defaults cannot combine unrelated parent and child records.
	groups := make(map[string][]recconf.SearchToolParamConfig)
	var roots []string
	for _, param := range r.toolParams() {
		if !param.FillMissingFromKnowledge || param.KnowledgeField == "" || !missingToolParam(current.ToolParams[param.Name]) {
			continue
		}
		root := param.Name
		for params[root].ParentParam != "" {
			root = params[root].ParentParam
		}
		if _, exists := groups[root]; !exists {
			roots = append(roots, root)
		}
		groups[root] = append(groups[root], param)
	}
	if len(roots) == 0 {
		return current
	}
	records := provider.Records()
	for _, root := range roots {
		for _, record := range records {
			candidate := current
			candidate.ToolParams = copyToolParamValues(current.ToolParams)
			complete := true
			for _, param := range groups[root] {
				value := knowledgeDefaultValue(record[param.KnowledgeField])
				if len(value) == 0 {
					complete = false
					break
				}
				candidate.ToolParams[param.Name] = value
			}
			if !complete {
				continue
			}
			// Other missing values may be filled by a later independent group.
			// Ignore them only while checking this candidate; retain their original
			// representation in the returned request and its normal validation.
			check := candidate
			check.ToolParams = copyToolParamValues(candidate.ToolParams)
			for name, value := range check.ToolParams {
				if missingToolParam(value) {
					delete(check.ToolParams, name)
				}
			}
			if r.ValidateToolParamSources(check, previous, evidence) != nil {
				continue
			}
			current = candidate
			break
		}
	}
	return current
}

func missingToolParam(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true
	}
	var values []string
	return json.Unmarshal(raw, &values) == nil && len(values) == 0
}

func knowledgeDefaultValue(raw interface{}) json.RawMessage {
	if value, ok := raw.(string); ok {
		raw = []string{value}
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	values, err := toolTermValues(encoded)
	if err != nil || len(values) == 0 {
		return nil
	}
	encoded, _ = json.Marshal(values)
	return encoded
}

func copyToolParamValues(values map[string]json.RawMessage) map[string]json.RawMessage {
	copied := make(map[string]json.RawMessage, len(values))
	for name, value := range values {
		copied[name] = value
	}
	return copied
}
