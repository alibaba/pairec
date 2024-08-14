// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: tensorflow/core/framework/resource_handle.proto

package framework

import (
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

// Protocol buffer representing a handle to a tensorflow resource. Handles are
// not valid across executions, but can be serialized back and forth from within
// a single run.
type ResourceHandleProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Unique name for the device containing the resource.
	Device string `protobuf:"bytes,1,opt,name=device,proto3" json:"device,omitempty"`
	// Container in which this resource is placed.
	Container string `protobuf:"bytes,2,opt,name=container,proto3" json:"container,omitempty"`
	// Unique name of this resource.
	Name string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	// Hash code for the type of the resource. Is only valid in the same device
	// and in the same execution.
	HashCode uint64 `protobuf:"varint,4,opt,name=hash_code,json=hashCode,proto3" json:"hash_code,omitempty"`
	// For debug-only, the name of the type pointed to by this handle, if
	// available.
	MaybeTypeName string `protobuf:"bytes,5,opt,name=maybe_type_name,json=maybeTypeName,proto3" json:"maybe_type_name,omitempty"`
}

func (x *ResourceHandleProto) Reset() {
	*x = ResourceHandleProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tensorflow_core_framework_resource_handle_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ResourceHandleProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResourceHandleProto) ProtoMessage() {}

func (x *ResourceHandleProto) ProtoReflect() protoreflect.Message {
	mi := &file_tensorflow_core_framework_resource_handle_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResourceHandleProto.ProtoReflect.Descriptor instead.
func (*ResourceHandleProto) Descriptor() ([]byte, []int) {
	return file_tensorflow_core_framework_resource_handle_proto_rawDescGZIP(), []int{0}
}

func (x *ResourceHandleProto) GetDevice() string {
	if x != nil {
		return x.Device
	}
	return ""
}

func (x *ResourceHandleProto) GetContainer() string {
	if x != nil {
		return x.Container
	}
	return ""
}

func (x *ResourceHandleProto) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ResourceHandleProto) GetHashCode() uint64 {
	if x != nil {
		return x.HashCode
	}
	return 0
}

func (x *ResourceHandleProto) GetMaybeTypeName() string {
	if x != nil {
		return x.MaybeTypeName
	}
	return ""
}

var File_tensorflow_core_framework_resource_handle_proto protoreflect.FileDescriptor

var file_tensorflow_core_framework_resource_handle_proto_rawDesc = []byte{
	0x0a, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x63, 0x6f, 0x72,
	0x65, 0x2f, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x5f, 0x68, 0x61, 0x6e, 0x64, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x0a, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x22, 0xa4, 0x01,
	0x0a, 0x13, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x1b, 0x0a, 0x09, 0x68, 0x61, 0x73, 0x68, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x08, 0x68, 0x61, 0x73, 0x68, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x26, 0x0a, 0x0f,
	0x6d, 0x61, 0x79, 0x62, 0x65, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x6d, 0x61, 0x79, 0x62, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x42, 0x6e, 0x0a, 0x18, 0x6f, 0x72, 0x67, 0x2e, 0x74, 0x65, 0x6e, 0x73,
	0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2e, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72, 0x6b,
	0x42, 0x0e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65,
	0x50, 0x01, 0x5a, 0x3d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x74,
	0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72,
	0x66, 0x6c, 0x6f, 0x77, 0x2f, 0x74, 0x65, 0x6e, 0x73, 0x6f, 0x72, 0x66, 0x6c, 0x6f, 0x77, 0x2f,
	0x67, 0x6f, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x2f, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72,
	0x6b, 0xf8, 0x01, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tensorflow_core_framework_resource_handle_proto_rawDescOnce sync.Once
	file_tensorflow_core_framework_resource_handle_proto_rawDescData = file_tensorflow_core_framework_resource_handle_proto_rawDesc
)

func file_tensorflow_core_framework_resource_handle_proto_rawDescGZIP() []byte {
	file_tensorflow_core_framework_resource_handle_proto_rawDescOnce.Do(func() {
		file_tensorflow_core_framework_resource_handle_proto_rawDescData = protoimpl.X.CompressGZIP(file_tensorflow_core_framework_resource_handle_proto_rawDescData)
	})
	return file_tensorflow_core_framework_resource_handle_proto_rawDescData
}

var file_tensorflow_core_framework_resource_handle_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_tensorflow_core_framework_resource_handle_proto_goTypes = []interface{}{
	(*ResourceHandleProto)(nil), // 0: tensorflow.ResourceHandleProto
}
var file_tensorflow_core_framework_resource_handle_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_tensorflow_core_framework_resource_handle_proto_init() }
func file_tensorflow_core_framework_resource_handle_proto_init() {
	if File_tensorflow_core_framework_resource_handle_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tensorflow_core_framework_resource_handle_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ResourceHandleProto); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tensorflow_core_framework_resource_handle_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tensorflow_core_framework_resource_handle_proto_goTypes,
		DependencyIndexes: file_tensorflow_core_framework_resource_handle_proto_depIdxs,
		MessageInfos:      file_tensorflow_core_framework_resource_handle_proto_msgTypes,
	}.Build()
	File_tensorflow_core_framework_resource_handle_proto = out.File
	file_tensorflow_core_framework_resource_handle_proto_rawDesc = nil
	file_tensorflow_core_framework_resource_handle_proto_goTypes = nil
	file_tensorflow_core_framework_resource_handle_proto_depIdxs = nil
}
