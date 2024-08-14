package eas

import (
	"encoding/json"
	"fmt"

	"github.com/alibaba/pairec/v2/algorithm/response"
)

type pssmartResponse struct {
	Score  float64 `json:"score"`
	Label  string  `json:"lable"`
	Label2 string  `json:"label"`
}

func (r *pssmartResponse) GetModuleType() bool {
	return false
}

func (r *pssmartResponse) GetScoreMap() map[string]float64 {
	var default_val map[string]float64
	return default_val
}

func (r *pssmartResponse) GetScore() float64 {
	if r.Label == "0" {
		return 1 - r.Score
	} else if r.Label2 == "0" {
		return 1 - r.Score
	}
	return r.Score
}

func pssmartResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {

	retstr, ok := data.(string)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	result := make([]*pssmartResponse, 0)
	err = json.Unmarshal([]byte(retstr), &result)
	if err != nil {
		return
	}

	for _, res := range result {
		ret = append(ret, res)
	}
	return
}
