package feature

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm"
)

type Normalizer interface {
	Apply(value interface{}) interface{}
}

func NewNormalizer(name, expression string) Normalizer {

	var normalize Normalizer
	if name == "hour_in_day" {
		normalize = &CreateHourNormalizer{}
	} else if name == "weekday" {
		normalize = &CreateDayNormalizer{}
	} else if name == "random" {
		normalize = NewCreateRandomNormalizer()
	} else if name == "const_value" {
		normalize = NewCreateConstValueNormalizer()
	} else if name == "expression" {
		normalize = NewExpressionNormalizer(expression)
	} else if name == "month" {
		normalize = &CreateMonthNormalizer{}
	} else if name == "week" {
		normalize = &CreateWeekNormalizer{}
	} else if name == "expr" {
		normalize = NewExprNormalizer(expression)
	}

	return normalize
}

type CreateHourNormalizer struct {
}

func (n *CreateHourNormalizer) Apply(value interface{}) interface{} {
	return time.Now().Hour()
}

type CreateDayNormalizer struct {
}

func (n *CreateDayNormalizer) Apply(value interface{}) interface{} {
	switch time.Now().Weekday() {
	case time.Monday:
		return int(0)
	case time.Tuesday:
		return int(1)
	case time.Wednesday:
		return int(2)
	case time.Thursday:
		return int(3)
	case time.Friday:
		return int(4)
	case time.Saturday:
		return int(5)
	default:
		return int(6)
	}
}

type CreateMonthNormalizer struct {
}

func (n *CreateMonthNormalizer) Apply(value interface{}) interface{} {
	return int(time.Now().Month())
}

type CreateWeekNormalizer struct {
}

func (n *CreateWeekNormalizer) Apply(value interface{}) interface{} {
	_, week := time.Now().ISOWeek()
	return week
}

type CreateRandomNormalizer struct {
}

func NewCreateRandomNormalizer() *CreateRandomNormalizer {
	rand.Seed(time.Now().UnixNano())
	return &CreateRandomNormalizer{}

}
func (n *CreateRandomNormalizer) Apply(value interface{}) interface{} {
	return rand.Intn(100)
}

type CreateConstValueNormalizer struct {
}

func NewCreateConstValueNormalizer() *CreateConstValueNormalizer {
	return &CreateConstValueNormalizer{}

}
func (n *CreateConstValueNormalizer) Apply(value interface{}) interface{} {
	return nil
}

type ExpressionNormalizer struct {
	evaluableExpression *govaluate.EvaluableExpression
}

func NewExpressionNormalizer(expression string) *ExpressionNormalizer {
	normalizer := &ExpressionNormalizer{}
	goExpression, err := govaluate.NewEvaluableExpressionWithFunctions(expression, utils.GovaluateFunctions())
	if err == nil {
		normalizer.evaluableExpression = goExpression
	} else {
		log.Error(fmt.Sprintf("event=ExpressionNormalizer\texpression=%s\terror=%v", expression, err))
	}

	return normalizer
}
func (n *ExpressionNormalizer) Apply(value interface{}) interface{} {
	if n.evaluableExpression == nil {
		return ""
	}

	if params, ok := value.(map[string]interface{}); ok {
		if result, err := n.evaluableExpression.Evaluate(params); err == nil {
			return result
		} else {
			log.Error(fmt.Sprintf("event=ExpressionNormalizer\texpression=%s\tparams={%s}\terror=%v", n.evaluableExpression.String(), describeExprParams(n.evaluableExpression.Vars(), params), err))
		}

	}

	return ""
}

// describeExprParams returns a human-readable description of the params that the
// expression actually references. It flags problematic values such as nil, NaN,
// Inf or missing keys, which are the common causes of evaluation errors.
func describeExprParams(vars []string, params map[string]interface{}) string {
	var sb strings.Builder
	for i, name := range vars {
		if i > 0 {
			sb.WriteString(", ")
		}
		if v, ok := params[name]; ok {
			sb.WriteString(fmt.Sprintf("%s=%s", name, describeParamValue(v)))
		} else {
			sb.WriteString(fmt.Sprintf("%s=<missing>", name))
		}
	}
	return sb.String()
}

// exprReferencedVars extracts the variable identifiers referenced by a compiled
// expression, excluding function callees. It lets ExprNormalizer log only the
// params the expression actually uses, avoiding dumping unrelated (potentially
// sensitive) values.
func exprReferencedVars(program *vm.Program) []string {
	if program == nil {
		return nil
	}
	collector := &exprVarCollector{
		idents: make(map[string]struct{}),
		funcs:  make(map[string]struct{}),
	}
	node := program.Node()
	if node == nil {
		return nil
	}
	ast.Walk(&node, collector)

	vars := make([]string, 0, len(collector.idents))
	for name := range collector.idents {
		if _, isFunc := collector.funcs[name]; isFunc {
			continue
		}
		vars = append(vars, name)
	}
	return vars
}

// exprVarCollector walks an expression AST, collecting identifier names and the
// names used as function callees so the latter can be excluded from vars.
type exprVarCollector struct {
	idents map[string]struct{}
	funcs  map[string]struct{}
}

func (c *exprVarCollector) Visit(node *ast.Node) {
	switch n := (*node).(type) {
	case *ast.IdentifierNode:
		c.idents[n.Value] = struct{}{}
	case *ast.CallNode:
		if callee, ok := n.Callee.(*ast.IdentifierNode); ok {
			c.funcs[callee.Value] = struct{}{}
		}
	}
}

// describeParamValue renders a single param value, highlighting nil/NaN/Inf.
func describeParamValue(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	var f float64
	switch n := v.(type) {
	case float64:
		f = n
	case float32:
		f = float64(n)
	default:
		return fmt.Sprintf("%v", v)
	}
	if math.IsNaN(f) {
		return "<NaN>"
	}
	if math.IsInf(f, 1) {
		return "<+Inf>"
	}
	if math.IsInf(f, -1) {
		return "<-Inf>"
	}
	return fmt.Sprintf("%v", f)
}

type ExprNormalizer struct {
	prog       *vm.Program
	expression string
	vars       []string
}

func NewExprNormalizer(expression string) *ExprNormalizer {
	normalizer := &ExprNormalizer{expression: expression}

	options := append([]expr.Option{expr.AllowUndefinedVariables()}, utils.ExprFunctions()...)
	if program, err := expr.Compile(expression, options...); err != nil {
		log.Error(fmt.Sprintf("event=ExprNormalizer\texpression=%s\terr=%v", expression, err))
	} else {
		normalizer.prog = program
		normalizer.vars = exprReferencedVars(program)
	}
	return normalizer
}
func (n *ExprNormalizer) Apply(value interface{}) interface{} {
	if n.prog == nil {
		return ""
	}

	if params, ok := value.(map[string]interface{}); ok {
		if result, err := expr.Run(n.prog, params); err == nil {
			return result
		} else {
			log.Error(fmt.Sprintf("event=ExprNormalizer\texpression=%s\tparams={%s}\terror=%v", n.expression, describeExprParams(n.vars, params), err))
		}
	}

	return ""
}
