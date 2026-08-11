package web

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/searchsuggestion"
	"github.com/alibaba/pairec/v2/service/shoppingknowledge"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"golang.org/x/text/unicode/norm"
)

const (
	searchSuggestionBodyLimit = 16 << 10
	searchSuggestionQueryMax  = 512
	searchSuggestionTimeout   = 5 * time.Second
)

type SearchSuggestionParam struct {
	SceneId  string `json:"scene_id"`
	Language string `json:"language"`
	Query    string `json:"query"`
}

type SearchSuggestionResponse struct {
	Response
	Suggestions []string `json:"suggestions,omitempty"`
}

type SearchSuggestionController struct {
	Controller
	param SearchSuggestionParam
}

func (c *SearchSuggestionController) Process(w http.ResponseWriter, r *http.Request) {
	c.Start = time.Now()
	c.RequestId = utils.UUID()
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if r.Method != http.MethodPost {
		c.writeResponse(w, ERROR_PARAMETER_CODE, "method must be POST", nil)
		return
	}
	body, err := readLimitedSuggestionBody(r, searchSuggestionBodyLimit)
	if err != nil {
		c.writeResponse(w, ERROR_PARAMETER_CODE, "invalid request body", nil)
		return
	}
	c.RequestBody = body
	if err := decodeSuggestionParam(body, &c.param); err != nil {
		c.writeResponse(w, ERROR_PARAMETER_CODE, "invalid request body", nil)
		return
	}
	if err := normalizeSuggestionParam(&c.param); err != nil {
		c.writeResponse(w, ERROR_PARAMETER_CODE, err.Error(), nil)
		return
	}
	log.Info(fmt.Sprintf("requestId=%s\tmodule=SearchSuggestion\tevent=begin\tscene=%s\tlanguage=%s\tqueryRunes=%d\tqueryHash=%s",
		c.RequestId, c.param.SceneId, c.param.Language, utf8.RuneCountInString(c.param.Query), suggestionQueryHash(c.param.Query)))

	ctx, cancel := context.WithTimeout(r.Context(), searchSuggestionTimeout)
	defer cancel()
	runtimeConfig, knowledgeRecall, resolveErr := searchsuggestion.ResolveStandalone(recconf.Config, c.param.SceneId, c.param.Language)
	if resolveErr != nil {
		c.writeSuggestionError(w, resolveErr)
		return
	}
	knowledgeResult, err := knowledgeRecall.SearchKnowledge(ctx, c.param.Query)
	if err != nil {
		c.writeSuggestionError(w, searchsuggestion.NewError(searchsuggestion.CodeKnowledgeFailed, true, err))
		return
	}
	evidence := shoppingknowledge.NewEvidence(knowledgeResult)
	if evidence == nil {
		c.writeSuggestionError(w, searchsuggestion.NewError(searchsuggestion.CodeKnowledgeEmpty, false, nil))
		return
	}
	input, inputErr := searchsuggestion.BuildStandaloneInput(c.param.Language, c.param.Query, evidence.SuggestionKnowledge(nil))
	if inputErr != nil {
		c.writeSuggestionError(w, inputErr)
		return
	}
	outcome := searchsuggestion.Generate(ctx, runtimeConfig, input)
	if outcome.Err != nil {
		c.writeSuggestionError(w, outcome.Err)
		return
	}
	c.writeResponse(w, SUCCESS_CODE, CODE_MAPS[SUCCESS_CODE], outcome.Suggestions)
}

func (c *SearchSuggestionController) writeSuggestionError(w http.ResponseWriter, suggestionErr *searchsuggestion.Error) {
	log.Warning(fmt.Sprintf("requestId=%s\tmodule=SearchSuggestion\tevent=end\tstatus=error\tcode=%s\tretryable=%t\terr=%v\tcost=%d",
		c.RequestId, suggestionErr.Code, suggestionErr.Retryable, suggestionErr.Cause, time.Since(c.Start).Milliseconds()))
	c.writeResponse(w, SERVER_ERROR_CODE, suggestionErr.PublicMessage(c.param.Language), nil)
}

func (c *SearchSuggestionController) writeResponse(w http.ResponseWriter, code int, message string, suggestions []string) {
	response := SearchSuggestionResponse{
		Response: Response{Code: code, Message: message, RequestId: c.RequestId},
	}
	if code == SUCCESS_CODE {
		response.Suggestions = suggestions
	}
	payload, _ := json.Marshal(response)
	_, _ = w.Write(payload)
	c.End = time.Now()
	if code == SUCCESS_CODE {
		log.Info(fmt.Sprintf("requestId=%s\tmodule=SearchSuggestion\tevent=end\tstatus=ok\tcount=%d\tcost=%d",
			c.RequestId, len(suggestions), c.cost()))
	}
}

func decodeSuggestionParam(body []byte, param *SearchSuggestionParam) error {
	if len(body) == 0 {
		return fmt.Errorf("request body is empty")
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(param); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values are not allowed")
		}
		return err
	}
	return nil
}

func normalizeSuggestionParam(param *SearchSuggestionParam) error {
	param.SceneId = strings.TrimSpace(norm.NFKC.String(param.SceneId))
	if param.SceneId == "" {
		param.SceneId = "search_suggestion"
	}
	param.Language = strings.TrimSpace(norm.NFKC.String(param.Language))
	if param.Language == "" {
		param.Language = "zh"
	}
	param.Query = strings.TrimSpace(norm.NFKC.String(param.Query))
	if param.Query == "" {
		return fmt.Errorf("query is empty")
	}
	if utf8.RuneCountInString(param.Query) > searchSuggestionQueryMax {
		return fmt.Errorf("query exceeds %d characters", searchSuggestionQueryMax)
	}
	return nil
}

func readLimitedSuggestionBody(r *http.Request, limit int64) ([]byte, error) {
	var reader io.Reader = r.Body
	var closer io.Closer
	switch r.Header.Get("Content-Encoding") {
	case "":
	case "gzip":
		decoded, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		reader, closer = decoded, decoded
	case "lz4":
		reader = lz4.NewReader(r.Body)
	case "zstd":
		decoded, err := zstd.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		reader, closer = decoded, decoded.IOReadCloser()
	default:
		return nil, fmt.Errorf("unsupported content encoding")
	}
	if closer != nil {
		defer closer.Close()
	}
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("request body too large")
	}
	return body, nil
}

func suggestionQueryHash(query string) string {
	sum := sha256.Sum256([]byte(query))
	return fmt.Sprintf("%x", sum[:6])
}
