package eas

import (
	"github.com/golang/protobuf/proto"
	"github.com/alibaba/pairec/v2/pkg/eas/types/torch_predict_protos"
)

// TorchRequest class for PyTorch data and requests
type TorchRequest struct {
	RequestData torch_predict_protos.PredictRequest
}

// AddFeedFloat32 function adds float values input data for torchrequest
func (tr *TorchRequest) AddFeedFloat32(index int, shape []int64, content []float32) {
	for len(tr.RequestData.Inputs) < index+1 {
		tr.RequestData.Inputs = append(tr.RequestData.Inputs, &torch_predict_protos.ArrayProto{})
	}
	requestProto := torch_predict_protos.ArrayProto{
		Dtype: TorchType_DT_FLOAT,
		ArrayShape: &torch_predict_protos.ArrayShape{
			Dim: shape,
		},
		FloatVal: content,
	}
	tr.RequestData.Inputs[index] = &requestProto
}

// AddFeedFloat64 function adds double values input data for torchrequest
func (tr *TorchRequest) AddFeedFloat64(index int, shape []int64, content []float64) {
	for len(tr.RequestData.Inputs) < index+1 {
		tr.RequestData.Inputs = append(tr.RequestData.Inputs, &torch_predict_protos.ArrayProto{})
	}
	requestProto := torch_predict_protos.ArrayProto{
		Dtype: TorchType_DT_DOUBLE,
		ArrayShape: &torch_predict_protos.ArrayShape{
			Dim: shape,
		},
		DoubleVal: content,
	}
	tr.RequestData.Inputs[index] = &requestProto
}

// AddFeedInt32 function adds int values input data for torchrequest
func (tr *TorchRequest) AddFeedInt32(index int, shape []int64, content []int32) {
	for len(tr.RequestData.Inputs) < index+1 {
		tr.RequestData.Inputs = append(tr.RequestData.Inputs, &torch_predict_protos.ArrayProto{})
	}
	requestProto := torch_predict_protos.ArrayProto{
		Dtype: TorchType_DT_INT32,
		ArrayShape: &torch_predict_protos.ArrayShape{
			Dim: shape,
		},
		IntVal: content,
	}
	tr.RequestData.Inputs[index] = &requestProto
}

// AddFeedInt64 function adds int64 values input data for torchrequest
func (tr *TorchRequest) AddFeedInt64(index int, shape []int64, content []int64) {
	for len(tr.RequestData.Inputs) < index+1 {
		tr.RequestData.Inputs = append(tr.RequestData.Inputs, &torch_predict_protos.ArrayProto{})
	}
	requestProto := torch_predict_protos.ArrayProto{
		Dtype: TorchType_DT_INT64,
		ArrayShape: &torch_predict_protos.ArrayShape{
			Dim: shape,
		},
		Int64Val: content,
	}
	tr.RequestData.Inputs[index] = &requestProto
}

// AddFetch add OutputFilter (outIndex) for response
func (tr *TorchRequest) AddFetch(outIndex int32) {
	tr.RequestData.OutputFilter = append(tr.RequestData.OutputFilter, outIndex)
}

// ToString for interface
func (tr TorchRequest) ToString() (string, error) {
	reqData, err := proto.Marshal(&tr.RequestData)
	if err != nil {
		return "", NewPredictError(-1, "", err.Error())
	}
	return string(reqData), nil
}

// TorchResponse class for PyTorch predicted results
type TorchResponse struct {
	Response torch_predict_protos.PredictResponse
}

// GetTensorShape returns []int64 slice as shape of tensor outindexed
func (resp *TorchResponse) GetTensorShape(outIndex int) []int64 {
	return resp.Response.Outputs[outIndex].ArrayShape.Dim
}

// GetFloatVal returns []float32 slice as output data
func (resp *TorchResponse) GetFloatVal(outIndex int) []float32 {
	return resp.Response.Outputs[outIndex].GetFloatVal()
}

// GetDoubleVal returns []float64 slice as output data
func (resp *TorchResponse) GetDoubleVal(outIndex int) []float64 {
	return resp.Response.Outputs[outIndex].GetDoubleVal()
}

// GetIntVal returns []int32 slice as output data
func (resp *TorchResponse) GetIntVal(outIndex int) []int32 {
	return resp.Response.Outputs[outIndex].GetIntVal()
}

// GetInt64Val returns []int64 slice as output data
func (resp *TorchResponse) GetInt64Val(outIndex int) []int64 {
	return resp.Response.Outputs[outIndex].GetInt64Val()
}

// Unmarshal for interface
func (resp *TorchResponse) unmarshal(body []byte) error {
	bd := &torch_predict_protos.PredictResponse{}
	err := proto.Unmarshal(body, bd)
	if err != nil {
		return err
	}
	resp.Response = *bd
	return nil
}
