// Code generated by protoc-gen-go. DO NOT EDIT.
// source: perf_boot_service.proto

package arc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	duration "github.com/golang/protobuf/ptypes/duration"
	empty "github.com/golang/protobuf/ptypes/empty"
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

type GetPerfValuesResponse struct {
	Values               []*GetPerfValuesResponse_PerfValue `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                           `json:"-"`
	XXX_unrecognized     []byte                             `json:"-"`
	XXX_sizecache        int32                              `json:"-"`
}

func (m *GetPerfValuesResponse) Reset()         { *m = GetPerfValuesResponse{} }
func (m *GetPerfValuesResponse) String() string { return proto.CompactTextString(m) }
func (*GetPerfValuesResponse) ProtoMessage()    {}
func (*GetPerfValuesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e5cf6399dafe784, []int{0}
}

func (m *GetPerfValuesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPerfValuesResponse.Unmarshal(m, b)
}
func (m *GetPerfValuesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPerfValuesResponse.Marshal(b, m, deterministic)
}
func (m *GetPerfValuesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPerfValuesResponse.Merge(m, src)
}
func (m *GetPerfValuesResponse) XXX_Size() int {
	return xxx_messageInfo_GetPerfValuesResponse.Size(m)
}
func (m *GetPerfValuesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPerfValuesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetPerfValuesResponse proto.InternalMessageInfo

func (m *GetPerfValuesResponse) GetValues() []*GetPerfValuesResponse_PerfValue {
	if m != nil {
		return m.Values
	}
	return nil
}

type GetPerfValuesResponse_PerfValue struct {
	Name                 string             `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Duration             *duration.Duration `protobuf:"bytes,2,opt,name=duration,proto3" json:"duration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *GetPerfValuesResponse_PerfValue) Reset()         { *m = GetPerfValuesResponse_PerfValue{} }
func (m *GetPerfValuesResponse_PerfValue) String() string { return proto.CompactTextString(m) }
func (*GetPerfValuesResponse_PerfValue) ProtoMessage()    {}
func (*GetPerfValuesResponse_PerfValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_7e5cf6399dafe784, []int{0, 0}
}

func (m *GetPerfValuesResponse_PerfValue) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPerfValuesResponse_PerfValue.Unmarshal(m, b)
}
func (m *GetPerfValuesResponse_PerfValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPerfValuesResponse_PerfValue.Marshal(b, m, deterministic)
}
func (m *GetPerfValuesResponse_PerfValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPerfValuesResponse_PerfValue.Merge(m, src)
}
func (m *GetPerfValuesResponse_PerfValue) XXX_Size() int {
	return xxx_messageInfo_GetPerfValuesResponse_PerfValue.Size(m)
}
func (m *GetPerfValuesResponse_PerfValue) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPerfValuesResponse_PerfValue.DiscardUnknown(m)
}

var xxx_messageInfo_GetPerfValuesResponse_PerfValue proto.InternalMessageInfo

func (m *GetPerfValuesResponse_PerfValue) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *GetPerfValuesResponse_PerfValue) GetDuration() *duration.Duration {
	if m != nil {
		return m.Duration
	}
	return nil
}

func init() {
	proto.RegisterType((*GetPerfValuesResponse)(nil), "tast.cros.arc.GetPerfValuesResponse")
	proto.RegisterType((*GetPerfValuesResponse_PerfValue)(nil), "tast.cros.arc.GetPerfValuesResponse.PerfValue")
}

func init() { proto.RegisterFile("perf_boot_service.proto", fileDescriptor_7e5cf6399dafe784) }

var fileDescriptor_7e5cf6399dafe784 = []byte{
	// 293 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x51, 0xdb, 0x4a, 0xc3, 0x30,
	0x18, 0x5e, 0x54, 0x86, 0xcb, 0x18, 0x42, 0xf0, 0x30, 0x2b, 0x48, 0xad, 0x5e, 0xf4, 0x2a, 0x81,
	0x8a, 0x2f, 0xb0, 0xcd, 0xc3, 0x9d, 0xa3, 0xb2, 0x09, 0xde, 0x8c, 0xb4, 0xfe, 0x9d, 0x85, 0xb6,
	0x7f, 0x49, 0xd2, 0x89, 0xef, 0xe4, 0xbd, 0xaf, 0x27, 0x3d, 0xac, 0x30, 0x75, 0xe0, 0x5d, 0xf8,
	0x4e, 0xf9, 0xf2, 0x85, 0x9e, 0xe4, 0xa0, 0xa2, 0x45, 0x80, 0x68, 0x16, 0x1a, 0xd4, 0x2a, 0x0e,
	0x81, 0xe7, 0x0a, 0x0d, 0xb2, 0x81, 0x91, 0xda, 0xf0, 0x50, 0xa1, 0xe6, 0x52, 0x85, 0xd6, 0xf9,
	0x12, 0x71, 0x99, 0x80, 0xa8, 0xc8, 0xa0, 0x88, 0xc4, 0x6b, 0xa1, 0xa4, 0x89, 0x31, 0xab, 0xe5,
	0xd6, 0xd9, 0x4f, 0x1e, 0xd2, 0xdc, 0x7c, 0xd4, 0xa4, 0xf3, 0x45, 0xe8, 0xd1, 0x3d, 0x98, 0x29,
	0xa8, 0x68, 0x2e, 0x93, 0x02, 0xb4, 0x0f, 0x3a, 0xc7, 0x4c, 0x03, 0xbb, 0xa3, 0xdd, 0x55, 0x85,
	0x0c, 0x89, 0xbd, 0xeb, 0xf6, 0x3d, 0xce, 0x37, 0xae, 0xe5, 0x7f, 0xba, 0x78, 0x0b, 0xf9, 0x8d,
	0xdb, 0x9a, 0xd3, 0x5e, 0x0b, 0x32, 0x46, 0xf7, 0x32, 0x99, 0xc2, 0x90, 0xd8, 0xc4, 0xed, 0xf9,
	0xd5, 0x99, 0xdd, 0xd0, 0xfd, 0x75, 0xe3, 0xe1, 0x8e, 0x4d, 0xdc, 0xbe, 0x77, 0xca, 0xeb, 0xca,
	0x7c, 0x5d, 0x99, 0x4f, 0x1a, 0x81, 0xdf, 0x4a, 0xbd, 0x4f, 0x42, 0x0f, 0xca, 0xe0, 0x11, 0xa2,
	0x79, 0xaa, 0xf7, 0x61, 0x0f, 0xf4, 0xf0, 0x59, 0xc6, 0x66, 0x96, 0x99, 0x38, 0x19, 0x4f, 0x67,
	0x63, 0xc4, 0x64, 0x82, 0xef, 0x19, 0x3b, 0xfe, 0x15, 0x78, 0x5b, 0x6e, 0x60, 0x6d, 0xc1, 0x9d,
	0x0e, 0x7b, 0xa4, 0x83, 0x8d, 0x07, 0x6e, 0x8d, 0xb8, 0xfa, 0xcf, 0x2c, 0x4e, 0x67, 0x74, 0xf9,
	0x72, 0x11, 0xbe, 0x29, 0x4c, 0xe3, 0x22, 0x45, 0x2d, 0x4a, 0x8f, 0x68, 0x7e, 0x55, 0x8b, 0xd2,
	0x2c, 0xa4, 0x0a, 0x83, 0x6e, 0x15, 0x7e, 0xfd, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x63, 0x4d, 0x25,
	0xc9, 0xfb, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PerfBootServiceClient is the client API for PerfBootService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PerfBootServiceClient interface {
	// WaitUntilCPUCoolDown internally calls power.WaitUntilCPUCoolDown on DUT
	// and waits until CPU is cooled down.
	WaitUntilCPUCoolDown(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error)
	// GetPerfValues signs in to DUT and measures Android boot performance metrics.
	GetPerfValues(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*GetPerfValuesResponse, error)
}

type perfBootServiceClient struct {
	cc *grpc.ClientConn
}

func NewPerfBootServiceClient(cc *grpc.ClientConn) PerfBootServiceClient {
	return &perfBootServiceClient{cc}
}

func (c *perfBootServiceClient) WaitUntilCPUCoolDown(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/tast.cros.arc.PerfBootService/WaitUntilCPUCoolDown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *perfBootServiceClient) GetPerfValues(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*GetPerfValuesResponse, error) {
	out := new(GetPerfValuesResponse)
	err := c.cc.Invoke(ctx, "/tast.cros.arc.PerfBootService/GetPerfValues", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PerfBootServiceServer is the server API for PerfBootService service.
type PerfBootServiceServer interface {
	// WaitUntilCPUCoolDown internally calls power.WaitUntilCPUCoolDown on DUT
	// and waits until CPU is cooled down.
	WaitUntilCPUCoolDown(context.Context, *empty.Empty) (*empty.Empty, error)
	// GetPerfValues signs in to DUT and measures Android boot performance metrics.
	GetPerfValues(context.Context, *empty.Empty) (*GetPerfValuesResponse, error)
}

// UnimplementedPerfBootServiceServer can be embedded to have forward compatible implementations.
type UnimplementedPerfBootServiceServer struct {
}

func (*UnimplementedPerfBootServiceServer) WaitUntilCPUCoolDown(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WaitUntilCPUCoolDown not implemented")
}
func (*UnimplementedPerfBootServiceServer) GetPerfValues(ctx context.Context, req *empty.Empty) (*GetPerfValuesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPerfValues not implemented")
}

func RegisterPerfBootServiceServer(s *grpc.Server, srv PerfBootServiceServer) {
	s.RegisterService(&_PerfBootService_serviceDesc, srv)
}

func _PerfBootService_WaitUntilCPUCoolDown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PerfBootServiceServer).WaitUntilCPUCoolDown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tast.cros.arc.PerfBootService/WaitUntilCPUCoolDown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PerfBootServiceServer).WaitUntilCPUCoolDown(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _PerfBootService_GetPerfValues_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PerfBootServiceServer).GetPerfValues(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tast.cros.arc.PerfBootService/GetPerfValues",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PerfBootServiceServer).GetPerfValues(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _PerfBootService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tast.cros.arc.PerfBootService",
	HandlerType: (*PerfBootServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WaitUntilCPUCoolDown",
			Handler:    _PerfBootService_WaitUntilCPUCoolDown_Handler,
		},
		{
			MethodName: "GetPerfValues",
			Handler:    _PerfBootService_GetPerfValues_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "perf_boot_service.proto",
}
