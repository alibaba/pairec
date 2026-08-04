package aichat

func SearchGoodsTool() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "search_goods",
			Description: "Search real products from the catalog. Use this for product recommendation requests, including first recommendations, changed conditions, category/color/material/style/occasion changes, or requests for more options.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keywords": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": `English catalog search keywords, 1-5 terms, e.g. ["Shirt","Blue","Long Sleeve"]. Translate product/category/color/material/style/occasion intent into English before calling.`,
					},
					"operator": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"AND", "OR"},
						"description": `Keyword boolean logic. For the first call with multiple keywords, pass "AND" for precise all-term matching. Only if that AND result returns total=0 may you call this tool again with "OR" to broaden recall. If AND returns any hits, even 1-3, do not call OR. OR is a fallback, not the default first call.`,
					},
					"exclude_keywords": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": `English keywords to exclude, e.g. ["Red"]. Set only when the user explicitly rejects an attribute, e.g. "不要红色", "no leather", or "不要短袖"; translate the rejected attribute to English; do not infer.`,
					},
					"min_price": map[string]interface{}{
						"type":        "number",
						"description": `Inclusive minimum price in CNY. Set only when the user explicitly states a lower bound, e.g. "100元以上" or "at least 100"; do not infer.`,
					},
					"max_price": map[string]interface{}{
						"type":        "number",
						"description": `Inclusive maximum price in CNY. Set only when the user explicitly states an upper bound, e.g. "200元以下" or "under 200"; do not infer.`,
					},
				},
				"required": []string{"keywords"},
			},
		},
	}
}

func FieldAwareSearchGoodsTool() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        "search_goods",
			Description: "Parse one product-shopping query and search real products from the catalog.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"keywords": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"minItems":    1,
						"maxItems":    1,
						"description": "The PAI-Rec-compatible positive search phrase. Preserve the full raw Query unless explicit price or exclusion syntax must be removed.",
					},
					"operator": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"AND"},
						"description": "Always AND, matching the current PAI-Rec Planner contract.",
					},
					"product_type_keywords": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"minItems":    1,
						"maxItems":    4,
						"description": "Exact product-type noun plus equivalent singular/plural/common catalog forms; never adjacent product types.",
					},
					"attribute_keywords": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"maxItems":    5,
						"description": "Positive color, material, style, occasion, gender, age, size, or feature descriptors; exclude product type, price, and negated values.",
					},
					"exclude_keywords": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"maxItems":    5,
						"description": "Only attributes explicitly rejected by the Query.",
					},
					"min_price": map[string]interface{}{
						"type":             "number",
						"exclusiveMinimum": 0,
						"description":      "Inclusive positive CNY minimum, emitted only when explicit.",
					},
					"max_price": map[string]interface{}{
						"type":             "number",
						"exclusiveMinimum": 0,
						"description":      "Inclusive positive CNY maximum, emitted only when explicit.",
					},
				},
				"required": []string{
					"keywords",
					"operator",
					"product_type_keywords",
					"attribute_keywords",
				},
				"additionalProperties": false,
			},
		},
	}
}
