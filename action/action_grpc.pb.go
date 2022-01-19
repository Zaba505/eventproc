// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package action

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// GatewayClient is the client API for Gateway service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type GatewayClient interface {
	ProcessAction(ctx context.Context, in *ActionRequest, opts ...grpc.CallOption) (*ActionResponse, error)
}

type gatewayClient struct {
	cc grpc.ClientConnInterface
}

func NewGatewayClient(cc grpc.ClientConnInterface) GatewayClient {
	return &gatewayClient{cc}
}

func (c *gatewayClient) ProcessAction(ctx context.Context, in *ActionRequest, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, "/event.Gateway/ProcessAction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GatewayServer is the server API for Gateway service.
// All implementations must embed UnimplementedGatewayServer
// for forward compatibility
type GatewayServer interface {
	ProcessAction(context.Context, *ActionRequest) (*ActionResponse, error)
	mustEmbedUnimplementedGatewayServer()
}

// UnimplementedGatewayServer must be embedded to have forward compatible implementations.
type UnimplementedGatewayServer struct {
}

func (UnimplementedGatewayServer) ProcessAction(context.Context, *ActionRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProcessAction not implemented")
}
func (UnimplementedGatewayServer) mustEmbedUnimplementedGatewayServer() {}

// UnsafeGatewayServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to GatewayServer will
// result in compilation errors.
type UnsafeGatewayServer interface {
	mustEmbedUnimplementedGatewayServer()
}

func RegisterGatewayServer(s grpc.ServiceRegistrar, srv GatewayServer) {
	s.RegisterService(&Gateway_ServiceDesc, srv)
}

func _Gateway_ProcessAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ActionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).ProcessAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/event.Gateway/ProcessAction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).ProcessAction(ctx, req.(*ActionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Gateway_ServiceDesc is the grpc.ServiceDesc for Gateway service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Gateway_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "event.Gateway",
	HandlerType: (*GatewayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessAction",
			Handler:    _Gateway_ProcessAction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "action.proto",
}

// ProcessorClient is the client API for Processor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProcessorClient interface {
	ProcessActions(ctx context.Context, opts ...grpc.CallOption) (Processor_ProcessActionsClient, error)
}

type processorClient struct {
	cc grpc.ClientConnInterface
}

func NewProcessorClient(cc grpc.ClientConnInterface) ProcessorClient {
	return &processorClient{cc}
}

func (c *processorClient) ProcessActions(ctx context.Context, opts ...grpc.CallOption) (Processor_ProcessActionsClient, error) {
	stream, err := c.cc.NewStream(ctx, &Processor_ServiceDesc.Streams[0], "/event.Processor/ProcessActions", opts...)
	if err != nil {
		return nil, err
	}
	x := &processorProcessActionsClient{stream}
	return x, nil
}

type Processor_ProcessActionsClient interface {
	Send(*ProcessorRequest) error
	Recv() (*ProcessorResponse, error)
	grpc.ClientStream
}

type processorProcessActionsClient struct {
	grpc.ClientStream
}

func (x *processorProcessActionsClient) Send(m *ProcessorRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *processorProcessActionsClient) Recv() (*ProcessorResponse, error) {
	m := new(ProcessorResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ProcessorServer is the server API for Processor service.
// All implementations must embed UnimplementedProcessorServer
// for forward compatibility
type ProcessorServer interface {
	ProcessActions(Processor_ProcessActionsServer) error
	mustEmbedUnimplementedProcessorServer()
}

// UnimplementedProcessorServer must be embedded to have forward compatible implementations.
type UnimplementedProcessorServer struct {
}

func (UnimplementedProcessorServer) ProcessActions(Processor_ProcessActionsServer) error {
	return status.Errorf(codes.Unimplemented, "method ProcessActions not implemented")
}
func (UnimplementedProcessorServer) mustEmbedUnimplementedProcessorServer() {}

// UnsafeProcessorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProcessorServer will
// result in compilation errors.
type UnsafeProcessorServer interface {
	mustEmbedUnimplementedProcessorServer()
}

func RegisterProcessorServer(s grpc.ServiceRegistrar, srv ProcessorServer) {
	s.RegisterService(&Processor_ServiceDesc, srv)
}

func _Processor_ProcessActions_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ProcessorServer).ProcessActions(&processorProcessActionsServer{stream})
}

type Processor_ProcessActionsServer interface {
	Send(*ProcessorResponse) error
	Recv() (*ProcessorRequest, error)
	grpc.ServerStream
}

type processorProcessActionsServer struct {
	grpc.ServerStream
}

func (x *processorProcessActionsServer) Send(m *ProcessorResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *processorProcessActionsServer) Recv() (*ProcessorRequest, error) {
	m := new(ProcessorRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Processor_ServiceDesc is the grpc.ServiceDesc for Processor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Processor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "event.Processor",
	HandlerType: (*ProcessorServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ProcessActions",
			Handler:       _Processor_ProcessActions_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "action.proto",
}