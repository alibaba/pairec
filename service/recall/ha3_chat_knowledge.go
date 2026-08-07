package recall

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/datasource/ha3engine/ha3client"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

const (
	defaultKnowledgeTopK          = 20
	defaultKnowledgeSearchTimeout = 2500
	knowledgeEmbeddingTimeout     = 5 * time.Second
	knowledgeEmbeddingAttempts    = 2
)

var validKnowledgeTypes = map[string]struct{}{
	"category":   {},
	"categories": {},
	"content":    {},
	"brand":      {},
	"tag":        {},
	"title":      {},
}

type KnowledgeHit struct {
	KnowledgeID   string      `json:"knowledge_id"`
	KnowledgeType string      `json:"knowledge_type"`
	Value         string      `json:"value"`
	Category      string      `json:"category"`
	Score         interface{} `json:"score,omitempty"`
}

type KnowledgeSearchResult struct {
	Total              int            `json:"total"`
	Hits               []KnowledgeHit `json:"hits"`
	EmbeddingDimension int            `json:"embedding_dimension"`
	EmbeddingAttempts  int            `json:"embedding_attempts"`
	EmbeddingCostMs    int64          `json:"embedding_cost_ms"`
	SearchCostMs       int64          `json:"search_cost_ms"`
}

func normalizeHa3KnowledgeVectorConfig(conf recconf.Ha3KnowledgeVectorConfig) recconf.Ha3KnowledgeVectorConfig {
	conf.FeatureStoreName = strings.TrimSpace(conf.FeatureStoreName)
	conf.LLMConfigName = strings.TrimSpace(conf.LLMConfigName)
	conf.IndexName = strings.TrimSpace(conf.IndexName)
	conf.VectorIndexName = strings.TrimSpace(conf.VectorIndexName)
	conf.KnowledgeIDField = strings.TrimSpace(conf.KnowledgeIDField)
	conf.KnowledgeTypeField = strings.TrimSpace(conf.KnowledgeTypeField)
	conf.ValueField = strings.TrimSpace(conf.ValueField)
	conf.CategoryField = strings.TrimSpace(conf.CategoryField)
	conf.QueryTemplate = strings.TrimSpace(conf.QueryTemplate)
	if conf.EmbeddingDelimiter == "" {
		conf.EmbeddingDelimiter = "\x1d"
	}
	if conf.TopK <= 0 {
		conf.TopK = defaultKnowledgeTopK
	}
	if conf.SearchTimeout <= 0 {
		conf.SearchTimeout = defaultKnowledgeSearchTimeout
	}
	conf.KnowledgeTypes = append([]string(nil), conf.KnowledgeTypes...)
	for index := range conf.KnowledgeTypes {
		conf.KnowledgeTypes[index] = strings.TrimSpace(conf.KnowledgeTypes[index])
	}
	return conf
}

func validateHa3KnowledgeVectorConfig(conf recconf.Ha3KnowledgeVectorConfig) {
	required := []struct {
		name  string
		value string
	}{
		{name: "FeatureStoreName", value: conf.FeatureStoreName},
		{name: "LLMConfigName", value: conf.LLMConfigName},
		{name: "IndexName", value: conf.IndexName},
		{name: "VectorIndexName", value: conf.VectorIndexName},
		{name: "KnowledgeIDField", value: conf.KnowledgeIDField},
		{name: "KnowledgeTypeField", value: conf.KnowledgeTypeField},
		{name: "ValueField", value: conf.ValueField},
		{name: "CategoryField", value: conf.CategoryField},
		{name: "QueryTemplate", value: conf.QueryTemplate},
	}
	for _, field := range required {
		if field.value == "" {
			panic(fmt.Sprintf("Ha3KnowledgeVectorConf.%s is required", field.name))
		}
	}
	fieldNames := required[2:8]
	for _, field := range fieldNames {
		if !validHa3FieldName(field.value) {
			panic(fmt.Sprintf("Ha3KnowledgeVectorConf.%s has invalid name %q", field.name, field.value))
		}
	}
	if len(conf.KnowledgeTypes) == 0 {
		panic("Ha3KnowledgeVectorConf.KnowledgeTypes is required")
	}
	seen := make(map[string]struct{}, len(conf.KnowledgeTypes))
	for _, knowledgeType := range conf.KnowledgeTypes {
		if _, ok := validKnowledgeTypes[knowledgeType]; !ok {
			panic(fmt.Sprintf("Ha3KnowledgeVectorConf.KnowledgeTypes contains unsupported value %q", knowledgeType))
		}
		if _, ok := seen[knowledgeType]; ok {
			panic(fmt.Sprintf("Ha3KnowledgeVectorConf.KnowledgeTypes contains duplicate value %q", knowledgeType))
		}
		seen[knowledgeType] = struct{}{}
	}
	if strings.Count(conf.QueryTemplate, "{query}") != 1 {
		panic("Ha3KnowledgeVectorConf.QueryTemplate must contain exactly one {query}")
	}
	if conf.TopK < 1 || conf.TopK > 100 {
		panic("Ha3KnowledgeVectorConf.TopK must be between 1 and 100")
	}
	if conf.SearchTimeout < 1 {
		panic("Ha3KnowledgeVectorConf.SearchTimeout must be positive")
	}
}

func (r *Ha3ChatRecall) KnowledgeEnabled() bool {
	return r.knowledgeConf != nil && r.knowledgeLLM != nil
}

func (r *Ha3ChatRecall) SearchKnowledge(ctx context.Context, query string) (*KnowledgeSearchResult, error) {
	if !r.KnowledgeEnabled() {
		return nil, nil
	}
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("knowledge query is empty")
	}
	conf := r.knowledgeConf
	embeddingInput := strings.Replace(conf.QueryTemplate, "{query}", query, 1)
	embeddingCtx, cancel := context.WithTimeout(ctx, knowledgeEmbeddingTimeout)
	defer cancel()
	embeddingStart := time.Now()
	var (
		vectors  [][]float32
		err      error
		attempts int
	)
	for attempts = 1; attempts <= knowledgeEmbeddingAttempts; attempts++ {
		vectors, err = r.knowledgeLLM.CreateMultiModalEmbeddings(embeddingCtx, []domain.MultiModalContent{{Text: embeddingInput}})
		if err == nil {
			break
		}
		if embeddingCtx.Err() != nil {
			return nil, embeddingCtx.Err()
		}
	}
	if err != nil {
		return nil, fmt.Errorf("create knowledge embedding failed after %d attempts: %w", attempts-1, err)
	}
	if len(vectors) != 1 {
		return nil, fmt.Errorf("knowledge embedding returned %d vectors, want 1", len(vectors))
	}
	vectorText, err := encodeKnowledgeVector(vectors[0], conf.EmbeddingDelimiter)
	if err != nil {
		return nil, err
	}
	result := &KnowledgeSearchResult{
		EmbeddingDimension: len(vectors[0]),
		EmbeddingAttempts:  attempts,
		EmbeddingCostMs:    time.Since(embeddingStart).Milliseconds(),
	}
	searchStart := time.Now()
	searchResult, err := r.searchKnowledgeVector(ctx, vectorText)
	result.SearchCostMs = time.Since(searchStart).Milliseconds()
	if err != nil {
		return nil, err
	}
	result.Total = searchResult.Total
	result.Hits = searchResult.Hits
	return result, nil
}

func encodeKnowledgeVector(vector []float32, delimiter string) (string, error) {
	if len(vector) == 0 {
		return "", fmt.Errorf("knowledge embedding vector is empty")
	}
	parts := make([]string, len(vector))
	norm := float64(0)
	for index, value := range vector {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return "", fmt.Errorf("knowledge embedding vector contains invalid value at index %d", index)
		}
		norm += float64(value) * float64(value)
		parts[index] = strconv.FormatFloat(float64(value), 'f', 9, 32)
	}
	if norm == 0 {
		return "", fmt.Errorf("knowledge embedding vector has zero norm")
	}
	return strings.Join(parts, delimiter), nil
}

func (r *Ha3ChatRecall) searchKnowledgeVector(ctx context.Context, vectorText string) (*KnowledgeSearchResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	conf := r.knowledgeConf
	filters := make([]string, 0, len(conf.KnowledgeTypes))
	for _, knowledgeType := range conf.KnowledgeTypes {
		filters = append(filters, fmt.Sprintf(`%s=%q`, conf.KnowledgeTypeField, knowledgeType))
	}
	body := map[string]interface{}{
		"query":  fmt.Sprintf("%s:'%s&n=%d'", conf.VectorIndexName, vectorText, conf.TopK),
		"filter": strings.Join(filters, " OR "),
		"config": map[string]interface{}{
			"start":   0,
			"hit":     conf.TopK,
			"format":  "json",
			"timeout": conf.SearchTimeout,
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Ha3Client.SearchRestWithOptions(
		tea.String(conf.IndexName),
		(&ha3client.SearchRequestModel{}).SetHeaders(map[string]*string{}).SetBody(string(payload)),
		r.client.Runtime(),
	)
	if ctxErr := ctx.Err(); ctxErr != nil {
		return nil, ctxErr
	}
	if err != nil {
		return nil, err
	}
	return r.parseKnowledgeResponse(resp)
}

func (r *Ha3ChatRecall) parseKnowledgeResponse(resp *ha3client.SearchResponseModel) (*KnowledgeSearchResult, error) {
	total, items, responseErrors, err := decodeHa3ChatResponse(resp)
	if err != nil {
		return nil, err
	}
	if hasHa3ResponseErrors(responseErrors) {
		payload, _ := json.Marshal(responseErrors)
		return nil, fmt.Errorf("ha3 knowledge search errors: %s", payload)
	}
	hits := make([]KnowledgeHit, 0, len(items))
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		fields, _ := itemMap["fields"].(map[string]interface{})
		if fields == nil {
			fields = itemMap
		}
		hit := KnowledgeHit{
			KnowledgeID:   strings.TrimSpace(stringFromAny(fields[r.knowledgeConf.KnowledgeIDField])),
			KnowledgeType: strings.TrimSpace(stringFromAny(fields[r.knowledgeConf.KnowledgeTypeField])),
			Value:         strings.TrimSpace(stringFromAny(fields[r.knowledgeConf.ValueField])),
			Category:      strings.TrimSpace(stringFromAny(fields[r.knowledgeConf.CategoryField])),
			Score:         firstAny(itemMap["sortExprValues"], itemMap["score"]),
		}
		if hit.KnowledgeID == "" || hit.KnowledgeType == "" || hit.Value == "" || hit.Category == "" {
			continue
		}
		hits = append(hits, hit)
	}
	if len(items) > 0 && len(hits) == 0 {
		return nil, fmt.Errorf("ha3 knowledge response contains %d items but none has all configured result fields", len(items))
	}
	return &KnowledgeSearchResult{Total: total, Hits: hits}, nil
}
