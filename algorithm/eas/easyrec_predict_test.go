package eas

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/pkg/eas"
	"github.com/alibaba/pairec/v2/utils"
	"google.golang.org/protobuf/proto"
)

func buildRequest() (request *easyrec.PBRequest, item_ids []string, err error) {
	content, err := os.ReadFile("data.json")
	if err != nil {
		return
	}
	data := make(map[string]any)

	json.Unmarshal(content, &data)

	builder := easyrec.NewEasyrecRequestBuilder()

	for k, v := range data {
		if k == "item_ids" {
			m := v.(map[string]any)
			for _, itemid := range m["values"].([]any) {
				builder.AddItemId(utils.ToString(itemid, ""))
				item_ids = append(item_ids, utils.ToString(itemid, ""))
			}
		} else {
			m := v.(map[string]any)
			values := m["values"].([]any)
			if len(values) == 1 {
				builder.AddUserFeature(k, utils.ToString(values[0], ""))
			} else {
				if m["dtype"].(string) == "int64" {
					iValues := make([]any, 0)
					for _, v := range values {
						switch val := v.(type) {
						case int64:
							iValues = append(iValues, val)
						case float64:
							iValues = append(iValues, utils.ToInt64(val, 0))
						default:
							iValues = append(iValues, "")
						}
					}
					builder.AddContextFeature(k, iValues)

				} else {
					builder.AddContextFeature(k, values)
				}
			}

		}
	}

	request = builder.EasyrecRequest()
	return

}
func TestTorchrecResponse(t *testing.T) {

	request, _, err := buildRequest()
	if err != nil {
		t.Fatal(err)
	}
	pbData, _ := proto.Marshal(request)
	client := eas.NewPredictClient("http://1730760139076263.cn-beijing.pai-eas.aliyuncs.com", "test_torch_rec_multi_tower_din")
	client.SetToken(os.Getenv("TEST_TORCHREC_DIN_TOKEN"))
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	respBody, err := client.BytesPredict(pbData)
	if err != nil {
		t.Fatal(err)
	}
	responseData := &easyrec.TorchRecPBResponse{}
	proto.Unmarshal(respBody, responseData)

	fmt.Println(responseData)
	for k, v := range responseData.GetMapOutputs() {
		fmt.Println(k, v)
	}

}
func TestTorchrecMutValResponseFunc(t *testing.T) {
	request, item_ids, err := buildRequest()
	if err != nil {
		t.Fatal(err)
	}
	pbData, _ := proto.Marshal(request)
	client := eas.NewPredictClient("http://1730760139076263.cn-beijing.pai-eas.aliyuncs.com", "test_torch_rec_multi_tower_din")
	client.SetToken(os.Getenv("TEST_TORCHREC_DIN_TOKEN"))
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	respBody, err := client.BytesPredict(pbData)
	if err != nil {
		t.Fatal(err)
	}
	responseData := &easyrec.TorchRecPBResponse{}
	proto.Unmarshal(respBody, responseData)
	responseData.ItemIds = item_ids

	alogResponse, err := torchrecMutValResponseFunc(responseData)
	if err != nil {
		t.Fatal(err)
	}
	for _, algo := range alogResponse {
		fmt.Println(algo.GetScoreMap())
	}

}

func TestTorchrecMutValResponseFuncDebug(t *testing.T) {
	request, item_ids, err := buildRequest()
	if err != nil {
		t.Fatal(err)
	}
	request.DebugLevel = 3
	pbData, _ := proto.Marshal(request)
	client := eas.NewPredictClient("http://1730760139076263.cn-beijing.pai-eas.aliyuncs.com", "test_torch_rec_multi_tower_din")
	client.SetToken(os.Getenv("TEST_TORCHREC_DIN_TOKEN"))
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	respBody, err := client.BytesPredict(pbData)
	if err != nil {
		t.Fatal(err)
	}
	responseData := &easyrec.TorchRecPBResponse{}
	proto.Unmarshal(respBody, responseData)
	responseData.ItemIds = item_ids

	alogResponse, err := torchrecMutValResponseFuncDebug(responseData)
	if err != nil {
		t.Fatal(err)
	}
	for _, algo := range alogResponse {
		fmt.Println(algo.(*EasyrecResponse).GenerateFeatures.String())
	}

}

func buildDataComplexTypeRequest() (request *easyrec.PBRequest, item_ids []string, err error) {
	content, err := os.ReadFile("data_complex.json")
	if err != nil {
		return
	}
	data := make(map[string]any)
	requestData := make(map[string]any)

	json.Unmarshal(content, &data)

	builder := easyrec.NewEasyrecRequestBuilder()

	for k, v := range data {
		if k == "item_ids" {
			m := v.(map[string]any)
			for _, itemid := range m["values"].([]any) {
				builder.AddItemId(utils.ToString(itemid, ""))
				item_ids = append(item_ids, utils.ToString(itemid, ""))
			}
		} else {
			m := v.(map[string]any)
			values := m["values"].([]any)
			dtype := m["dtype"].(string)
			if len(values) == 1 {
				switch dtype {
				case "int64":
					builder.AddUserFeature(k, utils.ToInt64(values[0], 0))
					requestData[k] = utils.ToInt64(values[0], 0)
				case "list<float>":
					vals := values[0].([]any)
					var features []float32
					for _, v := range vals {
						features = append(features, utils.ToFloat32(v, 0))
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				case "list<double>":
					vals := values[0].([]any)
					var features []float64
					for _, v := range vals {
						features = append(features, utils.ToFloat(v, 0))
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				case "map<int,float>":
					vals := values[0].(map[string]any)
					features := make(map[int]float32)
					for k, v := range vals {
						features[utils.ToInt(k, 0)] = utils.ToFloat32(v, 0)
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				case "map<int,int>":
					vals := values[0].(map[string]any)
					features := make(map[int]int)
					for k, v := range vals {
						features[utils.ToInt(k, 0)] = utils.ToInt(v, 0)
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				case "map<int,int64>":
					vals := values[0].(map[string]any)
					features := make(map[int]int64)
					for k, v := range vals {
						features[utils.ToInt(k, 0)] = utils.ToInt64(v, 0)
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				case "map<int64,string>":
					vals := values[0].(map[string]any)
					features := make(map[int64]string)
					for k, v := range vals {
						features[utils.ToInt64(k, 0)] = utils.ToString(v, "")
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				case "string":
					builder.AddUserFeature(k, utils.ToString(values[0], ""))
					requestData[k] = utils.ToString(values[0], "")
				case "list<int>":
					vals := values[0].([]any)
					var features []int
					for _, v := range vals {
						features = append(features, utils.ToInt(v, 0))
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				case "list<string>":
					vals := values[0].([]any)
					var features []string
					for _, v := range vals {
						features = append(features, utils.ToString(v, ""))
					}
					builder.AddUserFeature(k, features)
					requestData[k] = features
				default:
					fmt.Println(dtype, "not support")

				}
			} else {
				switch dtype {
				case "float":
					var features []any
					for _, v := range values {
						if v == nil {
							features = append(features, "")
						} else {
							features = append(features, utils.ToFloat32(v, 0))
						}
					}
					builder.AddContextFeature(k, features)
				case "int64":
					var features []any
					for _, v := range values {
						if v == nil {
							features = append(features, "")
						} else {
							features = append(features, utils.ToInt64(v, 0))
						}
					}
					builder.AddContextFeature(k, features)
				case "string":
					var features []any
					for _, v := range values {
						if v == nil {
							features = append(features, "")
						} else {
							features = append(features, utils.ToString(v, ""))
						}
					}
					builder.AddContextFeature(k, features)
				case "list<int>":
					var features []any
					for _, v := range values {
						if v == nil {
							features = append(features, "")
						} else {
							vals := v.([]any)
							var feas []int
							for _, v := range vals {
								feas = append(feas, utils.ToInt(v, 0))
							}

							features = append(features, feas)
						}
					}
					builder.AddContextFeature(k, features)
				case "list<int64>":
					var features []any
					for _, v := range values {
						if v == nil {
							features = append(features, "")
						} else {
							vals := v.([]any)
							var feas []int64
							for _, v := range vals {
								feas = append(feas, utils.ToInt64(v, 0))
							}

							features = append(features, feas)
						}
					}
					builder.AddContextFeature(k, features)
				default:
					fmt.Println(dtype, "not support")
				}
			}

		}
	}

	request = builder.EasyrecRequest()
	buf, _ := json.Marshal(requestData)
	fmt.Println(string(buf))
	return

}

/*
result:
tf request
pytorch request

	map_outputs {
	  key: "logits"
	  value {
	    dtype: DT_FLOAT
	    array_shape {
	      dim: 512
	    }
	    float_val: -0.016554709523916245
	    float_val: -0.03377682343125343
	    float_val: -0.03862795606255531
	    float_val: -0.006319593638181686
	    float_val: -0.03659265115857124
	    float_val: -0.043478984385728836
	    float_val: -0.03849106654524803
	    float_val: -0.02703693136572838
	    float_val: -0.04433808848261833
	    float_val: -0.02531815692782402
	    float_val: -0.03893169388175011
	    float_val: -0.040057118982076645
*/
func TestTorchrecComplexMutValResponseFunc(t *testing.T) {
	request, item_ids, err := buildDataComplexTypeRequest()
	//fmt.Println(request)
	if err != nil {
		t.Fatal(err)
	}
	pbData, _ := proto.Marshal(request)
	client := eas.NewPredictClient("http://1730760139076263.cn-beijing.pai-eas.aliyuncs.com", "test_torch_rec_multi_tower_din_gpu")
	client.SetToken(os.Getenv("TEST_TORCHREC_DIN_TOKEN"))
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	respBody, err := client.BytesPredict(pbData)
	if err != nil {
		t.Fatal(err)
	}
	responseData := &easyrec.TorchRecPBResponse{}
	proto.Unmarshal(respBody, responseData)
	responseData.ItemIds = item_ids

	alogResponse, err := torchrecMutValResponseFunc(responseData)
	if err != nil {
		t.Fatal(err)
	}
	for _, algo := range alogResponse {
		fmt.Println(algo.GetScoreMap())
	}

}

func buildUserEmbeddingRequest() (request *easyrec.PBRequest, item_ids []string, err error) {
	content, err := os.ReadFile("rec_data.json")
	if err != nil {
		return
	}
	data := make(map[string]any)

	json.Unmarshal(content, &data)

	builder := easyrec.NewEasyrecRequestBuilder()

	for k, v := range data {
		if strings.HasPrefix(k, "user_") || strings.HasPrefix(k, "context_") {
			m := v.(map[string]any)
			values := m["values"].([]any)
			builder.AddUserFeature(k, utils.ToString(values[0], ""))
		}
	}

	request = builder.EasyrecRequest()
	return

}

func TestTorchrecUserEmbeddingResponseFunc(t *testing.T) {
	request, _, err := buildUserEmbeddingRequest()
	//fmt.Println(request)
	if err != nil {
		t.Fatal(err)
	}
	pbData, _ := proto.Marshal(request)
	client := eas.NewPredictClient("http://1730760139076263.cn-beijing.pai-eas.aliyuncs.com", "test_dssm_rec")
	client.SetToken(os.Getenv("TEST_TORCHREC_DSSM_REC_TOKEN"))
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	respBody, err := client.BytesPredict(pbData)
	if err != nil {
		t.Fatal(err)
	}
	responseData := &easyrec.TorchRecPBResponse{}
	proto.Unmarshal(respBody, responseData)

	alogResponse, err := torchrecEmbeddingResponseFunc(responseData)
	if err != nil {
		t.Fatal(err)
	}
	if resp, ok := alogResponse[0].(*TorchrecEmbeddingResponse); !ok {
		t.Fatal("not torchrecEmbeddingResponse")
	} else {
		t.Log(resp.GetEmbedding(), resp.dimSize)
	}

}

func buildItemEmbeddingRequest() (request *easyrec.PBRequest, err error) {
	content, err := os.ReadFile("rec_data.json")
	if err != nil {
		return
	}
	data := make(map[string]any)

	json.Unmarshal(content, &data)

	builder := easyrec.NewEasyrecRequestBuilder()

	for k, v := range data {
		if !(strings.HasPrefix(k, "user_") || strings.HasPrefix(k, "context_")) {
			m := v.(map[string]any)
			values := m["values"].([]any)
			if m["dtype"].(string) == "int64" {
				iValues := make([]any, 0)
				for _, v := range values {
					switch val := v.(type) {
					case int64:
						iValues = append(iValues, val)
					case float64:
						iValues = append(iValues, utils.ToInt64(val, 0))
					default:
						iValues = append(iValues, "")
					}
				}
				builder.AddContextFeature(k, iValues)

			} else {
				builder.AddContextFeature(k, values)
			}
		}
	}

	request = builder.EasyrecRequest()
	return

}

func TestTorchrecItemEmbeddingResponseFunc(t *testing.T) {
	request, err := buildItemEmbeddingRequest()
	if err != nil {
		t.Fatal(err)
	}
	pbData, _ := proto.Marshal(request)
	client := eas.NewPredictClient("http://1730760139076263.cn-beijing.pai-eas.aliyuncs.com", "test_dssm_item")
	client.SetToken(os.Getenv("TEST_TORCHREC_DSSM_ITEM_TOKEN"))
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	respBody, err := client.BytesPredict(pbData)
	if err != nil {
		t.Fatal(err)
	}
	responseData := &easyrec.TorchRecPBResponse{}
	proto.Unmarshal(respBody, responseData)

	alogResponse, err := torchrecEmbeddingResponseFunc(responseData)
	if err != nil {
		t.Fatal(err)
	}
	if resp, ok := alogResponse[0].(*TorchrecEmbeddingResponse); !ok {
		t.Fatal("not torchrecEmbeddingResponse")
	} else {
		t.Log(resp.GetEmbedding(), resp.dimSize)
	}

}
