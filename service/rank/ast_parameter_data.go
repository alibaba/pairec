package rank

import (
	"github.com/alibaba/pairec/v2/context"
	"github.com/alibaba/pairec/v2/module"
	"github.com/alibaba/pairec/v2/utils/ast"
)

var _ ast.ParameterExprData = AstParameterData{}

type AstParameterData struct {
	context *context.RecommendContext
	item    *module.Item
}

// ExprData implements ast.ParameterExprData.
func (a AstParameterData) ExprData() map[string]interface{} {
	if a.context.ExperimentResult != nil {
		params := a.context.ExperimentResult.GetExperimentParams().ListParams()
		itemParams := a.item.ExprData()
		for k, v := range itemParams {
			params[k] = v
		}
		return params
	}
	return a.item.ExprData()
}

// FloatExprData implements ast.ParameterExprData.
func (a AstParameterData) FloatExprData(param string) (float64, error) {
	if a.context.ExperimentResult != nil {
		ret := a.context.ExperimentResult.GetExperimentParams().GetFloat(param, 0)
		if ret == 0 {
			return a.item.FloatExprData(param)
		} else {
			return ret, nil
		}
	}
	return a.item.FloatExprData(param)
}
