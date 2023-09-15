package tfserving

type ArrayDataType int32

const (
	// Not a legal value for DataType. Used to indicate a DataType field
	// has not been set.
	ArrayDataType_DT_INVALID ArrayDataType = 0
	// Data types that all computation devices are expected to be
	// capable to support.
	ArrayDataType_DT_FLOAT      ArrayDataType = 1
	ArrayDataType_DT_DOUBLE     ArrayDataType = 2
	ArrayDataType_DT_INT32      ArrayDataType = 3
	ArrayDataType_DT_UINT8      ArrayDataType = 4
	ArrayDataType_DT_INT16      ArrayDataType = 5
	ArrayDataType_DT_INT8       ArrayDataType = 6
	ArrayDataType_DT_STRING     ArrayDataType = 7
	ArrayDataType_DT_COMPLEX64  ArrayDataType = 8
	ArrayDataType_DT_INT64      ArrayDataType = 9
	ArrayDataType_DT_BOOL       ArrayDataType = 10
	ArrayDataType_DT_QINT8      ArrayDataType = 11
	ArrayDataType_DT_QUINT8     ArrayDataType = 12
	ArrayDataType_DT_QINT32     ArrayDataType = 13
	ArrayDataType_DT_BFLOAT16   ArrayDataType = 14
	ArrayDataType_DT_QINT16     ArrayDataType = 15
	ArrayDataType_DT_QUINT16    ArrayDataType = 16
	ArrayDataType_DT_UINT16     ArrayDataType = 17
	ArrayDataType_DT_COMPLEX128 ArrayDataType = 18
	ArrayDataType_DT_HALF       ArrayDataType = 19
	ArrayDataType_DT_RESOURCE   ArrayDataType = 20
	ArrayDataType_DT_VARIANT    ArrayDataType = 21
	ArrayDataType_DT_DIM_FLOAT  ArrayDataType = 22
	ArrayDataType_DT_DIM_DOUBLE ArrayDataType = 23
)

type PredictRequest struct {
	// A named signature to evaluate. If unspecified, the default signature
	// will be used
	SignatureName string `protobuf:"bytes,1,opt,name=signature_name,json=signatureName,proto3" json:"signature_name,omitempty"`
	// Input tensors.
	// Names of input tensor are alias names. The mapping from aliases to real
	// input tensor names is expected to be stored as named generic signature
	// under the key "inputs" in the model export.
	// Each alias listed in a generic signature named "inputs" should be provided
	// exactly once in order to run the prediction.
	Inputs map[string]interface{} `protobuf:"bytes,2,rep,name=inputs,proto3" json:"inputs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// Output filter.
	// Names specified are alias names. The mapping from aliases to real output
	// tensor names is expected to be stored as named generic signature under
	// the key "outputs" in the model export.
	// Only tensors specified here will be run/fetched and returned, with the
	// exception that when none is specified, all tensors specified in the
	// named signature will be run/fetched and returned.
	OutputFilter []string `protobuf:"bytes,3,rep,name=output_filter,json=outputFilter,proto3" json:"output_filter,omitempty"`
}

func NewPredictRequest() *PredictRequest {
	request := PredictRequest{}
	request.Inputs = make(map[string]interface{}, 0)

	return &request
}

type ArrayProto struct {
	// Data Type.
	Dtype ArrayDataType `protobuf:"varint,1,opt,name=dtype,proto3,enum=ArrayDataType" json:"dtype,omitempty"`

	// DT_FLOAT.
	FloatVal []float32 `protobuf:"fixed32,3,rep,packed,name=float_val,json=floatVal,proto3" json:"float_val,omitempty"`
	// DT_DOUBLE.
	DoubleVal []float64 `protobuf:"fixed64,4,rep,packed,name=double_val,json=doubleVal,proto3" json:"double_val,omitempty"`
	// DT_INT32, DT_INT16, DT_INT8, DT_UINT8.
	IntVal []int32 `protobuf:"varint,5,rep,packed,name=int_val,json=intVal,proto3" json:"int_val,omitempty"`
	// DT_STRING.
	StringVal []string `protobuf:"bytes,6,rep,name=string_val,json=stringVal,proto3" json:"string_val,omitempty"`
	// DT_INT64.
	Int64Val []int64 `protobuf:"varint,7,rep,packed,name=int64_val,json=int64Val,proto3" json:"int64_val,omitempty"`
	// DT_BOOL.
	BoolVal []bool `protobuf:"varint,8,rep,packed,name=bool_val,json=boolVal,proto3" json:"bool_val,omitempty"`

	DimFloatVal [][]float32 `protobuf:"fixed32,3,rep,packed,name=float_val,json=dimFloatVal,proto3" json:"dim_float_val,omitempty"`

	DimDoubleVal [][]float64 `protobuf:"fixed32,3,rep,packed,name=float_val,json=dimDoubleVal,proto3" json:"dim_double_val,omitempty"`
}

func (v ArrayProto) GetVal() interface{} {
	switch v.Dtype {
	case ArrayDataType_DT_BOOL:
		return v.BoolVal
	case ArrayDataType_DT_FLOAT:
		return v.FloatVal
	case ArrayDataType_DT_DOUBLE:
		return v.DoubleVal
	case ArrayDataType_DT_STRING:
		return v.StringVal
	case ArrayDataType_DT_INT32:
		return v.IntVal
	case ArrayDataType_DT_INT64:
		return v.Int64Val
	case ArrayDataType_DT_DIM_FLOAT:
		return v.DimFloatVal
	case ArrayDataType_DT_DIM_DOUBLE:
		return v.DimDoubleVal
	default:
		return nil
	}
}

type PredictResponse struct {
	// Output tensors.
	Outputs [][]float64 `protobuf:"bytes,1,rep,name=outputs,proto3" json:"outputs,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}
