package seldon

import (
	"bytes"
	"encoding/json"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"io/ioutil"
	"net/http"
	"strconv"
)

type Request struct {
	url          string
	responseFunc response.ResponseFunc
}

func (r *Request) SetUrl(url string) {
	r.url = url
}

func (r *Request) SetResponseFunc(name string) {
	r.responseFunc = ResponseFunc
}

func (r *Request) GetResponseFunc() response.ResponseFunc {
	return r.responseFunc
}

type PredictRequest struct {
	ReqJsonData ReqJsonData `json:"jsonData"`
}

type ReqJsonData struct {
	Inputs map[string][]interface{} `json:"inputs"`
}

func (r *Request) Invoke(requestData interface{}) (body interface{}, err error) {
	features := make(map[string][]interface{})
	if data, ok := requestData.([]map[string]interface{}); ok {
		for _, d := range data {
			for k, v := range d {
				if features[k] == nil {
					features[k] = make([]interface{}, 0)
				}
				switch value := v.(type) {
				case int:
					v = strconv.Itoa(value)
				case float64:
					v = strconv.FormatFloat(value, 'f', -1, 64)
				}
				features[k] = append(features[k], v)
			}
		}
	}

	predictRequest := PredictRequest{ReqJsonData: ReqJsonData{Inputs: features}}
	reqData, _ := json.Marshal(predictRequest)

	header := map[string][]string{
		"Content-Type": {"application/json"},
	}

	req, err := http.NewRequest("POST", r.url, bytes.NewBuffer(reqData))
	if err != nil {
		return
	}
	req.Header = header
	resp, err := seldonClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return body, nil
}
