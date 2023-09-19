package module

import (
	"fmt"
	"math"
	"strings"

	"github.com/alibaba/pairec/v2/recconf"
	"github.com/alibaba/pairec/v2/utils"
)

const (
	ITEM = "item"
	USER = "user"
)

type FilterOp interface {
	Evaluate(map[string]interface{}) (bool, error)
	OpDomain() string
}

type EqualFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewEqualFilterOp(config recconf.FilterParamConfig) *EqualFilterOp {

	equalFilterOp := &EqualFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}
	if config.Domain == "" {
		equalFilterOp.Domain = ITEM
	} else {
		equalFilterOp.Domain = config.Domain
	}

	return equalFilterOp
}

func (p *EqualFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "string":
		v1 := utils.ToString(left, "v1")
		v2 := utils.ToString(p.Value, "v2")

		return v1 == v2, nil
	case "int":
		v1 := utils.ToInt(left, -1)
		v2 := utils.ToInt(p.Value, -2)
		return v1 == v2, nil
	case "int64":
		v1 := utils.ToInt64(left, -1)
		v2 := utils.ToInt64(p.Value, -2)
		return v1 == v2, nil
	default:
		return false, nil
	}
}

func (p *EqualFilterOp) OpDomain() string {
	return p.Domain
}

type NotEqualFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewNotEqualFilterOp(config recconf.FilterParamConfig) *NotEqualFilterOp {

	notEqualFilterOp := &NotEqualFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}

	if config.Domain == "" {
		notEqualFilterOp.Domain = ITEM
	} else {
		notEqualFilterOp.Domain = config.Domain
	}

	return notEqualFilterOp
}

func (p *NotEqualFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return true, nil
	}

	switch p.Type {
	case "string":
		v1 := utils.ToString(left, "")
		v2 := utils.ToString(p.Value, "")

		return v1 != v2, nil
	case "int":
		v1 := utils.ToInt(left, 0)
		v2 := utils.ToInt(p.Value, 0)
		return v1 != v2, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		v2 := utils.ToInt64(p.Value, 0)
		return v1 != v2, nil
	default:
		return false, nil
	}
}

func (p *NotEqualFilterOp) OpDomain() string {
	return p.Domain
}

type InFilterOp struct {
	Name          string
	Domain        string
	Type          string
	int_values    []int
	string_values []string
}

func NewInFilterOp(config recconf.FilterParamConfig) *InFilterOp {

	op := &InFilterOp{
		Name: config.Name,
		Type: config.Type,
	}

	if config.Domain == "" {
		op.Domain = ITEM
	} else {
		op.Domain = config.Domain
	}

	values, ok := config.Value.([]interface{})
	if !ok {
		panic("InFilterOp type error")

	}
	switch op.Type {
	case "int":
		for _, val := range values {
			switch value := val.(type) {
			case float64:
				op.int_values = append(op.int_values, int(value))
			case int:
				op.int_values = append(op.int_values, value)
			case int32:
				op.int_values = append(op.int_values, int(value))
			case int64:
				op.int_values = append(op.int_values, int(value))
			}
		}

	case "string":
		for _, val := range values {
			if value, ok := val.(string); ok {
				op.string_values = append(op.string_values, value)
			}
		}
	}

	return op
}

func (p *InFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "string":
		v1 := utils.ToString(left, "")
		for _, val := range p.string_values {
			if v1 == val {
				return true, nil
			}
		}
	case "int":
		v1 := utils.ToInt(left, math.MinInt32)
		for _, val := range p.int_values {
			if v1 == val {
				return true, nil
			}
		}
	}
	return false, nil
}

func (p *InFilterOp) OpDomain() string {
	return p.Domain
}

type FilterParam struct {
	operators []FilterOp
}

func NewFilterParamWithConfig(configs []recconf.FilterParamConfig) *FilterParam {
	p := FilterParam{
		operators: make([]FilterOp, 0, 4),
	}

	for _, config := range configs {
		if config.Operator == "equal" {
			p.operators = append(p.operators, NewEqualFilterOp(config))
		} else if config.Operator == "not_equal" {
			p.operators = append(p.operators, NewNotEqualFilterOp(config))
		} else if config.Operator == "in" {
			p.operators = append(p.operators, NewInFilterOp(config))
		} else if config.Operator == "not_in" {
			p.operators = append(p.operators, NewNotInFilterOp(config))
		} else if config.Operator == "greater" {
			p.operators = append(p.operators, NewGreaterFilterOp(config))
		} else if config.Operator == "greaterThan" {
			p.operators = append(p.operators, NewGreaterThanFilterOp(config))
		} else if config.Operator == "less" {
			p.operators = append(p.operators, NewLessFilterOp(config))
		} else if config.Operator == "lessThan" {
			p.operators = append(p.operators, NewLessThanFilterOp(config))
		} else if config.Operator == "contains" {
			p.operators = append(p.operators, NewContainsFilterOp(config))
		} else if config.Operator == "not_contains" {
			p.operators = append(p.operators, NewNotContainsFilterOp(config))
		}
	}

	return &p
}

func (p *FilterParam) Evaluate(properties map[string]interface{}) (bool, error) {
	for _, op := range p.operators {
		ret, err := op.Evaluate(properties)
		if !ret || err != nil {
			return false, err
		}
	}
	return true, nil
}

func (p *FilterParam) EvaluateByDomain(userProperties, itemProperties map[string]interface{}) (bool, error) {
	for _, op := range p.operators {
		if containsOp, ok := op.(*ContainsFilterOp); ok {
			if containsOp.OpDomain() == ITEM {
				ret, err := containsOp.ContainsEvaluate(itemProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
			} else if containsOp.OpDomain() == USER {
				ret, err := containsOp.ContainsEvaluate(userProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
			} else {
				return false, fmt.Errorf("not support this domain:%s", op.OpDomain())
			}

		} else if notContainsOp, ok := op.(*NotContainsFilterOp); ok {
			if notContainsOp.OpDomain() == ITEM {
				ret, err := notContainsOp.ContainsEvaluate(itemProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
			} else if notContainsOp.OpDomain() == USER {
				ret, err := notContainsOp.ContainsEvaluate(userProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
			} else {
				return false, fmt.Errorf("not support this domain:%s", op.OpDomain())
			}
		} else if notInOp, ok := op.(*NotInFilterOp); ok {
			if notInOp.OpDomain() == ITEM {
				ret, err := notInOp.NotInEvaluate(itemProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
			} else if notInOp.OpDomain() == USER {
				ret, err := notInOp.NotInEvaluate(userProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
			} else {
				return false, fmt.Errorf("not support this domain:%s", op.OpDomain())
			}

		} else {
			if op.OpDomain() == ITEM {
				ret, err := op.Evaluate(itemProperties)
				if ret == false || err != nil {
					return false, err
				}
			} else if op.OpDomain() == USER {
				ret, err := op.Evaluate(userProperties)
				if ret == false || err != nil {
					return false, err
				}
			} else {
				return false, fmt.Errorf("not support this domain:%s", op.OpDomain())
			}

		}
	}
	return true, nil
}

type GreaterFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewGreaterFilterOp(config recconf.FilterParamConfig) *GreaterFilterOp {

	greaterFilterOp := &GreaterFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}

	if config.Domain == "" {
		greaterFilterOp.Domain = ITEM
	} else {
		greaterFilterOp.Domain = config.Domain
	}

	return greaterFilterOp
}

func (p *GreaterFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, 0)
		v2 := utils.ToFloat(p.Value, 0)

		return v1 > v2, nil
	case "int":
		v1 := utils.ToInt(left, 0)
		v2 := utils.ToInt(p.Value, 0)
		return v1 > v2, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		v2 := utils.ToInt64(p.Value, 0)
		return v1 > v2, nil
	default:
		return false, nil
	}
}

func (p *GreaterFilterOp) OpDomain() string {
	return p.Domain
}

type GreaterThanFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewGreaterThanFilterOp(config recconf.FilterParamConfig) *GreaterThanFilterOp {

	greaterThanFilterOp := &GreaterThanFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}

	if config.Domain == "" {
		greaterThanFilterOp.Domain = ITEM
	} else {
		greaterThanFilterOp.Domain = config.Domain
	}

	return greaterThanFilterOp
}

func (p *GreaterThanFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, 0)
		v2 := utils.ToFloat(p.Value, 0)

		return v1 >= v2, nil
	case "int":
		v1 := utils.ToInt(left, 0)
		v2 := utils.ToInt(p.Value, 0)
		return v1 >= v2, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		v2 := utils.ToInt64(p.Value, 0)
		return v1 >= v2, nil
	default:
		return false, nil
	}
}

func (p *GreaterThanFilterOp) OpDomain() string {
	return p.Domain
}

type LessFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewLessFilterOp(config recconf.FilterParamConfig) *LessFilterOp {

	lessFilterOp := &LessFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}

	if config.Domain == "" {
		lessFilterOp.Domain = ITEM
	} else {
		lessFilterOp.Domain = config.Domain
	}

	return lessFilterOp
}

func (p *LessFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, 0)
		v2 := utils.ToFloat(p.Value, 0)

		return v1 < v2, nil
	case "int":
		v1 := utils.ToInt(left, 0)
		v2 := utils.ToInt(p.Value, 0)
		return v1 < v2, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		v2 := utils.ToInt64(p.Value, 0)
		return v1 < v2, nil
	default:
		return false, nil
	}
}

func (p *LessFilterOp) OpDomain() string {
	return p.Domain
}

type LessThanFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewLessThanFilterOp(config recconf.FilterParamConfig) *LessThanFilterOp {

	lessThanFilterOp := &LessThanFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}

	if config.Domain == "" {
		lessThanFilterOp.Domain = ITEM
	} else {
		lessThanFilterOp.Domain = config.Domain
	}

	return lessThanFilterOp
}

func (p *LessThanFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, 0)
		v2 := utils.ToFloat(p.Value, 0)

		return v1 <= v2, nil
	case "int":
		v1 := utils.ToInt(left, 0)
		v2 := utils.ToInt(p.Value, 0)
		return v1 <= v2, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		v2 := utils.ToInt64(p.Value, 0)
		return v1 <= v2, nil
	default:
		return false, nil
	}
}

func (p *LessThanFilterOp) OpDomain() string {
	return p.Domain
}

type ContainsFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewContainsFilterOp(config recconf.FilterParamConfig) *ContainsFilterOp {

	containsFilterOp := &ContainsFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}

	if config.Domain == "" {
		containsFilterOp.Domain = ITEM
	} else {
		containsFilterOp.Domain = config.Domain
	}

	return containsFilterOp
}

func (p *ContainsFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {
	return false, nil
}

func (p *ContainsFilterOp) ContainsEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left1, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "[]string":
		var (
			left  []string
			right []string
		)
		if leftStrings, ok := left1.([]string); ok {
			left = leftStrings
		} else if leftInterfaces, ok := left1.([]interface{}); ok {
			for _, val := range leftInterfaces {
				if val1 := utils.ToString(val, ""); val1 != "" {
					left = append(left, val1)
				}
			}
		} else {
			return false, nil
		}
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.Contains(value, "user.") {
				val := value[5:]
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.Contains(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			} else {
				right1 = []string{value}
			}

			if right, ok = right1.([]string); !ok {
				return false, nil
			}
		} else if values, ok := p.Value.([]string); ok {
			right = values
		} else if values, ok := p.Value.([]any); ok {
			for _, val := range values {
				if val1 := utils.ToString(val, ""); val1 != "" {
					right = append(right, val1)
				}
			}
		}

		if len(left) == 0 || len(right) == 0 {
			return false, nil
		}

		return utils.StringContains(left, right), nil
	case "[]int":
		var (
			left  []int
			right []int
		)
		left, ok = left1.([]int)
		if !ok {
			return false, nil
		}
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.Contains(value, "user.") {
				val := value[5:]
				fmt.Println(val)
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.Contains(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			}
			if right, ok = right1.([]int); !ok {
				return false, nil
			}
		} else if values, ok := p.Value.([]int); ok {
			right = values
		}
		if len(left) == 0 || len(right) == 0 {
			return false, nil
		}

		return utils.IntContains(left, right), nil
	}
	return false, nil
}

func (p *ContainsFilterOp) OpDomain() string {
	return p.Domain
}

type NotContainsFilterOp struct {
	Name   string
	Domain string
	Type   string
	Value  interface{}
}

func NewNotContainsFilterOp(config recconf.FilterParamConfig) *NotContainsFilterOp {

	containsFilterOp := &NotContainsFilterOp{
		Name:  config.Name,
		Type:  config.Type,
		Value: config.Value,
	}

	if config.Domain == "" {
		containsFilterOp.Domain = ITEM
	} else {
		containsFilterOp.Domain = config.Domain
	}

	return containsFilterOp
}

func (p *NotContainsFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {
	return false, nil
}

func (p *NotContainsFilterOp) ContainsEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left1, ok := properties[p.Name]
	if !ok {
		return false, nil
	}
	switch p.Type {
	case "[]string":
		var (
			left  []string
			right []string
		)
		if leftStrings, ok := left1.([]string); ok {
			left = leftStrings
		} else if leftInterfaces, ok := left1.([]interface{}); ok {
			for _, val := range leftInterfaces {
				if val1 := utils.ToString(val, ""); val1 != "" {
					left = append(left, val1)
				}
			}
		} else {
			return false, nil
		}
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.Contains(value, "user.") {
				val := value[5:]
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.Contains(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			}
			if right, ok = right1.([]string); !ok {
				return false, nil
			}
		} else if values, ok := p.Value.([]string); ok {
			right = values
		}
		if len(left) == 0 || len(right) == 0 {
			return false, nil
		}

		return !utils.StringContains(left, right), nil
	case "[]int":
		var (
			left  []int
			right []int
		)
		left, ok = left1.([]int)
		if !ok {
			return false, nil
		}
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.Contains(value, "user.") {
				val := value[5:]
				fmt.Println(val)
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.Contains(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			}
			if right, ok = right1.([]int); !ok {
				return false, nil
			}
		} else if values, ok := p.Value.([]int); ok {
			right = values
		}
		if len(left) == 0 || len(right) == 0 {
			return false, nil
		}

		return !utils.IntContains(left, right), nil
	}
	return false, nil
}

func (p *NotContainsFilterOp) OpDomain() string {
	return p.Domain
}

type NotInFilterOp struct {
	Name          string
	Domain        string
	Type          string
	value         string
	int_values    []int
	string_values []string
}

func NewNotInFilterOp(config recconf.FilterParamConfig) *NotInFilterOp {

	op := &NotInFilterOp{
		Name: config.Name,
		Type: config.Type,
	}

	if config.Domain == "" {
		op.Domain = ITEM
	} else {
		op.Domain = config.Domain
	}
	if value, ok := config.Value.(string); ok {
		op.value = value
	} else {
		switch op.Type {
		case "int":
			if values, ok := config.Value.([]int); ok {
				op.int_values = append(op.int_values, values...)
			}
		case "string":
			if values, ok := config.Value.([]string); ok {
				op.string_values = append(op.string_values, values...)
			}

		}
	}

	return op
}
func (p *NotInFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {
	return false, nil
}

func (p *NotInFilterOp) NotInEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left1, ok := properties[p.Name]
	if !ok {
		return false, nil
	}
	switch p.Type {
	case "string":
		left := utils.ToString(left1, "")
		var right []string
		if p.value == "" {
			right = p.string_values
		} else {
			if strings.Contains(p.value, "user.") {
				val := p.value[5:]
				right1, ok := userProperties[val]
				if !ok {
					return true, nil
				}
				if rightVal, ok := right1.([]string); ok {
					right = rightVal
				} else {
					return false, nil
				}

			} else if strings.Contains(p.value, "item.") {
				val := p.value[5:]
				right1, ok := itemProperties[val]
				if !ok {
					return true, nil
				}
				if rightVal, ok := right1.([]string); ok {
					right = rightVal
				} else {
					return false, nil
				}
			}
		}
		for _, val := range right {
			if left == val {
				return false, nil
			}
		}

		return true, nil
	case "int":
		left := utils.ToInt(left1, math.MinInt32)
		var (
			right []int
		)
		if p.value == "" {
			right = p.int_values
		} else {
			if strings.Contains(p.value, "user.") {
				val := p.value[5:]
				right1, ok := userProperties[val]
				if !ok {
					return true, nil
				}
				if rightVal, ok := right1.([]int); ok {
					right = rightVal
				} else {
					return false, nil
				}

			} else if strings.Contains(p.value, "item.") {
				val := p.value[5:]
				right1, ok := itemProperties[val]
				if !ok {
					return true, nil
				}
				if rightVal, ok := right1.([]int); ok {
					right = rightVal
				} else {
					return false, nil
				}
			}
		}
		for _, val := range right {
			if left == val {
				return false, nil
			}
		}

		return true, nil
	}
	return false, nil
}

func (p *NotInFilterOp) OpDomain() string {
	return p.Domain
}
