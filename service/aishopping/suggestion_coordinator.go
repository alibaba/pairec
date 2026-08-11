package aishopping

import (
	"context"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/service/searchsuggestion"
)

const suggestionHistoryTurns = 3

type suggestionCoordinator struct {
	parent       context.Context
	runtime      *searchsuggestion.RuntimeConfig
	language     string
	currentQuery string
	conversation []searchsuggestion.ConversationTurn
	knowledge    *knowledgeEvidence
	prerequisite *searchsuggestion.Error
	result       chan searchsuggestion.Outcome
	cancel       context.CancelFunc
	started      bool
}

func newSuggestionCoordinator(parent context.Context, runtime *searchsuggestion.RuntimeConfig, language, currentQuery string, history []aichat.Message, knowledge *knowledgeEvidence, prerequisite *searchsuggestion.Error) *suggestionCoordinator {
	return &suggestionCoordinator{
		parent:       parent,
		runtime:      runtime,
		language:     language,
		currentQuery: currentQuery,
		conversation: suggestionConversation(history),
		knowledge:    knowledge,
		prerequisite: prerequisite,
		result:       make(chan searchsuggestion.Outcome, 1),
	}
}

func (c *suggestionCoordinator) OnFinalSearch(_ context.Context, snapshot *finalSearchSnapshot) {
	if c == nil || c.started || c.prerequisite != nil || snapshot == nil {
		return
	}
	c.started = true
	taskCtx, cancel := context.WithCancel(c.parent)
	c.cancel = cancel
	snapshotCopy := freezeFinalSearchSnapshot(snapshot)
	go func() {
		summary, err := searchsuggestion.BuildProductSummary(snapshotCopy.ReplyToolPayload, snapshotCopy.ProductIndexes, snapshotCopy.Total)
		if err != nil {
			c.result <- searchsuggestion.Outcome{Err: err}
			return
		}
		input := &searchsuggestion.GenerationInput{
			Language:              c.language,
			SuggestionCount:       searchsuggestion.EmbeddedCount,
			CurrentQuery:          c.currentQuery,
			Conversation:          append([]searchsuggestion.ConversationTurn(nil), c.conversation...),
			FinalSearchIntent:     &snapshotCopy.Request,
			Knowledge:             c.knowledge.SuggestionKnowledge(snapshotCopy.Request.SelectedKnowledgeCandidateIDs),
			CurrentProductSummary: summary,
		}
		c.result <- searchsuggestion.Generate(taskCtx, c.runtime, input)
	}()
}

func (c *suggestionCoordinator) Collect(loopResult *agentLoopResult) searchsuggestion.Outcome {
	if loopResult.MainReplyFallback {
		c.Cancel()
		return searchsuggestion.Outcome{Err: searchsuggestion.NewError(searchsuggestion.CodeMainReplyIneligible, true, nil)}
	}
	if c.prerequisite != nil {
		return searchsuggestion.Outcome{Err: c.prerequisite}
	}
	if !c.started {
		code := searchsuggestion.CodeProductContextUnavailable
		retryable := false
		switch loopResult.FinalSearchStatus {
		case finalSearchFailed:
			code = searchsuggestion.CodeProductSearchFailed
			retryable = true
		case finalSearchInvalid:
			code = searchsuggestion.CodeProductContextInvalid
			retryable = true
		}
		return searchsuggestion.Outcome{Err: searchsuggestion.NewError(code, retryable, nil)}
	}
	select {
	case outcome := <-c.result:
		c.Cancel()
		return outcome
	default:
		c.Cancel()
		return searchsuggestion.Outcome{Err: searchsuggestion.NewError(searchsuggestion.CodeNotReady, true, nil)}
	}
}

func (c *suggestionCoordinator) Cancel() {
	if c != nil && c.cancel != nil {
		c.cancel()
	}
}

func freezeFinalSearchSnapshot(snapshot *finalSearchSnapshot) finalSearchSnapshot {
	copy := *snapshot
	copy.ProductIndexes = append([]int(nil), snapshot.ProductIndexes...)
	copy.Request.Keywords = append([]string(nil), snapshot.Request.Keywords...)
	copy.Request.ProductTypeKeywords = append([]string(nil), snapshot.Request.ProductTypeKeywords...)
	copy.Request.AttributeKeywords = append([]string(nil), snapshot.Request.AttributeKeywords...)
	copy.Request.ExcludeKeywords = append([]string(nil), snapshot.Request.ExcludeKeywords...)
	copy.Request.SelectedKnowledgeCandidateIDs = append([]string(nil), snapshot.Request.SelectedKnowledgeCandidateIDs...)
	copy.Request.MinPrice = cloneFloat(snapshot.Request.MinPrice)
	copy.Request.MaxPrice = cloneFloat(snapshot.Request.MaxPrice)
	return copy
}

func suggestionConversation(messages []aichat.Message) []searchsuggestion.ConversationTurn {
	turns := 0
	start := len(messages)
	for index := len(messages) - 1; index >= 0; index-- {
		if messages[index].Role == "user" {
			turns++
			if turns == suggestionHistoryTurns {
				start = index
				break
			}
		}
	}
	if turns < suggestionHistoryTurns {
		start = 0
	}
	result := make([]searchsuggestion.ConversationTurn, 0, len(messages)-start)
	for _, message := range messages[start:] {
		if (message.Role != "user" && message.Role != "assistant") || len(message.ToolCalls) > 0 {
			continue
		}
		content := strings.TrimSpace(itemMarkerRegexp.ReplaceAllString(message.Content, ""))
		if content == "" {
			continue
		}
		result = append(result, searchsuggestion.ConversationTurn{Role: message.Role, Content: content})
	}
	return result
}
