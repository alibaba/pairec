package tfserving

import (
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

type TFservingModel struct {
	retryTimes int
	name       string
	request    ITFservingRequest
}

func NewTFservingModel(name string) *TFservingModel {
	return &TFservingModel{name: name, retryTimes: 2}
}
func (m *TFservingModel) Init(conf *recconf.AlgoConfig) error {
	req := TFservingRequest{}
	req.SetUrl(conf.TFservingConf.Url)
	req.SetTimeout(conf.TFservingConf.Timeout)
	req.SetResponseFunc(conf.TFservingConf.ResponseFuncName)
	m.request = &req

	if conf.TFservingConf.RetryTimes > 0 {
		m.retryTimes = conf.TFservingConf.RetryTimes
	}

	return nil
}
func (m *TFservingModel) Run(algoData interface{}) (interface{}, error) {
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
