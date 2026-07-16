package eas

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/config"
	"github.com/alibaba/pairec/v2/pkg/eas"
	proto "github.com/golang/protobuf/proto"
)

// marshalBufPool reuses proto.Buffer to reduce marshal allocations
var marshalBufPool = sync.Pool{
	New: func() interface{} {
		return proto.NewBuffer(make([]byte, 0, 4096))
	},
}

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

	buf := marshalBufPool.Get().(*proto.Buffer)
	buf.Reset()
	if err = buf.Marshal(request); err != nil {
		marshalBufPool.Put(buf)
		return
	}
	data := buf.Bytes()

	if config.AppConfig.WarmUpData {
		warmupFunc := func(data []byte) {
			if file, err := os.OpenFile("warm_up.bin", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0664); err == nil {
				file.Write(data)
				file.Close()
			}
		}
		config.AppConfig.Once.Do(func() { warmupFunc(data) })
	}

	body, err := r.EasClient.BytesPredict(data)
	// Only return normal-sized buffers to pool; let oversized ones be GC'd
	if cap(buf.Bytes()) <= 1<<20 {
		marshalBufPool.Put(buf)
	}
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

		if r.responseFuncName != "torchrecEmbeddingItemsResponseFunc" {
			responseData.ItemIds = request.ItemIds
		}

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
