package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba/pairec/v2/log"
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
	if request == nil {
		return nil, errors.New("aichat request is nil")
	}
	payload := *request
	if payload.Model == "" {
		payload.Model = m.conf.Model
	}
	payload.Stream = true
	body, err := json.Marshal(chatCompletionPayload{
		ChatCompletionRequest: &payload,
		ExtraBody: map[string]string{
			"channel": paiTokenChannel,
		},
	})
	if err != nil {
		return nil, err
	}
	attempts, callID, status := 0, rand.Uint64(), "error"
	defer func() {
		log.Info(fmt.Sprintf("module=AIChatModel\tcallId=%016x\tattempts=%d\tstatus=%s", callID, attempts, status))
	}()
	delivered := false
	// Replaying is safe only before any content reaches the caller.
	var handler DeltaHandler
	if onDelta != nil {
		handler = func(text string) error {
			delivered = true
			return onDelta(text)
		}
	}
	for retry := 0; ; retry++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		attempts++
		result, err := m.streamOnce(ctx, body, handler)
		if err == nil {
			status = "ok"
			return result, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if delivered || retry >= m.conf.RetryTimes {
			return nil, err
		}
		log.Warning(fmt.Sprintf("module=AIChatModel\tcallId=%016x\tevent=retry\tattempt=%d", callID, attempts+1))
	}
}

func (m *Model) streamOnce(ctx context.Context, body []byte, onDelta DeltaHandler) (*StreamResult, error) {
	endpoint := m.endpoint + "/chat/completions"
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
		if readErr != nil {
			return nil, fmt.Errorf("aichat upstream status:%d read body error:%v", resp.StatusCode, readErr)
		}
		if detail := strings.TrimSpace(string(body)); detail != "" {
			return nil, fmt.Errorf("aichat upstream status:%d body:%s", resp.StatusCode, detail)
		}
		return nil, fmt.Errorf("aichat upstream status:%d", resp.StatusCode)
	}
	return parseStream(resp.Body, onDelta)
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
