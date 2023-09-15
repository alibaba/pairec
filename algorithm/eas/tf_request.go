package eas

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	proto "github.com/golang/protobuf/proto"
	"github.com/alibaba/pairec/pkg/eas/types/tf_predict_protos"
)

type TFRequest struct {
	EasRequest
	SignatureName string
	Outputs       []string
}

func (r *TFRequest) SetSignatureName(name string) {
	r.SignatureName = name
}
func (r *TFRequest) SetOutputs(outputs []string) {
	r.Outputs = outputs
}
func (r *TFRequest) Invoke(requestData interface{}) (response interface{}, err error) {
	request, ok := requestData.(*tf_predict_protos.PredictRequest)
	if !ok {
		err = errors.New("requestData is not tf.PredictRequest type")
		return
	}
	request.SignatureName = r.SignatureName
	request.OutputFilter = r.Outputs
	data, _ := proto.Marshal(request)
	req, err := http.NewRequest("POST", r.url, bytes.NewBuffer(data))
	if err != nil {
		return
	}

	headers := map[string][]string{
		"Authorization": {r.auth},
	}
	req.Header = headers

	easClient.Timeout = r.timeout

	resp, err := easClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	responseData := &tf_predict_protos.PredictResponse{}
	err = proto.Unmarshal(body, responseData)
	if err != nil {
		err = fmt.Errorf("error:%s, body:%s", err.Error(), string(body))
		return
	}

	response = responseData
	return
}
