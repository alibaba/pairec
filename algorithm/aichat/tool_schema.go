package aichat

import "github.com/alibaba/pairec/v2/recconf"

const (
	SearchGoodsMaxKeywords          = 8
	SearchGoodsMaxPreferredKeywords = 5
	SuggestionMinLength             = 2
	SuggestionMaxLength             = 80
)

func SearchGoodsTool() Tool {
	return FieldAwareSearchGoodsTool(nil)
}

func FieldAwareSearchGoodsTool(conf *recconf.SearchGoodsConfig) Tool {
	properties := map[string]interface{}{
		"keywords": map[string]interface{}{
			"type":     "array",
			"items":    map[string]interface{}{"type": "string"},
			"minItems": 1,
			"maxItems": SearchGoodsMaxKeywords,
		},
		"preferred_keywords": map[string]interface{}{
			"type":     "array",
			"items":    map[string]interface{}{"type": "string"},
			"maxItems": SearchGoodsMaxPreferredKeywords,
		},
		"exclude_keywords": map[string]interface{}{
			"type":     "array",
			"items":    map[string]interface{}{"type": "string"},
			"maxItems": 5,
		},
		"min_price": map[string]interface{}{
			"type":             "number",
			"exclusiveMinimum": 0,
		},
		"max_price": map[string]interface{}{
			"type":             "number",
			"exclusiveMinimum": 0,
		},
	}
	description := ""
	required := []string{"keywords"}
	if conf != nil {
		description = conf.ToolDescription
		for name, property := range properties {
			if text := conf.ParameterDescriptions[name]; text != "" {
				property.(map[string]interface{})["description"] = text
			}
		}
		for _, param := range conf.ToolParams {
			items := map[string]interface{}{"type": "string"}
			if param.KnowledgeField == "" {
				items["enum"] = param.Values
			}
			property := map[string]interface{}{
				"type": "array", "items": items, "minItems": 1, "uniqueItems": true,
			}
			if param.Description != "" {
				property["description"] = param.Description
			}
			properties[param.Name] = property
			if param.Required {
				required = append(required, param.Name)
			}
		}
	}
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "search_goods",
			Description: description,
			Parameters: map[string]interface{}{
				"type":                 "object",
				"properties":           properties,
				"required":             required,
				"additionalProperties": false,
			},
		},
	}
}

func SuggestionTool(count int, description string) Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "emit_suggestions",
			Description: description,
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
