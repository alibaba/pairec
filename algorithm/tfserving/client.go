package tfserving

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/response"
)

var tfservingClient *http.Client

func init() {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   100 * time.Millisecond, // 100ms
			KeepAlive: 5 * time.Minute,
		}).DialContext,
		MaxIdleConnsPerHost:   200,
		MaxIdleConns:          200,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}

	tfservingClient = &http.Client{Transport: tr}
}

type ITFservingRequest interface {
	Invoke(requestData interface{}) (body interface{}, err error)
	GetResponseFunc() response.ResponseFunc
}

type TFservingRequest struct {
	url          string
	timeout      time.Duration
	responseFunc response.ResponseFunc

	SignatureName string
	Outputs       []string
}

func (r *TFservingRequest) SetUrl(url string) {
	r.url = url
}

func (r *TFservingRequest) SetTimeout(timeout int) {
	if timeout <= 0 {
		r.timeout = 100 * time.Millisecond
	} else {
		r.timeout = time.Millisecond * time.Duration(timeout)
	}
}

func (r *TFservingRequest) SetSignatureName(name string) {
	r.SignatureName = name
}
func (r *TFservingRequest) SetOutputs(outputs []string) {
	r.Outputs = outputs
}

func (r *TFservingRequest) SetResponseFunc(name string) {
	if name == "tfservingResponseFunc" {
		r.responseFunc = tfservingResponseFunc
	}
}

func (r *TFservingRequest) GetResponseFunc() response.ResponseFunc {
	return r.responseFunc
}

func (r *TFservingRequest) Invoke(requestData interface{}) (response interface{}, err error) {
	request, ok := requestData.(*PredictRequest)
	if !ok {
		err = errors.New("requestData is not PredictRequest type")
		return
	}
	request.SignatureName = r.SignatureName
	request.OutputFilter = r.Outputs
	data, _ := json.Marshal(request)

	req, err := http.NewRequest("POST", r.url, bytes.NewBuffer(data))
	if err != nil {
		return
	}

	tfservingClient.Timeout = r.timeout

	resp, err := tfservingClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	responseData := &PredictResponse{}
	err = json.Unmarshal(body, responseData)
	if err != nil {
		err = errors.New(fmt.Sprintf("error:%s, body:%s", err.Error(), string(body)))
		return
	}

	response = responseData
	return
}
