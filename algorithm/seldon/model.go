package seldon

import (
	"github.com/alibaba/pairec/v2/recconf"
)

type Model struct {
	request ISeldonRequest
}

func (r *Model) Init(conf *recconf.AlgoConfig) error {
	req := Request{}
	req.SetUrl(conf.SeldonConf.Url)
	req.SetResponseFunc(conf.SeldonConf.ResponseFuncName)
	r.request = &req
	return nil
}

func (r *Model) Run(algoData interface{}) (interface{}, error) {
	data, err := r.request.Invoke(algoData)
	if err != nil {
		return nil, err
	}

	if r.request.GetResponseFunc() != nil {
		return r.request.GetResponseFunc()(data)
	}

	return data, err
}
