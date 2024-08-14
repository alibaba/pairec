package seldon

import (
	"encoding/json"
	"github.com/alibaba/pairec/v2/algorithm/response"
)

type Response struct {
	score          float64
	scoreArr       map[string]float64
	multiValModule bool
}

func (r *Response) GetScore() float64 {
	return r.score
}

func (r *Response) GetScoreMap() map[string]float64 {
	return r.scoreArr
}

func (r *Response) GetModuleType() bool {
	return r.multiValModule
}

type PredictResponse struct {
	JsonData JsonData `json:"jsonData"`
}

type JsonData struct {
	Outputs map[string][]float64 `json:"outputs"`
}

func ResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	var predictResp PredictResponse

	if v, ok := data.([]byte); ok {
		err = json.Unmarshal(v, &predictResp)
		if err != nil {
			return nil, err
		}
	}

	retLen := 0
	for _, scores := range predictResp.JsonData.Outputs {
		if len(scores) > retLen {
			retLen = len(scores)
		}
	}

	for i := 0; i < retLen; i++ {
		scoreArr := make(map[string]float64)
		for property, scores := range predictResp.JsonData.Outputs {
			if i < len(scores) {
				scoreArr[property] = scores[i]
			}
		}
		ret = append(ret, &Response{
			scoreArr:       scoreArr,
			multiValModule: true,
		})
	}

	return
}
