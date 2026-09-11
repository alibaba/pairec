package aishopping

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/log"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/service/searchsuggestion"
	"github.com/alibaba/pairec/v2/utils"
)

const toolArgumentsLogLimit = 2048

type turnState struct {
	indexMap     map[int]string
	itemToIndex  map[string]int
	nextIndex    int
	replyItemIDs []string
}

type chatRecall interface {
	Search(context.Context, recallsvc.SearchGoodsRequest) (*recallsvc.SearchGoodsResult, error)
	SearchGoodsTool() aichat.Tool
	ValidateSearchGoodsRequest(recallsvc.SearchGoodsRequest) error
	ValidateToolParamSources(aichat.SearchGoodsParams, *aichat.SearchGoodsParams, recallsvc.ToolParamEvidence) error
	ToolParamsConfigID() string
}

type agentLoopResult struct {
	Reply               string
	IndexMap            map[int]string
	ReplyItemIDs        []string
	ReplyAlreadyEmitted bool
	MainReplyFallback   bool
	FinalSearchStatus   finalSearchStatus
	LastSearch          *searchsuggestion.SearchIntent
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
	content  string
	isSearch bool
	operator string
	empty    bool
	hasError bool
	search   *finalSearchSnapshot
	intent   *searchsuggestion.SearchIntent
}

func runAgentLoop(ctx context.Context, model *aichat.Model, recall chatRecall, messages []aichat.Message, cfg *chatConfig, rankRuntime *fineRankRuntime, knowledge *knowledgeEvidence, previous *searchsuggestion.SearchIntent, writer *StreamWriter, meta timingMeta, onFinalSearch func(context.Context, *finalSearchSnapshot)) (*agentLoopResult, error) {
	state := &turnState{
		indexMap:    make(map[int]string),
		itemToIndex: make(map[string]int),
		nextIndex:   1,
	}
	finalStatus := finalSearchNotAttempted
	var lastSearch *searchsuggestion.SearchIntent
	loopResult := func(reply string, emitted, fallback bool) *agentLoopResult {
		return &agentLoopResult{
			Reply:               reply,
			IndexMap:            state.indexMap,
			ReplyItemIDs:        append([]string(nil), state.replyItemIDs...),
			ReplyAlreadyEmitted: emitted,
			MainReplyFallback:   fallback,
			FinalSearchStatus:   finalStatus,
			LastSearch:          lastSearch,
		}
	}
	if err := writer.EmitStep("analyze_requirement"); err != nil {
		return nil, err
	}
	readyToReply := false
	fieldAwareSearch := cfg.fieldAware
	maxRounds := cfg.raw.ToolMaxRounds + 1 // Reserve one final round for Reply after Planner retries.
	plannerAttempts := 0
	plannerRetry := ""
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
		plannerMessages := messagesWithPrompt(messages, prompt)
		if !readyToReply && plannerRetry != "" {
			plannerMessages = append(plannerMessages, aichat.Message{
				Role:    "system",
				Content: compactJSON(map[string]string{"planner_validation_error": plannerRetry}),
			})
		}
		if !readyToReply {
			plannerMessages = messagesWithKnowledge(plannerMessages, cfg.raw.KnowledgePlannerInstruction, knowledge)
		}
		llmReq := &aichat.ChatCompletionRequest{
			Model:          "",
			Messages:       plannerMessages,
			Stream:         true,
			EnableThinking: false,
		}
		if !readyToReply {
			plannerAttempts++
			parallelToolCalls := false
			llmReq.Tools = []aichat.Tool{recall.SearchGoodsTool()}
			if fieldAwareSearch {
				temperature := 0.0
				llmReq.Temperature = &temperature
			}
			llmReq.Tools[0].Function.Strict = cfg.raw.PlannerToolStrict
			llmReq.ToolChoice = cfg.raw.PlannerToolChoice
			llmReq.ParallelToolCalls = &parallelToolCalls
		}
		streamer := newReplyStreamer(
			writer,
			state.indexMap,
			state.replyItemIDs,
			cfg.raw.DisplayItemCountMax,
			readyToReply && rankRuntime != nil,
		)
		var streamErr error
		replyStarted := false
		llmPhase := "planner_llm"
		if readyToReply {
			llmPhase = "reply_llm"
		}
		var onDelta aichat.DeltaHandler
		if readyToReply {
			onDelta = func(text string) error {
				replyStarted = true
				streamErr = streamer.Feed(text)
				return streamErr
			}
		}
		llmStart := time.Now()
		result, err := model.Stream(ctx, llmReq, onDelta)
		llmCost := utils.CostTime(llmStart)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}
			if streamErr != nil {
				log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\tevent=model_stream_callback_error\tcost=%d\tstreamErr=%+v\terr=%+v",
					meta.requestId, meta.uid, meta.sessionId, llmPhase, round, llmCost, streamErr, err))
				return nil, streamErr
			}
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\tevent=model_stream_error\tcost=%d\terr=%+v",
				meta.requestId, meta.uid, meta.sessionId, llmPhase, round, llmCost, err))
			if replyStarted {
				return nil, err
			}
			return loopResult(fallbackText(cfg.raw, cfg.language, "generic"), false, true), nil
		}
		log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\ttoolCalls=%d\tfinishReason=%s\tcontentBytes=%d\tcost=%d",
			meta.requestId, meta.uid, meta.sessionId, llmPhase, round, len(result.ToolCalls), result.FinishReason, len(result.Content), llmCost))
		if readyToReply && len(result.ToolCalls) > 0 {
			log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=reply_llm\tround=%d\tevent=unexpected_tool_calls_ignored\ttoolCalls=%d",
				meta.requestId, meta.uid, meta.sessionId, round, len(result.ToolCalls)))
			result.ToolCalls = nil
		}
		if !readyToReply {
			if err := normalizePlannerResponse(result, cfg, recall, knowledge, previous, meta.requestId); err != nil {
				plannerRetry = truncateLogValue(err.Error(), 256)
				log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=planner_retry\tround=%d\tattempt=%d\terr=%s\targs=%s",
					meta.requestId, meta.uid, meta.sessionId, round, plannerAttempts, compactLogError(err), fieldAwareToolArguments(result.ToolCalls)))
				if plannerAttempts < cfg.raw.ToolMaxRounds {
					continue
				}
				return loopResult(fallbackText(cfg.raw, cfg.language, "generic"), false, true), nil
			}
			plannerRetry = ""
			if len(result.ToolCalls) > 0 {
				result.Content = ""
			}
		}
		assistant := aichat.Message{Role: "assistant", Content: result.Content, ToolCalls: result.ToolCalls}
		messages = append(messages, assistant)
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
		noResults := false
		lastSearch = nil
		for i, toolCall := range result.ToolCalls {
			log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=tool_call_args\tround=%d\ttoolIndex=%d\ttool=%s\targs=%s",
				meta.requestId, meta.uid, meta.sessionId, round, i, toolCall.Function.Name, compactJSONString(toolCall.Function.Arguments, toolArgumentsLogLimit)))
			toolResult := dispatchTool(ctx, recall, toolCall, state, cfg, rankRuntime, fieldAwareSearch, onFinalSearch != nil, meta, round)
			if len(result.ToolCalls) == 1 {
				lastSearch = toolResult.intent
				noResults = toolResult.isSearch && !toolResult.hasError && toolResult.empty
			}
			if rankRuntime != nil {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
			}
			messages = append(messages, aichat.Message{
				Role:       "tool",
				ToolCallId: toolCall.ID,
				Content:    toolResult.content,
			})
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
		if roundSearchFailed && len(result.ToolCalls) == 1 {
			return loopResult(fallbackText(cfg.raw, cfg.language, "generic"), false, true), nil
		}
		if noResults {
			return loopResult(fallbackText(cfg.raw, cfg.language, "empty_after_tools"), false, false), nil
		}
	}
	return loopResult(fallbackText(cfg.raw, cfg.language, "empty_after_tools"), false, true), nil
}

func dispatchTool(ctx context.Context, recall chatRecall, toolCall aichat.ToolCall, state *turnState, cfg *chatConfig, rankRuntime *fineRankRuntime, fieldAware bool, captureFinalSearch bool, meta timingMeta, round int) toolDispatchResult {
	if toolCall.Function.Name != "search_goods" {
		return toolDispatchResult{content: `{"error":"unsupported tool"}`, hasError: true}
	}
	req, err := parseSearchGoodsRequest(toolCall.Function.Arguments, fieldAware)
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
	req.FieldAware = fieldAware
	intent := sanitizeSearchIntent(req)
	dispatchResult := toolDispatchResult{
		isSearch: true,
		operator: "AND",
		intent:   &intent,
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
		dispatchResult.content = `{"error":"search returned no response"}`
		dispatchResult.hasError = true
		return dispatchResult
	}
	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=opensearch_recall\tround=%d\tkeywords=%s\toperator=%s\tlimit=%d\ttotal=%d\thits=%d\titemIds=%s\tcost=%d",
		meta.requestId, meta.uid, meta.sessionId, round, compactJSON(req.Keywords), dispatchResult.operator, req.Limit, result.Total, len(result.Hits), compactJSON(searchResultItemIds(result)), recallCost))
	dispatchResult.empty = len(result.Hits) == 0
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
	if len(result.DroppedPreferredKeywords) > 0 {
		modelResult["dropped_preferred_keywords"] = result.DroppedPreferredKeywords
	}
	if dispatchResult.empty {
		modelResult["current_search"] = intent
	}
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
		Keywords:          append([]string(nil), req.Keywords...),
		PreferredKeywords: append([]string(nil), req.PreferredKeywords...),
		ExcludeKeywords:   append([]string(nil), req.ExcludeKeywords...),
		MinPrice:          cloneFloat(req.MinPrice),
		MaxPrice:          cloneFloat(req.MaxPrice),
		ToolParams:        cloneToolParams(req.ToolParams),
	}
}

func cloneToolParams(values map[string]json.RawMessage) map[string]json.RawMessage {
	if values == nil {
		return nil
	}
	result := make(map[string]json.RawMessage, len(values))
	for name, value := range values {
		result[name] = append(json.RawMessage(nil), value...)
	}
	return result
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

func normalizePlannerResponse(result *aichat.StreamResult, cfg *chatConfig, recall chatRecall, knowledge *knowledgeEvidence, previous *searchsuggestion.SearchIntent, requestID string) error {
	if result.FinishReason != "stop" && result.FinishReason != "tool_calls" {
		return fmt.Errorf("planner response did not finish normally: %q", result.FinishReason)
	}
	toolCalls := result.ToolCalls
	if len(toolCalls) == 0 {
		if cfg.raw.PlannerToolChoice == "required" {
			return fmt.Errorf("search_goods is required by tool_choice")
		}
		if result.FinishReason != "stop" || strings.TrimSpace(result.Content) == "" {
			return fmt.Errorf("a complete nonempty direct reply is required when no tool is called")
		}
		return nil
	}
	if len(toolCalls) != 1 || toolCalls[0].Function.Name != "search_goods" {
		return fmt.Errorf("exactly one search_goods call is required")
	}
	if strings.TrimSpace(toolCalls[0].ID) == "" {
		return fmt.Errorf("search_goods tool call id is required")
	}
	if toolCalls[0].Type != "function" {
		return fmt.Errorf("search_goods tool call type must be function")
	}
	req, err := parseSearchGoodsRequest(toolCalls[0].Function.Arguments, cfg.fieldAware)
	if err != nil {
		return err
	}
	if cfg.fieldAware {
		if filler, ok := recall.(interface {
			FillMissingToolParams(aichat.SearchGoodsParams, *aichat.SearchGoodsParams, recallsvc.ToolParamEvidence, string) aichat.SearchGoodsParams
		}); ok {
			req.SearchGoodsParams = filler.FillMissingToolParams(req.SearchGoodsParams, previous, knowledge, requestID)
		}
	}
	if err := recall.ValidateSearchGoodsRequest(req); err != nil {
		return err
	}
	if cfg.fieldAware {
		if err := recall.ValidateToolParamSources(req.SearchGoodsParams, previous, knowledge); err != nil {
			return err
		}
	}
	arguments, err := marshalSearchGoodsRequest(req)
	if err != nil {
		return err
	}
	toolCalls[0].Function.Arguments = string(arguments)
	return nil
}

func marshalSearchGoodsRequest(req recallsvc.SearchGoodsRequest) ([]byte, error) {
	return json.Marshal(req)
}

func fieldAwareToolArguments(toolCalls []aichat.ToolCall) string {
	if len(toolCalls) != 1 {
		return ""
	}
	return compactJSONString(toolCalls[0].Function.Arguments, toolArgumentsLogLimit)
}

func parseSearchGoodsRequest(arguments string, fieldAware bool) (recallsvc.SearchGoodsRequest, error) {
	var req recallsvc.SearchGoodsRequest
	if !fieldAware {
		if err := json.Unmarshal([]byte(arguments), &req); err != nil {
			return req, err
		}
		return recallsvc.NormalizeFieldAwareSearchGoodsRequest(req)
	}
	arguments = normalizeOptionalPriceLiterals(arguments)
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(arguments), &fields); err != nil {
		return req, err
	}
	for _, field := range []string{"min_price", "max_price"} {
		if raw, ok := fields[field]; ok {
			var price *float64
			if err := json.Unmarshal(raw, &price); err != nil {
				fields[field] = json.RawMessage("null")
				log.Info(fmt.Sprintf("module=AIShoppingChat\tevent=optional_price_ignored\tfield=%s\treason=invalid_value", field))
			}
		}
	}
	if raw, ok := fields["keywords"]; ok {
		var keyword string
		if json.Unmarshal(raw, &keyword) == nil {
			normalizedKeyword, err := json.Marshal([]string{keyword})
			if err != nil {
				return req, err
			}
			fields["keywords"] = normalizedKeyword
		}
	}
	payload, err := json.Marshal(fields)
	if err != nil {
		return req, err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
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
	return recallsvc.NormalizeFieldAwareSearchGoodsRequest(req)
}

func searchResultItemIds(result *recallsvc.SearchGoodsResult) []string {
	itemIds := make([]string, 0, len(result.Hits))
	for _, hit := range result.Hits {
		itemIds = append(itemIds, hit.ItemId)
	}
	return itemIds
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
		if len(hit.ConstraintEvidence) > 0 {
			modelHit["constraint_evidence"] = hit.ConstraintEvidence
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
