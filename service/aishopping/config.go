package aishopping

import (
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils/ast"
)

func resolveConfig(config *recconf.RecommendConfig, sceneId, language string) (*chatConfig, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	sceneConfs, ok := config.SceneConfs[sceneId]
	if !ok {
		return nil, fmt.Errorf("ai chat scene not found:%s", sceneId)
	}
	category, ok := sceneConfs["default"]
	if !ok || category.AIChatConfig == nil {
		return nil, fmt.Errorf("AIChatConfig not found for scene:%s", sceneId)
	}
	cfg := normalizeConfig(cloneAIChatConfig(category.AIChatConfig))
	if err := validateFineRankConfig(config.AlgoConfs, cfg); err != nil {
		return nil, err
	}
	if language == "" {
		language = cfg.DefaultLanguage
	}
	if !contains(cfg.OutputLanguages, language) {
		return nil, fmt.Errorf("language_unsupported:%s", language)
	}
	plannerPrompt := promptForLanguage(cfg.PlannerPromptTemplates, language)
	replyPrompt := promptForLanguage(cfg.ReplyPromptTemplates, language)
	if plannerPrompt == "" {
		return nil, fmt.Errorf("planner_prompt_missing:%s", language)
	}
	if replyPrompt == "" {
		return nil, fmt.Errorf("reply_prompt_missing:%s", language)
	}
	if _, ok := cfg.FallbackTemplates[language]; !ok {
		return nil, fmt.Errorf("fallback_missing:%s", language)
	}
	fieldAware := isFieldAwareRecall(config.RecallConfs, cfg.RecallName)
	knowledgeConfigured := isKnowledgeRecall(config.RecallConfs, cfg.RecallName)
	if knowledgeConfigured && strings.TrimSpace(cfg.KnowledgePlannerInstruction) == "" {
		return nil, fmt.Errorf("KnowledgePlannerInstruction is required when Ha3KnowledgeVectorConf is configured")
	}
	if knowledgeConfigured && !fieldAware {
		return nil, fmt.Errorf("Ha3KnowledgeVectorConf requires field-aware Ha3ChatRecallConf")
	}
	return &chatConfig{
		raw:                 cfg,
		language:            language,
		plannerPrompt:       plannerPrompt,
		replyPrompt:         replyPrompt,
		fieldAware:          fieldAware,
		knowledgeConfigured: knowledgeConfigured,
	}, nil
}

func isFieldAwareRecall(recallConfs []recconf.RecallConfig, recallName string) bool {
	for _, recallConf := range recallConfs {
		if recallConf.Name == recallName {
			conf := recallConf.Ha3ChatRecallConf
			return conf.TitleField != "" && conf.CategoryField != ""
		}
	}
	return false
}

func isKnowledgeRecall(recallConfs []recconf.RecallConfig, recallName string) bool {
	for _, recallConf := range recallConfs {
		if recallConf.Name == recallName {
			return recallConf.Ha3KnowledgeVectorConf != nil
		}
	}
	return false
}

func normalizeConfig(cfg *recconf.AIChatConfig) *recconf.AIChatConfig {
	if cfg.DefaultLanguage == "" {
		cfg.DefaultLanguage = "zh"
	}
	if len(cfg.OutputLanguages) == 0 {
		cfg.OutputLanguages = []string{cfg.DefaultLanguage}
	}
	if cfg.ToolMaxRounds <= 0 {
		cfg.ToolMaxRounds = 5
	}
	if cfg.DisplayItemCountMax <= 0 {
		cfg.DisplayItemCountMax = 12
	}
	if cfg.SessionMaxTurns <= 0 {
		cfg.SessionMaxTurns = 32
	}
	if cfg.SessionMaxTokens <= 0 {
		cfg.SessionMaxTokens = 24000
	}
	return cfg
}

func cloneAIChatConfig(cfg *recconf.AIChatConfig) *recconf.AIChatConfig {
	if cfg == nil {
		return nil
	}
	cloned := *cfg
	cloned.OutputLanguages = append([]string(nil), cfg.OutputLanguages...)
	cloned.PlannerPromptTemplates = cloneStringMap(cfg.PlannerPromptTemplates)
	cloned.ReplyPromptTemplates = cloneStringMap(cfg.ReplyPromptTemplates)
	cloned.FallbackTemplates = cloneNestedStringMap(cfg.FallbackTemplates)
	if cfg.FineRankConfig != nil {
		fineRank := *cfg.FineRankConfig
		fineRank.RankConf.RankAlgoList = append([]string(nil), cfg.FineRankConfig.RankConf.RankAlgoList...)
		fineRank.RankConf.ContextFeatures = append([]string(nil), cfg.FineRankConfig.RankConf.ContextFeatures...)
		fineRank.RankConf.ItemFeatures = append([]string(nil), cfg.FineRankConfig.RankConf.ItemFeatures...)
		cloned.FineRankConfig = &fineRank
	}
	return &cloned
}

func validateFineRankConfig(algoConfs []recconf.AlgoConfig, cfg *recconf.AIChatConfig) error {
	fineRank := cfg.FineRankConfig
	if fineRank == nil {
		return nil
	}
	if fineRank.CandidateCount < cfg.DisplayItemCountMax {
		return fmt.Errorf("FineRankConfig.CandidateCount must be >= DisplayItemCountMax")
	}
	rankConf := fineRank.RankConf
	if len(rankConf.RankAlgoList) != 1 {
		return fmt.Errorf("FineRankConfig.RankConf.RankAlgoList must contain exactly one algorithm")
	}
	algoName := strings.TrimSpace(rankConf.RankAlgoList[0])
	if algoName == "" || !containsAlgo(algoConfs, algoName) {
		return fmt.Errorf("FineRankConfig algorithm not found:%s", algoName)
	}
	fineRank.RankConf.RankAlgoList[0] = algoName
	if strings.TrimSpace(rankConf.RankScore) == "" {
		return fmt.Errorf("FineRankConfig.RankConf.RankScore is required")
	}
	if rankConf.BatchCount != 0 || len(rankConf.ScoreRewrite) != 0 {
		return fmt.Errorf("FineRankConfig does not support BatchCount or ScoreRewrite")
	}
	if _, err := ast.GetExpASTWithType(rankConf.RankScore, rankConf.ASTType); err != nil {
		return fmt.Errorf("invalid FineRankConfig.RankConf.RankScore: %w", err)
	}
	return nil
}

func containsAlgo(algoConfs []recconf.AlgoConfig, name string) bool {
	for _, algoConf := range algoConfs {
		if algoConf.Name == name {
			return true
		}
	}
	return false
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneNestedStringMap(values map[string]map[string]string) map[string]map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]map[string]string, len(values))
	for key, value := range values {
		cloned[key] = cloneStringMap(value)
	}
	return cloned
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func promptForLanguage(templates map[string]string, language string) string {
	if templates == nil {
		return ""
	}
	return templates[language]
}

func fallbackText(cfg *recconf.AIChatConfig, language, key string) string {
	langFallbacks := cfg.FallbackTemplates[language]
	if text := langFallbacks[key]; text != "" {
		return text
	}
	if text := langFallbacks["generic"]; text != "" {
		return text
	}
	return "抱歉，我这边出了点小问题，麻烦再试一次。"
}
