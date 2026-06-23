package aishopping

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/aichat"
	"github.com/alibaba/pairec/v2/persist/fs"
	"github.com/alibaba/pairec/v2/recconf"
	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

type SessionStore struct {
	cfg *recconf.AIChatConfig
}

func NewSessionStore(cfg *recconf.AIChatConfig) *SessionStore {
	return &SessionStore{cfg: cfg}
}

func (s *SessionStore) Load(sessionId, language, prompt string) (*SessionBlob, error) {
	blob, err := s.read(sessionId)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if blob == nil {
		blob = &SessionBlob{
			Language:     language,
			CreatedAt:    now,
			LastActiveAt: now,
			Messages:     []aichat.Message{{Role: "system", Content: prompt}},
		}
		return blob, nil
	}
	blob.Language = language
	blob.LastActiveAt = now
	if len(blob.Messages) == 0 {
		blob.Messages = []aichat.Message{{Role: "system", Content: prompt}}
	} else {
		blob.Messages[0] = aichat.Message{Role: "system", Content: prompt}
	}
	return blob, nil
}

func (s *SessionStore) Save(sessionId string, blob *SessionBlob) error {
	blob.LastActiveAt = time.Now().Unix()
	blob.TurnCount = countTurns(blob.Messages)
	blob.Messages = trimMessages(blob.Messages, s.cfg.SessionMaxTurns, s.cfg.SessionMaxTokens)
	payload, err := json.Marshal(blob)
	if err != nil {
		return err
	}
	featureView, err := s.featureView()
	if err != nil {
		return err
	}
	return featureView.WriteFeatures([]map[string]interface{}{
		{
			"session_id":   sessionId,
			"session_blob": string(payload),
			"event_time":   blob.LastActiveAt,
		},
	}, domain.WithDirect())
}

func (s *SessionStore) read(sessionId string) (*SessionBlob, error) {
	featureView, err := s.featureView()
	if err != nil {
		return nil, err
	}
	rows, err := featureView.GetOnlineFeatures([]interface{}{sessionId}, []string{"session_blob"}, nil)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	value, ok := rows[0]["session_blob"]
	if !ok || value == nil {
		return nil, nil
	}
	var payload string
	switch v := value.(type) {
	case string:
		payload = v
	case []byte:
		payload = string(v)
	default:
		payload = fmt.Sprint(v)
	}
	if payload == "" {
		return nil, nil
	}
	blob := &SessionBlob{}
	if err := json.Unmarshal([]byte(payload), blob); err != nil {
		return nil, err
	}
	return blob, nil
}

func (s *SessionStore) featureView() (interface {
	GetOnlineFeatures([]interface{}, []string, map[string]string) ([]map[string]interface{}, error)
	WriteFeatures([]map[string]interface{}, ...domain.WriteOption) error
}, error) {
	client, err := fs.GetFeatureStoreClient(s.cfg.SessionFeatureStoreName)
	if err != nil {
		return nil, err
	}
	featureView := client.GetProject().GetFeatureView(s.cfg.SessionFeatureView)
	if featureView == nil {
		return nil, fmt.Errorf("feature view not found:%s", s.cfg.SessionFeatureView)
	}
	return featureView, nil
}

func countTurns(messages []aichat.Message) int {
	count := 0
	for _, message := range messages {
		if message.Role == "user" {
			count++
		}
	}
	return count
}

func trimMessages(messages []aichat.Message, maxTurns, maxTokens int) []aichat.Message {
	if len(messages) <= 1 {
		return messages
	}
	system := messages[0]
	groups := messageGroups(messages[1:])
	for countGroupTurns(groups) > maxTurns && len(groups) > 1 {
		groups = groups[1:]
	}
	out := flattenMessageGroups(system, groups)
	for estimatedTokens(out) > maxTokens && len(groups) > 1 {
		groups = groups[1:]
		out = flattenMessageGroups(system, groups)
	}
	return out
}

func messageGroups(messages []aichat.Message) [][]aichat.Message {
	groups := make([][]aichat.Message, 0)
	var current []aichat.Message
	for _, message := range messages {
		if message.Role == "user" && len(current) > 0 {
			groups = append(groups, current)
			current = nil
		}
		current = append(current, message)
	}
	if len(current) > 0 {
		groups = append(groups, current)
	}
	return groups
}

func countGroupTurns(groups [][]aichat.Message) int {
	count := 0
	for _, group := range groups {
		for _, message := range group {
			if message.Role == "user" {
				count++
				break
			}
		}
	}
	return count
}

func flattenMessageGroups(system aichat.Message, groups [][]aichat.Message) []aichat.Message {
	out := []aichat.Message{system}
	for _, group := range groups {
		out = append(out, group...)
	}
	return out
}

func estimatedTokens(messages []aichat.Message) int {
	total := 0
	for _, message := range messages {
		total += len(message.Content) / 4
		for _, call := range message.ToolCalls {
			total += len(call.Function.Arguments) / 4
		}
	}
	return total
}
