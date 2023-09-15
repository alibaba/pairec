package tfserving

import (
	"fmt"

	"github.com/alibaba/pairec/algorithm/response"
)

type tfservingResponse struct {
	score          float64
	scoreArr       map[string]float64
	multiValModule bool
}

func (r *tfservingResponse) GetScore() float64 {
	return r.score
}

func (r *tfservingResponse) GetScoreMap() map[string]float64 {
	return r.scoreArr
}

func (r *tfservingResponse) GetModuleType() bool {
	return r.multiValModule
}

/**
func tfservingMutValResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*tf.PredictResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	var response []map[string]float64
	for name, arrayProto := range resp.GetOutputs() {
		for i, val := range arrayProto.FloatVal {
			if i >= len(response) {
				response = append(response, map[string]float64{name: float64(val)})
			} else {
				response[i][name] = float64(val)
			}
		}
	}
	for _, v := range response {
		ret = append(ret, &tfResponse{scoreArr: v, multiValModule: true})
	}
	return
}
**/

func tfservingResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*PredictResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}

	for _, val := range resp.Outputs {
		for _, score := range val {
			ret = append(ret, &tfservingResponse{score: float64(score)})
		}
	}
	return
}
