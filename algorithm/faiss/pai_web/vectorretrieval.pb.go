// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vectorretrieval.proto

package pai_web

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type VectorRequest struct {
	K                    uint32    `protobuf:"varint,1,opt,name=k,proto3" json:"k,omitempty"`
	Vector               []float32 `protobuf:"fixed32,2,rep,packed,name=vector,proto3" json:"vector,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *VectorRequest) Reset()         { *m = VectorRequest{} }
func (m *VectorRequest) String() string { return proto.CompactTextString(m) }
func (*VectorRequest) ProtoMessage()    {}
func (*VectorRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_41265fb1df0b580d, []int{0}
}

func (m *VectorRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VectorRequest.Unmarshal(m, b)
}
func (m *VectorRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VectorRequest.Marshal(b, m, deterministic)
}
func (m *VectorRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VectorRequest.Merge(m, src)
}
func (m *VectorRequest) XXX_Size() int {
	return xxx_messageInfo_VectorRequest.Size(m)
}
func (m *VectorRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_VectorRequest.DiscardUnknown(m)
}

var xxx_messageInfo_VectorRequest proto.InternalMessageInfo

func (m *VectorRequest) GetK() uint32 {
	if m != nil {
		return m.K
	}
	return 0
}

func (m *VectorRequest) GetVector() []float32 {
	if m != nil {
		return m.Vector
	}
	return nil
}

type VectorReply struct {
	Retval               []uint64  `protobuf:"varint,1,rep,packed,name=retval,proto3" json:"retval,omitempty"`
	Scores               []float32 `protobuf:"fixed32,2,rep,packed,name=scores,proto3" json:"scores,omitempty"`
	Labels               []string  `protobuf:"bytes,3,rep,name=labels,proto3" json:"labels,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *VectorReply) Reset()         { *m = VectorReply{} }
func (m *VectorReply) String() string { return proto.CompactTextString(m) }
func (*VectorReply) ProtoMessage()    {}
func (*VectorReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_41265fb1df0b580d, []int{1}
}

func (m *VectorReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VectorReply.Unmarshal(m, b)
}
func (m *VectorReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VectorReply.Marshal(b, m, deterministic)
}
func (m *VectorReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VectorReply.Merge(m, src)
}
func (m *VectorReply) XXX_Size() int {
	return xxx_messageInfo_VectorReply.Size(m)
}
func (m *VectorReply) XXX_DiscardUnknown() {
	xxx_messageInfo_VectorReply.DiscardUnknown(m)
}

var xxx_messageInfo_VectorReply proto.InternalMessageInfo

func (m *VectorReply) GetRetval() []uint64 {
	if m != nil {
		return m.Retval
	}
	return nil
}

func (m *VectorReply) GetScores() []float32 {
	if m != nil {
		return m.Scores
	}
	return nil
}

func (m *VectorReply) GetLabels() []string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func init() {
	proto.RegisterType((*VectorRequest)(nil), "pai.web.VectorRequest")
	proto.RegisterType((*VectorReply)(nil), "pai.web.VectorReply")
}

func init() { proto.RegisterFile("vectorretrieval.proto", fileDescriptor_41265fb1df0b580d) }

var fileDescriptor_41265fb1df0b580d = []byte{
	// 200 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x8f, 0x31, 0x4b, 0xc5, 0x30,
	0x14, 0x85, 0x4d, 0x23, 0x11, 0xaf, 0x3e, 0x94, 0xa0, 0x8f, 0xd0, 0x29, 0x74, 0xca, 0x94, 0x41,
	0x17, 0x5d, 0x5d, 0xdd, 0x22, 0xb8, 0x39, 0xa4, 0xe5, 0x82, 0xa5, 0x81, 0xc4, 0x24, 0x56, 0xfa,
	0xef, 0xa5, 0x4d, 0x0a, 0xca, 0x1b, 0xf3, 0xe5, 0xdc, 0x8f, 0x73, 0xe0, 0x7e, 0xc6, 0x21, 0xfb,
	0x18, 0x31, 0xc7, 0x11, 0x67, 0xeb, 0x74, 0x88, 0x3e, 0x7b, 0x7e, 0x11, 0xec, 0xa8, 0x7f, 0xb0,
	0xef, 0x9e, 0xe1, 0xf0, 0xbe, 0x25, 0x0c, 0x7e, 0x7d, 0x63, 0xca, 0xfc, 0x1a, 0xc8, 0x24, 0x88,
	0x24, 0xea, 0x60, 0xc8, 0xc4, 0x5b, 0x60, 0x45, 0x20, 0x1a, 0x49, 0x55, 0xf3, 0xd2, 0xdc, 0x12,
	0x53, 0x49, 0xf7, 0x01, 0x57, 0xfb, 0x69, 0x70, 0xcb, 0x1a, 0x8d, 0x98, 0x67, 0xeb, 0x04, 0x91,
	0x54, 0x9d, 0x97, 0x68, 0x21, 0xeb, 0x5f, 0x1a, 0x7c, 0xc4, 0xf4, 0x57, 0x53, 0x08, 0x3f, 0x02,
	0x73, 0xb6, 0x47, 0x97, 0x04, 0x95, 0x54, 0x5d, 0x9a, 0xfa, 0x7a, 0x78, 0x85, 0x9b, 0x5d, 0x5f,
	0xbb, 0xf3, 0x27, 0x60, 0x6f, 0x68, 0xe3, 0xf0, 0xc9, 0x8f, 0xba, 0x0e, 0xd0, 0xff, 0xda, 0xb7,
	0x77, 0x27, 0x3c, 0xb8, 0xa5, 0x3b, 0xeb, 0xd9, 0x36, 0xfb, 0xf1, 0x37, 0x00, 0x00, 0xff, 0xff,
	0x42, 0x02, 0x45, 0xe8, 0x0f, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// VectorRetrievalClient is the client API for VectorRetrieval service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type VectorRetrievalClient interface {
	Search(ctx context.Context, in *VectorRequest, opts ...grpc.CallOption) (*VectorReply, error)
}

type vectorRetrievalClient struct {
	cc *grpc.ClientConn
}

func NewVectorRetrievalClient(cc *grpc.ClientConn) VectorRetrievalClient {
	return &vectorRetrievalClient{cc}
}

func (c *vectorRetrievalClient) Search(ctx context.Context, in *VectorRequest, opts ...grpc.CallOption) (*VectorReply, error) {
	out := new(VectorReply)
	err := c.cc.Invoke(ctx, "/pai.web.VectorRetrieval/Search", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VectorRetrievalServer is the server API for VectorRetrieval service.
type VectorRetrievalServer interface {
	Search(context.Context, *VectorRequest) (*VectorReply, error)
}

// UnimplementedVectorRetrievalServer can be embedded to have forward compatible implementations.
type UnimplementedVectorRetrievalServer struct {
}

func (*UnimplementedVectorRetrievalServer) Search(ctx context.Context, req *VectorRequest) (*VectorReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Search not implemented")
}

func RegisterVectorRetrievalServer(s *grpc.Server, srv VectorRetrievalServer) {
	s.RegisterService(&_VectorRetrieval_serviceDesc, srv)
}

func _VectorRetrieval_Search_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VectorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VectorRetrievalServer).Search(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pai.web.VectorRetrieval/Search",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VectorRetrievalServer).Search(ctx, req.(*VectorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _VectorRetrieval_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pai.web.VectorRetrieval",
	HandlerType: (*VectorRetrievalServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Search",
			Handler:    _VectorRetrieval_Search_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "vectorretrieval.proto",
}
