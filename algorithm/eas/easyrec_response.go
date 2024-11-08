package eas

import (
	"bytes"
	json "encoding/json"
	"fmt"
	"strings"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/algorithm/response"
	"github.com/alibaba/pairec/v2/utils"
)

type EasyrecResponse struct {
	RawFeatures      string
	GenerateFeatures *bytes.Buffer
	ContextFeatures  string
	multiValModule   bool
	score            float64
	scoreArr         map[string]float64
}

func (r *EasyrecResponse) GetScore() float64 {
	return r.score
}

func (r *EasyrecResponse) GetScoreMap() map[string]float64 {
	return r.scoreArr
}

func (r *EasyrecResponse) GetModuleType() bool {
	return r.multiValModule
}

func easyrecMutValResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.PBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	outputs := resp.GetOutputs()
	var response []map[string]float64
	for _, itemId := range resp.ItemIds {
		if result, ok := resp.Results[itemId]; ok {
			scores := make(map[string]float64)
			if len(outputs) == len(result.Scores) {
				for k, score := range result.Scores {
					scores[outputs[k]] = score
				}
			} else {
				err = fmt.Errorf("outputs size is not equal scores")
				return
			}

			response = append(response, scores)
		} else {
			scores := make(map[string]float64)
			for _, out := range outputs {
				scores[out] = float64(0)
			}
			response = append(response, scores)
		}
	}

	for _, v := range response {
		ret = append(ret, &EasyrecResponse{scoreArr: v, multiValModule: true})
	}

	return
}

func easyrecMutValResponseFuncDebug(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.PBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	outputs := resp.GetOutputs()
	var response []map[string]float64
	var (
		itemFeatures     []string
		generateFeatures []*bytes.Buffer
		contextFeatures  []string
	)
	for _, itemId := range resp.ItemIds {
		scores := make(map[string]float64)
		for _, out := range outputs {
			scores[out] = float64(0)
		}
		response = append(response, scores)

		if f, ok := resp.RawFeatures[itemId]; ok {
			itemFeatures = append(itemFeatures, f)
		} else {
			itemFeatures = append(itemFeatures, "")
		}

		if g, ok := resp.GenerateFeatures[itemId]; ok {
			generateFeatures = append(generateFeatures, bytes.NewBufferString(g))
		} else {
			generateFeatures = append(generateFeatures, new(bytes.Buffer))
		}
		if c, ok := resp.ContextFeatures[itemId]; ok {
			features := c.Features
			j, _ := json.Marshal(features)
			contextFeatures = append(contextFeatures, string(j))
		} else {
			contextFeatures = append(contextFeatures, "")
		}
	}

	for i, v := range response {
		ret = append(ret, &EasyrecResponse{scoreArr: v, multiValModule: true, RawFeatures: itemFeatures[i], GenerateFeatures: generateFeatures[i], ContextFeatures: contextFeatures[i]})
	}

	return
}

func easyrecResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.PBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}

	for _, itemId := range resp.ItemIds {
		if results, ok := resp.Results[itemId]; ok {
			ret = append(ret, &EasyrecResponse{score: results.Scores[0]})
		} else {
			ret = append(ret, &EasyrecResponse{score: float64(0)})
		}
	}

	return
}
func easyrecResponseFuncDebug(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.PBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}

	for _, itemId := range resp.ItemIds {
		itemFeatures := ""
		generateFeatures := new(bytes.Buffer)
		contextFeatures := ""
		if f, ok := resp.RawFeatures[itemId]; ok {
			itemFeatures = f
		}

		if g, ok := resp.GenerateFeatures[itemId]; ok {
			generateFeatures.WriteString(g)
		}

		if c, ok := resp.ContextFeatures[itemId]; ok {
			features := c.Features
			j, _ := json.Marshal(features)
			contextFeatures = string(j)
		}

		if results, ok := resp.Results[itemId]; ok {
			ret = append(ret, &EasyrecResponse{score: results.Scores[0], RawFeatures: itemFeatures, GenerateFeatures: generateFeatures, ContextFeatures: contextFeatures})
		} else {
			ret = append(ret, &EasyrecResponse{score: float64(0), RawFeatures: itemFeatures, GenerateFeatures: generateFeatures, ContextFeatures: contextFeatures})
		}
	}

	return
}

type EasyrecUserEmbResponse struct {
	userEmb string
}

func (r *EasyrecUserEmbResponse) GetScore() float64 {
	return 0
}

func (r *EasyrecUserEmbResponse) GetScoreMap() map[string]float64 {
	return nil
}

func (r *EasyrecUserEmbResponse) GetModuleType() bool {
	return false
}
func (r *EasyrecUserEmbResponse) GetUserEmb() string {
	return r.userEmb
}

func easyrecUserEmbResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.PBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}

	for _, arrayProto := range resp.GetTfOutputs() {
		if arrayProto.GetDtype() == easyrec.ArrayDataType_DT_STRING {
			if len(arrayProto.GetStringVal()) > 0 {
				ret = append(ret, &EasyrecUserEmbResponse{userEmb: string(arrayProto.GetStringVal()[0])})
				return
			}
		}

	}

	return
}

type EmbeddingInfo struct {
	ItemId string
	Score  float64
}
type EasyrecUserRealtimeEmbeddingResponse struct {
	EmbeddingList    []*EmbeddingInfo
	UserEmbedding    string
	GenerateFeatures *bytes.Buffer
}

func (r *EasyrecUserRealtimeEmbeddingResponse) GetScore() float64 {
	return 0
}

func (r *EasyrecUserRealtimeEmbeddingResponse) GetScoreMap() map[string]float64 {
	return nil
}

func (r *EasyrecUserRealtimeEmbeddingResponse) GetModuleType() bool {
	return false
}

func (r *EasyrecUserRealtimeEmbeddingResponse) GetEmbeddingList() []*EmbeddingInfo {
	return r.EmbeddingList
}
func (r *EasyrecUserRealtimeEmbeddingResponse) GetUserEmbedding() string {
	return r.UserEmbedding
}

func easyrecUserRealtimeEmbeddingResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.PBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	var embeddings []string
	for _, output := range resp.TfOutputs {
		for _, val := range output.FloatVal {
			embeddings = append(embeddings, utils.ToString(val, ""))
		}
	}

	response := &EasyrecUserRealtimeEmbeddingResponse{
		UserEmbedding: strings.Join(embeddings, ","),
	}
	if len(resp.GenerateFeatures) > 0 {
		generateFeatures := new(bytes.Buffer)
		for _, v := range resp.GenerateFeatures {
			generateFeatures.WriteString(v)
			break
		}
		response.GenerateFeatures = generateFeatures
	}

	for itemid, result := range resp.Results {
		response.EmbeddingList = append(response.EmbeddingList, &EmbeddingInfo{ItemId: itemid, Score: result.Scores[0]})
	}

	ret = append(ret, response)

	return
}

type EasyrecUserRealtimeEmbeddingMindResponse struct {
	DimSize          int
	DimLength        int
	UserEmbedding    string
	EmbeddingList    []*EmbeddingInfo
	GenerateFeatures *bytes.Buffer
}

func (r *EasyrecUserRealtimeEmbeddingMindResponse) GetScore() float64 {
	return 0
}

func (r *EasyrecUserRealtimeEmbeddingMindResponse) GetScoreMap() map[string]float64 {
	return nil
}

func (r *EasyrecUserRealtimeEmbeddingMindResponse) GetModuleType() bool {
	return false
}

func (r *EasyrecUserRealtimeEmbeddingMindResponse) GetEmbeddingList() []*EmbeddingInfo {
	return r.EmbeddingList
}
func (r *EasyrecUserRealtimeEmbeddingMindResponse) GetDimSize() int {
	return r.DimSize
}
func (r *EasyrecUserRealtimeEmbeddingMindResponse) GetDimLength() int {
	return r.DimLength
}
func (r *EasyrecUserRealtimeEmbeddingMindResponse) GetUserEmbedding() string {
	return r.UserEmbedding
}

func easyrecUserRealtimeEmbeddingMindResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.PBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}

	itemIdOutput, itemIdok := resp.TfOutputs["match_item_ids"]
	itemScoresOutput, itemScoreok := resp.TfOutputs["match_item_scores"]

	response := &EasyrecUserRealtimeEmbeddingMindResponse{}
	if itemIdok && itemScoreok {
		dimLength := itemIdOutput.ArrayShape.Dim[0]
		dimSize := itemIdOutput.ArrayShape.Dim[1]
		embeddings := make([]*EmbeddingInfo, 0, dimLength*dimSize)

		for _, val := range itemIdOutput.StringVal {
			info := &EmbeddingInfo{
				ItemId: string(val),
			}
			embeddings = append(embeddings, info)
		}
		for i, val := range itemScoresOutput.FloatVal {
			embeddings[i].Score = float64(val)
		}

		response.DimLength = int(dimLength)
		response.DimSize = int(dimSize)
		response.EmbeddingList = embeddings

	}

	if userEmbOutput, ok := resp.TfOutputs["user_interests"]; ok {
		size := int(userEmbOutput.ArrayShape.Dim[len(userEmbOutput.ArrayShape.Dim)-1])
		var embeddings []string
		var embeddingList []string
		for i, val := range userEmbOutput.FloatVal {
			embeddings = append(embeddings, utils.ToString(val, ""))
			if (i+1)%size == 0 {
				embeddingList = append(embeddingList, strings.Join(embeddings, ","))
				embeddings = embeddings[:0]
			}
		}

		response.UserEmbedding = strings.Join(embeddingList, "|")
	}

	if len(resp.GenerateFeatures) > 0 {
		generateFeatures := new(bytes.Buffer)
		for _, v := range resp.GenerateFeatures {
			generateFeatures.WriteString(v)
			break
		}
		response.GenerateFeatures = generateFeatures
	}

	ret = append(ret, response)

	return
}

func torchrecMutValResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.TorchRecPBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	outputs := resp.GetMapOutputs()
	var response []map[string]float64
	for i := range resp.ItemIds {
		scores := make(map[string]float64)
		for output, arrayProto := range outputs {
			if arrayProto.Dtype == easyrec.ArrayDataType_DT_FLOAT {
				scores[output] = float64(arrayProto.FloatVal[i])
			} else if arrayProto.Dtype == easyrec.ArrayDataType_DT_DOUBLE {
				scores[output] = arrayProto.DoubleVal[i]
			}
		}
		response = append(response, scores)
	}

	for _, v := range response {
		ret = append(ret, &EasyrecResponse{scoreArr: v, multiValModule: true})
	}

	return
}

func torchrecMutValResponseFuncDebug(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.TorchRecPBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	outputs := resp.GetMapOutputs()
	var response []map[string]float64
	var (
		itemFeatures     []string
		generateFeatures []*bytes.Buffer
	)
	for i, itemId := range resp.ItemIds {
		scores := make(map[string]float64)
		for output, arrayProto := range outputs {
			if arrayProto.Dtype == easyrec.ArrayDataType_DT_FLOAT {
				scores[output] = float64(arrayProto.FloatVal[i])
			} else if arrayProto.Dtype == easyrec.ArrayDataType_DT_DOUBLE {
				scores[output] = arrayProto.DoubleVal[i]
			}
		}
		response = append(response, scores)

		if f, ok := resp.RawFeatures[itemId]; ok {
			itemFeatures = append(itemFeatures, f)
		} else {
			itemFeatures = append(itemFeatures, "")
		}

		if g, ok := resp.GenerateFeatures[itemId]; ok {
			generateFeatures = append(generateFeatures, bytes.NewBufferString(g))
		} else {
			generateFeatures = append(generateFeatures, new(bytes.Buffer))
		}
	}

	for i, v := range response {
		ret = append(ret, &EasyrecResponse{scoreArr: v, multiValModule: true, RawFeatures: itemFeatures[i], GenerateFeatures: generateFeatures[i]})
	}

	return
}

type TorchrecEmbeddingResponse struct {
	embeddings []float32
	dimSize    int
}

func (r *TorchrecEmbeddingResponse) GetScore() float64 {
	return 0
}

func (r *TorchrecEmbeddingResponse) GetScoreMap() map[string]float64 {
	return nil
}

func (r *TorchrecEmbeddingResponse) GetModuleType() bool {
	return false
}
func (r *TorchrecEmbeddingResponse) GetEmbedding() []float32 {
	return r.embeddings
}
func (r *TorchrecEmbeddingResponse) GetEmbeddingSize() int {
	return r.dimSize
}

func torchrecEmbeddingResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.TorchRecPBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	outputs := resp.GetMapOutputs()
	var embeddings []float32
	var dimSize int
	for _, arrayProto := range outputs {

		if len(arrayProto.ArrayShape.Dim) >= 2 {
			dimSize = int(arrayProto.ArrayShape.Dim[1])
		}
		if arrayProto.Dtype == easyrec.ArrayDataType_DT_FLOAT {
			embeddings = append(embeddings, arrayProto.FloatVal...)
		}
		break
	}

	ret = append(ret, &TorchrecEmbeddingResponse{embeddings: embeddings, dimSize: dimSize})

	return
}

type TorchrecEmbeddingItemsResponse struct {
	EmbeddingItems []*EmbeddingInfo
}

func (r *TorchrecEmbeddingItemsResponse) GetScore() float64 {
	return 0
}

func (r *TorchrecEmbeddingItemsResponse) GetScoreMap() map[string]float64 {
	return nil
}

func (r *TorchrecEmbeddingItemsResponse) GetModuleType() bool {
	return false
}
func (r *TorchrecEmbeddingItemsResponse) GetEmbeddingItems() []*EmbeddingInfo {
	return r.EmbeddingItems
}

func torchrecEmbeddingItemsResponseFunc(data interface{}) (ret []response.AlgoResponse, err error) {
	resp, ok := data.(*easyrec.TorchRecPBResponse)
	if !ok {
		err = fmt.Errorf("invalid data type, %v", data)
		return
	}
	outputs := resp.GetMapOutputs()
	var embeddingItems []*EmbeddingInfo
	var dimSize int
	itemScoresOutput, itemScoreok := outputs["match_item_scores"]
	if itemScoreok {
		if len(itemScoresOutput.ArrayShape.Dim) >= 2 {
			dimSize = int(itemScoresOutput.ArrayShape.Dim[1])
		}
		embeddingItems = make([]*EmbeddingInfo, 0, dimSize)
	}

	for i, itemId := range resp.ItemIds {
		info := EmbeddingInfo{
			ItemId: itemId,
		}

		if itemScoresOutput.Dtype == easyrec.ArrayDataType_DT_FLOAT {
			info.Score = float64(itemScoresOutput.FloatVal[i])
		} else if itemScoresOutput.Dtype == easyrec.ArrayDataType_DT_DOUBLE {
			info.Score = itemScoresOutput.DoubleVal[i]
		}

		embeddingItems = append(embeddingItems, &info)
	}

	ret = append(ret, &TorchrecEmbeddingItemsResponse{EmbeddingItems: embeddingItems})

	return
}
