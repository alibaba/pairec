// Protocol messages for describing the configuration of the ExampleParserOp.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: tensorflow/core/example/example_parser_configuration.proto

package example

import (
	framework "github.com/alibaba/pairec/pkg/tensorflow/core/framework"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type VarLenFeatureProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Dtype                   framework.DataType `protobuf:"varint,1,opt,name=dtype,proto3,enum=tensorflow.DataType" json:"dtype,omitempty"`
	ValuesOutputTensorName  string             `protobuf:"bytes,2,opt,name=values_output_tensor_name,json=valuesOutputTensorName,proto3" json:"values_output_tensor_name,omitempty"`
	IndicesOutputTensorName string             `protobuf:"bytes,3,opt,name=indices_output_tensor_name,json=indicesOutputTensorName,proto3" json:"indices_output_tensor_name,omitempty"`
	ShapesOutputTensorName  string             `protobuf:"bytes,4,opt,name=shapes_output_tensor_name,json=shapesOutputTensorName,proto3" json:"shapes_output_tensor_name,omitempty"`
}

func (x *VarLenFeatureProto) Reset() {
	*x = VarLenFeatureProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *VarLenFeatureProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*VarLenFeatureProto) ProtoMessage() {}

func (x *VarLenFeatureProto) ProtoReflect() protoreflect.Message {
	mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use VarLenFeatureProto.ProtoReflect.Descriptor instead.
func (*VarLenFeatureProto) Descriptor() ([]byte, []int) {
	return file_tensorflow_core_example_example_parser_configuration_proto_rawDescGZIP(), []int{0}
}

func (x *VarLenFeatureProto) GetDtype() framework.DataType {
	if x != nil {
		return x.Dtype
	}
	return framework.DataType(0)
}

func (x *VarLenFeatureProto) GetValuesOutputTensorName() string {
	if x != nil {
		return x.ValuesOutputTensorName
	}
	return ""
}

func (x *VarLenFeatureProto) GetIndicesOutputTensorName() string {
	if x != nil {
		return x.IndicesOutputTensorName
	}
	return ""
}

func (x *VarLenFeatureProto) GetShapesOutputTensorName() string {
	if x != nil {
		return x.ShapesOutputTensorName
	}
	return ""
}

type FixedLenFeatureProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Dtype                  framework.DataType          `protobuf:"varint,1,opt,name=dtype,proto3,enum=tensorflow.DataType" json:"dtype,omitempty"`
	Shape                  *framework.TensorShapeProto `protobuf:"bytes,2,opt,name=shape,proto3" json:"shape,omitempty"`
	DefaultValue           *framework.TensorProto      `protobuf:"bytes,3,opt,name=default_value,json=defaultValue,proto3" json:"default_value,omitempty"`
	ValuesOutputTensorName string                      `protobuf:"bytes,4,opt,name=values_output_tensor_name,json=valuesOutputTensorName,proto3" json:"values_output_tensor_name,omitempty"`
}

func (x *FixedLenFeatureProto) Reset() {
	*x = FixedLenFeatureProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FixedLenFeatureProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FixedLenFeatureProto) ProtoMessage() {}

func (x *FixedLenFeatureProto) ProtoReflect() protoreflect.Message {
	mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FixedLenFeatureProto.ProtoReflect.Descriptor instead.
func (*FixedLenFeatureProto) Descriptor() ([]byte, []int) {
	return file_tensorflow_core_example_example_parser_configuration_proto_rawDescGZIP(), []int{1}
}

func (x *FixedLenFeatureProto) GetDtype() framework.DataType {
	if x != nil {
		return x.Dtype
	}
	return framework.DataType(0)
}

func (x *FixedLenFeatureProto) GetShape() *framework.TensorShapeProto {
	if x != nil {
		return x.Shape
	}
	return nil
}

func (x *FixedLenFeatureProto) GetDefaultValue() *framework.TensorProto {
	if x != nil {
		return x.DefaultValue
	}
	return nil
}

func (x *FixedLenFeatureProto) GetValuesOutputTensorName() string {
	if x != nil {
		return x.ValuesOutputTensorName
	}
	return ""
}

type FeatureConfiguration struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Config:
	//
	//	*FeatureConfiguration_FixedLenFeature
	//	*FeatureConfiguration_VarLenFeature
	Config isFeatureConfiguration_Config `protobuf_oneof:"config"`
}

func (x *FeatureConfiguration) Reset() {
	*x = FeatureConfiguration{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FeatureConfiguration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FeatureConfiguration) ProtoMessage() {}

func (x *FeatureConfiguration) ProtoReflect() protoreflect.Message {
	mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FeatureConfiguration.ProtoReflect.Descriptor instead.
func (*FeatureConfiguration) Descriptor() ([]byte, []int) {
	return file_tensorflow_core_example_example_parser_configuration_proto_rawDescGZIP(), []int{2}
}

func (m *FeatureConfiguration) GetConfig() isFeatureConfiguration_Config {
	if m != nil {
		return m.Config
	}
	return nil
}

func (x *FeatureConfiguration) GetFixedLenFeature() *FixedLenFeatureProto {
	if x, ok := x.GetConfig().(*FeatureConfiguration_FixedLenFeature); ok {
		return x.FixedLenFeature
	}
	return nil
}

func (x *FeatureConfiguration) GetVarLenFeature() *VarLenFeatureProto {
	if x, ok := x.GetConfig().(*FeatureConfiguration_VarLenFeature); ok {
		return x.VarLenFeature
	}
	return nil
}

type isFeatureConfiguration_Config interface {
	isFeatureConfiguration_Config()
}

type FeatureConfiguration_FixedLenFeature struct {
	FixedLenFeature *FixedLenFeatureProto `protobuf:"bytes,1,opt,name=fixed_len_feature,json=fixedLenFeature,proto3,oneof"`
}

type FeatureConfiguration_VarLenFeature struct {
	VarLenFeature *VarLenFeatureProto `protobuf:"bytes,2,opt,name=var_len_feature,json=varLenFeature,proto3,oneof"`
}

func (*FeatureConfiguration_FixedLenFeature) isFeatureConfiguration_Config() {}

func (*FeatureConfiguration_VarLenFeature) isFeatureConfiguration_Config() {}

type ExampleParserConfiguration struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FeatureMap map[string]*FeatureConfiguration `protobuf:"bytes,1,rep,name=feature_map,json=featureMap,proto3" json:"feature_map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ExampleParserConfiguration) Reset() {
	*x = ExampleParserConfiguration{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExampleParserConfiguration) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExampleParserConfiguration) ProtoMessage() {}

func (x *ExampleParserConfiguration) ProtoReflect() protoreflect.Message {
	mi := &file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExampleParserConfiguration.ProtoReflect.Descriptor instead.
func (*ExampleParserConfiguration) Descriptor() ([]byte, []int) {
	return file_tensorflow_core_example_example_parser_configuration_proto_rawDescGZIP(), []int{3}
}

func (x *ExampleParserConfiguration) GetFeatureMap() map[string]*FeatureConfiguration {
	if x != nil {
		return x.FeatureMap
	}
	return nil
}

var File_tensorflow_core_example_example_parser_configuration_proto protoreflect.FileDescriptor

var file_tensorflow_core_example_example_parser_configuration_proto_rawDesc = []byte{
	0x0a, 0x3a, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x63, 0x6f, 0x72,
	0x65, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c,
	0x65, 0x5f, 0x70, 0x61, 0x72, 0x73, 0x65, 0x72, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x74, 0x65,
	0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x1a, 0x2c, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72,
	0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77,
	0x6f, 0x72, 0x6b, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x5f, 0x73, 0x68, 0x61, 0x70, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x26, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c,
	0x6f, 0x77, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72,
	0x6b, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x25,
	0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f,
	0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf3, 0x01, 0x0a, 0x12, 0x56, 0x61, 0x72, 0x4c, 0x65, 0x6e,
	0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2a, 0x0a, 0x05,
	0x64, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x74, 0x65,
	0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x05, 0x64, 0x74, 0x79, 0x70, 0x65, 0x12, 0x39, 0x0a, 0x19, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x73, 0x5f, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x5f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x16, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x54, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x4e,
	0x61, 0x6d, 0x65, 0x12, 0x3b, 0x0a, 0x1a, 0x69, 0x6e, 0x64, 0x69, 0x63, 0x65, 0x73, 0x5f, 0x6f,
	0x75, 0x74, 0x70, 0x75, 0x74, 0x5f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x17, 0x69, 0x6e, 0x64, 0x69, 0x63, 0x65, 0x73,
	0x4f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x54, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x39, 0x0a, 0x19, 0x73, 0x68, 0x61, 0x70, 0x65, 0x73, 0x5f, 0x6f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x5f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x16, 0x73, 0x68, 0x61, 0x70, 0x65, 0x73, 0x4f, 0x75, 0x74, 0x70, 0x75,
	0x74, 0x54, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0xef, 0x01, 0x0a, 0x14,
	0x46, 0x69, 0x78, 0x65, 0x64, 0x4c, 0x65, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2a, 0x0a, 0x05, 0x64, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77,
	0x2e, 0x44, 0x61, 0x74, 0x61, 0x54, 0x79, 0x70, 0x65, 0x52, 0x05, 0x64, 0x74, 0x79, 0x70, 0x65,
	0x12, 0x32, 0x0a, 0x05, 0x73, 0x68, 0x61, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1c, 0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x54, 0x65, 0x6e,
	0x73, 0x6f, 0x72, 0x53, 0x68, 0x61, 0x70, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x05, 0x73,
	0x68, 0x61, 0x70, 0x65, 0x12, 0x3c, 0x0a, 0x0d, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x5f,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x74, 0x65,
	0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x54, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x52, 0x0c, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x12, 0x39, 0x0a, 0x19, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x5f, 0x6f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x5f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x16, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x4f, 0x75, 0x74,
	0x70, 0x75, 0x74, 0x54, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0xba, 0x01,
	0x0a, 0x14, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4e, 0x0a, 0x11, 0x66, 0x69, 0x78, 0x65, 0x64, 0x5f,
	0x6c, 0x65, 0x6e, 0x5f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x20, 0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x46,
	0x69, 0x78, 0x65, 0x64, 0x4c, 0x65, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x48, 0x00, 0x52, 0x0f, 0x66, 0x69, 0x78, 0x65, 0x64, 0x4c, 0x65, 0x6e, 0x46,
	0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x48, 0x0a, 0x0f, 0x76, 0x61, 0x72, 0x5f, 0x6c, 0x65,
	0x6e, 0x5f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1e, 0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x56, 0x61, 0x72,
	0x4c, 0x65, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x48,
	0x00, 0x52, 0x0d, 0x76, 0x61, 0x72, 0x4c, 0x65, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x42, 0x08, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0xd6, 0x01, 0x0a, 0x1a, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x50, 0x61, 0x72, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x57, 0x0a, 0x0b, 0x66, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x5f, 0x6d, 0x61, 0x70, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x36,
	0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x45, 0x78, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x50, 0x61, 0x72, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x75,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x4d, 0x61,
	0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x4d,
	0x61, 0x70, 0x1a, 0x5f, 0x0a, 0x0f, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x4d, 0x61, 0x70,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x36, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66,
	0x6c, 0x6f, 0x77, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x42, 0x7c, 0x0a, 0x16, 0x6f, 0x72, 0x67, 0x2e, 0x74, 0x65, 0x6e, 0x73, 0x6f,
	0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x42, 0x20, 0x45,
	0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x50, 0x61, 0x72, 0x73, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x50,
	0x01, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74, 0x65,
	0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66,
	0x6c, 0x6f, 0x77, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x67,
	0x6f, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0xf8, 0x01,
	0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tensorflow_core_example_example_parser_configuration_proto_rawDescOnce sync.Once
	file_tensorflow_core_example_example_parser_configuration_proto_rawDescData = file_tensorflow_core_example_example_parser_configuration_proto_rawDesc
)

func file_tensorflow_core_example_example_parser_configuration_proto_rawDescGZIP() []byte {
	file_tensorflow_core_example_example_parser_configuration_proto_rawDescOnce.Do(func() {
		file_tensorflow_core_example_example_parser_configuration_proto_rawDescData = protoimpl.X.CompressGZIP(file_tensorflow_core_example_example_parser_configuration_proto_rawDescData)
	})
	return file_tensorflow_core_example_example_parser_configuration_proto_rawDescData
}

var file_tensorflow_core_example_example_parser_configuration_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_tensorflow_core_example_example_parser_configuration_proto_goTypes = []interface{}{
	(*VarLenFeatureProto)(nil),         // 0: tensorflow.VarLenFeatureProto
	(*FixedLenFeatureProto)(nil),       // 1: tensorflow.FixedLenFeatureProto
	(*FeatureConfiguration)(nil),       // 2: tensorflow.FeatureConfiguration
	(*ExampleParserConfiguration)(nil), // 3: tensorflow.ExampleParserConfiguration
	nil,                                // 4: tensorflow.ExampleParserConfiguration.FeatureMapEntry
	(framework.DataType)(0),            // 5: tensorflow.DataType
	(*framework.TensorShapeProto)(nil), // 6: tensorflow.TensorShapeProto
	(*framework.TensorProto)(nil),      // 7: tensorflow.TensorProto
}
var file_tensorflow_core_example_example_parser_configuration_proto_depIdxs = []int32{
	5, // 0: tensorflow.VarLenFeatureProto.dtype:type_name -> tensorflow.DataType
	5, // 1: tensorflow.FixedLenFeatureProto.dtype:type_name -> tensorflow.DataType
	6, // 2: tensorflow.FixedLenFeatureProto.shape:type_name -> tensorflow.TensorShapeProto
	7, // 3: tensorflow.FixedLenFeatureProto.default_value:type_name -> tensorflow.TensorProto
	1, // 4: tensorflow.FeatureConfiguration.fixed_len_feature:type_name -> tensorflow.FixedLenFeatureProto
	0, // 5: tensorflow.FeatureConfiguration.var_len_feature:type_name -> tensorflow.VarLenFeatureProto
	4, // 6: tensorflow.ExampleParserConfiguration.feature_map:type_name -> tensorflow.ExampleParserConfiguration.FeatureMapEntry
	2, // 7: tensorflow.ExampleParserConfiguration.FeatureMapEntry.value:type_name -> tensorflow.FeatureConfiguration
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_tensorflow_core_example_example_parser_configuration_proto_init() }
func file_tensorflow_core_example_example_parser_configuration_proto_init() {
	if File_tensorflow_core_example_example_parser_configuration_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*VarLenFeatureProto); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FixedLenFeatureProto); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FeatureConfiguration); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExampleParserConfiguration); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_tensorflow_core_example_example_parser_configuration_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*FeatureConfiguration_FixedLenFeature)(nil),
		(*FeatureConfiguration_VarLenFeature)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tensorflow_core_example_example_parser_configuration_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tensorflow_core_example_example_parser_configuration_proto_goTypes,
		DependencyIndexes: file_tensorflow_core_example_example_parser_configuration_proto_depIdxs,
		MessageInfos:      file_tensorflow_core_example_example_parser_configuration_proto_msgTypes,
	}.Build()
	File_tensorflow_core_example_example_parser_configuration_proto = out.File
	file_tensorflow_core_example_example_parser_configuration_proto_rawDesc = nil
	file_tensorflow_core_example_example_parser_configuration_proto_goTypes = nil
	file_tensorflow_core_example_example_parser_configuration_proto_depIdxs = nil
}
