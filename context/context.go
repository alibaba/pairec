package context

import "github.com/alibaba/pairec/v2/recconf"

type Context struct {
	RequestId string
	Param     IParam
	Config    *recconf.RecommendConfig
}

func NewContext() *Context {
	context := Context{}
	return &context
}
func (r *Context) GetParameter(name string) interface{} {
	return r.Param.GetParameter(name)
}
