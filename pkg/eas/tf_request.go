package eas

import (
	"github.com/golang/protobuf/proto"
	"github.com/alibaba/pairec/pkg/eas/types/tf_predict_protos"
)

// TFRequest class for tensorflow data and requests
type TFRequest struct {
	RequestData tf_predict_protos.PredictRequest
}

// SetSignatureName set signature name for TensorFlow request
func (tr *TFRequest) SetSignatureName(sigName string) {
	tr.RequestData.SignatureName = sigName
}

// AddFeedFloat32 function adds float values input data for TFRequest
func (tr *TFRequest) AddFeedFloat32(inputName string, shape []int64, content []float32) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: TfType_DT_FLOAT,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		FloatVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedFloat64 function adds double values input data for TFRequest
func (tr *TFRequest) AddFeedFloat64(inputName string, shape []int64, content []float64) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: TfType_DT_DOUBLE,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		DoubleVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedInt32 function adds int values input data for TFRequest
func (tr *TFRequest) AddFeedInt32(inputName string, shape []int64, content []int32) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: TfType_DT_INT32,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		IntVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedInt64 function adds int64 values input data for TFRequest
func (tr *TFRequest) AddFeedInt64(inputName string, shape []int64, content []int64) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: TfType_DT_INT64,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		Int64Val: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedBool function adds boolean values input data for TFRequest
func (tr *TFRequest) AddFeedBool(inputName string, shape []int64, content []bool) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: TfType_DT_BOOL,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		BoolVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFeedString function adds string values input data for TFRequest
func (tr *TFRequest) AddFeedString(inputName string, shape []int64, content [][]byte) {
	requestProto := tf_predict_protos.ArrayProto{
		Dtype: TfType_DT_STRING,
		ArrayShape: &tf_predict_protos.ArrayShape{
			Dim: shape,
		},
		StringVal: content,
	}
	if tr.RequestData.Inputs == nil {
		tr.RequestData.Inputs = make(map[string]*tf_predict_protos.ArrayProto)
	}
	tr.RequestData.Inputs[inputName] = &requestProto
}

// AddFetch adds output filter (outname) for TensorFlow request
func (tr *TFRequest) AddFetch(outName string) {
	tr.RequestData.OutputFilter = append(tr.RequestData.OutputFilter, outName)
}

// ToString for interface
func (tr TFRequest) ToString() (string, error) {
	reqdata, err := proto.Marshal(&tr.RequestData)
	if err != nil {
		return "", NewPredictError(-1, "", err.Error())
	}
	return string(reqdata), nil
}

// TFResponse class for Pytf predicted results
type TFResponse struct {
	Response tf_predict_protos.PredictResponse
}

// GetTensorShape returns []int64 slice as shape of tensor outindexed
func (tresp *TFResponse) GetTensorShape(outputName string) []int64 {
	// return tresp.PredictResponse.Outputs[outputName].ArrayShape.Dim
	return tresp.Response.Outputs[outputName].ArrayShape.Dim
}

// GetFloatVal returns []float32 slice as output data
func (tresp *TFResponse) GetFloatVal(outputName string) []float32 {
	return tresp.Response.Outputs[outputName].GetFloatVal()
}

// GetDoubleVal returns []float64 slice as output data
func (tresp *TFResponse) GetDoubleVal(outputName string) []float64 {
	return tresp.Response.Outputs[outputName].GetDoubleVal()
}

// GetIntVal returns []int32 slice as output data
func (tresp *TFResponse) GetIntVal(outputName string) []int32 {
	return tresp.Response.Outputs[outputName].GetIntVal()
}

// GetInt64Val returns []int64 slice as output data
func (tresp *TFResponse) GetInt64Val(outputName string) []int64 {
	return tresp.Response.Outputs[outputName].GetInt64Val()
}

// GetBoolVal returns []bool slice as output data
func (tresp *TFResponse) GetBoolVal(outputName string) []bool {
	return tresp.Response.Outputs[outputName].GetBoolVal()
}

// GetStringVal returns []string slice as output data
func (tresp *TFResponse) GetStringVal(outputName string) [][]byte {
	return tresp.Response.Outputs[outputName].GetStringVal()
}

// Unmarshal for interface
func (tresp *TFResponse) unmarshal(body []byte) error {
	bd := &tf_predict_protos.PredictResponse{}
	err := proto.Unmarshal(body, bd)
	if err != nil {
		return err
	}
	tresp.Response = *bd
	return nil
}
