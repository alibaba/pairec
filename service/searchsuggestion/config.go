package searchsuggestion

import (
	"context"
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm"
	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/recconf"
	recallsvc "github.com/alibaba/pairec/v2/service/recall"
)

type KnowledgeRecall interface {
	SearchKnowledge(context.Context, string) (*recallsvc.KnowledgeSearchResult, error)
}

type RuntimeConfig struct {
	Prompt string
	Model  *aichat.Model
}

func ResolveGenerator(config *recconf.RecommendConfig, sceneID, language string) (*RuntimeConfig, *Error) {
	suggestionConfig, err := findConfig(config, sceneID)
	if err != nil {
		return nil, NewError(CodeConfigUnavailable, false, err)
	}
	if language == "" {
		language = "zh"
	}
	prompt := strings.TrimSpace(suggestionConfig.PromptTemplates[language])
	if prompt == "" {
		return nil, NewError(CodePromptMissing, false, fmt.Errorf("prompt missing for language %q", language))
	}
	algo, err := algorithm.Get(strings.TrimSpace(suggestionConfig.LLMAlgoName))
	if err != nil {
		return nil, NewError(CodeModelUnavailable, false, err)
	}
	model, ok := algo.(*aichat.Model)
	if !ok {
		return nil, NewError(CodeModelUnavailable, false, fmt.Errorf("algorithm %q is not PAI_CHAT", suggestionConfig.LLMAlgoName))
	}
	return &RuntimeConfig{Prompt: prompt, Model: model}, nil
}

func ResolveStandalone(config *recconf.RecommendConfig, sceneID, language string) (*RuntimeConfig, KnowledgeRecall, *Error) {
	runtimeConfig, resolveErr := ResolveGenerator(config, sceneID, language)
	if resolveErr != nil {
		return nil, nil, resolveErr
	}
	suggestionConfig, err := findConfig(config, sceneID)
	if err != nil {
		return nil, nil, NewError(CodeConfigUnavailable, false, err)
	}
	recallName := strings.TrimSpace(suggestionConfig.RecallName)
	if recallName == "" {
		return nil, nil, NewError(CodeKnowledgeUnavailable, false, fmt.Errorf("SuggestionConfig.RecallName is empty"))
	}
	recall, err := recallsvc.GetRecall(recallName)
	if err != nil {
		return nil, nil, NewError(CodeKnowledgeUnavailable, false, err)
	}
	knowledgeRecall, ok := recall.(KnowledgeRecall)
	if !ok {
		return nil, nil, NewError(CodeKnowledgeUnavailable, false, fmt.Errorf("recall %q does not support knowledge search", recallName))
	}
	if enabled, ok := recall.(interface{ KnowledgeEnabled() bool }); ok && !enabled.KnowledgeEnabled() {
		return nil, nil, NewError(CodeKnowledgeUnavailable, false, fmt.Errorf("recall %q knowledge search is disabled", recallName))
	}
	return runtimeConfig, knowledgeRecall, nil
}

func findConfig(config *recconf.RecommendConfig, sceneID string) (*recconf.SuggestionConfig, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	scene, ok := config.SceneConfs[sceneID]
	if !ok {
		return nil, fmt.Errorf("scene not found: %s", sceneID)
	}
	category, ok := scene["default"]
	if !ok || category.SuggestionConfig == nil {
		return nil, fmt.Errorf("SuggestionConfig not found for scene: %s", sceneID)
	}
	cloned := *category.SuggestionConfig
	cloned.PromptTemplates = make(map[string]string, len(category.SuggestionConfig.PromptTemplates))
	for key, value := range category.SuggestionConfig.PromptTemplates {
		cloned.PromptTemplates[key] = value
	}
	if strings.TrimSpace(cloned.LLMAlgoName) == "" {
		return nil, fmt.Errorf("SuggestionConfig.LLMAlgoName is empty")
	}
	return &cloned, nil
}
