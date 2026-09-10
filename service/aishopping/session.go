package aishopping

import (
	"encoding/json"
	"fmt"
	"time"

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

func (s *SessionStore) Load(uid, sessionId, language string) (*SessionBlob, error) {
	blob, err := s.read(sessionStoreKey(uid, sessionId))
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if blob == nil {
		blob = &SessionBlob{
			Language:     language,
			CreatedAt:    now,
			LastActiveAt: now,
		}
		return blob, nil
	}
	blob.Language = language
	blob.LastActiveAt = now
	return blob, nil
}

func (s *SessionStore) Save(uid, sessionId string, blob *SessionBlob) error {
	blob.LastActiveAt = time.Now().Unix()
	blob.trim(s.cfg.SessionMaxTurns, s.cfg.SessionMaxTokens)
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
			"session_id":   sessionStoreKey(uid, sessionId),
			"session_blob": string(payload),
			"event_time":   blob.LastActiveAt,
		},
	}, domain.WithDirect())
}

func sessionStoreKey(uid, sessionId string) string {
	return "aishopping:v2:" + uid + ":" + sessionId
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
