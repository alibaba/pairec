package eas

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/log"
	proto "google.golang.org/protobuf/proto"
)

// GenRecRequestData carries the PB request and the sampling parameters of one
// GenRec invocation. The sampling parameters live in the outer JSON body, not
// in the proto, so they are passed alongside the PBRequest.
type GenRecRequestData struct {
	Request      *easyrec.PBRequest
	BeamWidth    int
	SidCount     int
	MaxNewTokens int
	Rid          string
}

type genRecSamplingParams struct {
	BeamWidth    int `json:"beam_width"`
	N            int `json:"n"`
	MaxNewTokens int `json:"max_new_tokens"`
}

type genRecHttpRequest struct {
	InputIds       []int                `json:"input_ids"`
	ImageData      string               `json:"image_data"`
	Stream         bool                 `json:"stream"`
	SamplingParams genRecSamplingParams `json:"sampling_params"`
	Rid            string               `json:"rid,omitempty"`
}

// GenrecRequest invokes the EAS GenRec /generate API: the PBRequest is
// serialized, base64 encoded into image_data and wrapped in a JSON body; the
// response is plain JSON instead of PBResponse. The configured URL is used
// as-is and must include the /generate path.
type GenrecRequest struct {
	EasRequest
}

// genRecBodySnippet caps response bodies embedded in error messages.
func genRecBodySnippet(body []byte) string {
	const maxLen = 4096
	if len(body) > maxLen {
		return string(body[:maxLen]) + "...(truncated)"
	}
	return string(body)
}

func (r *GenrecRequest) Invoke(requestData interface{}) (response interface{}, err error) {
	data, ok := requestData.(*GenRecRequestData)
	if !ok {
		return nil, errors.New("requestData is not eas.GenRecRequestData type")
	}

	pbData, err := proto.Marshal(data.Request)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(&genRecHttpRequest{
		InputIds:  []int{0},
		ImageData: base64.StdEncoding.EncodeToString(pbData),
		Stream:    false,
		SamplingParams: genRecSamplingParams{
			BeamWidth:    data.BeamWidth,
			N:            data.SidCount,
			MaxNewTokens: data.MaxNewTokens,
		},
		Rid: data.Rid,
	})
	if err != nil {
		return nil, err
	}

	// Per-request deadline; the shared easClient is left immutable.
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", r.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", r.auth)
	req.Header.Set("Content-Type", "application/json")

	resp, err := easClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("genrec request failed, url:%s, status:%d, body:%s", r.url, resp.StatusCode, genRecBodySnippet(respBody))
	}

	genRecResponse := &GenRecResponse{}
	if err := json.Unmarshal(respBody, genRecResponse); err != nil {
		return nil, fmt.Errorf("error:%s, body:%s", err.Error(), genRecBodySnippet(respBody))
	}
	log.Info(fmt.Sprintf("module=GenrecRequest\turl=%s\trid=%s\tstatus=%d", r.url, data.Rid, resp.StatusCode))
	return genRecResponse, nil
}
