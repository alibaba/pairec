package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/service/aishopping"
	"github.com/alibaba/pairec/v2/utils"
)

type ChatParam struct {
	SceneId          string                 `json:"scene_id"`
	SessionId        string                 `json:"session_id"`
	InputMessage     ChatMessage            `json:"input_message"`
	Uid              string                 `json:"uid"`
	Language         string                 `json:"language"`
	Features         map[string]interface{} `json:"features,omitempty"`
	EnableSuggestion bool                   `json:"enable_suggestion"`
}

type ChatMessage struct {
	Content []ChatMessageContent `json:"content"`
}

type ChatMessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (p *ChatParam) GetParameter(name string) interface{} {
	switch name {
	case "scene", "scene_id":
		return p.SceneId
	case "session_id":
		return p.SessionId
	case "uid":
		return p.Uid
	case "language":
		return p.Language
	case "category":
		return "default"
	default:
		return nil
	}
}

type ChatController struct {
	Controller
	param    ChatParam
	userText string
}

func (c *ChatController) Process(w http.ResponseWriter, r *http.Request) {
	c.Start = time.Now()
	var err error
	c.RequestBody, err = c.ReadRequestBody(r)
	if err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, "read parameter error")
		return
	}
	if len(c.RequestBody) == 0 {
		c.SendError(w, ERROR_PARAMETER_CODE, "request body empty")
		return
	}
	c.RequestId = utils.UUID()
	if err := c.CheckParameter(); err != nil {
		c.SendError(w, ERROR_PARAMETER_CODE, err.Error())
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.SendError(w, SERVER_ERROR_CODE, "streaming unsupported")
		return
	}
	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tevent=begin\turi=%s\taddress=%s\tbody=%s",
		c.RequestId, c.param.Uid, c.param.SessionId, r.RequestURI, r.RemoteAddr, string(c.RequestBody)))
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	// Chat streams can run longer than the server write timeout while waiting on the model.
	_ = http.NewResponseController(w).SetWriteDeadline(time.Time{})

	stream := aishopping.NewStreamWriter(c.RequestId, c.param.SessionId, w, flusher)
	req := &aishopping.Request{
		RequestId:        c.RequestId,
		SceneId:          c.param.SceneId,
		SessionId:        c.param.SessionId,
		Uid:              c.param.Uid,
		Language:         c.param.Language,
		UserText:         c.userText,
		Features:         c.param.Features,
		EnableSuggestion: c.param.EnableSuggestion,
		Config:           recconf.Config,
	}
	_ = aishopping.NewChatSearchOrchestrator().Run(r.Context(), req, stream)
	c.End = time.Now()
	log.Info(fmt.Sprintf("requestId=%s\tuid=%s\tsession_id=%s\tmodule=AIShoppingChat\tevent=end\turi=%s\tcost=%d",
		c.RequestId, c.param.Uid, c.param.SessionId, r.RequestURI, c.cost()))
}

func (c *ChatController) CheckParameter() error {
	if err := json.Unmarshal(c.RequestBody, &c.param); err != nil {
		return err
	}
	if c.param.SceneId == "" {
		c.param.SceneId = "ai_shopping"
	}
	c.param.SessionId = strings.TrimSpace(c.param.SessionId)
	if c.param.SessionId == "" {
		return errors.New("session_id is empty")
	}
	c.param.Uid = strings.TrimSpace(c.param.Uid)
	if c.param.Uid == "" {
		return errors.New("uid not empty")
	}
	parts := make([]string, 0, len(c.param.InputMessage.Content))
	for _, content := range c.param.InputMessage.Content {
		if content.Type != "text" {
			return errors.New("only text input is supported")
		}
		if strings.TrimSpace(content.Text) != "" {
			parts = append(parts, content.Text)
		}
	}
	c.userText = strings.TrimSpace(strings.Join(parts, "\n"))
	if c.userText == "" {
		return errors.New("input_message.content text not empty")
	}
	return nil
}
