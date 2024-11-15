package web

import (
	"github.com/alibaba/pairec/v2/utils"
	jsoniter "github.com/json-iterator/go"
)

var json_iter = jsoniter.ConfigCompatibleWithStandardLibrary

type Feature struct {
	Name   string `json:"name"`
	Values any    `json:"values"`
	Type   string `json:"type"`
}
type ComplexTypeFeatures struct {
	Features []Feature `json:"features"`

	FeaturesMap map[string]any
}

func (f *ComplexTypeFeatures) UnmarshalJSON(data []byte) error {
	if err := json_iter.Unmarshal(data, &f.Features); err != nil {
		return err
	}

	f.FeaturesMap = make(map[string]any)
	for _, feature := range f.Features {
		switch feature.Type {
		case "string":
			f.FeaturesMap[feature.Name] = utils.ToString(feature.Values, "")
		case "double":
			f.FeaturesMap[feature.Name] = utils.ToFloat(feature.Values, 0)
		case "float":
			f.FeaturesMap[feature.Name] = utils.ToFloat32(feature.Values, 0)
		case "int64":
			f.FeaturesMap[feature.Name] = utils.ToInt64(feature.Values, 0)
		case "int":
			f.FeaturesMap[feature.Name] = utils.ToInt(feature.Values, 0)
		case "list<string>":
			if vals, ok := feature.Values.([]any); ok {
				values := make([]string, 0, len(vals))
				for _, v := range vals {
					values = append(values, utils.ToString(v, ""))
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "list<int64>":
			if vals, ok := feature.Values.([]any); ok {
				values := make([]int64, 0, len(vals))
				for _, v := range vals {
					values = append(values, utils.ToInt64(v, 0))
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "list<int>":
			if vals, ok := feature.Values.([]any); ok {
				values := make([]int, 0, len(vals))
				for _, v := range vals {
					values = append(values, utils.ToInt(v, 0))
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "list<double>":
			if vals, ok := feature.Values.([]any); ok {
				values := make([]float64, 0, len(vals))
				for _, v := range vals {
					values = append(values, utils.ToFloat(v, 0))
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "list<float>":
			if vals, ok := feature.Values.([]any); ok {
				values := make([]float32, 0, len(vals))
				for _, v := range vals {
					values = append(values, utils.ToFloat32(v, 0))
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "list<list<float>>":
			if vals, ok := feature.Values.([]any); ok {
				values := make([][]float32, 0, len(vals))
				for _, vlist := range vals {
					if lists, ok := vlist.([]any); ok {
						list := make([]float32, 0, len(lists))
						for _, v := range lists {
							list = append(list, utils.ToFloat32(v, 0))
						}
						values = append(values, list)
					}
				}

				f.FeaturesMap[feature.Name] = values
			}
		case "list<list<double>>":
			if vals, ok := feature.Values.([]any); ok {
				values := make([][]float64, 0, len(vals))
				for _, vlist := range vals {
					if lists, ok := vlist.([]any); ok {
						list := make([]float64, 0, len(lists))
						for _, v := range lists {
							list = append(list, utils.ToFloat(v, 0))
						}
						values = append(values, list)
					}
				}

				f.FeaturesMap[feature.Name] = values
			}
		case "map<string,string>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[string]string, len(vals))
				for k, v := range vals {
					values[k] = utils.ToString(v, "")
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<string,int>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[string]int, len(vals))
				for k, v := range vals {
					values[k] = utils.ToInt(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<string,int64>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[string]int64, len(vals))
				for k, v := range vals {
					values[k] = utils.ToInt64(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<string,float>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[string]float32, len(vals))
				for k, v := range vals {
					values[k] = utils.ToFloat32(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<string,double>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[string]float64, len(vals))
				for k, v := range vals {
					values[k] = utils.ToFloat(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int,int>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int]int, len(vals))
				for k, v := range vals {
					values[utils.ToInt(k, 0)] = utils.ToInt(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int,int64>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int]int64, len(vals))
				for k, v := range vals {
					values[utils.ToInt(k, 0)] = utils.ToInt64(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int,double>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int]float64, len(vals))
				for k, v := range vals {
					values[utils.ToInt(k, 0)] = utils.ToFloat(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int,float>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int]float32, len(vals))
				for k, v := range vals {
					values[utils.ToInt(k, 0)] = utils.ToFloat32(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int,string>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int]string, len(vals))
				for k, v := range vals {
					values[utils.ToInt(k, 0)] = utils.ToString(v, "")
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int64,int>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int64]int, len(vals))
				for k, v := range vals {
					values[utils.ToInt64(k, 0)] = utils.ToInt(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int64,int64>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int64]int64, len(vals))
				for k, v := range vals {
					values[utils.ToInt64(k, 0)] = utils.ToInt64(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int64,double>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int64]float64, len(vals))
				for k, v := range vals {
					values[utils.ToInt64(k, 0)] = utils.ToFloat(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int64,float>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int64]float32, len(vals))
				for k, v := range vals {
					values[utils.ToInt64(k, 0)] = utils.ToFloat32(v, 0)
				}
				f.FeaturesMap[feature.Name] = values
			}
		case "map<int64,string>":
			if vals, ok := feature.Values.(map[string]any); ok {
				values := make(map[int64]string, len(vals))
				for k, v := range vals {
					values[utils.ToInt64(k, 0)] = utils.ToString(v, "")
				}
				f.FeaturesMap[feature.Name] = values
			}
		default:
			f.FeaturesMap[feature.Name] = feature.Values
		}
	}

	return nil
}
