package utils

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/cespare/xxhash/v2"
	"github.com/expr-lang/expr"
	"github.com/golang/geo/s2"
	"github.com/mmcloughlin/geohash"
	"github.com/spaolacci/murmur3"
)

const earthRadiusKm = 6371.0

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
		"hash": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return "", errors.New("args length not equal 1")
			}
			str := ToString(args[0], "")
			return xxhash.Sum64String(str), nil
		},
		"hash32": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return "", errors.New("args length not equal 1")
			}
			str := ToString(args[0], "")
			return float64(murmur3.Sum32(String2byte(str))), nil
		},
		"toFloat64": func(args ...interface{}) (interface{}, error) {
			if len(args) != 1 {
				return "", errors.New("args length not equal 1")
			}
			return ToFloat(args[0], 0), nil
		},
		"log": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 1 {
				return "", errors.New("args length not equal 1")
			}
			return math.Log(ToFloat(arguments[0], 0)), nil
		},
		"log10": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 1 {
				return "", errors.New("args length not equal 1")
			}
			return math.Log10(ToFloat(arguments[0], 0)), nil
		},
		"log2": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 1 {
				return "", errors.New("args length not equal 1")
			}
			return math.Log2(ToFloat(arguments[0], 0)), nil
		},
		"max": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 2 {
				return "", errors.New("args length not equal 2")
			}
			return math.Max(ToFloat(arguments[0], 0), ToFloat(arguments[1], 0)), nil
		},
		"min": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 2 {
				return "", errors.New("args length not equal 2")
			}
			return math.Min(ToFloat(arguments[0], 0), ToFloat(arguments[1], 0)), nil
		},
		"pow": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 2 {
				return "", errors.New("args length not equal 2")
			}
			return math.Pow(ToFloat(arguments[0], 0), ToFloat(arguments[1], 0)), nil
		},
		"s2CellID": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) < 2 {
				return "", errors.New("args must have lat and lng params")
			}
			lat := ToFloat(arguments[0], 0)
			lng := ToFloat(arguments[1], 0)
			ll := s2.LatLngFromDegrees(lat, lng)
			cellID := s2.CellIDFromLatLng(ll)
			level := 15
			if len(arguments) > 2 {
				level = ToInt(arguments[2], 15)
			}

			cellIDAtLevel := cellID.Parent(level)
			return int(cellIDAtLevel), nil

		},
		"geoHash": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) < 2 {
				return "", errors.New("args must have lat and lng params")
			}
			lat := ToFloat(arguments[0], 0)
			lng := ToFloat(arguments[1], 0)
			precision := 6
			if len(arguments) > 2 {
				precision = ToInt(arguments[2], 6)
			}
			return geohash.EncodeWithPrecision(lat, lng, uint(precision)), nil
		},
		"geoHashWithNeighbors": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) < 2 {
				return "", errors.New("args must have lat and lng params")
			}
			lat := ToFloat(arguments[0], 0)
			lng := ToFloat(arguments[1], 0)
			precision := 6
			if len(arguments) > 2 {
				precision = ToInt(arguments[2], 6)
			}

			hashCode := geohash.EncodeWithPrecision(lat, lng, uint(precision))
			neighbors := geohash.Neighbors(hashCode)
			neighbors = append(neighbors, hashCode)

			return neighbors, nil
		},
		"haversine": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 4 {
				return "", errors.New("args length not equal 4")
			}
			lng1 := ToFloat(arguments[0], 0)
			lat1 := ToFloat(arguments[1], 0)

			lng2 := ToFloat(arguments[2], 0)
			lat2 := ToFloat(arguments[3], 0)
			radLat1 := degreesToRadians(lat1)
			radLat2 := degreesToRadians(lat2)
			radLng1 := degreesToRadians(lng1)
			radLng2 := degreesToRadians(lng2)

			deltaLat := radLat2 - radLat1
			deltaLng := radLng2 - radLng1

			a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
				math.Cos(radLat1)*math.Cos(radLat2)*
					math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

			c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

			distance := earthRadiusKm * c

			return distance, nil
		},
		"sphereDistance": func(arguments ...interface{}) (interface{}, error) {
			if len(arguments) != 4 {
				return "", errors.New("args length not equal 4")
			}
			lng1 := ToFloat(arguments[0], 0)
			lat1 := ToFloat(arguments[1], 0)

			lng2 := ToFloat(arguments[2], 0)
			lat2 := ToFloat(arguments[3], 0)

			radLat1 := degreesToRadians(lat1)
			radLat2 := degreesToRadians(lat2)

			deltaLng := degreesToRadians(lng2 - lng1)

			cosVal := math.Sin(radLat1)*math.Sin(radLat2) +
				math.Cos(radLat1)*math.Cos(radLat2)*math.Cos(deltaLng)

			// 进行边界检查，防止 acos(x) 的 x 超出 [-1, 1] 范围
			if cosVal > 1.0 {
				cosVal = 1.0
			} else if cosVal < -1.0 {
				cosVal = -1.0
			}

			distance := math.Acos(cosVal) * earthRadiusKm

			return distance, nil
		},
	}
)

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

func GovaluateFunctions() map[string]govaluate.ExpressionFunction {
	return functions
}
func ExprFunctions() []expr.Option {
	var options []expr.Option
	for name, f := range functions {
		options = append(options, expr.Function(name, f))
	}
	return options
}
