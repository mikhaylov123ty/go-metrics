// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: internal/server/proto/handlers.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Handlers_PostUpdate_FullMethodName  = "/server_grpc.Handlers/PostUpdate"
	Handlers_PostUpdates_FullMethodName = "/server_grpc.Handlers/PostUpdates"
	Handlers_GetValue_FullMethodName    = "/server_grpc.Handlers/GetValue"
)

// HandlersClient is the client API for Handlers service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type HandlersClient interface {
	PostUpdate(ctx context.Context, in *PostUpdateRequest, opts ...grpc.CallOption) (*PostUpdateResponse, error)
	PostUpdates(ctx context.Context, in *PostUpdatesRequest, opts ...grpc.CallOption) (*PostUpdatesResponse, error)
	GetValue(ctx context.Context, in *GetValueRequest, opts ...grpc.CallOption) (*GetValueResponse, error)
}

type handlersClient struct {
	cc grpc.ClientConnInterface
}

func NewHandlersClient(cc grpc.ClientConnInterface) HandlersClient {
	return &handlersClient{cc}
}

func (c *handlersClient) PostUpdate(ctx context.Context, in *PostUpdateRequest, opts ...grpc.CallOption) (*PostUpdateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PostUpdateResponse)
	err := c.cc.Invoke(ctx, Handlers_PostUpdate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *handlersClient) PostUpdates(ctx context.Context, in *PostUpdatesRequest, opts ...grpc.CallOption) (*PostUpdatesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PostUpdatesResponse)
	err := c.cc.Invoke(ctx, Handlers_PostUpdates_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *handlersClient) GetValue(ctx context.Context, in *GetValueRequest, opts ...grpc.CallOption) (*GetValueResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetValueResponse)
	err := c.cc.Invoke(ctx, Handlers_GetValue_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HandlersServer is the server API for Handlers service.
// All implementations must embed UnimplementedHandlersServer
// for forward compatibility.
type HandlersServer interface {
	PostUpdate(context.Context, *PostUpdateRequest) (*PostUpdateResponse, error)
	PostUpdates(context.Context, *PostUpdatesRequest) (*PostUpdatesResponse, error)
	GetValue(context.Context, *GetValueRequest) (*GetValueResponse, error)
	mustEmbedUnimplementedHandlersServer()
}

// UnimplementedHandlersServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedHandlersServer struct{}

func (UnimplementedHandlersServer) PostUpdate(context.Context, *PostUpdateRequest) (*PostUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostUpdate not implemented")
}
func (UnimplementedHandlersServer) PostUpdates(context.Context, *PostUpdatesRequest) (*PostUpdatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostUpdates not implemented")
}
func (UnimplementedHandlersServer) GetValue(context.Context, *GetValueRequest) (*GetValueResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetValue not implemented")
}
func (UnimplementedHandlersServer) mustEmbedUnimplementedHandlersServer() {}
func (UnimplementedHandlersServer) testEmbeddedByValue()                  {}

// UnsafeHandlersServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to HandlersServer will
// result in compilation errors.
type UnsafeHandlersServer interface {
	mustEmbedUnimplementedHandlersServer()
}

func RegisterHandlersServer(s grpc.ServiceRegistrar, srv HandlersServer) {
	// If the following call pancis, it indicates UnimplementedHandlersServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Handlers_ServiceDesc, srv)
}

func _Handlers_PostUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlersServer).PostUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Handlers_PostUpdate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlersServer).PostUpdate(ctx, req.(*PostUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Handlers_PostUpdates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostUpdatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlersServer).PostUpdates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Handlers_PostUpdates_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlersServer).PostUpdates(ctx, req.(*PostUpdatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Handlers_GetValue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetValueRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HandlersServer).GetValue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Handlers_GetValue_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HandlersServer).GetValue(ctx, req.(*GetValueRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Handlers_ServiceDesc is the grpc.ServiceDesc for Handlers service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Handlers_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "server_grpc.Handlers",
	HandlerType: (*HandlersServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PostUpdate",
			Handler:    _Handlers_PostUpdate_Handler,
		},
		{
			MethodName: "PostUpdates",
			Handler:    _Handlers_PostUpdates_Handler,
		},
		{
			MethodName: "GetValue",
			Handler:    _Handlers_GetValue_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/server/proto/handlers.proto",
}
