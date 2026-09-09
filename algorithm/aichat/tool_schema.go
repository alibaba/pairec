package aichat

import (
	"sort"

	"github.com/alibaba/pairec/v2/recconf"
)

const (
	SuggestionMinLength = 2
	SuggestionMaxLength = 80
)

func SearchGoodsTool() Tool {
	return FieldAwareSearchGoodsTool(nil)
}

func FieldAwareSearchGoodsTool(conf *recconf.SearchGoodsConfig) Tool {
	properties := map[string]interface{}{
		"keywords": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"minItems":    1,
			"maxItems":    1,
			"description": "One required English catalog phrase: exact product noun and hard text conditions not represented in constraints. No singular/plural expansion, synonyms, prices, exclusions or optional preferences. Knowledge values are vocabulary references only.",
		},
		"preferred_keywords": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"maxItems":    5,
			"description": "Optional English preferences that may all be dropped after zero results, e.g. office for a general work outfit. Never put required attributes, explicit must/only conditions, exclusions or product type here.",
		},
		"exclude_keywords": map[string]interface{}{
			"type":        "array",
			"items":       map[string]interface{}{"type": "string"},
			"maxItems":    5,
			"description": "Explicitly rejected English text terms without a configured attribute. Use constraints for configured attributes; do not duplicate them here or infer exclusions.",
		},
		"min_price": map[string]interface{}{
			"type":             "number",
			"exclusiveMinimum": 0,
			"description":      "Inclusive positive minimum in catalog price units, only when explicit and supported; otherwise omit.",
		},
		"max_price": map[string]interface{}{
			"type":             "number",
			"exclusiveMinimum": 0,
			"description":      "Inclusive positive maximum in catalog price units, only when explicit and supported; otherwise omit.",
		},
	}
	constraints := map[string]interface{}{}
	if conf != nil {
		for _, param := range conf.ConstraintParams {
			values := make([]string, 0, len(param.Values)+len(param.EqualValues))
			for value := range param.Values {
				values = append(values, value)
			}
			for value := range param.EqualValues {
				values = append(values, value)
			}
			sort.Strings(values)
			property := map[string]interface{}{"description": param.Description}
			if param.Kind == "all_eq" {
				property["type"] = []string{"string", "null"}
				property["enum"] = appendNullable(values)
			} else {
				selection := func() map[string]interface{} {
					return map[string]interface{}{
						"type": []string{"array", "null"}, "items": map[string]interface{}{"type": "string", "enum": values}, "uniqueItems": true,
					}
				}
				property["type"] = []string{"object", "null"}
				property["additionalProperties"] = false
				property["properties"] = map[string]interface{}{
					"any": selection(), "exclude": selection(),
					"known": map[string]interface{}{"enum": []interface{}{true, nil}},
				}
			}
			constraints[param.Name] = property
		}
	}
	properties["constraints"] = map[string]interface{}{
		"type": []string{"object", "null"}, "properties": constraints, "additionalProperties": false,
		"description": "Hard attributes, never relaxed. any: one allowed value; exclude: reject those values among known options; known=true: require evidence without selecting a value. any/exclude imply known. Omit unspecified fields. For option sets, the same selectable option must satisfy any and exclude.",
	}
	required := []string{"keywords"}
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "search_goods",
			Description: "Parse one product-shopping query in any language and search real products from the English-language catalog.",
			Parameters: map[string]interface{}{
				"type":                 "object",
				"properties":           properties,
				"required":             required,
				"additionalProperties": false,
			},
		},
	}
}

func appendNullable(values []string) []interface{} {
	result := make([]interface{}, 0, len(values)+1)
	for _, value := range values {
		result = append(result, value)
	}
	return append(result, nil)
}

func SuggestionTool(count int) Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "emit_suggestions",
			Description: "Return the required number of executable product-search queries that change search conditions and recommend more products.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"suggestions": map[string]interface{}{
						"type":     "array",
						"minItems": count,
						"maxItems": count,
						"items": map[string]interface{}{
							"type":      "string",
							"minLength": SuggestionMinLength,
							"maxLength": SuggestionMaxLength,
						},
					},
				},
				"required":             []string{"suggestions"},
				"additionalProperties": false,
			},
		},
	}
}
