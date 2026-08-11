package recall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	pairecctx "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/ha3engine"
	"github.com/alibaba/pairec/v2/datasource/ha3engine/ha3client"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	maxSearchKeywordCount = 8
	maxSearchKeywordRunes = 64
)

type Ha3ChatRecall struct {
	*BaseRecall
	client    *ha3engine.Ha3EngineClient
	conf      recconf.Ha3ChatRecallConfig
	knowledge *ha3KnowledgeSearcher
}

type SearchGoodsRequest struct {
	Keywords              []string `json:"keywords"`
	ProductTypeKeywords   []string `json:"product_type_keywords"`
	AttributeKeywords     []string `json:"attribute_keywords"`
	Operator              string   `json:"operator"`
	ExcludeKeywords       []string `json:"exclude_keywords,omitempty"`
	MinPrice              *float64 `json:"min_price,omitempty"`
	MaxPrice              *float64 `json:"max_price,omitempty"`
	Limit                 int      `json:"-"`
	MultiFieldFallback    bool     `json:"-"`
	KnowledgeCandidateIDs []string `json:"knowledge_candidate_ids,omitempty"`
}

type GoodsHit struct {
	ItemId     string                 `json:"item_id"`
	Title      string                 `json:"title,omitempty"`
	Content    string                 `json:"content,omitempty"`
	Score      interface{}            `json:"score,omitempty"`
	Properties map[string]interface{} `json:"raw,omitempty"`
}

type SearchGoodsResult struct {
	Total        int        `json:"total"`
	Hits         []GoodsHit `json:"hits"`
	FallbackUsed bool       `json:"-"`
	RouteErrors  []string   `json:"-"`
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
	fieldAware := req.MultiFieldFallback && r.fieldAwareEnabled()
	req.MultiFieldFallback = fieldAware
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
	result, err := search(ctx, r.conf.DefaultField, req.Keywords, req.Operator, req, req.Limit)
	if err != nil {
		return nil, err
	}
	if len(result.Hits) > 0 || !fieldAware {
		return result, nil
	}
	return r.searchMultiFieldFallback(ctx, req)
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
	resp, err := r.client.Ha3Client.SearchRestWithOptions(
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
	if req.MultiFieldFallback && r.fieldAwareEnabled() {
		return r.parseConfiguredResponse(resp)
	}
	return parseHa3ChatResponse(resp)
}

func (r *Ha3ChatRecall) buildQueryExpr(req SearchGoodsRequest) (string, error) {
	return r.buildFieldQueryExpr(r.conf.DefaultField, req.Keywords, req.Operator, req.ExcludeKeywords)
}

func (r *Ha3ChatRecall) buildFieldQueryExpr(field string, keywords []string, operator string, excludeKeywords []string) (string, error) {
	if field == "" {
		return "", fmt.Errorf("search field is empty")
	}
	keywords = normalizeKeywords(keywords)
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
	excludes := normalizeKeywords(excludeKeywords)
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

func normalizeKeywords(keywords []string) []string {
	out := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		keyword = sanitizeSearchKeyword(keyword)
		if keyword != "" {
			out = append(out, keyword)
			if len(out) >= maxSearchKeywordCount {
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
	total, items, responseErrors, err := decodeHa3ChatResponse(resp)
	if err != nil {
		return nil, err
	}
	if hasHa3ResponseErrors(responseErrors) {
		payload, _ := json.Marshal(responseErrors)
		return nil, fmt.Errorf("ha3 search errors: %s", payload)
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
	if resp == nil || resp.Body == nil {
		return nil, fmt.Errorf("ha3 search response body is empty")
	}
	total, items, responseErrors, err := decodeHa3ChatResponse(resp)
	if err != nil {
		return nil, err
	}
	if hasHa3ResponseErrors(responseErrors) {
		payload, _ := json.Marshal(responseErrors)
		return nil, fmt.Errorf("ha3 search errors: %s", payload)
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

func decodeHa3ChatResponse(resp *ha3client.SearchResponseModel) (int, []interface{}, interface{}, error) {
	if resp == nil || resp.Body == nil {
		return 0, nil, nil, nil
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(tea.StringValue(resp.Body)), &body); err != nil {
		return 0, nil, nil, err
	}
	resultMap := body
	responseErrors := body["errors"]
	if v, ok := body["result"].(map[string]interface{}); ok {
		resultMap = v
		if responseErrors == nil {
			responseErrors = v["errors"]
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
	return total, items, responseErrors, nil
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
