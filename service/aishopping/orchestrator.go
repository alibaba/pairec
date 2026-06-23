package aishopping

import (
	"context"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/log"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
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
		log.Info(fmt.Sprintf("requestId=%s\tmodule=AIShoppingChat\tphase=total\tscene=%s\tsessionId=%s\tlanguage=%s\tstatus=%s\tcost=%d",
			req.RequestId, req.SceneId, req.SessionId, req.Language, status, utils.CostTime(totalStart)))
	}()
	cfg, err := resolveConfig(req.Config, req.SceneId, req.Language)
	if err != nil {
		_ = writer.EmitStop("error", err.Error())
		return err
	}
	store := NewSessionStore(cfg.raw)
	blob, err := store.Load(req.SessionId, cfg.language, cfg.plannerPrompt)
	if err != nil {
		log.Error(fmt.Sprintf("ai shopping session read failed, requestId:%s, sessionId:%s, err:%v", req.RequestId, req.SessionId, err))
		_ = writer.EmitStop("error", "session_read_failed")
		return err
	}
	blob.Messages = append(blob.Messages, aichat.Message{Role: "user", Content: req.UserText})

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
		sessionId: req.SessionId,
		sceneId:   req.SceneId,
		language:  cfg.language,
	}
	loopResult, err := runAgentLoop(ctx, model, chatRecall, blob, cfg, writer, meta)
	if err != nil {
		log.Error(fmt.Sprintf("ai shopping upstream failed, requestId:%s, sessionId:%s, err:%v", req.RequestId, req.SessionId, err))
		_ = writer.EmitStop("error", "upstream_error")
		return err
	}
	reply := loopResult.Reply
	canonical, events := resolveReplyEvents(reply, loopResult.IndexMap, cfg.raw.DisplayItemCountMax)
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
	if err := store.Save(req.SessionId, blob); err != nil {
		log.Error(fmt.Sprintf("ai shopping session write failed, requestId:%s, sessionId:%s, err:%v", req.RequestId, req.SessionId, err))
		_ = writer.EmitStop("error", "session_write_failed")
		return err
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
