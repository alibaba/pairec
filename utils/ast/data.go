package ast

type ParameterExprData interface {
	FloatExprData(string) (float64, error)
	ExprData() map[string]interface{}
}
