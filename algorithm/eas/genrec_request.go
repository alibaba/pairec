package eas

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/pkg/eas"
	proto "google.golang.org/protobuf/proto"
)

// GenRecRequestData carries the PB request and the sampling parameters of one
// GenRec invocation. The sampling parameters live in the outer JSON body, not
// in the proto, so they are passed alongside the PBRequest.
type GenRecRequestData struct {
	Request           *easyrec.PBRequest
	BeamWidth         int
	DynamicBeamWidths []int
	SidCount          int
	MaxNewTokens      int
	FilterParams      map[string]interface{}
	Rid               string
}

type genRecSamplingParams struct {
	BeamWidth    *int                   `json:"beam_width,omitempty"`
	N            *int                   `json:"n,omitempty"`
	MaxNewTokens int                    `json:"max_new_tokens"`
	CustomParams map[string]interface{} `json:"custom_params,omitempty"`
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
// response is plain JSON instead of PBResponse. The configured URL must
// include the /generate path, which is passed to the SDK without suffix completion.
type GenrecRequest struct {
	EasRequest
	EasClient *eas.PredictClient
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

	// Keep custom parameters local to this invocation; never mutate caller maps.
	customParams := make(map[string]interface{}, len(data.FilterParams)+1)
	for name, value := range data.FilterParams {
		// Beam parameters must come from configuration, not filter parameters.
		if name == "dynamic_beam_widths" {
			continue
		}
		customParams[name] = value
	}
	samplingParams := genRecSamplingParams{
		N:            &data.SidCount,
		MaxNewTokens: data.MaxNewTokens,
		CustomParams: customParams,
	}
	if len(data.DynamicBeamWidths) > 0 {
		customParams["dynamic_beam_widths"] = data.DynamicBeamWidths
		if data.SidCount == 0 {
			samplingParams.N = nil
		}
	} else {
		// Preserve the fixed beam request, including explicitly encoded zero values.
		samplingParams.BeamWidth = &data.BeamWidth
	}

	body, err := json.Marshal(&genRecHttpRequest{
		InputIds:       []int{0},
		ImageData:      base64.StdEncoding.EncodeToString(pbData),
		Stream:         false,
		SamplingParams: samplingParams,
		Rid:            data.Rid,
	})
	if err != nil {
		return nil, err
	}

	respBody, err := r.EasClient.BytesPredict(body)
	if err != nil {
		var predictErr *eas.PredictError
		if errors.As(err, &predictErr) {
			err = eas.NewPredictError(predictErr.Code, predictErr.RequestURL, genRecBodySnippet([]byte(predictErr.Message)))
		}
		return nil, fmt.Errorf("genrec request failed, url:%s, err:%w", r.url, err)
	}

	genRecResponse := &GenRecResponse{}
	if err := json.Unmarshal(respBody, genRecResponse); err != nil {
		return nil, fmt.Errorf("error:%s, body:%s", err.Error(), genRecBodySnippet(respBody))
	}
	return genRecResponse, nil
}
