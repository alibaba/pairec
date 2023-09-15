package eas

import (
	"github.com/alibaba/pairec/pkg/eas/types/tf_predict_protos"
)

// type tf_predict_protos.ArrayDataType tf_predict_protos.ArrayDataType

// TfType_DT_INVALID and listed types use ALL_CAPS names here to consist with other language's sdk.
const (
	TfType_DT_FLOAT  tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_FLOAT
	TfType_DT_DOUBLE tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_DOUBLE
	TfType_DT_INT32  tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_INT32
	TfType_DT_UINT8  tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_UINT8
	TfType_DT_INT16  tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_INT16
	TfType_DT_INT8   tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_INT8
	TfType_DT_STRING tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_STRING
	TfType_DT_INT64  tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_INT64
	TfType_DT_BOOL   tf_predict_protos.ArrayDataType = tf_predict_protos.ArrayDataType_DT_BOOL
)
