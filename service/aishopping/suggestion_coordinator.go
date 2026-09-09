package aishopping

import (
	"context"
	"strings"

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

func newSuggestionCoordinator(parent context.Context, runtime *searchsuggestion.RuntimeConfig, language, currentQuery string, history []SessionQuery, knowledge *knowledgeEvidence, prerequisite *searchsuggestion.Error) *suggestionCoordinator {
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
	if snapshot.Total == 0 && len(snapshot.ProductIndexes) == 0 {
		c.result <- searchsuggestion.Outcome{Suggestions: []string{}}
		return
	}
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
			Knowledge:             c.knowledge.SuggestionKnowledge(),
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
	copy.Request.PreferredKeywords = append([]string(nil), snapshot.Request.PreferredKeywords...)
	copy.Request.Constraints = cloneConstraints(snapshot.Request.Constraints)
	copy.Request.ExcludeKeywords = append([]string(nil), snapshot.Request.ExcludeKeywords...)
	copy.Request.MinPrice = cloneFloat(snapshot.Request.MinPrice)
	copy.Request.MaxPrice = cloneFloat(snapshot.Request.MaxPrice)
	return copy
}

func suggestionConversation(queries []SessionQuery) []searchsuggestion.ConversationTurn {
	if len(queries) > suggestionHistoryTurns {
		queries = queries[len(queries)-suggestionHistoryTurns:]
	}
	result := make([]searchsuggestion.ConversationTurn, 0, len(queries))
	for _, query := range queries {
		if strings.TrimSpace(query.Query) != "" {
			result = append(result, searchsuggestion.ConversationTurn{Role: "user", Content: query.Query})
		}
	}
	return result
}
