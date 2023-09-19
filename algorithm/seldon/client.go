package seldon

import (
	"github.com/alibaba/pairec/v2/algorithm/response"
	"net"
	"net/http"
	"time"
)

var seldonClient *http.Client

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

	seldonClient = &http.Client{Transport: tr}
}

type ISeldonRequest interface {
	Invoke(requestData interface{}) (body interface{}, err error)
	GetResponseFunc() response.ResponseFunc
}
