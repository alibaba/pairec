package recall

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/recconf"
)

// ToolParamEvidence contains only knowledge records actually shown to the model.
type ToolParamEvidence interface {
	Contains(field, value string) bool
	ContainsPair(field, value, otherField, otherValue string) bool
}

func validateSearchToolParams(conf *recconf.SearchGoodsConfig, knowledge *recconf.Ha3KnowledgeVectorConfig) {
	if conf == nil {
		return
	}
	visible := make(map[string]bool)
	if knowledge != nil {
		for _, field := range knowledge.ModelFields {
			visible[field.Name] = true
		}
	}
	names := make(map[string]bool)
	for _, param := range conf.ConstraintParams {
		names[param.Name] = true
	}
	params := make(map[string]recconf.SearchToolParamConfig)
	for _, param := range conf.ToolParams {
		if !validHa3FieldName(param.Name) || strings.Contains(param.Name, ".") || names[param.Name] || aichat.IsBuiltinSearchParam(param.Name) {
			panic(fmt.Sprintf("invalid or conflicting tool parameter name %q", param.Name))
		}
		names[param.Name] = true
		if param.Kind != "terms" || !validHa3FieldName(param.Field) {
			panic(fmt.Sprintf("tool parameter %q requires kind terms and a valid field", param.Name))
		}
		if (param.KnowledgeField != "") == (len(param.Values) > 0) {
			panic(fmt.Sprintf("tool parameter %q requires exactly one of KnowledgeField or Values", param.Name))
		}
		if param.KnowledgeField != "" && !visible[param.KnowledgeField] {
			panic(fmt.Sprintf("tool parameter %q requires a configured knowledge ModelField", param.Name))
		}
		seen := make(map[string]bool)
		for _, value := range param.Values {
			validateConstraintLiteral(value)
			if seen[value] {
				panic(fmt.Sprintf("tool parameter %q contains duplicate Values", param.Name))
			}
			seen[value] = true
		}
		params[param.Name] = param
	}
	for _, param := range conf.ToolParams {
		seen := map[string]bool{param.Name: true}
		for current := param; current.ParentParam != ""; {
			parent, ok := params[current.ParentParam]
			if !ok || seen[parent.Name] || current.KnowledgeField == "" || parent.KnowledgeField == "" {
				panic(fmt.Sprintf("tool parameter %q has an invalid ParentParam", param.Name))
			}
			seen[parent.Name] = true
			current = parent
		}
	}
}

func (r *Ha3ChatRecall) toolParams() []recconf.SearchToolParamConfig {
	if r.conf.SearchGoodsConf == nil {
		return nil
	}
	return r.conf.SearchGoodsConf.ToolParams
}

// ToolParamsConfigID prevents inheriting values under changed field mappings.
func (r *Ha3ChatRecall) ToolParamsConfigID() string {
	if len(r.toolParams()) == 0 {
		return ""
	}
	var knowledge *recconf.Ha3KnowledgeVectorConfig
	if r.knowledge != nil {
		knowledge = &r.knowledge.conf
	}
	data, _ := json.Marshal(struct {
		Search    recconf.Ha3ChatRecallConfig
		Knowledge *recconf.Ha3KnowledgeVectorConfig
	}{r.conf, knowledge})
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func toolTermValues(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, fmt.Errorf("expected a string array or null")
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\x00\r\n\t\\\"") {
			return nil, fmt.Errorf("invalid term value")
		}
		if !slices.Contains(result, value) {
			result = append(result, value)
		}
	}
	return result, nil
}

func (r *Ha3ChatRecall) buildToolParamsExpr(values map[string]json.RawMessage) (string, error) {
	params := make(map[string]recconf.SearchToolParamConfig)
	for _, param := range r.toolParams() {
		params[param.Name] = param
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	var clauses []string
	for _, name := range names {
		param, ok := params[name]
		if !ok {
			return "", fmt.Errorf("unknown tool parameter %q", name)
		}
		terms, err := toolTermValues(values[name])
		if err != nil {
			return "", fmt.Errorf("tool parameter %s: %w", name, err)
		}
		var choices []string
		for _, term := range terms {
			if param.KnowledgeField == "" && !slices.Contains(param.Values, term) {
				return "", fmt.Errorf("tool parameter %s: value %q is not in Values", name, term)
			}
			choices = append(choices, param.Field+" = "+strconv.Quote(term))
		}
		if len(choices) > 0 {
			clauses = append(clauses, "("+strings.Join(choices, " OR ")+")")
		}
	}
	return strings.Join(clauses, " AND "), nil
}

func (r *Ha3ChatRecall) ValidateToolParamSources(current aichat.SearchGoodsParams, previous *aichat.SearchGoodsParams, evidence ToolParamEvidence) error {
	values := make(map[string][]string)
	oldValues := make(map[string][]string)
	params := make(map[string]recconf.SearchToolParamConfig)
	for _, param := range r.toolParams() {
		params[param.Name] = param
		var err error
		values[param.Name], err = toolTermValues(current.ToolParams[param.Name])
		if err != nil {
			return fmt.Errorf("tool parameter %s: %w", param.Name, err)
		}
		if previous != nil {
			oldValues[param.Name], _ = toolTermValues(previous.ToolParams[param.Name])
		}
		if param.KnowledgeField == "" {
			continue // Static Values are checked by ValidateSearchGoodsRequest.
		}
		for _, value := range values[param.Name] {
			if slices.Contains(oldValues[param.Name], value) || evidence != nil && evidence.Contains(param.KnowledgeField, value) {
				continue
			}
			return fmt.Errorf("tool parameter %s: value %q is not in visible knowledge field %s; select a supported value or retain a valid prior value", param.Name, value, param.KnowledgeField)
		}
	}
	for _, param := range r.toolParams() {
		parent := params[param.ParentParam]
		parents := values[parent.Name]
		if param.ParentParam == "" || len(parents) == 0 {
			continue
		}
		for _, value := range values[param.Name] {
			related := false
			for _, parentValue := range parents {
				if evidence != nil && evidence.ContainsPair(param.KnowledgeField, value, parent.KnowledgeField, parentValue) {
					related = true
					break
				}
			}
			// A previously validated selection remains valid when none of its
			// parent choices has been removed. Subsets need fresh pair evidence.
			if !related && slices.Contains(oldValues[param.Name], value) && len(oldValues[parent.Name]) > 0 {
				related = true
				for _, oldParent := range oldValues[parent.Name] {
					if !slices.Contains(parents, oldParent) {
						related = false
						break
					}
				}
			}
			if !related {
				return fmt.Errorf("tool parameter %s: value %q must be associated with the selected %s in visible knowledge", param.Name, value, parent.Name)
			}
		}
	}
	return nil
}
