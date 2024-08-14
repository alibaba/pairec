package eas

import (
	"github.com/alibaba/pairec/v2/pkg/eas/types/torch_predict_protos"
)

// type torch_predict_protos.ArrayDataType torch_predict_protos.ArrayDataType

const (
	// TorchType_DT_FLOAT and listed types use ALL_CAPS names here to consist with other language's sdk.
	TorchType_DT_FLOAT  torch_predict_protos.ArrayDataType = torch_predict_protos.ArrayDataType_DT_FLOAT
	TorchType_DT_DOUBLE torch_predict_protos.ArrayDataType = torch_predict_protos.ArrayDataType_DT_DOUBLE
	TorchType_DT_INT32  torch_predict_protos.ArrayDataType = torch_predict_protos.ArrayDataType_DT_INT32
	TorchType_DT_UINT8  torch_predict_protos.ArrayDataType = torch_predict_protos.ArrayDataType_DT_UINT8
	TorchType_DT_INT16  torch_predict_protos.ArrayDataType = torch_predict_protos.ArrayDataType_DT_INT16
	TorchType_DT_INT8   torch_predict_protos.ArrayDataType = torch_predict_protos.ArrayDataType_DT_INT8
	TorchType_DT_INT64  torch_predict_protos.ArrayDataType = torch_predict_protos.ArrayDataType_DT_INT64
)
