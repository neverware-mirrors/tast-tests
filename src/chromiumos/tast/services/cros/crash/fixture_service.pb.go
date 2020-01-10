// Code generated by protoc-gen-go. DO NOT EDIT.
// source: fixture_service.proto

/*
Package crash is a generated protocol buffer package.

It is generated from these files:
	fixture_service.proto

It has these top-level messages:
*/
package crash

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/empty"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for FixtureService service

type FixtureServiceClient interface {
	// SetUp sets up the DUT for a crash test.
	// *NOTE*: If the DUT reboots during the test, it will clear
	// crash_test_in_progress state.
	// After the test is complete, you must call TearDown to clean up the
	// associated resources.
	SetUp(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
	// Close releases the resources obtained by New.
	TearDown(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*google_protobuf.Empty, error)
}

type fixtureServiceClient struct {
	cc *grpc.ClientConn
}

func NewFixtureServiceClient(cc *grpc.ClientConn) FixtureServiceClient {
	return &fixtureServiceClient{cc}
}

func (c *fixtureServiceClient) SetUp(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/tast.cros.crash.FixtureService/SetUp", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fixtureServiceClient) TearDown(ctx context.Context, in *google_protobuf.Empty, opts ...grpc.CallOption) (*google_protobuf.Empty, error) {
	out := new(google_protobuf.Empty)
	err := grpc.Invoke(ctx, "/tast.cros.crash.FixtureService/TearDown", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for FixtureService service

type FixtureServiceServer interface {
	// SetUp sets up the DUT for a crash test.
	// *NOTE*: If the DUT reboots during the test, it will clear
	// crash_test_in_progress state.
	// After the test is complete, you must call TearDown to clean up the
	// associated resources.
	SetUp(context.Context, *google_protobuf.Empty) (*google_protobuf.Empty, error)
	// Close releases the resources obtained by New.
	TearDown(context.Context, *google_protobuf.Empty) (*google_protobuf.Empty, error)
}

func RegisterFixtureServiceServer(s *grpc.Server, srv FixtureServiceServer) {
	s.RegisterService(&_FixtureService_serviceDesc, srv)
}

func _FixtureService_SetUp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FixtureServiceServer).SetUp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tast.cros.crash.FixtureService/SetUp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FixtureServiceServer).SetUp(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _FixtureService_TearDown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(google_protobuf.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FixtureServiceServer).TearDown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tast.cros.crash.FixtureService/TearDown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FixtureServiceServer).TearDown(ctx, req.(*google_protobuf.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _FixtureService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "tast.cros.crash.FixtureService",
	HandlerType: (*FixtureServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetUp",
			Handler:    _FixtureService_SetUp_Handler,
		},
		{
			MethodName: "TearDown",
			Handler:    _FixtureService_TearDown_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fixture_service.proto",
}

func init() { proto.RegisterFile("fixture_service.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 171 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4d, 0xcb, 0xac, 0x28,
	0x29, 0x2d, 0x4a, 0x8d, 0x2f, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e, 0xd5, 0x2b, 0x28, 0xca, 0x2f,
	0xc9, 0x17, 0xe2, 0x2f, 0x49, 0x2c, 0x2e, 0xd1, 0x4b, 0x2e, 0xca, 0x2f, 0xd6, 0x4b, 0x2e, 0x4a,
	0x2c, 0xce, 0x90, 0x92, 0x4e, 0xcf, 0xcf, 0x4f, 0xcf, 0x49, 0xd5, 0x07, 0x4b, 0x27, 0x95, 0xa6,
	0xe9, 0xa7, 0xe6, 0x16, 0x94, 0x54, 0x42, 0x54, 0x1b, 0x75, 0x32, 0x72, 0xf1, 0xb9, 0x41, 0xcc,
	0x09, 0x86, 0x18, 0x23, 0x64, 0xc9, 0xc5, 0x1a, 0x9c, 0x5a, 0x12, 0x5a, 0x20, 0x24, 0xa6, 0x07,
	0xd1, 0xa9, 0x07, 0xd3, 0xa9, 0xe7, 0x0a, 0xd2, 0x29, 0x85, 0x43, 0x5c, 0x89, 0x41, 0xc8, 0x86,
	0x8b, 0x23, 0x24, 0x35, 0xb1, 0xc8, 0x25, 0xbf, 0x3c, 0x8f, 0x74, 0xdd, 0x4e, 0xaa, 0x51, 0xca,
	0xc9, 0x19, 0x45, 0xf9, 0xb9, 0x99, 0xa5, 0xb9, 0xf9, 0xc5, 0xfa, 0x20, 0x6f, 0xe8, 0x43, 0xbd,
	0x56, 0xac, 0x0f, 0xf2, 0x8f, 0x3e, 0xd8, 0x3f, 0x49, 0x6c, 0x60, 0x8d, 0xc6, 0x80, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x3f, 0x98, 0x27, 0xa3, 0x00, 0x01, 0x00, 0x00,
}
