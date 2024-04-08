package eas

import (
	"encoding/json"
	"fmt"
	"os"
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
