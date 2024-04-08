package eas

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/config"
	proto "github.com/golang/protobuf/proto"
)

type HttpEasyrecRequest struct {
	EasRequest
}

func (r *HttpEasyrecRequest) Invoke(requestData interface{}) (response interface{}, err error) {
	request, ok := requestData.(*easyrec.PBRequest)
	if !ok {
		err = errors.New("requestData is not easyrec.PBRequest type")
		return
	}

	data, _ := proto.Marshal(request)
	if config.AppConfig.WarmUpData {
		warmupFunc := func(data []byte) {
			if file, err := os.OpenFile("warm_up.bin", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664); err == nil {
				file.WriteString(base64.StdEncoding.EncodeToString(data))
				file.Close()
			}
		}
		config.AppConfig.Once.Do(func() { warmupFunc(data) })
	}
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if r.responseFuncName != "" && strings.HasPrefix(r.responseFuncName, "torchrec") {
		responseData := &easyrec.TorchRecPBResponse{}
		err = proto.Unmarshal(body, responseData)
		if err != nil {
			err = fmt.Errorf("error:%s, body:%s", err.Error(), string(body))
			return
		}

		responseData.ItemIds = request.ItemIds
		response = responseData
		return

	}
	responseData := &easyrec.PBResponse{}
	err = proto.Unmarshal(body, responseData)
	if err != nil {
		err = fmt.Errorf("error:%s, body:%s", err.Error(), string(body))
		return
	}

	if responseData.StatusCode != easyrec.StatusCode_OK {
		err = errors.New(responseData.ErrorMsg)
		return
	}

	responseData.ItemIds = request.ItemIds
	response = responseData
	return
}
