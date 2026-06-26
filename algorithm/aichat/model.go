package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/recconf"
)

const paiModelBaseURL = "https://cn-beijing.pai-token.aliyuncs.com/v1"

type Model struct {
	name   string
	conf   recconf.PAIModelConfig
	client *http.Client
	apiKey string
}

func NewModel(name string) *Model {
	return &Model{name: name}
}

func (m *Model) Init(conf *recconf.AlgoConfig) error {
	m.conf = conf.PAIModelConf
	if m.conf.Model == "" {
		return errors.New("PAIModelConf.Model is empty")
	}
	if m.conf.APIKey == "" {
		return errors.New("PAIModelConf.APIKey is empty")
	}
	m.apiKey = m.conf.APIKey
	timeout := m.conf.Timeout
	if timeout <= 0 {
		timeout = 60000
	}
	m.client = &http.Client{Timeout: time.Duration(timeout) * time.Millisecond}
	return nil
}

func (m *Model) Run(algoData interface{}) (interface{}, error) {
	req, ok := algoData.(*ChatCompletionRequest)
	if !ok {
		return nil, errors.New("aichat model expects *ChatCompletionRequest")
	}
	return m.Stream(context.Background(), req, nil)
}

func (m *Model) Stream(ctx context.Context, request *ChatCompletionRequest, onDelta DeltaHandler) (*StreamResult, error) {
	if request.Model == "" {
		request.Model = m.conf.Model
	}
	request.Stream = true
	endpoint := paiModelBaseURL + "/chat/completions"
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		detail := strings.TrimSpace(string(body))
		if detail != "" {
			return nil, fmt.Errorf("aichat upstream status:%d body:%s", resp.StatusCode, detail)
		}
		return nil, fmt.Errorf("aichat upstream status:%d", resp.StatusCode)
	}
	result, err := parseStream(resp.Body, onDelta)
	if err != nil {
		return nil, err
	}
	return result, nil
}
