package utils

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Knetic/govaluate"
)

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

			str := ToString(args[0], "")
			cutset := ToString(args[1], "")
			fmt.Println(strings.TrimPrefix(str, cutset), str, cutset)
			return strings.Trim(str, cutset), nil
		},
		"trimPrefix": func(args ...interface{}) (interface{}, error) {
			if len(args) != 2 {
				return "", errors.New("args length not equal 2")
			}

			str := ToString(args[0], "")
			cutset := ToString(args[1], "")
			return strings.TrimPrefix(str, cutset), nil
		},
		"replace": func(args ...interface{}) (interface{}, error) {
			if len(args) != 3 {
				return "", errors.New("args length not equal 3")
			}

			str := ToString(args[0], "")
			old := ToString(args[1], "")
			new := ToString(args[2], "")
			return strings.ReplaceAll(str, old, new), nil
		},
		"round": func(args ...interface{}) (interface{}, error) {
			if len(args) == 1 {
				return math.Round(ToFloat(args[0], 0)), nil
			} else if len(args) == 2 {
				f := ToFloat(args[0], 0)
				n := ToFloat(args[1], 0)
				multiplier := math.Pow(10, n)
				return math.Trunc(f*multiplier) / multiplier, nil

			} else {
				return nil, errors.New("wrong number of arguments")
			}
		},
	}
)

func GovaluateFunctions() map[string]govaluate.ExpressionFunction {
	return functions
}
