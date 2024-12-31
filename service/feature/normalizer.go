package feature

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/utils"
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

var (
	functions = map[string]govaluate.ExpressionFunction{
		"getString": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return "", errors.New("args should not empty")
			}
			if args[0] != "" {
				return args[0], nil
			}
			if len(args) > 1 {
				return args[1], nil
			}
			return "", nil
		},
		"trim": func(args ...interface{}) (interface{}, error) {
			if len(args) != 2 {
				return "", errors.New("args length not equal 2")
			}

			str := utils.ToString(args[0], "")
			cutset := utils.ToString(args[1], "")
			fmt.Println(strings.TrimPrefix(str, cutset), str, cutset)
			return strings.Trim(str, cutset), nil
		},
		"trimPrefix": func(args ...interface{}) (interface{}, error) {
			if len(args) != 2 {
				return "", errors.New("args length not equal 2")
			}

			str := utils.ToString(args[0], "")
			cutset := utils.ToString(args[1], "")
			return strings.TrimPrefix(str, cutset), nil
		},
	}
)

func NewExpressionNormalizer(expression string) *ExpressionNormalizer {
	normalizer := &ExpressionNormalizer{}
	goExpression, err := govaluate.NewEvaluableExpressionWithFunctions(expression, functions)
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
