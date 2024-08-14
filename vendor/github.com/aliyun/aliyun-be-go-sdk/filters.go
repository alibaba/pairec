package be

import (
	"fmt"
	"github.com/go-kit/kit/log/level"
	"reflect"
	"strings"
)

type Filter interface {
	GetConditionValue() string
	Validate() error
}

type FilterOperator string

const (
	EQ FilterOperator = "="
	NE FilterOperator = "!="
	LT FilterOperator = "<"
	GT FilterOperator = ">"
	LE FilterOperator = "<="
	GE FilterOperator = ">="
)

type FilterConnector string

const (
	FilterConnectorAnd FilterConnector = "AND"
	FilterConnectorOR  FilterConnector = "OR"
)

type SingleFilter struct {
	Left     string         `json:"left"`
	Operator FilterOperator `json:"operator"`
	Right    string         `json:"right"`
}

type MultiFilter struct {
	Filters   []Filter        `json:"filters"`
	Connector FilterConnector `json:"connector"`
}

func (f *SingleFilter) Validate() error {
	if f.Left == "" || f.Right == "" || f.Operator == "" {
		return InvalidParamsError{fmt.Sprintf("Invalid params, left[%s], op[%s], right[%s]",
			f.Left, f.Operator, f.Right)}
	}
	return nil
}

func (f *SingleFilter) GetConditionValue() string {
	return f.Left + string(f.Operator) + f.Right
}

func (f *MultiFilter) Validate() error {
	if f.Filters == nil || len(f.Filters) == 0 {
		return InvalidParamsError{"Empty filters"}
	}
	return nil
}

func (f *MultiFilter) GetConditionValue() string {
	var conditions []string
	for _, filter := range f.Filters {
		if filter == nil {
			continue
		}
		filterType := reflect.TypeOf(filter)
		if filterType == reflect.TypeOf(new(MultiFilter)) {
			conditions = append(conditions, "("+filter.GetConditionValue()+")")
		} else if filterType == reflect.TypeOf(new(SingleFilter)) {
			conditions = append(conditions, filter.GetConditionValue())
		} else {
			level.Warn(Logger).Log(fmt.Printf("Unsupported filter[%s], ignore", filterType))
		}
	}
	return strings.Join(conditions[:], " "+string(f.Connector)+" ")
}
