package eas

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/golang/snappy"
	"github.com/alibaba/pairec/utils/jsonutil"
)

type FMRequest struct {
	EasRequest
	pool *sync.Pool
}

func NewFMRequest() *FMRequest {
	return &FMRequest{
		pool: &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, 40960)
				return bytes.NewBuffer(buf)
			},
		},
	}
}
func (r *FMRequest) Invoke(requestData interface{}) (body interface{}, err error) {
	//data, _ := json.Marshal(requestData)
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic error:%v", e)
		}
	}()

	buf := r.pool.Get().(*bytes.Buffer)
	buf.Reset()
	if err = jsonutil.MarshalSliceWithByteBuffer(requestData, buf); err != nil {
		return
	}

	dst := snappy.Encode(nil, buf.Bytes())
	buf.Reset()
	buf.Write(dst)

	defer r.pool.Put(buf)

	req, err := http.NewRequest("POST", r.url, buf)
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
