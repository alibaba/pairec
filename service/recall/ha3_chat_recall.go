package recall

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	pairecctx "github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/datasource/ha3engine"
	"github.com/alibaba/pairec/v2/datasource/ha3engine/ha3client"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibabacloud-go/tea/tea"
)

type Ha3ChatRecall struct {
	*BaseRecall
	client *ha3engine.Ha3EngineClient
	conf   recconf.Ha3ChatRecallConfig
}

type SearchGoodsRequest struct {
	Keywords        []string `json:"keywords"`
	Operator        string   `json:"operator"`
	ExcludeKeywords []string `json:"exclude_keywords"`
	MinPrice        *float64 `json:"min_price"`
	MaxPrice        *float64 `json:"max_price"`
	Limit           int
}

type GoodsHit struct {
	ItemId     string                 `json:"item_id"`
	Title      string                 `json:"title,omitempty"`
	Content    string                 `json:"content,omitempty"`
	Score      interface{}            `json:"score,omitempty"`
	Properties map[string]interface{} `json:"raw,omitempty"`
}

type SearchGoodsResult struct {
	Total int        `json:"total"`
	Hits  []GoodsHit `json:"hits"`
}

func NewHa3ChatRecall(config recconf.RecallConfig) *Ha3ChatRecall {
	client, err := ha3engine.GetHa3EngineClient(config.Ha3ChatRecallConf.EngineName)
	if err != nil {
		panic(err)
	}
	conf := config.Ha3ChatRecallConf
	if conf.DefaultField == "" {
		conf.DefaultField = "default"
	}
	if conf.PriceField == "" {
		conf.PriceField = "price"
	}
	return &Ha3ChatRecall{
		BaseRecall: NewBaseRecall(config),
		client:     client,
		conf:       conf,
	}
}

func (r *Ha3ChatRecall) GetCandidateItems(user *module.User, context *pairecctx.RecommendContext) []*module.Item {
	return nil
}

func (r *Ha3ChatRecall) Search(ctx context.Context, req SearchGoodsRequest) (*SearchGoodsResult, error) {
	queryExpr, err := r.buildQueryExpr(req)
	if err != nil {
		return nil, err
	}
	filterExpr := r.buildFilterExpr(req)
	body := map[string]interface{}{
		"query": queryExpr,
		"config": map[string]interface{}{
			"start":  0,
			"hit":    req.Limit,
			"format": "json",
		},
		"analyzer": map[string]interface{}{
			r.conf.DefaultField: r.conf.Analyzer,
		},
	}
	if filterExpr != "" {
		body["filter"] = filterExpr
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
	if err != nil {
		return nil, err
	}
	return parseHa3ChatResponse(resp)
}

func (r *Ha3ChatRecall) buildQueryExpr(req SearchGoodsRequest) (string, error) {
	keywords := normalizeKeywords(req.Keywords)
	if len(keywords) == 0 {
		return "", fmt.Errorf("keywords is empty")
	}
	sep := " & "
	if strings.EqualFold(req.Operator, "OR") {
		sep = " | "
	}
	first, rest := keywords[0], keywords[1:]
	pos := fmt.Sprintf("%s:'%s'", r.conf.DefaultField, first)
	for _, kw := range rest {
		pos += sep + fmt.Sprintf("'%s'", kw)
	}
	excludes := normalizeKeywords(req.ExcludeKeywords)
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
		keyword = strings.TrimSpace(strings.ReplaceAll(keyword, "'", " "))
		if keyword != "" {
			out = append(out, keyword)
		}
	}
	return out
}

func parseHa3ChatResponse(resp *ha3client.SearchResponseModel) (*SearchGoodsResult, error) {
	if resp == nil || resp.Body == nil {
		return &SearchGoodsResult{}, nil
	}
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(tea.StringValue(resp.Body)), &body); err != nil {
		return nil, err
	}
	resultMap := body
	if v, ok := body["result"].(map[string]interface{}); ok {
		resultMap = v
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
