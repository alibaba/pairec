package eas

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/log"
	proto "github.com/golang/protobuf/proto"
)

// genRecGeneratePath is the fixed path of the GenRec generation API. It is
// appended to the configured service address when missing.
const genRecGeneratePath = "/generate"

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
// response is plain JSON instead of PBResponse.
type GenrecRequest struct {
	EasRequest
}

// normalizeGenRecUrl appends the /generate suffix to the configured service
// address. It is idempotent: addresses already carrying the suffix are used
// as-is, so the suffix is never duplicated.
func normalizeGenRecUrl(url string) string {
	url = strings.TrimRight(url, "/")
	if !strings.HasSuffix(url, genRecGeneratePath) {
		url += genRecGeneratePath
	}
	return url
}

func (r *GenrecRequest) SetUrl(url string) {
	r.EasRequest.SetUrl(normalizeGenRecUrl(url))
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

	req, err := http.NewRequest("POST", r.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", r.auth)
	req.Header.Set("Content-Type", "application/json")

	easClient.Timeout = r.timeout
	resp, err := easClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("genrec request failed, url:%s, status:%d, body:%s", r.url, resp.StatusCode, string(respBody))
	}

	genRecResponse := &GenRecResponse{}
	if err := json.Unmarshal(respBody, genRecResponse); err != nil {
		return nil, fmt.Errorf("error:%s, body:%s", err.Error(), string(respBody))
	}
	log.Info(fmt.Sprintf("module=GenrecRequest\turl=%s\trid=%s\tstatus=%d", r.url, data.Rid, resp.StatusCode))
	return genRecResponse, nil
}
