package eas

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alibaba/pairec/log"
	"github.com/alibaba/pairec/pkg/eas"
	tensorflow_serving "github.com/alibaba/pairec/pkg/tensorflow_serving/apis"
	"github.com/alibaba/pairec/recconf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	Eas_Processor_FM        = "ALINK_FM"
	Eas_Processor_PMML      = "PMML"
	Eas_Processor_TF        = "TensorFlow"
	Eas_Processor_TFServing = "TFServing"
	Eas_Processor_EASYREC   = "EasyRec"
	Eas_Processor_LINUCB    = "Linucb"
)

type EasModel struct {
	retryTimes int
	name       string
	request    IEasRequest
}

func NewEasModel(name string) *EasModel {
	return &EasModel{name: name, retryTimes: 2}
}
func (m *EasModel) Init(conf *recconf.AlgoConfig) error {
	if conf.EasConf.Processor == Eas_Processor_PMML {
		req := PMMLRequest{}
		req.SetUrl(conf.EasConf.Url)
		req.SetAuth(conf.EasConf.Auth)
		req.SetTimeout(conf.EasConf.Timeout)
		req.SetResponseFunc(conf.EasConf.ResponseFuncName)
		m.request = &req

		if conf.EasConf.RetryTimes > 0 {
			m.retryTimes = conf.EasConf.RetryTimes
		}
		return nil
	} else if conf.EasConf.Processor == Eas_Processor_TF {
		req := TFRequest{}
		req.SetUrl(conf.EasConf.Url)
		req.SetAuth(conf.EasConf.Auth)
		req.SetSignatureName(conf.EasConf.SignatureName)
		req.SetTimeout(conf.EasConf.Timeout)
		req.SetResponseFunc(conf.EasConf.ResponseFuncName)
		if len(conf.EasConf.Outputs) > 0 {
			req.SetOutputs(conf.EasConf.Outputs)
		}
		m.request = &req

		if conf.EasConf.RetryTimes > 0 {
			m.retryTimes = conf.EasConf.RetryTimes
		}
		return nil
	} else if conf.EasConf.Processor == Eas_Processor_FM {
		req := NewFMRequest()
		req.SetUrl(conf.EasConf.Url)
		req.SetAuth(conf.EasConf.Auth)
		req.SetTimeout(conf.EasConf.Timeout)
		req.SetResponseFunc(conf.EasConf.ResponseFuncName)
		m.request = req

		if conf.EasConf.RetryTimes > 0 {
			m.retryTimes = conf.EasConf.RetryTimes
		}
		return nil

	} else if conf.EasConf.Processor == Eas_Processor_EASYREC && conf.EasConf.EndpointType == eas.EndpointTypeDocker {
		req := HttpEasyrecRequest{}
		req.SetUrl(conf.EasConf.Url)
		req.SetAuth(conf.EasConf.Auth)
		req.SetTimeout(conf.EasConf.Timeout)
		req.SetResponseFunc(conf.EasConf.ResponseFuncName)
		m.request = &req

		if conf.EasConf.RetryTimes > 0 {
			m.retryTimes = conf.EasConf.RetryTimes
		}
		return nil
	} else if conf.EasConf.Processor == Eas_Processor_EASYREC {
		req := EasyrecRequest{}
		req.SetUrl(conf.EasConf.Url)
		req.SetAuth(conf.EasConf.Auth)
		req.SetTimeout(conf.EasConf.Timeout)
		req.SetResponseFunc(conf.EasConf.ResponseFuncName)

		var client *eas.PredictClient
		if conf.EasConf.EndpointType == eas.EndpointTypeDirect {
			url := strings.ReplaceAll(conf.EasConf.Url, "http://", "")
			index := strings.Index(url, "/api/predict/")
			endpoint := url[:index]
			strs := strings.Split(endpoint, ".")
			region := strs[2]
			for i := 0; i < len(strs); i++ {
				if strs[i] == "pai-eas" {
					region = strs[i-1]
					break
				}
			}
			name := url[index+len("/api/predict/"):]
			client = eas.NewPredictClient(fmt.Sprintf("pai-eas-vpc.%s.aliyuncs.com", region), name)
			client.SetEndpointType(eas.EndpointTypeDirect)
		} else {
			url := strings.ReplaceAll(conf.EasConf.Url, "http://", "")
			index := strings.Index(url, "/api/predict/")
			endpoint := url[:index]
			name := url[index+len("/api/predict/"):]
			client = eas.NewPredictClient(endpoint, name)
		}
		client.SetToken(conf.EasConf.Auth)
		client.SetTimeout(conf.EasConf.Timeout)
		client.SetHttpTransport(&http.Transport{
			MaxConnsPerHost:       300,
			MaxIdleConnsPerHost:   300,
			MaxIdleConns:          300,
			TLSHandshakeTimeout:   100 * time.Millisecond,
			ExpectContinueTimeout: 200 * time.Millisecond,
			DialContext: (&net.Dialer{
				Timeout:   100 * time.Millisecond, // 100ms
				KeepAlive: 10 * time.Minute,
			}).DialContext,
		})
		if conf.EasConf.RetryTimes > 0 {
			m.retryTimes = conf.EasConf.RetryTimes
		}
		client.SetRetryCount(m.retryTimes - 1)
		// if use eas sdk, we should not retry
		m.retryTimes = 1
		client.Init()
		req.EasClient = client
		m.request = &req

		return nil
	} else if conf.EasConf.Processor == Eas_Processor_LINUCB {
		req := LincubRequest{}
		req.SetUrl(conf.EasConf.Url)
		req.SetAuth(conf.EasConf.Auth)
		req.SetTimeout(conf.EasConf.Timeout)
		req.SetResponseFunc(conf.EasConf.ResponseFuncName)
		m.request = &req

		if conf.EasConf.RetryTimes > 0 {
			m.retryTimes = conf.EasConf.RetryTimes
		}
		return nil

	} else if conf.EasConf.Processor == Eas_Processor_TFServing {
		req := TFServingRequest{}
		req.SetUrl(conf.EasConf.Url)
		req.SetModelName(conf.EasConf.ModelName)
		req.SetAuth(conf.EasConf.Auth)
		req.SetSignatureName(conf.EasConf.SignatureName)
		req.SetTimeout(conf.EasConf.Timeout)
		req.SetResponseFunc(conf.EasConf.ResponseFuncName)
		if len(conf.EasConf.Outputs) > 0 {
			req.SetOutputs(conf.EasConf.Outputs)
		}

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

		conn, err := grpc.Dial(req.url, opts...)
		if err != nil {
			panic(fmt.Sprintf("fail to dial: %v", err))
		}

		client := tensorflow_serving.NewPredictionServiceClient(conn)
		req.Client = client
		m.request = &req

		if conf.EasConf.RetryTimes > 0 {
			m.retryTimes = conf.EasConf.RetryTimes
		}
		return nil
	}

	return errors.New("not found eas Processor:" + conf.EasConf.Processor)
}
func (m *EasModel) Run(algoData interface{}) (interface{}, error) {
	retryTimes := m.retryTimes

	var (
		data interface{}
		err  error
	)

	for {
		retryTimes--
		data, err = m.request.Invoke(algoData)
		if err != nil && retryTimes == 0 {
			return data, err
		} else if err == nil {
			break
		} else if err != nil {
			log.Warning("eas request fail :" + err.Error())
		}
	}

	if m.request.GetResponseFunc() != nil {
		return m.request.GetResponseFunc()(data)
	}

	return data, err
}
