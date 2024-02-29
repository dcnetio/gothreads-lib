// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.3
// source: threadsnet.proto

package threads_net_pb

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

// APIClient is the client API for API service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type APIClient interface {
	GetHostID(ctx context.Context, in *GetHostIDRequest, opts ...grpc.CallOption) (*GetHostIDReply, error)
	GetToken(ctx context.Context, opts ...grpc.CallOption) (API_GetTokenClient, error)
	CreateThread(ctx context.Context, in *CreateThreadRequest, opts ...grpc.CallOption) (*ThreadInfoReply, error)
	AddThread(ctx context.Context, in *AddThreadRequest, opts ...grpc.CallOption) (*ThreadInfoReply, error)
	GetThread(ctx context.Context, in *GetThreadRequest, opts ...grpc.CallOption) (*ThreadInfoReply, error)
	PullThread(ctx context.Context, in *PullThreadRequest, opts ...grpc.CallOption) (*PullThreadReply, error)
	DeleteThread(ctx context.Context, in *DeleteThreadRequest, opts ...grpc.CallOption) (*DeleteThreadReply, error)
	AddReplicator(ctx context.Context, in *AddReplicatorRequest, opts ...grpc.CallOption) (*AddReplicatorReply, error)
	CreateRecord(ctx context.Context, in *CreateRecordRequest, opts ...grpc.CallOption) (*NewRecordReply, error)
	AddRecord(ctx context.Context, in *AddRecordRequest, opts ...grpc.CallOption) (*AddRecordReply, error)
	GetRecord(ctx context.Context, in *GetRecordRequest, opts ...grpc.CallOption) (*GetRecordReply, error)
	Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (API_SubscribeClient, error)
}

type aPIClient struct {
	cc grpc.ClientConnInterface
}

func NewAPIClient(cc grpc.ClientConnInterface) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) GetHostID(ctx context.Context, in *GetHostIDRequest, opts ...grpc.CallOption) (*GetHostIDReply, error) {
	out := new(GetHostIDReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/GetHostID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) GetToken(ctx context.Context, opts ...grpc.CallOption) (API_GetTokenClient, error) {
	stream, err := c.cc.NewStream(ctx, &API_ServiceDesc.Streams[0], "/threads.net.pb.API/GetToken", opts...)
	if err != nil {
		return nil, err
	}
	x := &aPIGetTokenClient{stream}
	return x, nil
}

type API_GetTokenClient interface {
	Send(*GetTokenRequest) error
	Recv() (*GetTokenReply, error)
	grpc.ClientStream
}

type aPIGetTokenClient struct {
	grpc.ClientStream
}

func (x *aPIGetTokenClient) Send(m *GetTokenRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *aPIGetTokenClient) Recv() (*GetTokenReply, error) {
	m := new(GetTokenReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *aPIClient) CreateThread(ctx context.Context, in *CreateThreadRequest, opts ...grpc.CallOption) (*ThreadInfoReply, error) {
	out := new(ThreadInfoReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/CreateThread", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) AddThread(ctx context.Context, in *AddThreadRequest, opts ...grpc.CallOption) (*ThreadInfoReply, error) {
	out := new(ThreadInfoReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/AddThread", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) GetThread(ctx context.Context, in *GetThreadRequest, opts ...grpc.CallOption) (*ThreadInfoReply, error) {
	out := new(ThreadInfoReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/GetThread", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) PullThread(ctx context.Context, in *PullThreadRequest, opts ...grpc.CallOption) (*PullThreadReply, error) {
	out := new(PullThreadReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/PullThread", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) DeleteThread(ctx context.Context, in *DeleteThreadRequest, opts ...grpc.CallOption) (*DeleteThreadReply, error) {
	out := new(DeleteThreadReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/DeleteThread", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) AddReplicator(ctx context.Context, in *AddReplicatorRequest, opts ...grpc.CallOption) (*AddReplicatorReply, error) {
	out := new(AddReplicatorReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/AddReplicator", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) CreateRecord(ctx context.Context, in *CreateRecordRequest, opts ...grpc.CallOption) (*NewRecordReply, error) {
	out := new(NewRecordReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/CreateRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) AddRecord(ctx context.Context, in *AddRecordRequest, opts ...grpc.CallOption) (*AddRecordReply, error) {
	out := new(AddRecordReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/AddRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) GetRecord(ctx context.Context, in *GetRecordRequest, opts ...grpc.CallOption) (*GetRecordReply, error) {
	out := new(GetRecordReply)
	err := c.cc.Invoke(ctx, "/threads.net.pb.API/GetRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) Subscribe(ctx context.Context, in *SubscribeRequest, opts ...grpc.CallOption) (API_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, &API_ServiceDesc.Streams[1], "/threads.net.pb.API/Subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &aPISubscribeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type API_SubscribeClient interface {
	Recv() (*NewRecordReply, error)
	grpc.ClientStream
}

type aPISubscribeClient struct {
	grpc.ClientStream
}

func (x *aPISubscribeClient) Recv() (*NewRecordReply, error) {
	m := new(NewRecordReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// APIServer is the server API for API service.
// All implementations must embed UnimplementedAPIServer
// for forward compatibility
type APIServer interface {
	GetHostID(context.Context, *GetHostIDRequest) (*GetHostIDReply, error)
	GetToken(API_GetTokenServer) error
	CreateThread(context.Context, *CreateThreadRequest) (*ThreadInfoReply, error)
	AddThread(context.Context, *AddThreadRequest) (*ThreadInfoReply, error)
	GetThread(context.Context, *GetThreadRequest) (*ThreadInfoReply, error)
	PullThread(context.Context, *PullThreadRequest) (*PullThreadReply, error)
	DeleteThread(context.Context, *DeleteThreadRequest) (*DeleteThreadReply, error)
	AddReplicator(context.Context, *AddReplicatorRequest) (*AddReplicatorReply, error)
	CreateRecord(context.Context, *CreateRecordRequest) (*NewRecordReply, error)
	AddRecord(context.Context, *AddRecordRequest) (*AddRecordReply, error)
	GetRecord(context.Context, *GetRecordRequest) (*GetRecordReply, error)
	Subscribe(*SubscribeRequest, API_SubscribeServer) error
	mustEmbedUnimplementedAPIServer()
}

// UnimplementedAPIServer must be embedded to have forward compatible implementations.
type UnimplementedAPIServer struct {
}

func (UnimplementedAPIServer) GetHostID(context.Context, *GetHostIDRequest) (*GetHostIDReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHostID not implemented")
}
func (UnimplementedAPIServer) GetToken(API_GetTokenServer) error {
	return status.Errorf(codes.Unimplemented, "method GetToken not implemented")
}
func (UnimplementedAPIServer) CreateThread(context.Context, *CreateThreadRequest) (*ThreadInfoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateThread not implemented")
}
func (UnimplementedAPIServer) AddThread(context.Context, *AddThreadRequest) (*ThreadInfoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddThread not implemented")
}
func (UnimplementedAPIServer) GetThread(context.Context, *GetThreadRequest) (*ThreadInfoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetThread not implemented")
}
func (UnimplementedAPIServer) PullThread(context.Context, *PullThreadRequest) (*PullThreadReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PullThread not implemented")
}
func (UnimplementedAPIServer) DeleteThread(context.Context, *DeleteThreadRequest) (*DeleteThreadReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteThread not implemented")
}
func (UnimplementedAPIServer) AddReplicator(context.Context, *AddReplicatorRequest) (*AddReplicatorReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddReplicator not implemented")
}
func (UnimplementedAPIServer) CreateRecord(context.Context, *CreateRecordRequest) (*NewRecordReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateRecord not implemented")
}
func (UnimplementedAPIServer) AddRecord(context.Context, *AddRecordRequest) (*AddRecordReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddRecord not implemented")
}
func (UnimplementedAPIServer) GetRecord(context.Context, *GetRecordRequest) (*GetRecordReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRecord not implemented")
}
func (UnimplementedAPIServer) Subscribe(*SubscribeRequest, API_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "method Subscribe not implemented")
}
func (UnimplementedAPIServer) mustEmbedUnimplementedAPIServer() {}

// UnsafeAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to APIServer will
// result in compilation errors.
type UnsafeAPIServer interface {
	mustEmbedUnimplementedAPIServer()
}

func RegisterAPIServer(s grpc.ServiceRegistrar, srv APIServer) {
	s.RegisterService(&API_ServiceDesc, srv)
}


func _API_GetHostID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetHostIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).GetHostID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/GetHostID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).GetHostID(ctx, req.(*GetHostIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_GetToken_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(APIServer).GetToken(&aPIGetTokenServer{stream})
}

type API_GetTokenServer interface {
	Send(*GetTokenReply) error
	Recv() (*GetTokenRequest, error)
	grpc.ServerStream
}

type aPIGetTokenServer struct {
	grpc.ServerStream
}

func (x *aPIGetTokenServer) Send(m *GetTokenReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *aPIGetTokenServer) Recv() (*GetTokenRequest, error) {
	m := new(GetTokenRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _API_CreateThread_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateThreadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).CreateThread(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/CreateThread",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).CreateThread(ctx, req.(*CreateThreadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_AddThread_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddThreadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).AddThread(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/AddThread",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).AddThread(ctx, req.(*AddThreadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_GetThread_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetThreadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).GetThread(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/GetThread",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).GetThread(ctx, req.(*GetThreadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_PullThread_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PullThreadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).PullThread(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/PullThread",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).PullThread(ctx, req.(*PullThreadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_DeleteThread_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteThreadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).DeleteThread(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/DeleteThread",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).DeleteThread(ctx, req.(*DeleteThreadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_AddReplicator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddReplicatorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).AddReplicator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/AddReplicator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).AddReplicator(ctx, req.(*AddReplicatorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_CreateRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).CreateRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/CreateRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).CreateRecord(ctx, req.(*CreateRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_AddRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).AddRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/AddRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).AddRecord(ctx, req.(*AddRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_GetRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRecordRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).GetRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/threads.net.pb.API/GetRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).GetRecord(ctx, req.(*GetRecordRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SubscribeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(APIServer).Subscribe(m, &aPISubscribeServer{stream})
}

type API_SubscribeServer interface {
	Send(*NewRecordReply) error
	grpc.ServerStream
}

type aPISubscribeServer struct {
	grpc.ServerStream
}

func (x *aPISubscribeServer) Send(m *NewRecordReply) error {
	return x.ServerStream.SendMsg(m)
}

// API_ServiceDesc is the grpc.ServiceDesc for API service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var API_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "threads.net.pb.API",
	HandlerType: (*APIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetHostID",
			Handler:    _API_GetHostID_Handler,
		},
		{
			MethodName: "CreateThread",
			Handler:    _API_CreateThread_Handler,
		},
		{
			MethodName: "AddThread",
			Handler:    _API_AddThread_Handler,
		},
		{
			MethodName: "GetThread",
			Handler:    _API_GetThread_Handler,
		},
		{
			MethodName: "PullThread",
			Handler:    _API_PullThread_Handler,
		},
		{
			MethodName: "DeleteThread",
			Handler:    _API_DeleteThread_Handler,
		},
		{
			MethodName: "AddReplicator",
			Handler:    _API_AddReplicator_Handler,
		},
		{
			MethodName: "CreateRecord",
			Handler:    _API_CreateRecord_Handler,
		},
		{
			MethodName: "AddRecord",
			Handler:    _API_AddRecord_Handler,
		},
		{
			MethodName: "GetRecord",
			Handler:    _API_GetRecord_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetToken",
			Handler:       _API_GetToken_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "Subscribe",
			Handler:       _API_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "threadsnet.proto",
}
