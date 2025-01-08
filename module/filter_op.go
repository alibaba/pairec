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
type FilterByDomainOp interface {
	FilterOp
	DomainEvaluate(map[string]any, map[string]any, map[string]any) (bool, error)
}

type EqualFilterOp struct {
	Name        string
	Domain      string
	Type        string
	Value       interface{}
	DomainValue string
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
	if v, ok := config.Value.(string); ok {
		equalFilterOp.DomainValue = v
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
func (p *EqualFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {

	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "string":
		v1 := utils.ToString(left, "v1")
		var right string
		if p.DomainValue == "" {
			right = utils.ToString(p.Value, "v2")
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "v2")

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "v2")

		} else {
			right = utils.ToString(p.Value, "v2")
		}

		return v1 == right, nil
	case "int":
		v1 := utils.ToInt(left, -1)
		var right int
		if p.DomainValue == "" {
			right = utils.ToInt(p.Value, -2)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, -2)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, -2)

		} else {
			right = utils.ToInt(p.Value, -2)
		}
		return v1 == right, nil
	case "int64":
		v1 := utils.ToInt64(left, -1)
		var right int64
		if p.DomainValue == "" {
			right = utils.ToInt64(p.Value, -2)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, -2)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, -2)

		} else {
			right = utils.ToInt64(p.Value, -2)
		}
		return v1 == right, nil
	default:
		return false, nil
	}
}

func (p *EqualFilterOp) OpDomain() string {
	return p.Domain
}

type NotEqualFilterOp struct {
	Name        string
	Domain      string
	Type        string
	Value       interface{}
	DomainValue string
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

	if v, ok := config.Value.(string); ok {
		notEqualFilterOp.DomainValue = v
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
func (p *NotEqualFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left, ok := properties[p.Name]
	if !ok {
		return true, nil
	}

	switch p.Type {
	case "string":
		v1 := utils.ToString(left, "")
		var right string
		if p.DomainValue == "" {
			right = utils.ToString(p.Value, "")
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToString(right1, "")

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToString(right1, "")
		} else {
			right = utils.ToString(p.Value, "")
		}

		return v1 != right, nil
	case "int":
		v1 := utils.ToInt(left, 0)
		var right int
		if p.DomainValue == "" {
			right = utils.ToInt(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToInt(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToInt(right1, 0)
		} else {
			right = utils.ToInt(p.Value, 0)
		}
		return v1 != right, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		var right int64
		if p.DomainValue == "" {
			right = utils.ToInt64(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToInt64(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToInt64(right1, 0)
		} else {
			right = utils.ToInt64(p.Value, 0)
		}
		return v1 != right, nil
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
	value         string
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

	if val, ok := config.Value.(string); ok {
		op.value = val
	}

	switch op.Type {
	case "int":
		op.int_values = utils.ToIntArray(config.Value)

	case "string":
		op.string_values = utils.ToStringArray(config.Value)
	}

	return op
}
func (p *InFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "string":
		v1 := utils.ToString(left, "")
		var right []string
		if p.value == "" {
			right = p.string_values
		} else {
			if strings.HasPrefix(p.value, "user.") {
				val := p.value[5:]
				right1, ok := userProperties[val]
				if !ok {
					return false, nil
				}
				right = utils.ToStringArray(right1)

			} else if strings.HasPrefix(p.value, "item.") {
				val := p.value[5:]
				right1, ok := itemProperties[val]
				if !ok {
					return true, nil
				}
				right = utils.ToStringArray(right1)
			}
		}
		for _, val := range right {
			if v1 == val {
				return true, nil
			}
		}
	case "int":
		v1 := utils.ToInt(left, math.MinInt32)
		var right []int
		if p.value == "" {
			right = p.int_values
		} else {
			if strings.HasPrefix(p.value, "user.") {
				val := p.value[5:]
				right1, ok := userProperties[val]
				if !ok {
					return false, nil
				}
				right = utils.ToIntArray(right1)

			} else if strings.HasPrefix(p.value, "item.") {
				val := p.value[5:]
				right1, ok := itemProperties[val]
				if !ok {
					return true, nil
				}
				right = utils.ToIntArray(right1)
			}

		}
		for _, val := range right {
			if v1 == val {
				return true, nil
			}
		}
	}
	return false, nil
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
		} else if config.Operator == "is_null" {
			p.operators = append(p.operators, NewIsNullFilterOp(config))
		} else if config.Operator == "is_not_null" {
			p.operators = append(p.operators, NewIsNotNullFilterOp(config))
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
		if domainFilterOp, ok := op.(FilterByDomainOp); ok {
			if domainFilterOp.OpDomain() == ITEM {
				ret, err := domainFilterOp.DomainEvaluate(itemProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
			} else if domainFilterOp.OpDomain() == USER {
				ret, err := domainFilterOp.DomainEvaluate(userProperties, userProperties, itemProperties)
				if !ret || err != nil {
					return false, err
				}
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
	Name        string
	Domain      string
	Type        string
	Value       interface{}
	DomainValue string
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

	if v, ok := config.Value.(string); ok {
		greaterFilterOp.DomainValue = v
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
func (p *GreaterFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, 0)
		var right float64
		if p.DomainValue == "" {
			right = utils.ToFloat(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToFloat(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToFloat(right1, 0)
		} else {
			right = utils.ToFloat(p.Value, 0)
		}

		return v1 > right, nil
	case "int":
		v1 := utils.ToInt(left, 0)
		var right int
		if p.DomainValue == "" {
			right = utils.ToInt(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, 0)
		} else {
			right = utils.ToInt(p.Value, 0)
		}
		return v1 > right, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		var right int64
		if p.DomainValue == "" {
			right = utils.ToInt64(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)
		} else {
			right = utils.ToInt64(p.Value, 0)
		}
		return v1 > right, nil
	case "time":
		v1 := utils.ToString(left, "")
		var right string
		if p.DomainValue == "" {
			right = utils.ToString(p.Value, "")
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")
		} else {
			right = utils.ToString(p.Value, "")
		}
		leftTime, leftOk := utils.TryParseTime(v1)
		if !leftOk {
			return false, nil
		}
		rightTime, rightOk := utils.TryParseTime(right)
		if !rightOk {
			return false, nil
		}
		return leftTime.UnixNano() > rightTime.UnixNano(), nil
	default:
		return false, nil
	}
}

func (p *GreaterFilterOp) OpDomain() string {
	return p.Domain
}

type GreaterThanFilterOp struct {
	Name        string
	Domain      string
	Type        string
	Value       interface{}
	DomainValue string
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
	if v, ok := config.Value.(string); ok {
		greaterThanFilterOp.DomainValue = v
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
func (p *GreaterThanFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, math.SmallestNonzeroFloat64)
		var right float64
		if p.DomainValue == "" {
			right = utils.ToFloat(p.Value, math.MaxFloat64)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToFloat(right1, math.MaxFloat64)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToFloat(right1, math.MaxFloat64)
		} else {
			right = utils.ToFloat(p.Value, math.MaxFloat64)
		}

		return v1 >= right, nil
	case "int":
		v1 := utils.ToInt(left, math.MinInt)
		var right int
		if p.DomainValue == "" {
			right = utils.ToInt(p.Value, math.MaxInt)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, math.MaxInt)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, math.MaxInt)
		} else {
			right = utils.ToInt(p.Value, math.MaxInt)
		}
		return v1 >= right, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		var right int64
		if p.DomainValue == "" {
			right = utils.ToInt64(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)
		} else {
			right = utils.ToInt64(p.Value, 0)
		}
		return v1 >= right, nil
	case "time":
		v1 := utils.ToString(left, "")
		var right string
		if p.DomainValue == "" {
			right = utils.ToString(p.Value, "")
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")
		} else {
			right = utils.ToString(p.Value, "")
		}
		leftTime, leftOk := utils.TryParseTime(v1)
		if !leftOk {
			return false, nil
		}
		rightTime, rightOk := utils.TryParseTime(right)
		if !rightOk {
			return false, nil
		}
		return leftTime.UnixNano() >= rightTime.UnixNano(), nil
	default:
		return false, nil
	}
}

func (p *GreaterThanFilterOp) OpDomain() string {
	return p.Domain
}

type LessFilterOp struct {
	Name        string
	Domain      string
	Type        string
	Value       interface{}
	DomainValue string
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

	if v, ok := config.Value.(string); ok {
		lessFilterOp.DomainValue = v
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
func (p *LessFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, math.MaxFloat64)
		var right float64
		if p.DomainValue == "" {
			right = utils.ToFloat(p.Value, math.SmallestNonzeroFloat64)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToFloat(right1, math.SmallestNonzeroFloat64)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToFloat(right1, math.SmallestNonzeroFloat64)
		} else {
			right = utils.ToFloat(p.Value, math.SmallestNonzeroFloat64)
		}

		return v1 < right, nil
	case "int":
		v1 := utils.ToInt(left, math.MaxInt)
		var right int
		if p.DomainValue == "" {
			right = utils.ToInt(p.Value, math.MinInt)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, math.MinInt)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, math.MinInt)
		} else {
			right = utils.ToInt(p.Value, math.MinInt)
		}
		return v1 < right, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		var right int64
		if p.DomainValue == "" {
			right = utils.ToInt64(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)
		} else {
			right = utils.ToInt64(p.Value, 0)
		}
		return v1 < right, nil
	case "time":
		v1 := utils.ToString(left, "")
		var right string
		if p.DomainValue == "" {
			right = utils.ToString(p.Value, "")
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")
		} else {
			right = utils.ToString(p.Value, "")
		}
		leftTime, leftOk := utils.TryParseTime(v1)
		if !leftOk {
			return false, nil
		}
		rightTime, rightOk := utils.TryParseTime(right)
		if !rightOk {
			return false, nil
		}
		return leftTime.UnixNano() < rightTime.UnixNano(), nil
	default:
		return false, nil
	}
}

func (p *LessFilterOp) OpDomain() string {
	return p.Domain
}

type LessThanFilterOp struct {
	Name        string
	Domain      string
	Type        string
	Value       interface{}
	DomainValue string
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

	if v, ok := config.Value.(string); ok {
		lessThanFilterOp.DomainValue = v
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
func (p *LessThanFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
	left, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	switch p.Type {
	case "float":
		v1 := utils.ToFloat(left, math.MaxFloat64)
		var right float64
		if p.DomainValue == "" {
			right = utils.ToFloat(p.Value, math.SmallestNonzeroFloat64)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToFloat(right1, math.SmallestNonzeroFloat64)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return true, nil
			}
			right = utils.ToFloat(right1, math.SmallestNonzeroFloat64)
		} else {
			right = utils.ToFloat(p.Value, math.SmallestNonzeroFloat64)
		}

		return v1 <= right, nil
	case "int":
		v1 := utils.ToInt(left, math.MaxInt)
		var right int
		if p.DomainValue == "" {
			right = utils.ToInt(p.Value, math.MinInt)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, math.MinInt)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt(right1, math.MinInt)
		} else {
			right = utils.ToInt(p.Value, math.MinInt)
		}
		return v1 <= right, nil
	case "int64":
		v1 := utils.ToInt64(left, 0)
		var right int64
		if p.DomainValue == "" {
			right = utils.ToInt64(p.Value, 0)
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToInt64(right1, 0)
		} else {
			right = utils.ToInt64(p.Value, 0)
		}
		return v1 <= right, nil
	case "time":
		v1 := utils.ToString(left, "")
		var right string
		if p.DomainValue == "" {
			right = utils.ToString(p.Value, "")
		} else if strings.HasPrefix(p.DomainValue, "user.") {
			val := p.DomainValue[5:]
			right1, ok := userProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")

		} else if strings.HasPrefix(p.DomainValue, "item.") {
			val := p.DomainValue[5:]
			right1, ok := itemProperties[val]
			if !ok {
				return false, nil
			}
			right = utils.ToString(right1, "")
		} else {
			right = utils.ToString(p.Value, "")
		}
		leftTime, leftOk := utils.TryParseTime(v1)
		if !leftOk {
			return false, nil
		}
		rightTime, rightOk := utils.TryParseTime(right)
		if !rightOk {
			return false, nil
		}
		return leftTime.UnixNano() <= rightTime.UnixNano(), nil
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

func (p *ContainsFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
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
		left = utils.ToStringArray(left1)
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.HasPrefix(value, "user.") {
				val := value[5:]
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.HasPrefix(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			} else {
				right1 = []string{value}
			}

			right = utils.ToStringArray(right1)
		} else {
			right = utils.ToStringArray(p.Value)
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
		left = utils.ToIntArray(left1)
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.HasPrefix(value, "user.") {
				val := value[5:]
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.HasPrefix(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			}
			right = utils.ToIntArray(right1)
		} else {
			right = utils.ToIntArray(p.Value)
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

func (p *NotContainsFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
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
		left = utils.ToStringArray(left1)
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.HasPrefix(value, "user.") {
				val := value[5:]
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.HasPrefix(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			}
			right = utils.ToStringArray(right1)
		} else {
			right = utils.ToStringArray(p.Value)
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
		left = utils.ToIntArray(left1)
		if value, ok := p.Value.(string); ok {
			var right1 interface{}
			if strings.HasPrefix(value, "user.") {
				val := value[5:]
				right1, ok = userProperties[val]
				if !ok {
					return false, nil
				}

			} else if strings.HasPrefix(value, "item.") {
				val := value[5:]
				right1, ok = itemProperties[val]
				if !ok {
					return false, nil
				}
			}
			right = utils.ToIntArray(right1)
		} else {
			right = utils.ToIntArray(p.Value)
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
			op.int_values = utils.ToIntArray(config.Value)
		case "string":
			op.string_values = utils.ToStringArray(config.Value)

		}
	}

	return op
}
func (p *NotInFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {
	return false, nil
}

func (p *NotInFilterOp) DomainEvaluate(properties map[string]interface{}, userProperties map[string]interface{}, itemProperties map[string]interface{}) (bool, error) {
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
			if strings.HasPrefix(p.value, "user.") {
				val := p.value[5:]
				right1, ok := userProperties[val]
				if !ok {
					return true, nil
				}
				right = utils.ToStringArray(right1)

			} else if strings.HasPrefix(p.value, "item.") {
				val := p.value[5:]
				right1, ok := itemProperties[val]
				if !ok {
					return true, nil
				}
				right = utils.ToStringArray(right1)
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
			if strings.HasPrefix(p.value, "user.") {
				val := p.value[5:]
				right1, ok := userProperties[val]
				if !ok {
					return true, nil
				}

				right = utils.ToIntArray(right1)

			} else if strings.HasPrefix(p.value, "item.") {
				val := p.value[5:]
				right1, ok := itemProperties[val]
				if !ok {
					return true, nil
				}

				right = utils.ToIntArray(right1)
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

type IsNullFilterOp struct {
	Name   string
	Domain string
}

func NewIsNullFilterOp(config recconf.FilterParamConfig) *IsNullFilterOp {

	isNullFilterOp := &IsNullFilterOp{
		Name: config.Name,
	}
	if config.Domain == "" {
		isNullFilterOp.Domain = ITEM
	} else {
		isNullFilterOp.Domain = config.Domain
	}

	return isNullFilterOp
}

func (p *IsNullFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	_, ok := properties[p.Name]
	if !ok {
		return true, nil
	}

	return false, nil
}

func (p *IsNullFilterOp) OpDomain() string {
	return p.Domain
}

type IsNotNullFilterOp struct {
	Name   string
	Domain string
}

func NewIsNotNullFilterOp(config recconf.FilterParamConfig) *IsNotNullFilterOp {

	isNotNullFilterOp := &IsNotNullFilterOp{
		Name: config.Name,
	}
	if config.Domain == "" {
		isNotNullFilterOp.Domain = ITEM
	} else {
		isNotNullFilterOp.Domain = config.Domain
	}

	return isNotNullFilterOp
}

func (p *IsNotNullFilterOp) Evaluate(properties map[string]interface{}) (bool, error) {

	_, ok := properties[p.Name]
	if !ok {
		return false, nil
	}

	return true, nil
}

func (p *IsNotNullFilterOp) OpDomain() string {
	return p.Domain
}
