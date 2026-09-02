package aishopping

import (
	"context"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/log"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
	"github.com/alibaba/pairec/v2/service/searchsuggestion"
	"github.com/alibaba/pairec/v2/utils"
)

type ChatSearchOrchestrator struct{}

func NewChatSearchOrchestrator() *ChatSearchOrchestrator {
	return &ChatSearchOrchestrator{}
}

func (o *ChatSearchOrchestrator) Run(ctx context.Context, req *Request, writer *StreamWriter) error {
	totalStart := time.Now()
	status := "error"
	defer func() {
		log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=total\tscene=%s\tlanguage=%s\tstatus=%s\tcost=%d",
			req.RequestId, req.Uid, req.SessionId, req.SceneId, req.Language, status, utils.CostTime(totalStart)))
	}()
	cfg, err := resolveConfig(req.Config, req.SceneId, req.Language)
	if err != nil {
		_ = writer.EmitStop("error", err.Error())
		return err
	}
	store := NewSessionStore(cfg.raw)
	blob, err := store.Load(req.Uid, req.SessionId, cfg.language, cfg.plannerPrompt)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=session_read\terr=%v",
			req.RequestId, req.Uid, req.SessionId, err))
		_ = writer.EmitStop("error", "session_read_failed")
		return err
	}
	history := append([]aichat.Message(nil), blob.Messages...)
	blob.Messages = append(blob.Messages, aichat.Message{Role: "user", Content: req.UserText})

	var suggestionRuntime *searchsuggestion.RuntimeConfig
	var suggestionPrerequisite *searchsuggestion.Error
	if req.EnableSuggestion {
		suggestionRuntime, suggestionPrerequisite = searchsuggestion.ResolveGenerator(req.Config, req.SceneId, cfg.language)
	}

	model, err := getChatModel(cfg.raw.LLMAlgoName)
	if err != nil {
		_ = writer.EmitStop("error", "llm_not_found")
		return err
	}
	chatRecall, err := getChatRecall(cfg.raw.RecallName)
	if err != nil {
		_ = writer.EmitStop("error", "recall_not_found")
		return err
	}
	meta := timingMeta{
		requestId: req.RequestId,
		uid:       req.Uid,
		sessionId: req.SessionId,
		sceneId:   req.SceneId,
		language:  cfg.language,
	}
	var knowledge *knowledgeEvidence
	if req.EnableSuggestion && suggestionPrerequisite == nil && !cfg.knowledgeConfigured {
		suggestionPrerequisite = searchsuggestion.NewError(searchsuggestion.CodeKnowledgeUnavailable, false, nil)
	}
	if cfg.knowledgeConfigured {
		knowledgeRecall, ok := chatRecall.(interface {
			SearchKnowledge(context.Context, string) (*recallsvc.KnowledgeSearchResult, error)
		})
		if !ok {
			err := fmt.Errorf("recall %s does not support configured knowledge vector search", cfg.raw.RecallName)
			log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=knowledge_recall\terr=%v",
				meta.requestId, meta.uid, meta.sessionId, err))
			_ = writer.EmitStop("error", "knowledge_recall_unsupported")
			return err
		}
		knowledgeStart := time.Now()
		knowledgeResult, knowledgeErr := knowledgeRecall.SearchKnowledge(ctx, req.UserText)
		knowledgeCost := utils.CostTime(knowledgeStart)
		if knowledgeErr != nil {
			if req.EnableSuggestion && suggestionPrerequisite == nil {
				suggestionPrerequisite = searchsuggestion.NewError(searchsuggestion.CodeKnowledgeFailed, true, knowledgeErr)
			}
			log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=knowledge_recall\tstatus=degraded\tcost=%d\terr=%s",
				meta.requestId, meta.uid, meta.sessionId, knowledgeCost, compactLogError(knowledgeErr)))
		} else {
			if knowledgeResult == nil {
				knowledgeResult = &recallsvc.KnowledgeSearchResult{}
			}
			knowledge = newKnowledgeEvidence(knowledgeResult)
			candidateCount := 0
			candidateSummary := "[]"
			if knowledge != nil {
				candidateCount = knowledge.Len()
				candidateSummary = compactJSON(knowledge.LogSummary())
			} else if req.EnableSuggestion && suggestionPrerequisite == nil {
				suggestionPrerequisite = searchsuggestion.NewError(searchsuggestion.CodeKnowledgeEmpty, false, nil)
			}
			log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=knowledge_recall\tstatus=ok\ttotal=%d\thits=%d\tcandidates=%d\tembeddingDimension=%d\tembeddingAttempts=%d\tembeddingCost=%d\tsearchCost=%d\tcost=%d\tcandidateSummary=%s",
				meta.requestId, meta.uid, meta.sessionId, knowledgeResult.Total, len(knowledgeResult.Hits), candidateCount,
				knowledgeResult.EmbeddingDimension, knowledgeResult.EmbeddingAttempts, knowledgeResult.EmbeddingCostMs,
				knowledgeResult.SearchCostMs, knowledgeCost, compactJSONString(candidateSummary, toolArgumentsLogLimit)))
		}
	}
	var coordinator *suggestionCoordinator
	var onFinalSearch func(context.Context, *finalSearchSnapshot)
	if req.EnableSuggestion {
		coordinator = newSuggestionCoordinator(ctx, suggestionRuntime, cfg.language, req.UserText, history, knowledge, suggestionPrerequisite)
		defer coordinator.Cancel()
		if suggestionPrerequisite == nil {
			onFinalSearch = coordinator.OnFinalSearch
		}
	}
	var rankRuntime *fineRankRuntime
	if cfg.raw.FineRankConfig != nil {
		rankRuntime = newFineRankRuntime(req)
	}
	loopResult, err := runAgentLoop(ctx, model, chatRecall, blob, cfg, rankRuntime, knowledge, writer, meta, onFinalSearch)
	if err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=upstream\terr=%v",
			req.RequestId, req.Uid, req.SessionId, err))
		_ = writer.EmitStop("error", "upstream_error")
		return err
	}
	reply := loopResult.Reply
	canonical, events := resolveReplyEvents(
		reply,
		loopResult.IndexMap,
		loopResult.ReplyItemIDs,
		cfg.raw.DisplayItemCountMax,
		rankRuntime != nil,
	)
	if !loopResult.ReplyAlreadyEmitted {
		if err := writer.EmitStep("reply"); err != nil {
			return err
		}
		for _, event := range events {
			if event.Content != "" {
				if err := writer.EmitContent(event.Content); err != nil {
					return err
				}
			}
			if event.ItemId != "" {
				if err := writer.EmitCitation(event.ItemId); err != nil {
					return err
				}
			}
		}
	}
	recordAssistant(blob, canonical)
	if err := store.Save(req.Uid, req.SessionId, blob); err != nil {
		log.Error(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=session_write\terr=%v",
			req.RequestId, req.Uid, req.SessionId, err))
		_ = writer.EmitStop("error", "session_write_failed")
		return err
	}
	if coordinator != nil {
		outcome := coordinator.Collect(loopResult)
		if outcome.Err != nil {
			log.Warning(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=suggestion\tstatus=error\tcode=%s\tretryable=%t\terr=%s",
				req.RequestId, req.Uid, req.SessionId, outcome.Err.Code, outcome.Err.Retryable, compactLogError(outcome.Err.Cause)))
			if err := writer.EmitSuggestionError(outcome.Err.Code, outcome.Err.PublicMessage(cfg.language), outcome.Err.Retryable); err != nil {
				return err
			}
		} else {
			log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tphase=suggestion\tstatus=ok\tcount=%d",
				req.RequestId, req.Uid, req.SessionId, len(outcome.Suggestions)))
			if err := writer.EmitSuggestions(outcome.Suggestions); err != nil {
				return err
			}
		}
	}
	status = "ok"
	return writer.EmitStop("stop", "")
}

func getChatModel(name string) (*aichat.Model, error) {
	algo, err := algorithm.Get(name)
	if err != nil {
		return nil, err
	}
	model, ok := algo.(*aichat.Model)
	if !ok {
		return nil, fmt.Errorf("algorithm %s is not PAI_CHAT", name)
	}
	return model, nil
}

func getChatRecall(name string) (chatRecall, error) {
	recall, err := recallsvc.GetRecall(name)
	if err != nil {
		return nil, err
	}
	chatRecall, ok := recall.(chatRecall)
	if !ok {
		return nil, fmt.Errorf("recall %s is not chat recall", name)
	}
	return chatRecall, nil
}

func recordAssistant(blob *SessionBlob, content string) {
	if len(blob.Messages) > 0 {
		last := &blob.Messages[len(blob.Messages)-1]
		if last.Role == "assistant" && len(last.ToolCalls) == 0 {
			last.Content = content
			return
		}
	}
	blob.Messages = append(blob.Messages, aichat.Message{Role: "assistant", Content: content})
}
