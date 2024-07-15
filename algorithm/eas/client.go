package eas

import (
	"net"
	"net/http"
	"time"

	"github.com/alibaba/pairec/v2/algorithm/response"
)

var easClient *http.Client

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

	easClient = &http.Client{Transport: tr}
}

type IEasRequest interface {
	Invoke(requestData interface{}) (body interface{}, err error)
	GetResponseFunc() response.ResponseFunc
}
type EasRequest struct {
	serviceName      string
	auth             string
	url              string
	timeout          time.Duration
	responseFunc     response.ResponseFunc
	responseFuncName string
}

func (r *EasRequest) SetUrl(url string) {
	r.url = url
}

func (r *EasRequest) SetServiceName(name string) {
	r.serviceName = name
}
func (r *EasRequest) SetAuth(auth string) {
	r.auth = auth
}

func (r *EasRequest) SetTimeout(timeout int) {
	if timeout <= 0 {
		r.timeout = 100 * time.Millisecond
	} else {
		r.timeout = time.Millisecond * time.Duration(timeout)
	}
}
func (r *EasRequest) SetResponseFunc(name string) {
	if name == "pssmartResponseFunc" {
		r.responseFunc = pssmartResponseFunc
	} else if name == "tfResponseFunc" {
		r.responseFunc = tfResponseFunc
	} else if name == "alinkFMResponseFunc" {
		r.responseFunc = alinkFMResponseFunc
	} else if name == "tfMutValResponseFunc" {
		r.responseFunc = tfMutValResponseFunc
	} else if name == "easyrecResponseFunc" {
		r.responseFunc = easyrecResponseFunc
	} else if name == "easyrecResponseFuncDebug" {
		r.responseFunc = easyrecResponseFuncDebug
	} else if name == "easyrecMutValResponseFunc" {
		r.responseFunc = easyrecMutValResponseFunc
	} else if name == "easyrecMutValResponseFuncDebug" {
		r.responseFunc = easyrecMutValResponseFuncDebug
	} else if name == "easyrecUserEmbResponseFunc" {
		r.responseFunc = easyrecUserEmbResponseFunc
	} else if name == "easyrecUserRealtimeEmbeddingResponseFunc" {
		r.responseFunc = easyrecUserRealtimeEmbeddingResponseFunc
	} else if name == "easyrecUserRealtimeEmbeddingMindResponseFunc" {
		r.responseFunc = easyrecUserRealtimeEmbeddingMindResponseFunc
	} else if name == "tfServingResponseFunc" {
		r.responseFunc = tfServingResponseFunc
	} else if name == "torchrecMutValResponseFunc" {
		r.responseFunc = torchrecMutValResponseFunc
	} else if name == "torchrecMutValResponseFuncDebug" {
		r.responseFunc = torchrecMutValResponseFuncDebug
	} else if name == "torchrecEmbeddingResponseFunc" {
		r.responseFunc = torchrecEmbeddingResponseFunc
	}

	r.responseFuncName = name
}

func (r *EasRequest) GetResponseFunc() response.ResponseFunc {
	return r.responseFunc
}
