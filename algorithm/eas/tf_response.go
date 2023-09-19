package eas

import (
	"fmt"

	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/pkg/eas/types/tf_predict_protos"
)

type tfResponse struct {
	score          float64
	scoreArr       map[string]float64
	multiValModule bool
}

func (r *tfResponse) GetScore() float64 {
	return r.score
}

func (r *tfResponse) GetScoreMap() map[string]float64 {
	return r.scoreArr
}

func (r *tfResponse) GetModuleType() bool {
	return r.multiValModule
}

func tfMutValResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*tf_predict_protos.PredictResponse)
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

func tfResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*tf_predict_protos.PredictResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	for _, arrayProto := range resp.GetOutputs() {
		for _, val := range arrayProto.FloatVal {
			ret = append(ret, &tfResponse{score: float64(val)})
		}
		break
	}
	return
}
