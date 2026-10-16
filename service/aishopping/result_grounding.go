package aishopping

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/recconf"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

// Grounding metadata describes a dialogue action and verifiable attribute
// constraints. It is removed before executing business search parameters.
type groundingRequest struct {
	CandidateCount   int                    `json:"candidate_count,omitempty"`
	Mode             string                 `json:"mode"`
	ClarifyAttribute string                 `json:"clarify_attribute"`
	Requirements     []attributeRequirement `json:"requirements"`
	OptionID         string                 `json:"option_id"`
}
type attributeRequirement struct {
	Attribute string `json:"attribute"`
	ValueID   string `json:"value_id"`
}
type verifiedAttributeOption struct {
	ID              string                   `json:"option_id"`
	Label           string                   `json:"label"`
	Requirements    []attributeRequirement   `json:"requirements"`
	EvidenceItemIDs []string                 `json:"evidence_item_ids"`
	Search          aichat.SearchGoodsParams `json:"search"`
}
type pendingClarification struct {
	Attribute  string                    `json:"attribute"`
	ConfigID   string                    `json:"config_id"`
	VerifiedAt int64                     `json:"verified_at"`
	Options    []verifiedAttributeOption `json:"options"`
}

func validateGroundingConfig(c *recconf.AIShoppingGroundingConfig) error {
	if c == nil {
		return nil
	}
	if len(c.Attributes) == 0 || c.MaxAgeSeconds <= 0 || c.Language != "zh" {
		return fmt.Errorf("GroundingConfig requires attributes, positive MaxAgeSeconds and Language zh")
	}
	if c.CandidateCountDefault < 0 || c.CandidateCountMax < 0 || c.CandidateCountDefault > c.CandidateCountMax || c.CandidateCountMax > 200 {
		return fmt.Errorf("grounding candidate counts must satisfy 0 <= default <= max <= 200")
	}
	names := map[string]bool{}
	validID := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	for _, a := range c.Attributes {
		if !validID.MatchString(a.Name) || names[a.Name] || strings.TrimSpace(a.Label) == "" || len(a.Fields) == 0 || len(a.Options) == 0 {
			return fmt.Errorf("invalid or duplicate grounding attribute %q", a.Name)
		}
		names[a.Name] = true
		if a.MatchMode != "" && a.MatchMode != "exact" && a.MatchMode != "prefix" {
			return fmt.Errorf("attribute %q has unsupported match mode", a.Name)
		}
		for _, field := range a.Fields {
			if strings.TrimSpace(field) == "" {
				return fmt.Errorf("attribute %q has an empty evidence field", a.Name)
			}
		}
		ids := map[string]bool{}
		for _, v := range a.Options {
			if !validID.MatchString(v.ID) || ids[v.ID] || strings.TrimSpace(v.Label) == "" || len(v.Values) == 0 || (len(v.Keywords) == 0 && len(v.ToolParams) == 0) {
				return fmt.Errorf("invalid or duplicate option %q for attribute %q", v.ID, a.Name)
			}
			ids[v.ID] = true
			for _, s := range append(append([]string{}, v.Values...), v.Keywords...) {
				if strings.TrimSpace(s) == "" || strings.EqualFold(s, "None") {
					return fmt.Errorf("invalid value for attribute %q", a.Name)
				}
			}
			for param, values := range v.ToolParams {
				if aichat.IsBuiltinSearchParam(param) || strings.TrimSpace(param) == "" || len(values) == 0 {
					return fmt.Errorf("invalid tool-parameter binding for attribute %q", a.Name)
				}
			}
		}
	}
	return nil
}
func groundingConfigID(c *recconf.AIShoppingGroundingConfig, recallID string) string {
	b, _ := json.Marshal(c)
	return fmt.Sprintf("%x", sha256.Sum256(append(b, []byte(recallID)...)))
}
func validPendingClarification(p *pendingClarification, c *recconf.AIShoppingGroundingConfig, recallID string, now time.Time) bool {
	return p != nil && c != nil && findGroundingAttribute(c, p.Attribute) != nil && len(p.Options) > 0 && p.ConfigID == groundingConfigID(c, recallID) && p.VerifiedAt <= now.Unix() && now.Unix()-p.VerifiedAt <= int64(c.MaxAgeSeconds)
}
func findGroundingAttribute(c *recconf.AIShoppingGroundingConfig, name string) *recconf.AIShoppingGroundingAttribute {
	for i := range c.Attributes {
		if c.Attributes[i].Name == name {
			return &c.Attributes[i]
		}
	}
	return nil
}
func findGroundingValue(a *recconf.AIShoppingGroundingAttribute, id string) *recconf.AIShoppingAttributeOption {
	if a == nil {
		return nil
	}
	for i := range a.Options {
		if a.Options[i].ID == id {
			return &a.Options[i]
		}
	}
	return nil
}
func hasRequirement(list []attributeRequirement, r attributeRequirement) bool {
	for _, x := range list {
		if x == r {
			return true
		}
	}
	return false
}

func groundedSearchTool(tool aichat.Tool, c *recconf.AIShoppingGroundingConfig) aichat.Tool {
	if c == nil {
		return tool
	}
	b, _ := json.Marshal(tool)
	var cloned aichat.Tool
	_ = json.Unmarshal(b, &cloned)
	tool = cloned
	attributes := []string{}
	valueIDs := []string{}
	guide := map[string]map[string]string{}
	seen := map[string]bool{}
	for _, a := range c.Attributes {
		attributes = append(attributes, a.Name)
		guide[a.Name] = map[string]string{}
		for _, v := range a.Options {
			guide[a.Name][v.ID] = v.Label
			if !seen[v.ID] {
				valueIDs = append(valueIDs, v.ID)
				seen[v.ID] = true
			}
		}
	}
	props := tool.Function.Parameters["properties"].(map[string]interface{})
	metaProps := map[string]interface{}{
		"mode":              map[string]interface{}{"type": "string", "enum": []string{"recommend", "clarify"}},
		"clarify_attribute": map[string]interface{}{"type": "string", "enum": append([]string{""}, attributes...)},
		"requirements": map[string]interface{}{"type": "array", "description": "Verifiable requested attribute values. Use an empty array if no configured value is requested. Supported attribute/value IDs: " + compactJSON(guide), "items": map[string]interface{}{
			"type": "object", "properties": map[string]interface{}{"attribute": map[string]interface{}{"type": "string", "enum": attributes}, "value_id": map[string]interface{}{"type": "string", "enum": valueIDs}}, "required": []string{"attribute", "value_id"}, "additionalProperties": false}},
		"option_id": map[string]interface{}{"type": "string"},
	}
	requiredMeta := []string{"mode", "clarify_attribute", "requirements", "option_id"}
	if c.CandidateCountMax > 0 {
		requiredMeta = append(requiredMeta, "candidate_count")
		metaProps["candidate_count"] = map[string]interface{}{"type": "integer", "minimum": 1, "maximum": c.CandidateCountMax, "description": fmt.Sprintf("Candidates to inspect, not items to display. For clarify use %d to discover evidence-backed values. For recommend normally use %d unless more candidates are needed. Always explicitly choose a positive count.", c.CandidateCountMax, c.CandidateCountDefault)}
	}
	props["grounding"] = map[string]interface{}{"type": "object", "description": "Use clarify only when a user decision is needed for one configured attribute, and name it in clarify_attribute. Search unchanged requirements without guessing the unresolved value. Missing optional attributes do not automatically require clarification. Use recommend when the current intent is executable; clarify_attribute must then be empty. For an unchanged pending_clarification choice, copy its option_id and complete search; for explicitly changed conditions leave option_id empty and update the query. Attribute constraints are verified by the backend using configured fields and bindings.", "properties": metaProps, "required": requiredMeta, "additionalProperties": false}
	required, _ := tool.Function.Parameters["required"].([]interface{})
	tool.Function.Parameters["required"] = append(required, "grounding")
	return tool
}

func parseGroundedSearch(arguments string, fieldAware bool, cfg *chatConfig, recallID string) (recallsvc.SearchGoodsRequest, groundingRequest, error) {
	if cfg.raw.GroundingConfig == nil {
		r, e := parseSearchGoodsRequest(arguments, fieldAware)
		return r, groundingRequest{}, e
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal([]byte(arguments), &fields); e != nil {
		return recallsvc.SearchGoodsRequest{}, groundingRequest{}, e
	}
	var g groundingRequest
	raw, ok := fields["grounding"]
	if !ok {
		return recallsvc.SearchGoodsRequest{}, g, fmt.Errorf("grounding metadata is required")
	}
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if e := dec.Decode(&g); e != nil {
		return recallsvc.SearchGoodsRequest{}, g, e
	}
	delete(fields, "grounding")
	b, _ := json.Marshal(fields)
	req, e := parseSearchGoodsRequest(string(b), fieldAware)
	if e != nil {
		return req, g, e
	}
	c := cfg.raw.GroundingConfig
	if c.CandidateCountMax > 0 && g.CandidateCount == 0 {
		return req, g, fmt.Errorf("candidate_count is required and must be positive")
	}
	if g.CandidateCount < 0 || (g.CandidateCount > 0 && c.CandidateCountMax == 0) {
		return req, g, fmt.Errorf("invalid candidate_count")
	}
	if g.CandidateCount > c.CandidateCountMax {
		g.CandidateCount = c.CandidateCountMax
	}
	if g.Mode != "recommend" && g.Mode != "clarify" {
		return req, g, fmt.Errorf("invalid grounding mode")
	}
	if g.Requirements == nil {
		return req, g, fmt.Errorf("requirements must be an array, empty when unspecified")
	}
	seen := map[attributeRequirement]bool{}
	for _, r := range g.Requirements {
		if findGroundingValue(findGroundingAttribute(c, r.Attribute), r.ValueID) == nil || seen[r] {
			return req, g, fmt.Errorf("invalid or duplicate attribute requirement")
		}
		seen[r] = true
	}
	if g.Mode == "clarify" {
		if findGroundingAttribute(c, g.ClarifyAttribute) == nil || g.OptionID != "" {
			return req, g, fmt.Errorf("clarify requires one configured clarify_attribute and no selected option")
		}
		for _, r := range g.Requirements {
			if r.Attribute == g.ClarifyAttribute {
				return req, g, fmt.Errorf("the attribute being clarified cannot already be selected")
			}
		}
	} else if g.ClarifyAttribute != "" {
		return req, g, fmt.Errorf("recommend must not set clarify_attribute")
	}
	if g.OptionID != "" {
		if !validPendingClarification(cfg.pendingClarification, c, recallID, time.Now()) {
			return req, g, fmt.Errorf("pending clarification expired; search updated conditions without option_id")
		}
		var selected *verifiedAttributeOption
		for i := range cfg.pendingClarification.Options {
			if cfg.pendingClarification.Options[i].ID == g.OptionID {
				selected = &cfg.pendingClarification.Options[i]
				break
			}
		}
		if selected == nil {
			return req, g, fmt.Errorf("unknown pending option")
		}
		for _, r := range g.Requirements {
			if !hasRequirement(selected.Requirements, r) {
				return req, g, fmt.Errorf("selected option conflicts with requirements; omit option_id for changed conditions")
			}
		}
		g.Requirements = append([]attributeRequirement{}, selected.Requirements...)
		if e = bindGroundingRequirements(&req, g.Requirements, c); e != nil {
			return req, g, e
		}
		if canonicalSearch(req.SearchGoodsParams) != canonicalSearch(selected.Search) {
			return req, g, fmt.Errorf("pending option conditions changed; omit option_id to search updated conditions")
		}
		req.SearchGoodsParams = cloneGroundingSearch(selected.Search)
	}
	if e = bindGroundingRequirements(&req, g.Requirements, c); e != nil {
		return req, g, e
	}
	for _, k := range req.Keywords {
		for _, a := range c.Attributes {
			for _, bad := range a.UnsupportedPhrases {
				if phrasePresent(k, bad) {
					return req, g, fmt.Errorf("unresolved %s must use clarify without the unsupported literal %q", a.Name, bad)
				}
			}
			for _, v := range a.Options {
				for _, word := range v.Keywords {
					if phrasePresent(k, word) {
						if g.Mode == "clarify" && a.Name == g.ClarifyAttribute {
							return req, g, fmt.Errorf("clarify must not retain a concrete value of %s", a.Name)
						}
						typed := false
						for _, r := range g.Requirements {
							if r.Attribute == a.Name {
								typed = true
							}
						}
						if typed && !hasRequirement(g.Requirements, attributeRequirement{Attribute: a.Name, ValueID: v.ID}) {
							return req, g, fmt.Errorf("keyword value conflicts with requirement for %s", a.Name)
						}
					}
				}
			}
		}
	}
	return req, g, nil
}

func bindGroundingRequirements(req *recallsvc.SearchGoodsRequest, requirements []attributeRequirement, c *recconf.AIShoppingGroundingConfig) error {
	for _, r := range requirements {
		a := findGroundingAttribute(c, r.Attribute)
		v := findGroundingValue(a, r.ValueID)
		if v == nil {
			return fmt.Errorf("unknown attribute/value pair")
		}
		for _, word := range v.Keywords {
			found := false
			for _, k := range req.Keywords {
				if phrasePresent(k, word) {
					found = true
					break
				}
			}
			if !found {
				req.Keywords = append(req.Keywords, word)
			}
		}
		if req.ToolParams == nil {
			req.ToolParams = map[string]json.RawMessage{}
		}
		for param, values := range v.ToolParams {
			if old, ok := req.ToolParams[param]; ok {
				var existing []string
				if json.Unmarshal(old, &existing) != nil || !sameTerms(existing, values) {
					return fmt.Errorf("attribute %s conflicts with business parameter %s", r.Attribute, param)
				}
			} else {
				req.ToolParams[param], _ = json.Marshal(values)
			}
		}
	}
	return nil
}
func sameTerms(a, b []string) bool {
	a = append([]string{}, a...)
	b = append([]string{}, b...)
	sort.Strings(a)
	sort.Strings(b)
	return compactJSON(a) == compactJSON(b)
}
func marshalGroundedSearch(req recallsvc.SearchGoodsRequest, g groundingRequest, enabled bool) ([]byte, error) {
	b, e := marshalSearchGoodsRequest(req)
	if e != nil || !enabled {
		return b, e
	}
	var fields map[string]json.RawMessage
	if e = json.Unmarshal(b, &fields); e != nil {
		return nil, e
	}
	if g.Requirements == nil {
		g.Requirements = []attributeRequirement{}
	}
	fields["grounding"], e = json.Marshal(g)
	if e != nil {
		return nil, e
	}
	return json.Marshal(fields)
}
func phrasePresent(text, phrase string) bool {
	if strings.TrimSpace(phrase) == "" {
		return false
	}
	return regexp.MustCompile(`(?i)(^|[^a-z0-9])` + regexp.QuoteMeta(phrase) + `($|[^a-z0-9])`).MatchString(text)
}
func propertyTokens(value interface{}) []string {
	var out []string
	switch v := value.(type) {
	case string:
		for _, s := range strings.Split(v, "|") {
			s = strings.TrimSpace(s)
			if s != "" && !strings.EqualFold(s, "none") && !strings.EqualFold(s, "null") {
				out = append(out, s)
			}
		}
	case []string:
		for _, s := range v {
			out = append(out, propertyTokens(s)...)
		}
	case []interface{}:
		for _, s := range v {
			out = append(out, propertyTokens(s)...)
		}
	}
	return out
}
func anyExact(actual, allowed []string) bool {
	for _, a := range actual {
		for _, b := range allowed {
			if strings.EqualFold(a, b) {
				return true
			}
		}
	}
	return false
}
func matchesGroundingValue(hit recallsvc.GoodsHit, a *recconf.AIShoppingGroundingAttribute, v *recconf.AIShoppingAttributeOption) bool {
	for _, field := range a.Fields {
		tokens := propertyTokens(hit.Properties[field])
		if a.MatchMode == "prefix" {
			for _, t := range tokens {
				for _, prefix := range v.Values {
					if strings.HasPrefix(strings.ToLower(t), strings.ToLower(prefix)) {
						return true
					}
				}
			}
		} else if anyExact(tokens, v.Values) {
			return true
		}
	}
	return false
}
func matchesGroundingBase(hit recallsvc.GoodsHit, req recallsvc.SearchGoodsRequest, c *recconf.AIShoppingGroundingConfig) bool {
	for param, field := range c.ParamFields {
		raw, ok := req.ToolParams[param]
		if !ok {
			continue
		}
		var wanted []string
		if json.Unmarshal(raw, &wanted) != nil || len(wanted) == 0 {
			return false
		}
		var allowed []string
		for _, v := range wanted {
			if mapped := c.ParamValueMapping[param][v]; len(mapped) > 0 {
				allowed = append(allowed, mapped...)
			} else {
				allowed = append(allowed, v)
			}
		}
		if !anyExact(propertyTokens(hit.Properties[field]), allowed) {
			return false
		}
	}
	// The same configured attribute rules also verify values carried by keywords.
	for _, a := range c.Attributes {
		for _, v := range a.Options {
			requested := false
			for _, word := range v.Keywords {
				for _, k := range req.Keywords {
					if phrasePresent(k, word) {
						requested = true
						break
					}
				}
			}
			if requested && !matchesGroundingValue(hit, &a, &v) {
				return false
			}
		}
	}
	return true
}
func matchesGroundingRequirements(hit recallsvc.GoodsHit, requirements []attributeRequirement, c *recconf.AIShoppingGroundingConfig) bool {
	for _, r := range requirements {
		a := findGroundingAttribute(c, r.Attribute)
		v := findGroundingValue(a, r.ValueID)
		if v == nil || !matchesGroundingValue(hit, a, v) {
			return false
		}
	}
	return true
}
func cloneGroundingSearch(s aichat.SearchGoodsParams) aichat.SearchGoodsParams {
	b, _ := json.Marshal(s)
	var out aichat.SearchGoodsParams
	_ = json.Unmarshal(b, &out)
	return out
}
func canonicalSearch(s aichat.SearchGoodsParams) string {
	s = cloneGroundingSearch(s)
	sort.Strings(s.Keywords)
	sort.Strings(s.PreferredKeywords)
	sort.Strings(s.ExcludeKeywords)
	for k, v := range s.ToolParams {
		var values []string
		if json.Unmarshal(v, &values) == nil {
			sort.Strings(values)
			s.ToolParams[k], _ = json.Marshal(values)
		}
	}
	return compactJSON(s)
}

// Only returned, field-verified rows can become options or citations. All
// configured enumerated attributes share this execution path.
func groundSearchResult(result *recallsvc.SearchGoodsResult, req recallsvc.SearchGoodsRequest, g groundingRequest, c *recconf.AIShoppingGroundingConfig, recallID string) (*recallsvc.SearchGoodsResult, *pendingClarification, string) {
	out := *result
	out.Hits = nil
	for _, h := range result.Hits {
		if matchesGroundingBase(h, req, c) && matchesGroundingRequirements(h, g.Requirements, c) {
			out.Hits = append(out.Hits, h)
		}
	}
	if g.Mode != "clarify" {
		if len(g.Requirements) > 0 && len(out.Hits) == 0 {
			return &out, nil, "本次返回的商品中，没有找到能确认同时满足这些属性要求的商品。是否调整其中一个条件？我会保留其他要求。"
		}
		return &out, nil, ""
	}
	a := findGroundingAttribute(c, g.ClarifyAttribute)
	if a == nil {
		return &out, nil, "暂时无法验证这个属性，请告诉我更具体的需求。"
	}
	pending := &pendingClarification{Attribute: a.Name, ConfigID: groundingConfigID(c, recallID), VerifiedAt: time.Now().Unix()}
	for _, v := range a.Options {
		requirements := append(append([]attributeRequirement{}, g.Requirements...), attributeRequirement{Attribute: a.Name, ValueID: v.ID})
		option := verifiedAttributeOption{ID: a.Name + "_" + v.ID, Label: v.Label, Requirements: requirements, Search: cloneGroundingSearch(req.SearchGoodsParams)}
		bound := recallsvc.SearchGoodsRequest{SearchGoodsParams: option.Search}
		if bindGroundingRequirements(&bound, requirements, c) != nil {
			continue
		}
		option.Search = bound.SearchGoodsParams
		for _, h := range out.Hits {
			if matchesGroundingValue(h, a, &v) {
				option.EvidenceItemIDs = append(option.EvidenceItemIDs, h.ItemId)
			}
		}
		if len(option.EvidenceItemIDs) == 0 {
			continue
		}
		if len(option.EvidenceItemIDs) > 3 {
			option.EvidenceItemIDs = option.EvidenceItemIDs[:3]
		}
		pending.Options = append(pending.Options, option)
		if len(pending.Options) == 3 {
			break
		}
	}
	if len(pending.Options) == 0 {
		return &out, nil, "本次返回的商品中，暂时没有能确认的“" + a.Label + "”选项。你对" + a.Label + "有什么具体要求？我会保留其他条件继续查。"
	}
	labels := []string{}
	for _, o := range pending.Options {
		labels = append(labels, o.Label)
	}
	if len(labels) == 1 {
		return &out, pending, "按当前条件，本次查到的商品中已确认有“" + labels[0] + "”。你想看这种，还是告诉我其他具体的" + a.Label + "要求再查？"
	}
	return &out, pending, "按当前条件，本次查到的商品中已确认有这些" + a.Label + "选项：" + strings.Join(labels, "、") + "。你更喜欢哪一种？也可以告诉我其他具体要求再查。"
}
func groundingCandidateLimit(g groundingRequest, c *recconf.AIShoppingGroundingConfig, current int) int {
	if c == nil || c.CandidateCountMax <= 0 {
		return current
	}
	n := g.CandidateCount
	if n <= 0 {
		n = c.CandidateCountDefault
	}
	if n <= 0 {
		n = current
	}
	if n > c.CandidateCountMax {
		n = c.CandidateCountMax
	}
	return n
}
