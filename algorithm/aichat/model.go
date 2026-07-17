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

const (
	defaultPAIModelRegion  = "cn-beijing"
	defaultPAIModelTimeout = 60000
	paiModelEndpointFormat = "https://%s.pai-token.aliyuncs.com/v1"
	paiTokenChannel        = "pairec_agent"
)

type chatCompletionPayload struct {
	*ChatCompletionRequest
	ExtraBody map[string]string `json:"extra_body"`
}

type Model struct {
	name     string
	conf     recconf.PAIModelConfig
	client   *http.Client
	apiKey   string
	endpoint string
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
	region := strings.TrimSpace(m.conf.Region)
	if region == "" {
		region = defaultPAIModelRegion
	}
	m.endpoint = fmt.Sprintf(paiModelEndpointFormat, region)
	m.client = newHTTPClient(m.conf.Timeout)
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
	endpoint := m.endpoint + "/chat/completions"
	body, err := json.Marshal(chatCompletionPayload{
		ChatCompletionRequest: request,
		ExtraBody: map[string]string{
			"channel": paiTokenChannel,
		},
	})
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
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		if readErr != nil {
			return nil, fmt.Errorf("aichat upstream status:%d read body error:%v", resp.StatusCode, readErr)
		}
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

func newHTTPClient(timeout int) *http.Client {
	if timeout <= 0 {
		timeout = defaultPAIModelTimeout
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxConnsPerHost = 2000
	tr.MaxIdleConnsPerHost = 2000
	tr.MaxIdleConns = 2000
	tr.ResponseHeaderTimeout = time.Duration(timeout) * time.Millisecond
	return &http.Client{Transport: tr}
}
