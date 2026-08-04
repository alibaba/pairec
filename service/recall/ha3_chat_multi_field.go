package recall

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	multiFieldRouteHit       = 36
	multiFieldSearchAttempts = 5
	multiFieldRRFK           = 60.0
	multiFieldAttributeBonus = 0.08
)

type ha3FieldRoute struct {
	name     string
	field    string
	keywords []string
	weight   float64
}

type ha3FieldRouteResult struct {
	route  ha3FieldRoute
	result *SearchGoodsResult
	err    error
}

type ha3FusedCandidate struct {
	hit        GoodsHit
	rrfScore   float64
	fusedScore float64
	routes     map[string]int
}

func validateHa3ChatFieldConfig(conf recconf.Ha3ChatRecallConfig) {
	titleConfigured := conf.TitleField != ""
	categoryConfigured := conf.CategoryField != ""
	if titleConfigured != categoryConfigured {
		panic("Ha3ChatRecallConf.TitleField and CategoryField must be configured together")
	}
	if !titleConfigured {
		if conf.CategoriesField != "" || conf.ContentField != "" || conf.TagsField != "" {
			panic("Ha3ChatRecallConf optional semantic fields require TitleField and CategoryField")
		}
	} else {
		requiredFields := []struct {
			name  string
			value string
		}{
			{name: "DefaultField", value: conf.DefaultField},
			{name: "ItemIdField", value: conf.ItemIdField},
			{name: "PriceField", value: conf.PriceField},
		}
		for _, field := range requiredFields {
			if field.value == "" {
				panic(fmt.Sprintf("Ha3ChatRecallConf.%s is required for field-aware fallback", field.name))
			}
		}
	}

	configuredFields := []struct {
		name  string
		value string
	}{
		{name: "ItemIdField", value: conf.ItemIdField},
		{name: "DefaultField", value: conf.DefaultField},
		{name: "TitleField", value: conf.TitleField},
		{name: "CategoryField", value: conf.CategoryField},
		{name: "CategoriesField", value: conf.CategoriesField},
		{name: "ContentField", value: conf.ContentField},
		{name: "TagsField", value: conf.TagsField},
		{name: "PriceField", value: conf.PriceField},
	}
	for _, field := range configuredFields {
		if field.value != "" && !validHa3FieldName(field.value) {
			panic(fmt.Sprintf("Ha3ChatRecallConf.%s has invalid field name %q", field.name, field.value))
		}
	}
}

func ha3ChatFieldAwareConfigured(conf recconf.Ha3ChatRecallConfig) bool {
	return conf.TitleField != "" && conf.CategoryField != ""
}

func (r *Ha3ChatRecall) fieldAwareEnabled() bool {
	return ha3ChatFieldAwareConfigured(r.conf)
}

func validHa3FieldName(field string) bool {
	for index := 0; index < len(field); index++ {
		char := field[index]
		if index == 0 {
			if !isASCIIAlpha(char) && char != '_' {
				return false
			}
			continue
		}
		if !isASCIIAlpha(char) && (char < '0' || char > '9') && char != '_' && char != '.' {
			return false
		}
	}
	return field != ""
}

func isASCIIAlpha(char byte) bool {
	return char >= 'A' && char <= 'Z' || char >= 'a' && char <= 'z'
}

func normalizeFieldAwareRequest(req SearchGoodsRequest) (SearchGoodsRequest, error) {
	if len(req.Keywords) != 1 {
		return req, fmt.Errorf("keywords must contain exactly one phrase")
	}
	if len(req.ProductTypeKeywords) < 1 || len(req.ProductTypeKeywords) > 4 {
		return req, fmt.Errorf("product_type_keywords must contain 1-4 values")
	}
	if req.AttributeKeywords == nil {
		return req, fmt.Errorf("attribute_keywords is required")
	}
	if len(req.AttributeKeywords) > 5 {
		return req, fmt.Errorf("attribute_keywords must contain at most 5 values")
	}
	if len(req.ExcludeKeywords) > 5 {
		return req, fmt.Errorf("exclude_keywords must contain at most 5 values")
	}
	if req.Operator != "AND" {
		return req, fmt.Errorf("operator must be AND")
	}

	var err error
	if req.Keywords, err = trimUniqueNonEmpty("keywords", req.Keywords); err != nil {
		return req, err
	}
	if req.ProductTypeKeywords, err = trimUniqueNonEmpty("product_type_keywords", req.ProductTypeKeywords); err != nil {
		return req, err
	}
	if req.AttributeKeywords, err = trimUniqueNonEmpty("attribute_keywords", req.AttributeKeywords); err != nil {
		return req, err
	}
	if req.ExcludeKeywords, err = trimUniqueNonEmpty("exclude_keywords", req.ExcludeKeywords); err != nil {
		return req, err
	}
	if req.MinPrice != nil && !isFinitePrice(*req.MinPrice) {
		return req, fmt.Errorf("min_price must be finite")
	}
	if req.MaxPrice != nil && !isFinitePrice(*req.MaxPrice) {
		return req, fmt.Errorf("max_price must be finite")
	}
	if req.MinPrice != nil && req.MaxPrice != nil && *req.MinPrice > *req.MaxPrice {
		return req, fmt.Errorf("min_price exceeds max_price")
	}
	if req.MinPrice != nil && *req.MinPrice <= 0 {
		return req, fmt.Errorf("min_price must be positive")
	}
	if req.MaxPrice != nil && *req.MaxPrice <= 0 {
		return req, fmt.Errorf("max_price must be positive")
	}
	return req, nil
}

// NormalizeFieldAwareSearchGoodsRequest validates and canonicalizes the
// field-aware Planner contract before any search is executed.
func NormalizeFieldAwareSearchGoodsRequest(req SearchGoodsRequest) (SearchGoodsRequest, error) {
	return normalizeFieldAwareRequest(req)
}

func trimUniqueNonEmpty(field string, values []string) ([]string, error) {
	trimmed := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("%s contains an invalid string", field)
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		trimmed = append(trimmed, value)
	}
	return trimmed, nil
}

func (r *Ha3ChatRecall) searchMultiFieldFallback(ctx context.Context, req SearchGoodsRequest) (*SearchGoodsResult, error) {
	routes := r.multiFieldRoutes(req)
	results := make([]ha3FieldRouteResult, len(routes))
	var wg sync.WaitGroup
	for index, route := range routes {
		wg.Add(1)
		go func(index int, route ha3FieldRoute) {
			defer wg.Done()
			result, err := r.searchFieldWithRetry(ctx, route.field, route.keywords, "OR", req, multiFieldRouteHit)
			results[index] = ha3FieldRouteResult{route: route, result: result, err: err}
		}(index, route)
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	successCount := 0
	errors := make([]string, 0, len(results))
	for _, routeResult := range results {
		if routeResult.err == nil {
			successCount++
			continue
		}
		errors = append(errors, fmt.Sprintf("%s: %v", routeResult.route.name, routeResult.err))
		log.Warning(fmt.Sprintf("module=Ha3ChatRecall\tevent=field_route_failed\troute=%s\terr=%v", routeResult.route.name, routeResult.err))
	}
	if successCount == 0 {
		return nil, fmt.Errorf("all multi-field routes failed: %s", strings.Join(errors, "; "))
	}
	result := r.fuseFieldRoutes(results, req.AttributeKeywords, req.Limit)
	result.FallbackUsed = true
	result.RouteErrors = errors
	return result, nil
}

func (r *Ha3ChatRecall) searchFieldWithRetry(ctx context.Context, field string, keywords []string, operator string, req SearchGoodsRequest, hit int) (*SearchGoodsResult, error) {
	if hit <= 0 {
		return nil, fmt.Errorf("search hit must be positive")
	}
	if _, err := r.buildFieldQueryExpr(field, keywords, operator, req.ExcludeKeywords); err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 1; attempt <= multiFieldSearchAttempts; attempt++ {
		result, err := r.searchField(ctx, field, keywords, operator, req, hit)
		if err == nil {
			return result, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		lastErr = err
		if attempt == multiFieldSearchAttempts || !retryableHa3SearchError(err) {
			break
		}
		delay := multiFieldRetryDelay(attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func retryableHa3SearchError(err error) bool {
	if err == nil {
		return false
	}
	var sdkErr *tea.SDKError
	if errors.As(err, &sdkErr) {
		statusCode := tea.IntValue(sdkErr.StatusCode)
		if statusCode == 0 {
			statusCode, _ = strconv.Atoi(tea.StringValue(sdkErr.Code))
		}
		return statusCode == http.StatusRequestTimeout ||
			statusCode == http.StatusTooManyRequests ||
			statusCode >= http.StatusInternalServerError
	}
	return tea.BoolValue(tea.Retryable(err))
}

func multiFieldRetryDelay(attempt int) time.Duration {
	base := time.Second << (attempt - 1)
	maxJitter := base / 4
	if maxJitter > 500*time.Millisecond {
		maxJitter = 500 * time.Millisecond
	}
	jitter := time.Duration(rand.Int63n(int64(maxJitter) + 1))
	return base + jitter
}

func isFinitePrice(price float64) bool {
	return !math.IsNaN(price) && !math.IsInf(price, 0)
}

func (r *Ha3ChatRecall) multiFieldRoutes(req SearchGoodsRequest) []ha3FieldRoute {
	routes := make([]ha3FieldRoute, 0, 4)
	if r.conf.CategoriesField != "" {
		routes = append(routes, ha3FieldRoute{
			name:     "categories_type",
			field:    r.conf.CategoriesField,
			keywords: req.ProductTypeKeywords,
			weight:   12,
		})
	}
	routes = append(routes,
		ha3FieldRoute{
			name:     "category_type",
			field:    r.conf.CategoryField,
			keywords: req.ProductTypeKeywords,
			weight:   10,
		},
		ha3FieldRoute{
			name:     "title_raw",
			field:    r.conf.TitleField,
			keywords: req.Keywords,
			weight:   9,
		},
		ha3FieldRoute{
			name:     "title_type",
			field:    r.conf.TitleField,
			keywords: req.ProductTypeKeywords,
			weight:   7,
		},
	)
	return routes
}

func (r *Ha3ChatRecall) fuseFieldRoutes(results []ha3FieldRouteResult, attributes []string, limit int) *SearchGoodsResult {
	byID := make(map[string]*ha3FusedCandidate)
	for _, routeResult := range results {
		if routeResult.err != nil || routeResult.result == nil {
			continue
		}
		for index, hit := range routeResult.result.Hits {
			candidate := byID[hit.ItemId]
			if candidate == nil {
				candidate = &ha3FusedCandidate{
					hit:    hit,
					routes: make(map[string]int),
				}
				byID[hit.ItemId] = candidate
			}
			rank := index + 1
			candidate.rrfScore += routeResult.route.weight / (multiFieldRRFK + float64(rank))
			candidate.routes[routeResult.route.name] = rank
		}
	}

	ranked := make([]*ha3FusedCandidate, 0, len(byID))
	for _, candidate := range byID {
		candidate.fusedScore = candidate.rrfScore + multiFieldAttributeBonus*r.attributeMatchFraction(candidate.hit, attributes)
		ranked = append(ranked, candidate)
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].fusedScore != ranked[j].fusedScore {
			return ranked[i].fusedScore > ranked[j].fusedScore
		}
		if len(ranked[i].routes) != len(ranked[j].routes) {
			return len(ranked[i].routes) > len(ranked[j].routes)
		}
		return ranked[i].hit.ItemId < ranked[j].hit.ItemId
	})

	if limit < 0 {
		limit = 0
	}
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	hits := make([]GoodsHit, 0, len(ranked))
	for _, candidate := range ranked {
		hits = append(hits, candidate.hit)
	}
	return &SearchGoodsResult{Total: len(byID), Hits: hits}
}

func (r *Ha3ChatRecall) attributeText(hit GoodsHit) string {
	fields := []string{
		r.conf.TitleField,
		r.conf.ContentField,
		r.conf.CategoryField,
		r.conf.CategoriesField,
		r.conf.TagsField,
		r.conf.DefaultField,
	}
	parts := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		if value := textFromAny(hit.Properties[field]); value != "" {
			parts = append(parts, value)
		}
	}
	return strings.ToLower(strings.Join(parts, " "))
}

func (r *Ha3ChatRecall) attributeMatchFraction(hit GoodsHit, attributes []string) float64 {
	if len(attributes) == 0 {
		return 0
	}
	text := r.attributeText(hit)
	matched := 0
	for _, attribute := range attributes {
		terms := normalizeKeywords([]string{attribute})
		if len(terms) > 0 && strings.Contains(text, strings.ToLower(terms[0])) {
			matched++
		}
	}
	return float64(matched) / float64(len(attributes))
}

func textFromAny(value interface{}) string {
	if value == nil {
		return ""
	}
	if values, ok := value.([]interface{}); ok {
		parts := make([]string, 0, len(values))
		for _, item := range values {
			if text := textFromAny(item); text != "" {
				parts = append(parts, text)
			}
		}
		return strings.Join(parts, " ")
	}
	if value, ok := value.([]string); ok {
		return strings.Join(value, " ")
	}
	return fmt.Sprint(value)
}
