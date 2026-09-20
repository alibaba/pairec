package eas

import (
	"fmt"
	"math"

	"github.com/alibaba/pairec/v2/algorithm/eas/easyrec"
	"github.com/alibaba/pairec/v2/algorithm/response"
)

func newTorchrecMultiEmbeddingResponseFunc(output string) response.ResponseFunc {
	return func(data interface{}) ([]response.AlgoResponse, error) {
		resp, ok := data.(*easyrec.TorchRecPBResponse)
		if !ok || resp == nil {
			return nil, fmt.Errorf("multi-embedding response is not TorchRecPBResponse")
		}
		tensor := resp.GetMapOutputs()[output]
		if tensor == nil || tensor.ArrayShape == nil {
			return nil, fmt.Errorf("multi-embedding output %q is missing its tensor or shape", output)
		}
		shape := tensor.ArrayShape.Dim
		if (len(shape) != 2 && len(shape) != 3) || shape[0] != 1 {
			return nil, fmt.Errorf("multi-embedding output %q expects [1,D] or [1,K,D], got %v", output, shape)
		}
		k, d := int64(1), shape[len(shape)-1]
		if len(shape) == 3 {
			k = shape[1]
		}
		n := int64(len(tensor.FloatVal))
		if tensor.Dtype != easyrec.ArrayDataType_DT_FLOAT || k <= 0 || d <= 0 || n%k != 0 || n/k != d {
			return nil, fmt.Errorf("multi-embedding output %q has invalid dtype or shape/data length", output)
		}
		multiEmbeddings := make([][]float32, int(k))
		for i := 0; i < int(k); i++ {
			row := tensor.FloatVal[i*int(d) : (i+1)*int(d)]
			for _, v := range row {
				if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
					return nil, fmt.Errorf("multi-embedding output %q has a non-finite value", output)
				}
			}
			multiEmbeddings[i] = row
		}
		return []response.AlgoResponse{&TorchrecEmbeddingResponse{multiEmbeddings: multiEmbeddings, dimSize: int(d), passThroughData: resp.GetPassThroughData()}}, nil
	}
}
