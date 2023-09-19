package eas

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	proto "github.com/golang/protobuf/proto"
	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/config"
	"github.com/alibaba/pairec/v2/pkg/eas"
)

type EasyrecRequest struct {
	EasRequest
	EasClient *eas.PredictClient
}

func (r *EasyrecRequest) Invoke(requestData interface{}) (response interface{}, err error) {
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

	body, err := r.EasClient.BytesPredict(data)
	if err != nil {
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
