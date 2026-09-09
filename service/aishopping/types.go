package aishopping

import (
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/searchsuggestion"
)

type Request struct {
	RequestId        string
	SceneId          string
	SessionId        string
	Uid              string
	Language         string
	UserText         string
	Features         map[string]interface{}
	EnableSuggestion bool
	Config           *recconf.RecommendConfig
}

func (r *Request) GetParameter(name string) interface{} {
	switch name {
	case "scene", "scene_id":
		return r.SceneId
	case "session_id":
		return r.SessionId
	case "uid":
		return r.Uid
	case "language":
		return r.Language
	case "features":
		return r.Features
	case "category":
		return "default"
	default:
		return nil
	}
}

type SessionBlob struct {
	Language         string                         `json:"language"`
	CreatedAt        int64                          `json:"created_at"`
	LastActiveAt     int64                          `json:"last_active_at"`
	TurnCount        int                            `json:"turn_count"`
	UserQueries      []SessionQuery                 `json:"user_queries"`
	LastSearch       *searchsuggestion.SearchIntent `json:"last_search,omitempty"`
	LastSearchTurnID int                            `json:"last_search_turn_id,omitempty"`
}

type SessionQuery struct {
	TurnID int    `json:"turn_id"`
	Query  string `json:"query"`
}

type chatConfig struct {
	raw                 *recconf.AIChatConfig
	language            string
	plannerPrompt       string
	replyPrompt         string
	fieldAware          bool
	knowledgeConfigured bool
}
