package aishopping

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/log"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/utils"
)

type turnState struct {
	indexMap    map[int]string
	itemToIndex map[string]int
	nextIndex   int
}

type chatRecall interface {
	Search(context.Context, recallsvc.SearchGoodsRequest) (*recallsvc.SearchGoodsResult, error)
}

type agentLoopResult struct {
	Reply               string
	IndexMap            map[int]string
	ReplyAlreadyEmitted bool
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
}

func (r toolDispatchResult) shouldRetryWithFallback(tried bool) bool {
	return r.isSearch && !r.hasError && r.operator == "AND" && r.total == 0 && r.keywordCount > 1 && !tried
}

func runAgentLoop(ctx context.Context, model *aichat.Model, recall chatRecall, blob *SessionBlob, cfg *chatConfig, writer *StreamWriter, meta timingMeta) (*agentLoopResult, error) {
	state := &turnState{
		indexMap:    make(map[int]string),
		itemToIndex: make(map[string]int),
		nextIndex:   1,
	}
	if err := writer.EmitStep("analyze_requirement"); err != nil {
		return nil, err
	}
	readyToReply := false
	fallbackSearchTried := false
	for round := 1; round <= cfg.raw.ToolMaxRounds; round++ {
		if readyToReply {
			if err := writer.EmitStep("analyze_results"); err != nil {
				return nil, err
			}
		}
		prompt := cfg.plannerPrompt
		if readyToReply {
			prompt = cfg.replyPrompt
		}
		llmReq := &aichat.ChatCompletionRequest{
			Model:          "",
			Messages:       messagesWithPrompt(maskHistoryMessages(blob.Messages), prompt),
			Tools:          []aichat.Tool{aichat.SearchGoodsTool()},
			Stream:         true,
			EnableThinking: false,
		}
		if readyToReply || round == cfg.raw.ToolMaxRounds {
			llmReq.ToolChoice = "none"
		}
		streamer := newReplyStreamer(writer, state.indexMap, cfg.raw.DisplayItemCountMax)
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
			if streamErr != nil {
				log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\tevent=model_stream_callback_error\tcost=%d\tstreamErr=%+v\terr=%+v",
					meta.requestId, meta.uid, meta.sessionId, llmPhase, round, llmCost, streamErr, err))
				return nil, streamErr
			}
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\tevent=model_stream_error\tcost=%d\terr=%+v",
				meta.requestId, meta.uid, meta.sessionId, llmPhase, round, llmCost, err))
			return &agentLoopResult{
				Reply:    fallbackText(cfg.raw, cfg.language, "generic"),
				IndexMap: state.indexMap,
			}, nil
		}
		log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=%s\tround=%d\ttoolCalls=%d\tfinishReason=%s\tcontentBytes=%d\tcost=%d",
			meta.requestId, meta.uid, meta.sessionId, llmPhase, round, len(result.ToolCalls), result.FinishReason, len(result.Content), llmCost))
		assistant := aichat.Message{Role: "assistant", Content: result.Content, ToolCalls: result.ToolCalls}
		blob.Messages = append(blob.Messages, assistant)
		if len(result.ToolCalls) == 0 {
			if result.Content == "" {
				return &agentLoopResult{
					Reply:    fallbackText(cfg.raw, cfg.language, "empty_after_tools"),
					IndexMap: state.indexMap,
				}, nil
			}
			if !readyToReply {
				return &agentLoopResult{
					Reply:    result.Content,
					IndexMap: state.indexMap,
				}, nil
			}
			if err := streamer.Finish(); err != nil {
				return nil, err
			}
			return &agentLoopResult{
				Reply:               result.Content,
				IndexMap:            state.indexMap,
				ReplyAlreadyEmitted: true,
			}, nil
		}
		if err := writer.EmitStep("tool_call"); err != nil {
			return nil, err
		}
		readyToReply = true
		for i, toolCall := range result.ToolCalls {
			log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=tool_call_args\tround=%d\ttoolIndex=%d\ttool=%s\targs=%s",
				meta.requestId, meta.uid, meta.sessionId, round, i, toolCall.Function.Name, compactJSONString(toolCall.Function.Arguments, 512)))
			toolResult := dispatchTool(ctx, recall, toolCall, state, cfg.raw.DisplayItemCountMax, meta, round)
			blob.Messages = append(blob.Messages, aichat.Message{
				Role:       "tool",
				ToolCallId: toolCall.ID,
				Content:    toolResult.content,
			})
			if toolResult.shouldRetryWithFallback(fallbackSearchTried) {
				fallbackSearchTried = true
				readyToReply = false
			}
		}
		if err := writer.EmitStep("get_results"); err != nil {
			return nil, err
		}
	}
	return &agentLoopResult{
		Reply:    fallbackText(cfg.raw, cfg.language, "empty_after_tools"),
		IndexMap: state.indexMap,
	}, nil
}

func dispatchTool(ctx context.Context, recall chatRecall, toolCall aichat.ToolCall, state *turnState, limit int, meta timingMeta, round int) toolDispatchResult {
	if toolCall.Function.Name != "search_goods" {
		return toolDispatchResult{content: `{"error":"unsupported tool"}`, hasError: true}
	}
	var req recallsvc.SearchGoodsRequest
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &req); err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=tool_parse\tround=%d\terr=%v\targs=%s",
			meta.requestId, meta.uid, meta.sessionId, round, err, compactJSONString(toolCall.Function.Arguments, 512)))
		return toolDispatchResult{content: fmt.Sprintf(`{"error":%q}`, err.Error()), isSearch: true, hasError: true}
	}
	req.Limit = limit
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
	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=opensearch_recall\tround=%d\tkeywords=%s\toperator=%s\tlimit=%d\ttotal=%d\thits=%d\tcost=%d",
		meta.requestId, meta.uid, meta.sessionId, round, compactJSON(req.Keywords), dispatchResult.operator, req.Limit, result.Total, len(result.Hits), recallCost))
	dispatchResult.total = result.Total
	modelResult := annotateSearchResult(result, state)
	payload, err := json.Marshal(modelResult)
	if err != nil {
		dispatchResult.content = fmt.Sprintf(`{"error":%q}`, err.Error())
		dispatchResult.hasError = true
		return dispatchResult
	}
	dispatchResult.content = string(payload)
	return dispatchResult
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

func annotateSearchResult(result *recallsvc.SearchGoodsResult, state *turnState) map[string]interface{} {
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
		if hit.Score != nil {
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
