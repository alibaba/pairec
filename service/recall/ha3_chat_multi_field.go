package recall

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibabacloud-go/tea/tea"
)

const fieldAwareSearchAttempts = 5

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
				panic(fmt.Sprintf("Ha3ChatRecallConf.%s is required for field-aware search", field.name))
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
	if len(req.Keywords) == 0 || len(req.Keywords) > maxSearchKeywordCount {
		return req, fmt.Errorf("keywords must contain 1 to %d terms or phrases", maxSearchKeywordCount)
	}
	if len(req.PreferredKeywords) > maxSearchPreferredKeywordCount {
		return req, fmt.Errorf("preferred_keywords must contain at most %d values", maxSearchPreferredKeywordCount)
	}
	if len(req.ExcludeKeywords) > 5 {
		return req, fmt.Errorf("exclude_keywords must contain at most 5 values")
	}

	var err error
	if req.Keywords, err = trimUniqueNonEmpty("keywords", req.Keywords); err != nil {
		return req, err
	}
	if req.PreferredKeywords, err = trimUniqueNonEmpty("preferred_keywords", req.PreferredKeywords); err != nil {
		return req, err
	}
	if req.ExcludeKeywords, err = trimUniqueNonEmpty("exclude_keywords", req.ExcludeKeywords); err != nil {
		return req, err
	}
	if req.MinPrice != nil && (!isFinitePrice(*req.MinPrice) || *req.MinPrice <= 0) {
		req.MinPrice = nil
		log.Info("module=Ha3ChatRecall\tevent=optional_price_ignored\tfield=min_price\treason=nonpositive_or_nonfinite")
	}
	if req.MaxPrice != nil && (!isFinitePrice(*req.MaxPrice) || *req.MaxPrice <= 0) {
		req.MaxPrice = nil
		log.Info("module=Ha3ChatRecall\tevent=optional_price_ignored\tfield=max_price\treason=nonpositive_or_nonfinite")
	}
	if req.MinPrice != nil && req.MaxPrice != nil && *req.MinPrice > *req.MaxPrice {
		req.MinPrice, req.MaxPrice = nil, nil
		log.Info("module=Ha3ChatRecall\tevent=optional_price_ignored\tfield=min_price,max_price\treason=conflicting_bounds")
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

func (r *Ha3ChatRecall) searchFieldWithRetry(ctx context.Context, field string, keywords []string, operator string, req SearchGoodsRequest, hit int) (*SearchGoodsResult, error) {
	ctx, cancel := context.WithTimeout(ctx, fieldAwareSearchTimeout)
	defer cancel()
	if hit <= 0 {
		return nil, fmt.Errorf("search hit must be positive")
	}
	if _, err := r.buildFieldQueryExpr(field, keywords, operator, req.ExcludeKeywords); err != nil {
		return nil, err
	}
	var lastErr error
	for attempt := 1; attempt <= fieldAwareSearchAttempts; attempt++ {
		result, err := r.searchField(ctx, field, keywords, operator, req, hit)
		if err == nil {
			return result, nil
		}
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		lastErr = err
		if attempt == fieldAwareSearchAttempts || !retryableHa3SearchError(err) {
			break
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

func isFinitePrice(price float64) bool {
	return !math.IsNaN(price) && !math.IsInf(price, 0)
}
