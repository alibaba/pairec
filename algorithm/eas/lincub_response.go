package eas

import (
	json "encoding/json"
	"fmt"
)

type LinucbResponse struct {
	ErrorCode    int                   `json:"error_code"`
	ErrorMessage string                `json:"error_message"`
	Data         []*LinucbResponseItem `json:"data"`
}

type LinucbResponseItem struct {
	ItemId string  `json:"id"`
	Score  float64 `json:"score"`
}

func (r *LinucbResponseItem) GetModuleType() bool {
	return false
}

func (r *LinucbResponseItem) GetScoreMap() map[string]float64 {
	var default_val map[string]float64
	return default_val
}

func (r *LinucbResponseItem) GetScore() float64 {
	return r.Score
}

func LinucbResponseFunc(data interface{}) (ret []*LinucbResponseItem, err error) {
	retstr, ok := data.(string)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	response := LinucbResponse{}
	err = json.Unmarshal([]byte(retstr), &response)
	if err != nil {
		err = fmt.Errorf("error:%v, body:%s", err, bodyFormat(retstr, 512))
		return
	}
	if response.ErrorCode != 0 {
		err = fmt.Errorf("error:%s, body:%s", response.ErrorMessage, bodyFormat(retstr, 512))
		return
	}

	ret = response.Data

	return
}
