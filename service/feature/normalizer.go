package feature

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/utils"
	"github.com/expr-lang/expr"
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
		log.Error(fmt.Sprintf("event=ExpressionNormalizer\terror=%v", err))
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
			log.Error(fmt.Sprintf("event=ExpressionNormalizer\terror=%v", err))
		}

	}

	return ""
}

type ExprNormalizer struct {
	prog *vm.Program
}

func NewExprNormalizer(expression string) *ExprNormalizer {
	normalizer := &ExprNormalizer{}

	options := append([]expr.Option{expr.AllowUndefinedVariables()}, utils.ExprFunctions()...)
	if program, err := expr.Compile(expression, options...); err != nil {
		log.Error(fmt.Sprintf("event=ExprNormalizer\terr=%v", err))
	} else {
		normalizer.prog = program
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
			log.Error(fmt.Sprintf("event=ExprNormalizer\terror=%v", err))
		}
	}

	return ""
}
