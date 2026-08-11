package aishopping

import (
	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/recconf"
)

type Request struct {
	RequestId        string
	SceneId          string
	SessionId        string
	Uid              string
	Language         string
	UserText         string
	EnableSuggestion bool
	Config           *recconf.RecommendConfig
}

type SessionBlob struct {
	Language     string           `json:"language"`
	CreatedAt    int64            `json:"created_at"`
	LastActiveAt int64            `json:"last_active_at"`
	TurnCount    int              `json:"turn_count"`
	Messages     []aichat.Message `json:"messages"`
}

type chatConfig struct {
	raw                 *recconf.AIChatConfig
	language            string
	plannerPrompt       string
	replyPrompt         string
	fieldAware          bool
	knowledgeConfigured bool
}
