package eas

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type LincubRequestData struct {
	RequestId    string                   `json:"request_id"`
	Scene        string                   `json:"scene"`
	AlgoName     string                   `json:"algo"`
	UserId       string                   `json:"user_id"`
	Items        []string                 `json:"items"`
	Limit        int                      `json:"limit"`
	UserFeature  map[string]interface{}   `json:"user_feature"`
	ItemFeatures []map[string]interface{} `json:"item_features"`
}
type LincubRequest struct {
	EasRequest
}

func (r *LincubRequest) Invoke(requestData interface{}) (body interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic error:%v", e)
		}
	}()
	request, ok := requestData.(*LincubRequestData)
	if !ok {
		err = errors.New("requestData is not LincubRequestData type")
		return
	}

	data, _ := json.Marshal(request)

	req, err := http.NewRequest("POST", r.url, bytes.NewBuffer(data))
	if err != nil {
		return
	}

	headers := map[string][]string{
		"Authorization": {r.auth},
		// "Content-Encoding": {"gzip"},
	}
	req.Header = headers

	if easClient.Timeout == 0 {
		easClient.Timeout = r.timeout
	}

	response, err := easClient.Do(req)
	if err != nil {
		return
	}

	defer response.Body.Close()

	reqBody, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}

	body = string(reqBody)

	return
}
