package ast

import (
	"errors"
	"fmt"
	"math"

	"github.com/alibaba/pairec/v2/utils"
	"github.com/bruceding/go-antlr-valuate/parser"
	"github.com/cespare/xxhash/v2"
	"github.com/spaolacci/murmur3"
)

var (
	functions = map[string]parser.ExpressionFunction{
		"hash": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return "", errors.New("args length not equal 1")
			}
			str := utils.ToString(args[0], "")
			return xxhash.Sum64String(str), nil
		},
		"hash32": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return "", errors.New("args length not equal 1")
			}
			str := utils.ToString(args[0], "")
			return float64(murmur3.Sum32(utils.String2byte(str))), nil
		},
		"maxIndex": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("maxIndex: expects exactly one argument")
			}

			slice := utils.ToFloatArray(args[0])
			if len(slice) == 0 {
				return nil, errors.New("maxIndex: argument must not be empty")
			}
			// Use the reflection-based helper which can handle any slice type
			maxIndex, _, err := findMax(slice)
			if err != nil {
				return nil, fmt.Errorf("maxIndex: %w", err)
			}

			return maxIndex, nil
		},
		"maxValue": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, errors.New("maxValue: expects exactly one argument")
			}

			slice := utils.ToFloatArray(args[0])
			if len(slice) == 0 {
				return nil, errors.New("maxValue: argument must not be empty")
			}
			// Use the reflection-based helper which can handle any slice type
			_, maxValue, err := findMax(slice)
			if err != nil {
				return nil, fmt.Errorf("maxValue: %w", err)
			}

			return maxValue, nil
		},
	}
)

// findMax is a helper that iterates through a slice to find the maximum value and its index.
// This avoids code duplication between argmax and max_value.
func findMax(data []float64) (index int, value float64, err error) {
	if len(data) == 0 {
		return -1, 0, errors.New("cannot find max in an empty array")
	}

	// Initialize with the first element
	maxVal := data[0]
	maxIndex := 0

	// Iterate from the second element
	for i := 1; i < len(data); i++ {
		currentVal := data[i]
		if currentVal > maxVal {
			maxVal = currentVal
			maxIndex = i
		}
	}

	return maxIndex, maxVal, nil
}
func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

func AntlrFunctions() map[string]parser.ExpressionFunction {
	return functions
}
