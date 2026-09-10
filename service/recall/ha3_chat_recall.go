package recall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	pairecctx "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/ha3engine"
	"github.com/alibaba/pairec/v2/datasource/ha3engine/ha3client"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	maxSearchKeywordCount          = aichat.SearchGoodsMaxKeywords
	maxSearchPreferredKeywordCount = aichat.SearchGoodsMaxPreferredKeywords
	maxSearchKeywordRunes          = 64
	fieldAwareSearchTimeout        = 2 * time.Second
)

type Ha3ChatRecall struct {
	*BaseRecall
	client    *ha3engine.Ha3EngineClient
	conf      recconf.Ha3ChatRecallConfig
	knowledge *ha3KnowledgeSearcher
}

type SearchGoodsRequest struct {
	aichat.SearchGoodsParams
	Limit      int  `json:"-"`
	FieldAware bool `json:"-"`
}

type GoodsHit struct {
	ItemId             string                 `json:"item_id"`
	Title              string                 `json:"title,omitempty"`
	Content            string                 `json:"content,omitempty"`
	Score              interface{}            `json:"score,omitempty"`
	Properties         map[string]interface{} `json:"raw,omitempty"`
	ConstraintEvidence map[string]interface{} `json:"constraint_evidence,omitempty"`
}

type SearchGoodsResult struct {
	Total                    int        `json:"total"`
	Hits                     []GoodsHit `json:"hits"`
	DroppedPreferredKeywords []string   `json:"dropped_preferred_keywords,omitempty"`
}

func NewHa3ChatRecall(config recconf.RecallConfig) *Ha3ChatRecall {
	client, err := ha3engine.GetHa3EngineClient(config.Ha3ChatRecallConf.EngineName)
	if err != nil {
		panic(err)
	}
	conf := config.Ha3ChatRecallConf
	if !ha3ChatFieldAwareConfigured(conf) {
		if conf.DefaultField == "" {
			conf.DefaultField = "default"
		}
		if conf.PriceField == "" {
			conf.PriceField = "price"
		}
	}
	validateHa3ChatFieldConfig(conf)
	conf.SearchGoodsConf = cloneSearchGoodsConfig(conf.SearchGoodsConf)
	validateSearchToolParams(conf.SearchGoodsConf, config.Ha3KnowledgeVectorConf)
	conf.DistinctConf = normalizeHa3ChatDistinctConfig(conf.DistinctConf)
	if config.Ha3KnowledgeVectorConf != nil && !ha3ChatFieldAwareConfigured(conf) {
		panic("Ha3KnowledgeVectorConf requires field-aware Ha3ChatRecallConf")
	}
	recall := &Ha3ChatRecall{
		BaseRecall: NewBaseRecall(config),
		client:     client,
		conf:       conf,
	}
	if config.Ha3KnowledgeVectorConf != nil {
		recall.knowledge = newHa3KnowledgeSearcher(client, *config.Ha3KnowledgeVectorConf)
	}
	return recall
}

func (r *Ha3ChatRecall) GetCandidateItems(user *module.User, context *pairecctx.RecommendContext) []*module.Item {
	return nil
}

func (r *Ha3ChatRecall) Search(ctx context.Context, req SearchGoodsRequest) (*SearchGoodsResult, error) {
	fieldAware := req.FieldAware && r.fieldAwareEnabled()
	req.FieldAware = fieldAware
	var err error
	if fieldAware {
		req, err = normalizeFieldAwareRequest(req)
		if err != nil {
			return nil, err
		}
		if req.Limit <= 0 {
			return nil, fmt.Errorf("limit must be positive")
		}
	}
	search := r.searchField
	if fieldAware {
		search = r.searchFieldWithRetry
	}
	if err := r.ValidateSearchGoodsRequest(req); err != nil {
		return nil, err
	}
	keywords := append(append([]string(nil), req.Keywords...), req.PreferredKeywords...)
	result, err := search(ctx, r.conf.DefaultField, keywords, "AND", req, req.Limit)
	if err != nil || result == nil || result.Total != 0 || len(result.Hits) != 0 || len(req.PreferredKeywords) == 0 || r.conf.SearchGoodsConf == nil || !r.conf.SearchGoodsConf.DropPreferredOnEmpty {
		if err == nil {
			r.attachToolParamEvidence(result, req.ToolParams)
		}
		return result, err
	}
	result, err = search(ctx, r.conf.DefaultField, req.Keywords, "AND", req, req.Limit)
	if err == nil && result != nil {
		result.DroppedPreferredKeywords = append([]string(nil), req.PreferredKeywords...)
		r.attachToolParamEvidence(result, req.ToolParams)
	}
	return result, err
}

func (r *Ha3ChatRecall) searchField(ctx context.Context, field string, keywords []string, operator string, req SearchGoodsRequest, hit int) (*SearchGoodsResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	queryExpr, err := r.buildFieldQueryExpr(field, keywords, operator, req.ExcludeKeywords)
	if err != nil {
		return nil, err
	}
	filterExpr := r.buildFilterExpr(req)
	toolExpr, err := r.buildToolParamsExpr(req.ToolParams)
	if err != nil {
		return nil, err
	}
	if toolExpr != "" {
		if filterExpr != "" {
			filterExpr += " AND "
		}
		filterExpr += toolExpr
	}
	body := map[string]interface{}{
		"query": queryExpr,
		"config": map[string]interface{}{
			"start":  0,
			"hit":    hit,
			"format": "json",
		},
	}
	if r.conf.Analyzer != "" {
		body["analyzer"] = map[string]interface{}{
			field: r.conf.Analyzer,
		}
	}
	if filterExpr != "" {
		body["filter"] = filterExpr
	}
	if r.conf.DistinctConf != nil {
		body["distinct"] = buildHa3ChatDistinctClause(r.conf.DistinctConf)
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Ha3Client.SearchRestWithContext(
		ctx,
		tea.String(r.conf.IndexName),
		(&ha3client.SearchRequestModel{}).SetHeaders(map[string]*string{}).SetBody(string(bodyBytes)),
		r.client.Runtime(),
	)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	if err != nil {
		return nil, err
	}
	if req.FieldAware && r.fieldAwareEnabled() {
		return r.parseConfiguredResponse(resp)
	}
	return parseHa3ChatResponse(resp)
}

func (r *Ha3ChatRecall) buildQueryExpr(req SearchGoodsRequest) (string, error) {
	keywords := append(append([]string(nil), req.Keywords...), req.PreferredKeywords...)
	return r.buildFieldQueryExpr(r.conf.DefaultField, keywords, "AND", req.ExcludeKeywords)
}

func (r *Ha3ChatRecall) buildFieldQueryExpr(field string, keywords []string, operator string, excludeKeywords []string) (string, error) {
	if field == "" {
		return "", fmt.Errorf("search field is empty")
	}
	keywordLimit := maxSearchKeywordCount
	if r.fieldAwareEnabled() {
		keywordLimit += maxSearchPreferredKeywordCount
	}
	keywords = normalizeKeywords(keywords, keywordLimit)
	if len(keywords) == 0 {
		return "", fmt.Errorf("keywords is empty")
	}
	sep := " & "
	if strings.EqualFold(operator, "OR") {
		sep = " | "
	}
	first, rest := keywords[0], keywords[1:]
	pos := fmt.Sprintf("%s:'%s'", field, first)
	for _, kw := range rest {
		pos += sep + fmt.Sprintf("'%s'", kw)
	}
	excludes := normalizeKeywords(excludeKeywords, maxSearchKeywordCount)
	if len(excludes) == 0 {
		return pos, nil
	}
	neg := fmt.Sprintf("%s:'%s'", r.conf.DefaultField, excludes[0])
	for _, kw := range excludes[1:] {
		neg += " | " + fmt.Sprintf("'%s'", kw)
	}
	return fmt.Sprintf("(%s) ANDNOT (%s)", pos, neg), nil
}

func (r *Ha3ChatRecall) buildFilterExpr(req SearchGoodsRequest) string {
	conds := make([]string, 0, 3)
	if req.MinPrice != nil || req.MaxPrice != nil {
		conds = append(conds, r.conf.PriceField+" > 0")
	}
	if req.MinPrice != nil {
		conds = append(conds, fmt.Sprintf("%s >= %v", r.conf.PriceField, *req.MinPrice))
	}
	if req.MaxPrice != nil {
		conds = append(conds, fmt.Sprintf("%s <= %v", r.conf.PriceField, *req.MaxPrice))
	}
	return strings.Join(conds, " AND ")
}

func normalizeKeywords(keywords []string, limit int) []string {
	out := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		keyword = sanitizeSearchKeyword(keyword)
		if keyword != "" {
			out = append(out, keyword)
			if len(out) >= limit {
				break
			}
		}
	}
	return out
}

func sanitizeSearchKeyword(keyword string) string {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return ""
	}
	var b strings.Builder
	written := 0
	lastSpace := false
	for _, r := range keyword {
		if written >= maxSearchKeywordRunes {
			break
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastSpace = false
			written++
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
			lastSpace = false
			written++
		case unicode.IsSpace(r):
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
				written++
			}
		default:
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
				written++
			}
		}
	}
	return strings.TrimSpace(b.String())
}

func parseHa3ChatResponse(resp *ha3client.SearchResponseModel) (*SearchGoodsResult, error) {
	total, items, err := decodeHa3ChatResponse(resp, false)
	if err != nil {
		return nil, err
	}
	hits := make([]GoodsHit, 0, len(items))
	for index, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		fields, _ := itemMap["fields"].(map[string]interface{})
		if fields == nil {
			fields = itemMap
		}
		itemID := stringFromAny(fields["item_id"])
		if itemID == "" {
			itemID = stringFromAny(fields["id"])
		}
		if itemID == "" {
			itemID = stringFromAny(itemMap["id"])
		}
		if itemID == "" {
			log.Warning(fmt.Sprintf("module=Ha3ChatRecall\tevent=empty_item_id\titemIndex=%d", index))
			continue
		}
		hits = append(hits, GoodsHit{
			ItemId:     itemID,
			Title:      stringFromAny(fields["title"]),
			Content:    firstString(fields["default"], fields["content_desc"], fields["content"]),
			Score:      firstAny(itemMap["sortExprValues"], itemMap["score"]),
			Properties: fields,
		})
	}
	return &SearchGoodsResult{Total: total, Hits: hits}, nil
}

func (r *Ha3ChatRecall) parseConfiguredResponse(resp *ha3client.SearchResponseModel) (*SearchGoodsResult, error) {
	total, items, err := decodeHa3ChatResponse(resp, true)
	if err != nil {
		return nil, err
	}
	if total > 0 && len(items) == 0 {
		return nil, fmt.Errorf("ha3 search response has total=%d but no items", total)
	}
	hits := make([]GoodsHit, 0, len(items))
	for index, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		fields, _ := itemMap["fields"].(map[string]interface{})
		if fields == nil {
			fields = itemMap
		}
		itemID := stringFromAny(fields[r.conf.ItemIdField])
		if itemID == "" {
			log.Warning(fmt.Sprintf("module=Ha3ChatRecall\tevent=empty_configured_item_id\tfield=%s\titemIndex=%d", r.conf.ItemIdField, index))
			continue
		}
		properties := make(map[string]interface{}, len(fields))
		for key, value := range fields {
			if key != r.conf.ItemIdField {
				properties[key] = value
			}
		}
		hits = append(hits, GoodsHit{
			ItemId:     itemID,
			Title:      configuredString(fields, r.conf.TitleField),
			Content:    configuredString(fields, r.conf.ContentField),
			Score:      firstAny(itemMap["sortExprValues"], itemMap["score"]),
			Properties: properties,
		})
	}
	if len(items) > 0 && len(hits) == 0 {
		return nil, fmt.Errorf("ha3 search response contains %d items but none has configured item id field %q", len(items), r.conf.ItemIdField)
	}
	return &SearchGoodsResult{Total: total, Hits: hits}, nil
}

func decodeHa3ChatResponse(resp *ha3client.SearchResponseModel, strict bool) (int, []interface{}, error) {
	if resp == nil || resp.Body == nil {
		if strict {
			return 0, nil, fmt.Errorf("ha3 search response body is empty")
		}
		return 0, nil, nil
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(tea.StringValue(resp.Body)), &body); err != nil {
		return 0, nil, err
	}
	resultMap := body
	if v, ok := body["result"].(map[string]interface{}); ok {
		resultMap = v
	}
	if strict {
		if err := validateHa3ChatResponse(body, resultMap); err != nil {
			return 0, nil, err
		}
	}
	items, _ := resultMap["items"].([]interface{})
	if len(items) == 0 {
		items, _ = resultMap["hits"].([]interface{})
	}
	total := intFromAny(resultMap["totalHits"])
	if total == 0 {
		total = intFromAny(resultMap["total"])
	}
	if total == 0 {
		total = intFromAny(resultMap["numHits"])
	}
	if total == 0 {
		total = len(items)
	}
	return total, items, nil
}

func validateHa3ChatResponse(body, result map[string]interface{}) error {
	if status, ok := body["status"].(string); ok && status != "" && !strings.EqualFold(status, "OK") && !strings.EqualFold(status, "SUCCESS") {
		return fmt.Errorf("ha3 search status %q", status)
	}
	for _, responseErrors := range []interface{}{body["errors"], result["errors"]} {
		if hasHa3ResponseErrors(responseErrors) {
			payload, _ := json.Marshal(responseErrors)
			return fmt.Errorf("ha3 search errors: %s", payload)
		}
	}
	if covered, ok := result["coveredPercent"].(float64); ok && covered < 100 {
		return fmt.Errorf("ha3 search response is incomplete: coveredPercent=%v", covered)
	}
	hasCount := false
	for _, field := range []string{"totalHits", "total", "numHits"} {
		if raw, exists := result[field]; exists {
			hasCount = true
			count, ok := raw.(float64)
			if !ok || count < 0 || count != float64(int(count)) {
				return fmt.Errorf("ha3 search response has invalid %s", field)
			}
		}
	}
	if !hasCount {
		return fmt.Errorf("ha3 search response is missing hit count")
	}
	return nil
}

func hasHa3ResponseErrors(value interface{}) bool {
	switch value := value.(type) {
	case nil:
		return false
	case string:
		return value != ""
	case []interface{}:
		return len(value) > 0
	case map[string]interface{}:
		return len(value) > 0
	case bool:
		return value
	case float64:
		return value != 0
	default:
		return true
	}
}

func configuredString(fields map[string]interface{}, field string) string {
	if field == "" {
		return ""
	}
	return stringFromAny(fields[field])
}

func intFromAny(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	default:
		return 0
	}
}

func stringFromAny(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	default:
		return ""
	}
}

func firstString(values ...interface{}) string {
	for _, value := range values {
		if s := stringFromAny(value); s != "" {
			return s
		}
	}
	return ""
}

func firstAny(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
