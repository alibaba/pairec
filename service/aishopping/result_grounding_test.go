package aishopping

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/recconf"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/service/searchsuggestion"
)

func testGroundingConfig() *recconf.AIShoppingGroundingConfig {
	return &recconf.AIShoppingGroundingConfig{
		Language: "zh", MaxAgeSeconds: 900,
		Attributes: []recconf.AIShoppingGroundingAttribute{
			{Name: "pattern", Label: "图案", Fields: []string{"content_all_over_print", "content_placement_print"}, UnsupportedPhrases: []string{"cute pattern", "cute print"}, Options: []recconf.AIShoppingAttributeOption{
				{ID: "striped", Label: "条纹图案", Keywords: []string{"striped"}, Values: []string{"Striped"}},
				{ID: "floral", Label: "花卉图案", Keywords: []string{"floral"}, Values: []string{"Small Floral", "Big Floral", "Watercolour Floral"}},
				{ID: "bow", Label: "蝴蝶结印花", Keywords: []string{"bow print"}, Values: []string{"Bow"}},
				{ID: "heart", Label: "爱心印花", Keywords: []string{"heart print"}, Values: []string{"Heart", "Hearts"}},
				{ID: "cartoon", Label: "卡通动物印花", Keywords: []string{"cartoon animal"}, Values: []string{"Cartoon", "Animal"}},
			}},
			{Name: "color", Label: "颜色", Fields: []string{"sellable_colors"}, MatchMode: "prefix", Options: []recconf.AIShoppingAttributeOption{
				{ID: "pink", Label: "粉色", Keywords: []string{"pink"}, Values: []string{"Pink_"}},
				{ID: "blue", Label: "蓝色", Keywords: []string{"blue"}, Values: []string{"Blue_"}},
			}},
		},
		ParamFields:       map[string]string{"category": "category", "categories": "categories", "sleeve_length": "sleeve_length"},
		ParamValueMapping: map[string]map[string][]string{"sleeve_length": {"Long Sleeve": {"Long Sleeve", "Wrist-Length Sleeve", "Extra-Long Sleeve"}}},
	}
}
func testAttributeRequirements(attribute, value string) []attributeRequirement {
	return []attributeRequirement{{Attribute: attribute, ValueID: value}}
}
func testGroundingRequest() recallsvc.SearchGoodsRequest {
	var p aichat.SearchGoodsParams
	_ = json.Unmarshal([]byte(`{"keywords":["pink"],"category":["Blouses"],"categories":["Women/Clothing/Blouses/Placket Blouses"],"sleeve_length":["Long Sleeve"],"max_price":1000,"exclude_keywords":["lace"]}`), &p)
	return recallsvc.SearchGoodsRequest{SearchGoodsParams: p}
}
func testGroundingHit(id string) recallsvc.GoodsHit {
	return recallsvc.GoodsHit{ItemId: id, Properties: map[string]interface{}{"category": "Blouses", "categories": "Women/Clothing/Blouses/Placket Blouses", "sellable_colors": "Pink_Light Pink", "sleeve_length": "Long Sleeve", "content_all_over_print": "Striped", "content_placement_print": "None"}}
}

type groundingRecallFixture struct{ result *recallsvc.SearchGoodsResult }

func (r groundingRecallFixture) Search(context.Context, recallsvc.SearchGoodsRequest) (*recallsvc.SearchGoodsResult, error) {
	return r.result, nil
}
func (r groundingRecallFixture) SearchGoodsTool() aichat.Tool { return aichat.SearchGoodsTool() }
func (r groundingRecallFixture) ValidateSearchGoodsRequest(recallsvc.SearchGoodsRequest) error {
	return nil
}
func (r groundingRecallFixture) ValidateToolParamSources(aichat.SearchGoodsParams, *aichat.SearchGoodsParams, recallsvc.ToolParamEvidence) error {
	return nil
}
func (r groundingRecallFixture) ToolParamsConfigID() string { return "fixture" }
func groundingState() *turnState {
	return &turnState{indexMap: map[int]string{}, itemToIndex: map[string]int{}, nextIndex: 1}
}

func TestGroundingRejectsDecorationNecklineMissingAndWrongBase(t *testing.T) {
	c := testGroundingConfig()
	req := testGroundingRequest()
	for _, name := range []string{"decoration", "neckline", "missing", "wrong_color", "wrong_category", "wrong_sleeve", "token_substring"} {
		t.Run(name, func(t *testing.T) {
			h := testGroundingHit("candidate")
			pattern := "bow"
			switch name {
			case "decoration":
				h.Properties["content_all_over_print"] = "None"
				h.Properties["content_design_detail"] = "Bow | Tie Up"
			case "neckline":
				h.Properties["content_all_over_print"] = ""
				h.Properties["content_neckline"] = "Sweetheart Neck"
				pattern = "heart"
			case "missing":
				delete(h.Properties, "content_all_over_print")
			case "wrong_color":
				h.Properties["sellable_colors"] = "White_White"
				pattern = "striped"
			case "wrong_category":
				h.Properties["category"] = "Dresses"
				pattern = "striped"
			case "wrong_sleeve":
				h.Properties["sleeve_length"] = "Short Sleeve"
				pattern = "striped"
			case "token_substring":
				h.Properties["content_all_over_print"] = "Not Striped"
				pattern = "striped"
			}
			result, _, reply := groundSearchResult(&recallsvc.SearchGoodsResult{Total: 99, Hits: []recallsvc.GoodsHit{h}}, req, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", pattern)}, c, "fixture")
			if len(result.Hits) != 0 || reply == "" {
				t.Fatal("unsupported item reached recommendation")
			}
		})
	}
}
func TestGroundingChoicesBindFullSearchAndInvalidateChanges(t *testing.T) {
	c := testGroundingConfig()
	req := testGroundingRequest()
	h := testGroundingHit("stripe")
	_, pending, reply := groundSearchResult(&recallsvc.SearchGoodsResult{Hits: []recallsvc.GoodsHit{h}}, req, groundingRequest{Mode: "clarify", ClarifyAttribute: "pattern"}, c, "fixture")
	if pending == nil || len(pending.Options) != 1 || !strings.Contains(reply, "条纹图案") || strings.Contains(reply, "花卉") {
		t.Fatal("unverified choice")
	}
	bound := pending.Options[0].Search
	if bound.MaxPrice == nil || *bound.MaxPrice != 1000 || len(bound.ExcludeKeywords) != 1 || len(bound.ToolParams) != 3 {
		t.Fatal("lost business constraints")
	}
	cfg := &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: c}, pendingClarification: pending}
	g := groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", "striped"), OptionID: "pattern_striped"}
	payload, _ := marshalGroundedSearch(recallsvc.SearchGoodsRequest{SearchGoodsParams: bound}, g, true)
	parsed, _, err := parseGroundedSearch(string(payload), true, cfg, "fixture")
	if err != nil || canonicalSearch(parsed.SearchGoodsParams) != canonicalSearch(bound) {
		t.Fatalf("selection lost: %v", err)
	}
	changed := cloneGroundingSearch(bound)
	changed.Keywords = []string{"blue", "striped"}
	payload, _ = marshalGroundedSearch(recallsvc.SearchGoodsRequest{SearchGoodsParams: changed}, g, true)
	if _, _, err = parseGroundedSearch(string(payload), true, cfg, "fixture"); err == nil {
		t.Fatal("stale pink option accepted blue conditions")
	}
	g.OptionID = ""
	payload, _ = marshalGroundedSearch(recallsvc.SearchGoodsRequest{SearchGoodsParams: changed}, g, true)
	if _, _, err = parseGroundedSearch(string(payload), true, cfg, "fixture"); err != nil {
		t.Fatal("explicit changed conditions rejected", err)
	}
	if validPendingClarification(pending, c, "different-config", time.Now()) || validPendingClarification(pending, c, "fixture", time.Now().Add(901*time.Second)) {
		t.Fatal("stale proof retained")
	}
	blob := &SessionBlob{PendingClarification: pending}
	data, _ := json.Marshal(blob)
	var loaded SessionBlob
	_ = json.Unmarshal(data, &loaded)
	if !strings.Contains(loaded.messages("条纹的")[1].Content, "pending_clarification") {
		t.Fatal("pending choice missing from model context")
	}
}
func TestGroundingForcesSearchAndKeepsMetadataOutOfFilters(t *testing.T) {
	cfg := &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: testGroundingConfig()}, fieldAware: true}
	recall := groundingRecallFixture{}
	direct := &aichat.StreamResult{FinishReason: "stop", Content: "爱心还是蝴蝶结？"}
	if normalizePlannerResponse(direct, cfg, recall, nil, nil) == nil {
		t.Fatal("unverified direct question accepted")
	}
	original := aichat.SearchGoodsTool()
	tool := groundedSearchTool(original, cfg.raw.GroundingConfig)
	if _, ok := original.Function.Parameters["properties"].(map[string]interface{})["grounding"]; ok {
		t.Fatal("cached schema mutated")
	}
	if _, ok := tool.Function.Parameters["properties"].(map[string]interface{})["grounding"]; !ok {
		t.Fatal("missing schema")
	}
	b, _ := marshalGroundedSearch(testGroundingRequest(), groundingRequest{Mode: "clarify", ClarifyAttribute: "pattern"}, true)
	r, _, err := parseGroundedSearch(string(b), true, cfg, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := r.ToolParams["grounding"]; ok {
		t.Fatal("dialogue metadata leaked to search filters")
	}
	cfg.raw.GroundingConfig = nil
	if normalizePlannerResponse(direct, cfg, recall, nil, nil) != nil {
		t.Fatal("disabled path changed")
	}
}
func TestGroundingFiltersBeforeCitationsAndSeparatesCounts(t *testing.T) {
	c := testGroundingConfig()
	req := testGroundingRequest()
	req.Keywords = append(req.Keywords, "striped")
	good, bad := testGroundingHit("good"), testGroundingHit("bad")
	bad.Properties["content_all_over_print"] = "None"
	state := groundingState()
	r := groundingRecallFixture{&recallsvc.SearchGoodsResult{Total: 32, Hits: []recallsvc.GoodsHit{good, bad}}}
	b, _ := marshalGroundedSearch(req, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", "striped")}, true)
	out := dispatchTool(context.Background(), r, aichat.ToolCall{Function: aichat.FunctionCall{Name: "search_goods", Arguments: string(b)}}, state, &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: c, DisplayItemCountMax: 12}}, nil, true, true, timingMeta{}, 1)
	if out.search == nil {
		t.Fatal("missing suggestion snapshot")
	}
	if _, err := searchsuggestion.BuildProductSummary(out.search.ReplyToolPayload, out.search.ProductIndexes, out.search.Total); err != nil {
		t.Fatal("suggestion snapshot contract broken", err)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(out.content), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["raw_total"] != float64(32) || payload["verified_returned_count"] != float64(1) || payload["total"] != float64(1) || len(state.indexMap) != 1 || state.indexMap[1] != "good" {
		t.Fatal("wrong count or citation set", out.content, state.indexMap)
	}
}
func TestGroundingRealReleaseFixtures(t *testing.T) {
	base := os.Getenv("GROUNDING_FIXTURE_DIR")
	if base == "" {
		t.Skip("set GROUNDING_FIXTURE_DIR for saved release-result replay")
	}
	load := func(name string) *recallsvc.SearchGoodsResult {
		data, err := os.ReadFile(filepath.Join(base, name))
		if err != nil {
			t.Fatal(err)
		}
		var raw struct {
			Total int                      `json:"total_hits"`
			Hits  []map[string]interface{} `json:"hits"`
		}
		if err = json.Unmarshal(data, &raw); err != nil {
			t.Fatal(err)
		}
		out := &recallsvc.SearchGoodsResult{Total: raw.Total}
		for _, row := range raw.Hits {
			out.Hits = append(out.Hits, recallsvc.GoodsHit{ItemId: row["item_id"].(string), Properties: row})
		}
		return out
	}
	c := testGroundingConfig()
	req := testGroundingRequest()
	for _, tt := range []struct {
		name, pattern string
		want          int
	}{{"pink_long_striped_verified_query.json", "striped", 9}, {"pink_long_floral_verified_query.json", "floral", 10}, {"pink_long_heart.json", "heart", 0}, {"pink_long_bow_print_terms.json", "bow", 0}} {
		t.Run(tt.pattern, func(t *testing.T) {
			out, _, _ := groundSearchResult(load(tt.name), req, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", tt.pattern)}, c, "fixture")
			if len(out.Hits) != tt.want {
				t.Fatalf("got %d want %d", len(out.Hits), tt.want)
			}
		})
	}
	initial := load("pink_long_pattern_candidates.json")
	initial.Hits = initial.Hits[:12]
	_, pending, _ := groundSearchResult(initial, req, groundingRequest{Mode: "clarify", ClarifyAttribute: "pattern"}, c, "fixture")
	if pending == nil || len(pending.Options) != 1 || !hasRequirement(pending.Options[0].Requirements, attributeRequirement{Attribute: "pattern", ValueID: "striped"}) {
		t.Fatal("claimed unseen pattern in first 12")
	}
	data, err := os.ReadFile(filepath.Join(base, "E2E_FINDINGS.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Case  string                   `json:"case"`
		Items []map[string]interface{} `json:"returned_item_field_evidence"`
	}
	if err = json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, item := range cases {
		if item.Case != "bow" {
			continue
		}
		res := &recallsvc.SearchGoodsResult{Total: len(item.Items)}
		for _, row := range item.Items {
			res.Hits = append(res.Hits, recallsvc.GoodsHit{ItemId: row["item_id"].(string), Properties: row})
		}
		out, _, reply := groundSearchResult(res, req, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", "bow")}, c, "fixture")
		if len(res.Hits) != 10 || len(out.Hits) != 0 || reply == "" {
			t.Fatal("live ten-item false-positive not rejected")
		}
	}
}

func TestGroundingAddsRequiredPatternBeforeRetrieval(t *testing.T) {
	cfg := &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: testGroundingConfig()}}
	req := testGroundingRequest()
	args, _ := marshalGroundedSearch(req, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", "bow")}, true)
	parsed, _, err := parseGroundedSearch(string(args), true, cfg, "fixture")
	if err != nil || !strings.Contains(strings.Join(parsed.Keywords, "|"), "bow print") {
		t.Fatal("explicit pattern missing from retrieval", err, parsed.Keywords)
	}
	args, _ = marshalGroundedSearch(parsed, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", "heart")}, true)
	if _, _, err = parseGroundedSearch(string(args), true, cfg, "fixture"); err == nil {
		t.Fatal("mismatched print ID and keywords accepted")
	}
}

type countingGroundingRecall struct {
	groundingRecallFixture
	limit int
}

func (r *countingGroundingRecall) Search(_ context.Context, req recallsvc.SearchGoodsRequest) (*recallsvc.SearchGoodsResult, error) {
	r.limit = req.Limit
	return r.result, nil
}

func TestGroundingCandidateBudgetAndDisplayLimit(t *testing.T) {
	c := testGroundingConfig()
	c.CandidateCountDefault = 24
	c.CandidateCountMax = 96
	cfg := &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: c, DisplayItemCountMax: 12}}
	req := testGroundingRequest()
	req.Keywords = append(req.Keywords, "striped")
	missing, _ := marshalGroundedSearch(req, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", "striped")}, true)
	if _, _, err := parseGroundedSearch(string(missing), true, cfg, "fixture"); err == nil {
		t.Fatal("model silently omitted enabled candidate budget")
	}
	args, _ := marshalGroundedSearch(req, groundingRequest{Mode: "recommend", Requirements: testAttributeRequirements("pattern", "striped"), CandidateCount: 9999}, true)
	parsed, g, err := parseGroundedSearch(string(args), true, cfg, "fixture")
	if err != nil || g.CandidateCount != 96 {
		t.Fatal("candidate budget not capped", err, g)
	}
	if _, ok := parsed.ToolParams["candidate_count"]; ok {
		t.Fatal("operational count leaked into business parameters")
	}
	if n := groundingCandidateLimit(groundingRequest{}, c, 12); n != 24 {
		t.Fatal("wrong default", n)
	}
	if n := groundingCandidateLimit(groundingRequest{}, nil, 12); n != 12 {
		t.Fatal("legacy limit changed", n)
	}
	hits := make([]recallsvc.GoodsHit, 96)
	for i := range hits {
		hits[i] = testGroundingHit(fmt.Sprintf("item_%d", i))
	}
	r := &countingGroundingRecall{groundingRecallFixture: groundingRecallFixture{result: &recallsvc.SearchGoodsResult{Total: 364, Hits: hits}}}
	state := groundingState()
	out := dispatchTool(context.Background(), r, aichat.ToolCall{Function: aichat.FunctionCall{Name: "search_goods", Arguments: string(args)}}, state, cfg, nil, true, true, timingMeta{}, 1)
	if r.limit != 96 {
		t.Fatal("retrieval did not use model's bounded limit", r.limit)
	}
	var payload struct {
		Hits     []interface{} `json:"hits"`
		Total    int           `json:"total"`
		RawTotal int           `json:"raw_total"`
	}
	if err = json.Unmarshal([]byte(out.content), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Hits) != 12 || len(state.indexMap) != 12 || payload.Total != 12 || payload.RawTotal != 364 {
		t.Fatal("large candidate set leaked into reply or cards", len(payload.Hits), len(state.indexMap), payload)
	}
	if _, err := searchsuggestion.BuildProductSummary(out.search.ReplyToolPayload, out.search.ProductIndexes, out.search.Total); err != nil {
		t.Fatal("large-candidate suggestion snapshot broken", err)
	}
}

func TestGroundingLargerRealSampleFindsMorePatterns(t *testing.T) {
	base := os.Getenv("GROUNDING_FIXTURE_DIR")
	if base == "" {
		t.Skip("saved live catalog response required")
	}
	b, err := os.ReadFile(filepath.Join(base, "pink_long_complete_pattern_snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Hits []map[string]interface{} `json:"hits"`
	}
	if err = json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ n, want int }{{12, 1}, {24, 1}, {48, 1}, {96, 2}} {
		res := &recallsvc.SearchGoodsResult{Total: 364}
		for _, row := range raw.Hits[:tt.n] {
			res.Hits = append(res.Hits, recallsvc.GoodsHit{ItemId: row["item_id"].(string), Properties: row})
		}
		_, pending, _ := groundSearchResult(res, testGroundingRequest(), groundingRequest{Mode: "clarify", ClarifyAttribute: "pattern"}, testGroundingConfig(), "fixture")
		if pending == nil || len(pending.Options) != tt.want {
			t.Fatalf("%d rows: expected %d verified pattern choices, got %+v", tt.n, tt.want, pending)
		}
	}
}

func testSleeveAttribute() recconf.AIShoppingGroundingAttribute {
	return recconf.AIShoppingGroundingAttribute{Name: "sleeve_length", Label: "袖长", Fields: []string{"sleeve_length"}, Options: []recconf.AIShoppingAttributeOption{
		{ID: "long", Label: "长袖", Values: []string{"Long Sleeve", "Wrist-Length Sleeve", "Extra-Long Sleeve"}, ToolParams: map[string][]string{"sleeve_length": {"Long Sleeve"}}},
		{ID: "short", Label: "短袖", Values: []string{"Short Sleeve"}, ToolParams: map[string][]string{"sleeve_length": {"Short Sleeve"}}},
	}}
}
func TestGroundingGenericSleeveClarificationAndSelection(t *testing.T) {
	c := testGroundingConfig()
	c.Attributes = append(c.Attributes, testSleeveAttribute())
	if err := validateGroundingConfig(c); err != nil {
		t.Fatal(err)
	}
	req := testGroundingRequest()
	delete(req.ToolParams, "sleeve_length")
	long, short, unknown := testGroundingHit("long"), testGroundingHit("short"), testGroundingHit("unknown")
	short.Properties["sleeve_length"] = "Short Sleeve"
	delete(unknown.Properties, "sleeve_length")
	g := groundingRequest{Mode: "clarify", ClarifyAttribute: "sleeve_length", Requirements: []attributeRequirement{}}
	_, pending, reply := groundSearchResult(&recallsvc.SearchGoodsResult{Hits: []recallsvc.GoodsHit{long, short, unknown}}, req, g, c, "fixture")
	if pending == nil || pending.Attribute != "sleeve_length" || len(pending.Options) != 2 || !strings.Contains(reply, "袖长") || strings.Contains(reply, "图案") {
		t.Fatal("attribute-specific branch still leaked", pending, reply)
	}
	chosen := pending.Options[1]
	cfg := &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: c}, pendingClarification: pending}
	args, _ := marshalGroundedSearch(recallsvc.SearchGoodsRequest{SearchGoodsParams: chosen.Search}, groundingRequest{Mode: "recommend", OptionID: chosen.ID}, true)
	selected, meta, err := parseGroundedSearch(string(args), true, cfg, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	if string(selected.ToolParams["sleeve_length"]) != `["Short Sleeve"]` || !hasRequirement(meta.Requirements, attributeRequirement{Attribute: "sleeve_length", ValueID: "short"}) {
		t.Fatal("option did not bind structured sleeve parameter", selected, meta)
	}
	result, _, _ := groundSearchResult(&recallsvc.SearchGoodsResult{Hits: []recallsvc.GoodsHit{long, short, unknown}}, selected, meta, c, "fixture")
	if len(result.Hits) != 1 || result.Hits[0].ItemId != "short" {
		t.Fatal("selected attribute not verified", result)
	}
	if selected.MaxPrice == nil || *selected.MaxPrice != 1000 || len(selected.ExcludeKeywords) != 1 {
		t.Fatal("other constraints lost")
	}
}
func TestGroundingGenericAttributeValidationAndMultipleRequirements(t *testing.T) {
	c := testGroundingConfig()
	c.Attributes = append(c.Attributes, testSleeveAttribute())
	cfg := &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: c}}
	req := testGroundingRequest()
	for _, g := range []groundingRequest{
		{Mode: "clarify", ClarifyAttribute: "not_configured"},
		{Mode: "recommend", ClarifyAttribute: "sleeve_length"},
		{Mode: "recommend", Requirements: testAttributeRequirements("sleeve_length", "striped")},
		{Mode: "clarify", ClarifyAttribute: "sleeve_length", Requirements: testAttributeRequirements("sleeve_length", "short")},
	} {
		b, _ := marshalGroundedSearch(req, g, true)
		if _, _, err := parseGroundedSearch(string(b), true, cfg, "fixture"); err == nil {
			t.Fatal("invalid generic action accepted", g)
		}
	}
	g := groundingRequest{Mode: "recommend", Requirements: []attributeRequirement{{Attribute: "pattern", ValueID: "striped"}, {Attribute: "color", ValueID: "pink"}, {Attribute: "sleeve_length", ValueID: "long"}}}
	b, _ := marshalGroundedSearch(req, g, true)
	parsed, meta, err := parseGroundedSearch(string(b), true, cfg, "fixture")
	if err != nil {
		t.Fatal(err)
	}
	out, _, _ := groundSearchResult(&recallsvc.SearchGoodsResult{Hits: []recallsvc.GoodsHit{testGroundingHit("match")}}, parsed, meta, c, "fixture")
	if len(out.Hits) != 1 {
		t.Fatal("combined attributes failed")
	}
}

// A clarification does not update LastSearch. Its verified option must remain
// usable even when the next knowledge sample omits one of its categories.
type priorSourceGroundingRecall struct {
	groundingRecallFixture
	previous *aichat.SearchGoodsParams
}

func (r *priorSourceGroundingRecall) ValidateToolParamSources(current aichat.SearchGoodsParams, previous *aichat.SearchGoodsParams, _ recallsvc.ToolParamEvidence) error {
	r.previous = previous
	if previous == nil || string(previous.ToolParams["categories"]) != string(current.ToolParams["categories"]) {
		return fmt.Errorf("category absent from this turn's knowledge and previous search")
	}
	return nil
}
func TestGroundingPendingSelectionRetainsValidatedSearchSource(t *testing.T) {
	c := testGroundingConfig()
	c.Attributes = append(c.Attributes, testSleeveAttribute())
	req := testGroundingRequest()
	delete(req.ToolParams, "sleeve_length")
	h := testGroundingHit("short")
	h.Properties["sleeve_length"] = "Short Sleeve"
	_, pending, _ := groundSearchResult(&recallsvc.SearchGoodsResult{Hits: []recallsvc.GoodsHit{h}}, req, groundingRequest{Mode: "clarify", ClarifyAttribute: "sleeve_length"}, c, "fixture")
	chosen := pending.Options[0]
	cfg := &chatConfig{raw: &recconf.AIChatConfig{GroundingConfig: c}, pendingClarification: pending, fieldAware: true}
	args, _ := marshalGroundedSearch(recallsvc.SearchGoodsRequest{SearchGoodsParams: chosen.Search}, groundingRequest{Mode: "recommend", OptionID: chosen.ID}, true)
	r := &priorSourceGroundingRecall{}
	response := &aichat.StreamResult{FinishReason: "tool_calls", ToolCalls: []aichat.ToolCall{{ID: "choice", Type: "function", Function: aichat.FunctionCall{Name: "search_goods", Arguments: string(args)}}}}
	if err := normalizePlannerResponse(response, cfg, r, nil, nil); err != nil {
		t.Fatal("verified selection lost its category source", err)
	}
	if r.previous == nil || canonicalSearch(*r.previous) != canonicalSearch(chosen.Search) {
		t.Fatal("validation did not use the complete verified choice")
	}
	changed := cloneGroundingSearch(chosen.Search)
	changed.Keywords = []string{"dress"}
	args, _ = marshalGroundedSearch(recallsvc.SearchGoodsRequest{SearchGoodsParams: changed}, groundingRequest{Mode: "recommend", OptionID: chosen.ID}, true)
	response.ToolCalls[0].Function.Arguments = string(args)
	if normalizePlannerResponse(response, cfg, r, nil, nil) == nil {
		t.Fatal("changed conditions reused stale category evidence")
	}
}
