package aishopping

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/log"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/service/searchsuggestion"
	"github.com/alibaba/pairec/v2/utils"
)

const (
	toolArgumentsLogLimit        = 2048
	fieldAwareSearchRetryMessage = "The previous search_goods call was invalid. Return exactly one tool call and no prose. Follow all required fields and array constraints, and omit optional prices when absent."
)

var optionalPriceNullLikePattern = regexp.MustCompile(
	`("(?:min_price|max_price)"\s*:\s*)(?:None|NaN|-?Infinity)\b`,
)

type turnState struct {
	indexMap     map[int]string
	itemToIndex  map[string]int
	nextIndex    int
	replyItemIDs []string
}

type chatRecall interface {
	Search(context.Context, recallsvc.SearchGoodsRequest) (*recallsvc.SearchGoodsResult, error)
}

type agentLoopResult struct {
	Reply               string
	IndexMap            map[int]string
	ReplyItemIDs        []string
	ReplyAlreadyEmitted bool
	MainReplyFallback   bool
	FinalSearchStatus   finalSearchStatus
}

type finalSearchStatus string

const (
	finalSearchNotAttempted finalSearchStatus = "not_attempted"
	finalSearchReady        finalSearchStatus = "ready"
	finalSearchFailed       finalSearchStatus = "failed"
	finalSearchInvalid      finalSearchStatus = "invalid"
)

type finalSearchSnapshot struct {
	Request          searchsuggestion.SearchIntent
	ReplyToolPayload string
	Total            int
	ProductIndexes   []int
	ProductSetHash   string
}

type timingMeta struct {
	requestId string
	uid       string
	sessionId string
	sceneId   string
	language  string
}

type toolDispatchResult struct {
	content      string
	isSearch     bool
	operator     string
	keywordCount int
	total        int
	hasError     bool
	search       *finalSearchSnapshot
}

func (r toolDispatchResult) shouldRetryWithFallback(tried bool) bool {
	return r.isSearch && !r.hasError && r.operator == "AND" && r.total == 0 && r.keywordCount > 1 && !tried
}

func runAgentLoop(ctx context.Context, model *aichat.Model, recall chatRecall, blob *SessionBlob, cfg *chatConfig, rankRuntime *fineRankRuntime, knowledge *knowledgeEvidence, writer *StreamWriter, meta timingMeta, onFinalSearch func(context.Context, *finalSearchSnapshot)) (*agentLoopResult, error) {
	state := &turnState{
		indexMap:    make(map[int]string),
		itemToIndex: make(map[string]int),
		nextIndex:   1,
	}
	finalStatus := finalSearchNotAttempted
	loopResult := func(reply string, emitted, fallback bool) *agentLoopResult {
		return &agentLoopResult{
			Reply:               reply,
			IndexMap:            state.indexMap,
			ReplyItemIDs:        append([]string(nil), state.replyItemIDs...),
			ReplyAlreadyEmitted: emitted,
			MainReplyFallback:   fallback,
			FinalSearchStatus:   finalStatus,
		}
	}
	if err := writer.EmitStep("analyze_requirement"); err != nil {
		return nil, err
	}
	readyToReply := false
	fallbackSearchTried := false
	fieldAwareSearch := cfg.fieldAware
	maxRounds := cfg.raw.ToolMaxRounds
	if fieldAwareSearch {
		maxRounds++ // Reserve one final round for Reply after Planner retries.
	}
	plannerAttempts := 0
	plannerRetry := false
	for round := 1; round <= maxRounds; round++ {
		if readyToReply {
			if err := writer.EmitStep("analyze_results"); err != nil {
				return nil, err
			}
		}
		prompt := cfg.plannerPrompt
		if readyToReply {
			prompt = cfg.replyPrompt
		}
		messages := maskHistoryMessages(blob.Messages)
		if fieldAwareSearch && !readyToReply && plannerRetry {
			messages = append(messages, aichat.Message{
				Role:    "system",
				Content: fieldAwareSearchRetryMessage,
			})
		}
		plannerMessages := messagesWithPrompt(messages, prompt)
		if !readyToReply {
			plannerMessages = messagesWithKnowledge(plannerMessages, cfg.raw.KnowledgePlannerInstruction, knowledge)
		}
		llmReq := &aichat.ChatCompletionRequest{
			Model:          "",
			Messages:       plannerMessages,
			Tools:          []aichat.Tool{aichat.SearchGoodsTool()},
			Stream:         true,
			EnableThinking: false,
		}
		if !readyToReply && fieldAwareSearch {
			plannerAttempts++
			temperature := 0.0
			parallelToolCalls := false
			llmReq.Tools = []aichat.Tool{aichat.FieldAwareSearchGoodsTool(knowledge.CandidateIDs())}
			llmReq.Temperature = &temperature
			llmReq.ParallelToolCalls = &parallelToolCalls
		}
		if readyToReply {
			llmReq.Tools = nil
		} else if !fieldAwareSearch && round == cfg.raw.ToolMaxRounds {
			llmReq.ToolChoice = "none"
		}
		streamer := newReplyStreamer(
			writer,
			state.indexMap,
			state.replyItemIDs,
			cfg.raw.DisplayItemCountMax,
			readyToReply && rankRuntime != nil,
		)
		var streamErr error
		llmPhase := "planner_llm"
		if readyToReply {
			llmPhase = "reply_llm"
		}
		llmStart := time.Now()
		result, err := model.Stream(ctx, llmReq, func(text string) error {
			if !readyToReply {
				return nil
			}
			if err := streamer.Feed(text); err != nil {
				streamErr = err
				return err
			}
			return nil
		})
		llmCost := utils.CostTime(llmStart)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			if fieldAwareSearch && !readyToReply && plannerAttempts < cfg.raw.ToolMaxRounds {
				plannerRetry = true
				log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=planner_retry\tround=%d\tattempt=%d\terr=%s",
					meta.requestId, meta.uid, meta.sessionId, round, plannerAttempts, compactLogError(err)))
				continue
			}
			if streamErr != nil {
				log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\tevent=model_stream_callback_error\tcost=%d\tstreamErr=%+v\terr=%+v",
					meta.requestId, meta.uid, meta.sessionId, llmPhase, round, llmCost, streamErr, err))
				return nil, streamErr
			}
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\tevent=model_stream_error\tcost=%d\terr=%+v",
				meta.requestId, meta.uid, meta.sessionId, llmPhase, round, llmCost, err))
			return loopResult(fallbackText(cfg.raw, cfg.language, "generic"), false, true), nil
		}
		log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\ttoolCalls=%d\tfinishReason=%s\tcontentBytes=%d\tcost=%d",
			meta.requestId, meta.uid, meta.sessionId, llmPhase, round, len(result.ToolCalls), result.FinishReason, len(result.Content), llmCost))
		if readyToReply && len(result.ToolCalls) > 0 {
			log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=reply_llm\tround=%d\tevent=unexpected_tool_calls_ignored\ttoolCalls=%d",
				meta.requestId, meta.uid, meta.sessionId, round, len(result.ToolCalls)))
			result.ToolCalls = nil
		}
		if fieldAwareSearch && !readyToReply {
			if err := normalizeFieldAwareToolCalls(result.ToolCalls, knowledge); err != nil {
				plannerRetry = true
				log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=planner_retry\tround=%d\tattempt=%d\terr=%s\targs=%s",
					meta.requestId, meta.uid, meta.sessionId, round, plannerAttempts, compactLogError(err), fieldAwareToolArguments(result.ToolCalls)))
				if plannerAttempts < cfg.raw.ToolMaxRounds {
					continue
				}
				return loopResult(fallbackText(cfg.raw, cfg.language, "generic"), false, true), nil
			}
			plannerRetry = false
			result.Content = ""
		}
		assistant := aichat.Message{Role: "assistant", Content: result.Content, ToolCalls: result.ToolCalls}
		blob.Messages = append(blob.Messages, assistant)
		if len(result.ToolCalls) == 0 {
			if result.Content == "" {
				return loopResult(fallbackText(cfg.raw, cfg.language, "empty_after_tools"), false, true), nil
			}
			if !readyToReply {
				return loopResult(result.Content, false, false), nil
			}
			if err := streamer.Finish(); err != nil {
				return nil, err
			}
			return loopResult(result.Content, true, false), nil
		}
		if err := writer.EmitStep("tool_call"); err != nil {
			return nil, err
		}
		readyToReply = true
		roundSearches := make([]*finalSearchSnapshot, 0, 1)
		roundSearchFailed := false
		for i, toolCall := range result.ToolCalls {
			log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=tool_call_args\tround=%d\ttoolIndex=%d\ttool=%s\targs=%s",
				meta.requestId, meta.uid, meta.sessionId, round, i, toolCall.Function.Name, compactJSONString(toolCall.Function.Arguments, toolArgumentsLogLimit)))
			toolResult := dispatchTool(ctx, recall, toolCall, state, cfg, rankRuntime, fieldAwareSearch, knowledge, onFinalSearch != nil, meta, round)
			if rankRuntime != nil {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
			}
			blob.Messages = append(blob.Messages, aichat.Message{
				Role:       "tool",
				ToolCallId: toolCall.ID,
				Content:    toolResult.content,
			})
			if toolResult.shouldRetryWithFallback(fallbackSearchTried) {
				fallbackSearchTried = true
				readyToReply = false
			}
			if toolResult.isSearch {
				if toolResult.hasError {
					roundSearchFailed = true
				} else if toolResult.search != nil {
					roundSearches = append(roundSearches, toolResult.search)
				}
			}
		}
		if err := writer.EmitStep("get_results"); err != nil {
			return nil, err
		}
		if readyToReply && onFinalSearch != nil {
			switch {
			case roundSearchFailed:
				finalStatus = finalSearchFailed
			case len(roundSearches) == 1 && len(result.ToolCalls) == 1:
				finalStatus = finalSearchReady
				log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=suggestion_snapshot\ttotal=%d\thits=%d\tproductSetHash=%s",
					meta.requestId, meta.uid, meta.sessionId, roundSearches[0].Total, len(roundSearches[0].ProductIndexes), roundSearches[0].ProductSetHash))
				if onFinalSearch != nil {
					onFinalSearch(ctx, roundSearches[0])
				}
			case len(roundSearches) == 0:
				finalStatus = finalSearchInvalid
			default:
				finalStatus = finalSearchInvalid
			}
		}
	}
	return loopResult(fallbackText(cfg.raw, cfg.language, "empty_after_tools"), false, true), nil
}

func dispatchTool(ctx context.Context, recall chatRecall, toolCall aichat.ToolCall, state *turnState, cfg *chatConfig, rankRuntime *fineRankRuntime, fieldAware bool, knowledge *knowledgeEvidence, captureFinalSearch bool, meta timingMeta, round int) toolDispatchResult {
	if toolCall.Function.Name != "search_goods" {
		return toolDispatchResult{content: `{"error":"unsupported tool"}`, hasError: true}
	}
	req, err := parseSearchGoodsRequest(toolCall.Function.Arguments, fieldAware, knowledge)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=tool_parse\tround=%d\terr=%v\targs=%s",
			meta.requestId, meta.uid, meta.sessionId, round, err, compactJSONString(toolCall.Function.Arguments, toolArgumentsLogLimit)))
		return toolDispatchResult{content: fmt.Sprintf(`{"error":%q}`, err.Error()), isSearch: true, hasError: true}
	}
	displayLimit := cfg.raw.DisplayItemCountMax
	fineRank := cfg.raw.FineRankConfig
	req.Limit = displayLimit
	if fineRank != nil {
		req.Limit = fineRank.CandidateCount
	}
	req.MultiFieldFallback = fieldAware
	dispatchResult := toolDispatchResult{
		isSearch:     true,
		operator:     effectiveOperator(req.Operator),
		keywordCount: countSearchKeywords(req.Keywords),
	}
	recallStart := time.Now()
	result, err := recall.Search(ctx, req)
	recallCost := utils.CostTime(recallStart)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=opensearch_recall\tround=%d\tkeywords=%s\toperator=%s\tlimit=%d\tcost=%d\terr=%v",
			meta.requestId, meta.uid, meta.sessionId, round, compactJSON(req.Keywords), dispatchResult.operator, req.Limit, recallCost, err))
		dispatchResult.content = fmt.Sprintf(`{"error":%q}`, err.Error())
		dispatchResult.hasError = true
		return dispatchResult
	}
	if result == nil {
		result = &recallsvc.SearchGoodsResult{}
	}
	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=opensearch_recall\tround=%d\tkeywords=%s\toperator=%s\tlimit=%d\ttotal=%d\thits=%d\tfallbackUsed=%t\titemIds=%s\trouteErrors=%s\tcost=%d",
		meta.requestId, meta.uid, meta.sessionId, round, compactJSON(req.Keywords), dispatchResult.operator, req.Limit, result.Total, len(result.Hits), result.FallbackUsed, compactJSON(searchResultItemIds(result)), compactJSON(result.RouteErrors), recallCost))
	dispatchResult.total = result.Total
	if fineRank != nil {
		if len(result.Hits) > 1 {
			rankedHits, rankErr := fineRankGoods(ctx, result.Hits, rankRuntime, fineRank)
			if rankErr != nil {
				log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=fine_rank\tround=%d\terr=%s",
					meta.requestId, meta.uid, meta.sessionId, round, compactLogError(rankErr)))
			} else {
				result.Hits = rankedHits
			}
		}
		if len(result.Hits) > displayLimit {
			result.Hits = result.Hits[:displayLimit]
		}
	}
	modelResult := annotateSearchResult(result, state, fineRank == nil)
	payload, err := json.Marshal(modelResult)
	if err != nil {
		dispatchResult.content = fmt.Sprintf(`{"error":%q}`, err.Error())
		dispatchResult.hasError = true
		return dispatchResult
	}
	dispatchResult.content = string(payload)
	if fineRank != nil {
		state.replyItemIDs = searchResultItemIds(result)
	}
	if !captureFinalSearch {
		return dispatchResult
	}
	dispatchResult.search = &finalSearchSnapshot{
		Request:          sanitizeSearchIntent(req),
		ReplyToolPayload: dispatchResult.content,
		Total:            result.Total,
		ProductIndexes:   indexesFromAnnotatedHits(modelResult),
		ProductSetHash:   hashOrderedItemIDs(result),
	}
	return dispatchResult
}

func sanitizeSearchIntent(req recallsvc.SearchGoodsRequest) searchsuggestion.SearchIntent {
	return searchsuggestion.SearchIntent{
		Keywords:                      append([]string(nil), req.Keywords...),
		ProductTypeKeywords:           append([]string(nil), req.ProductTypeKeywords...),
		AttributeKeywords:             append([]string(nil), req.AttributeKeywords...),
		Operator:                      req.Operator,
		ExcludeKeywords:               append([]string(nil), req.ExcludeKeywords...),
		MinPrice:                      cloneFloat(req.MinPrice),
		MaxPrice:                      cloneFloat(req.MaxPrice),
		SelectedKnowledgeCandidateIDs: append([]string(nil), req.KnowledgeCandidateIDs...),
	}
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func indexesFromAnnotatedHits(result map[string]interface{}) []int {
	hits, _ := result["hits"].([]map[string]interface{})
	indexes := make([]int, 0, len(hits))
	for _, hit := range hits {
		index, _ := hit["index"].(int)
		indexes = append(indexes, index)
	}
	return indexes
}

func hashOrderedItemIDs(result *recallsvc.SearchGoodsResult) string {
	hash := sha256.New()
	for _, hit := range result.Hits {
		_, _ = hash.Write([]byte(hit.ItemId))
		_, _ = hash.Write([]byte{0})
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func normalizeFieldAwareToolCalls(toolCalls []aichat.ToolCall, knowledge *knowledgeEvidence) error {
	if len(toolCalls) != 1 || toolCalls[0].Function.Name != "search_goods" {
		return fmt.Errorf("exactly one search_goods call is required")
	}
	if strings.TrimSpace(toolCalls[0].ID) == "" {
		return fmt.Errorf("search_goods tool call id is required")
	}
	if toolCalls[0].Type != "function" {
		return fmt.Errorf("search_goods tool call type must be function")
	}
	req, err := parseSearchGoodsRequest(toolCalls[0].Function.Arguments, true, knowledge)
	if err != nil {
		return err
	}
	arguments, err := marshalSearchGoodsRequest(req, knowledge)
	if err != nil {
		return err
	}
	toolCalls[0].Function.Arguments = string(arguments)
	return nil
}

func marshalSearchGoodsRequest(req recallsvc.SearchGoodsRequest, knowledge *knowledgeEvidence) ([]byte, error) {
	payload, err := json.Marshal(req)
	if err != nil || knowledge == nil || len(req.KnowledgeCandidateIDs) > 0 {
		return payload, err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		return nil, err
	}
	fields["knowledge_candidate_ids"] = json.RawMessage(`[]`)
	return json.Marshal(fields)
}

func fieldAwareToolArguments(toolCalls []aichat.ToolCall) string {
	if len(toolCalls) != 1 {
		return ""
	}
	return compactJSONString(toolCalls[0].Function.Arguments, toolArgumentsLogLimit)
}

func parseSearchGoodsRequest(arguments string, fieldAware bool, knowledge *knowledgeEvidence) (recallsvc.SearchGoodsRequest, error) {
	var req recallsvc.SearchGoodsRequest
	if !fieldAware {
		return req, json.Unmarshal([]byte(arguments), &req)
	}
	arguments = optionalPriceNullLikePattern.ReplaceAllString(arguments, `${1}null`)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(arguments), &fields); err != nil {
		return req, err
	}
	if raw, ok := fields["keywords"]; ok {
		var keyword string
		if json.Unmarshal(raw, &keyword) == nil {
			normalizedKeyword, err := json.Marshal([]string{keyword})
			if err != nil {
				return req, err
			}
			fields["keywords"] = normalizedKeyword
			payload, err := json.Marshal(fields)
			if err != nil {
				return req, err
			}
			arguments = string(payload)
		}
	}
	decoder := json.NewDecoder(strings.NewReader(arguments))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		return req, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values are not allowed")
		}
		return req, err
	}
	if err := knowledge.Apply(&req); err != nil {
		return req, err
	}
	return recallsvc.NormalizeFieldAwareSearchGoodsRequest(req)
}

func searchResultItemIds(result *recallsvc.SearchGoodsResult) []string {
	itemIds := make([]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		itemIds = append(itemIds, hit.ItemId)
	}
	return itemIds
}

func effectiveOperator(operator string) string {
	if strings.EqualFold(operator, "OR") {
		return "OR"
	}
	return "AND"
}

func countSearchKeywords(keywords []string) int {
	count := 0
	for _, keyword := range keywords {
		if strings.TrimSpace(keyword) != "" {
			count++
		}
	}
	return count
}

func compactJSON(value interface{}) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(payload)
}

func compactJSONString(value string, limit int) string {
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err == nil {
		return truncateLogValue(compactJSON(parsed), limit)
	}
	return truncateLogValue(strings.Join(strings.Fields(value), " "), limit)
}

func truncateLogValue(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "...(truncated)"
}

func compactLogError(err error) string {
	if err == nil {
		return ""
	}
	return truncateLogValue(strings.Join(strings.Fields(err.Error()), " "), toolArgumentsLogLimit)
}

func annotateSearchResult(result *recallsvc.SearchGoodsResult, state *turnState, includeScore bool) map[string]interface{} {
	modelHits := make([]map[string]interface{}, 0, len(result.Hits))
	for _, hit := range result.Hits {
		index := state.itemToIndex[hit.ItemId]
		if index == 0 {
			index = state.nextIndex
			state.nextIndex++
			state.itemToIndex[hit.ItemId] = index
			state.indexMap[index] = hit.ItemId
		}
		modelHit := map[string]interface{}{
			"index": index,
		}
		if hit.Title != "" {
			modelHit["title"] = hit.Title
		}
		if hit.Content != "" {
			modelHit["content"] = hit.Content
		}
		if includeScore && hit.Score != nil {
			modelHit["score"] = hit.Score
		}
		raw := make(map[string]interface{}, len(hit.Properties))
		for key, value := range hit.Properties {
			if key == "item_id" || key == "id" {
				continue
			}
			raw[key] = value
		}
		if len(raw) > 0 {
			modelHit["raw"] = raw
		}
		modelHits = append(modelHits, modelHit)
	}
	return map[string]interface{}{"total": result.Total, "hits": modelHits}
}
