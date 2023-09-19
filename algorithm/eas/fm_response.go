package eas

import (
	"errors"
	"fmt"

	"github.com/alibaba/pairec/v2/algorithm/response"
)

//easyjson:json
type alinkFMResponseList []*alinkFMResponse

type alinkFMResponse struct {
	Result float64 `json:"prediction_result"`
	Score  float64 `json:"prediction_score"`
	//Details string  `json:"prediction_detail"`
}

func (r *alinkFMResponse) GetModuleType() bool {
	return false
}

func (r *alinkFMResponse) GetScoreMap() map[string]float64 {
	var default_val map[string]float64
	return default_val
}

func (r *alinkFMResponse) GetScore() float64 {
	if r.Result == float64(0) {
		return 1 - r.Score
	}

	return r.Score
}

func alinkFMResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	retstr, ok := data.(string)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	var result alinkFMResponseList
	err = result.UnmarshalJSON([]byte(retstr))
	if err != nil {
		err = errors.New(fmt.Sprintf("error:%v, body:%s", err, bodyFormat(retstr, 512)))
		return
	}

	for _, res := range result {
		ret = append(ret, res)
	}
	return
}

func bodyFormat(body string, size int) string {
	if len(body) <= size {
		return body
	}

	return body[:size]
}
