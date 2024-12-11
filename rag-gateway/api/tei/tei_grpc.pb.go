// Source : https://github.com/huggingface/text-embeddings-inference/blob/main/proto/tei.proto

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: rag-gateway/api/tei/tei.proto

package tei

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

const (
	Info_Info_FullMethodName = "/tei.v1.Info/Info"
)

// InfoClient is the client API for Info service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type InfoClient interface {
	Info(ctx context.Context, in *InfoRequest, opts ...grpc.CallOption) (*InfoResponse, error)
}

type infoClient struct {
	cc grpc.ClientConnInterface
}

func NewInfoClient(cc grpc.ClientConnInterface) InfoClient {
	return &infoClient{cc}
}

func (c *infoClient) Info(ctx context.Context, in *InfoRequest, opts ...grpc.CallOption) (*InfoResponse, error) {
	out := new(InfoResponse)
	err := c.cc.Invoke(ctx, Info_Info_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// InfoServer is the server API for Info service.
// All implementations should embed UnimplementedInfoServer
// for forward compatibility
type InfoServer interface {
	Info(context.Context, *InfoRequest) (*InfoResponse, error)
}

// UnimplementedInfoServer should be embedded to have forward compatible implementations.
type UnimplementedInfoServer struct {
}

func (UnimplementedInfoServer) Info(context.Context, *InfoRequest) (*InfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Info not implemented")
}

// UnsafeInfoServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to InfoServer will
// result in compilation errors.
type UnsafeInfoServer interface {
	mustEmbedUnimplementedInfoServer()
}

func RegisterInfoServer(s grpc.ServiceRegistrar, srv InfoServer) {
	s.RegisterService(&Info_ServiceDesc, srv)
}

func _Info_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(InfoServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Info_Info_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(InfoServer).Info(ctx, req.(*InfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Info_ServiceDesc is the grpc.ServiceDesc for Info service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Info_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tei.v1.Info",
	HandlerType: (*InfoServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Info",
			Handler:    _Info_Info_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "rag-gateway/api/tei/tei.proto",
}

const (
	Embed_Embed_FullMethodName             = "/tei.v1.Embed/Embed"
	Embed_EmbedStream_FullMethodName       = "/tei.v1.Embed/EmbedStream"
	Embed_EmbedSparse_FullMethodName       = "/tei.v1.Embed/EmbedSparse"
	Embed_EmbedSparseStream_FullMethodName = "/tei.v1.Embed/EmbedSparseStream"
	Embed_EmbedAll_FullMethodName          = "/tei.v1.Embed/EmbedAll"
	Embed_EmbedAllStream_FullMethodName    = "/tei.v1.Embed/EmbedAllStream"
)

// EmbedClient is the client API for Embed service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type EmbedClient interface {
	Embed(ctx context.Context, in *EmbedRequest, opts ...grpc.CallOption) (*EmbedResponse, error)
	EmbedStream(ctx context.Context, opts ...grpc.CallOption) (Embed_EmbedStreamClient, error)
	EmbedSparse(ctx context.Context, in *EmbedSparseRequest, opts ...grpc.CallOption) (*EmbedSparseResponse, error)
	EmbedSparseStream(ctx context.Context, opts ...grpc.CallOption) (Embed_EmbedSparseStreamClient, error)
	EmbedAll(ctx context.Context, in *EmbedAllRequest, opts ...grpc.CallOption) (*EmbedAllResponse, error)
	EmbedAllStream(ctx context.Context, opts ...grpc.CallOption) (Embed_EmbedAllStreamClient, error)
}

type embedClient struct {
	cc grpc.ClientConnInterface
}

func NewEmbedClient(cc grpc.ClientConnInterface) EmbedClient {
	return &embedClient{cc}
}

func (c *embedClient) Embed(ctx context.Context, in *EmbedRequest, opts ...grpc.CallOption) (*EmbedResponse, error) {
	out := new(EmbedResponse)
	err := c.cc.Invoke(ctx, Embed_Embed_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *embedClient) EmbedStream(ctx context.Context, opts ...grpc.CallOption) (Embed_EmbedStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Embed_ServiceDesc.Streams[0], Embed_EmbedStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &embedEmbedStreamClient{stream}
	return x, nil
}

type Embed_EmbedStreamClient interface {
	Send(*EmbedRequest) error
	Recv() (*EmbedResponse, error)
	grpc.ClientStream
}

type embedEmbedStreamClient struct {
	grpc.ClientStream
}

func (x *embedEmbedStreamClient) Send(m *EmbedRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *embedEmbedStreamClient) Recv() (*EmbedResponse, error) {
	m := new(EmbedResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *embedClient) EmbedSparse(ctx context.Context, in *EmbedSparseRequest, opts ...grpc.CallOption) (*EmbedSparseResponse, error) {
	out := new(EmbedSparseResponse)
	err := c.cc.Invoke(ctx, Embed_EmbedSparse_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *embedClient) EmbedSparseStream(ctx context.Context, opts ...grpc.CallOption) (Embed_EmbedSparseStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Embed_ServiceDesc.Streams[1], Embed_EmbedSparseStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &embedEmbedSparseStreamClient{stream}
	return x, nil
}

type Embed_EmbedSparseStreamClient interface {
	Send(*EmbedSparseRequest) error
	Recv() (*EmbedSparseResponse, error)
	grpc.ClientStream
}

type embedEmbedSparseStreamClient struct {
	grpc.ClientStream
}

func (x *embedEmbedSparseStreamClient) Send(m *EmbedSparseRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *embedEmbedSparseStreamClient) Recv() (*EmbedSparseResponse, error) {
	m := new(EmbedSparseResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *embedClient) EmbedAll(ctx context.Context, in *EmbedAllRequest, opts ...grpc.CallOption) (*EmbedAllResponse, error) {
	out := new(EmbedAllResponse)
	err := c.cc.Invoke(ctx, Embed_EmbedAll_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *embedClient) EmbedAllStream(ctx context.Context, opts ...grpc.CallOption) (Embed_EmbedAllStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Embed_ServiceDesc.Streams[2], Embed_EmbedAllStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &embedEmbedAllStreamClient{stream}
	return x, nil
}

type Embed_EmbedAllStreamClient interface {
	Send(*EmbedAllRequest) error
	Recv() (*EmbedAllResponse, error)
	grpc.ClientStream
}

type embedEmbedAllStreamClient struct {
	grpc.ClientStream
}

func (x *embedEmbedAllStreamClient) Send(m *EmbedAllRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *embedEmbedAllStreamClient) Recv() (*EmbedAllResponse, error) {
	m := new(EmbedAllResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// EmbedServer is the server API for Embed service.
// All implementations should embed UnimplementedEmbedServer
// for forward compatibility
type EmbedServer interface {
	Embed(context.Context, *EmbedRequest) (*EmbedResponse, error)
	EmbedStream(Embed_EmbedStreamServer) error
	EmbedSparse(context.Context, *EmbedSparseRequest) (*EmbedSparseResponse, error)
	EmbedSparseStream(Embed_EmbedSparseStreamServer) error
	EmbedAll(context.Context, *EmbedAllRequest) (*EmbedAllResponse, error)
	EmbedAllStream(Embed_EmbedAllStreamServer) error
}

// UnimplementedEmbedServer should be embedded to have forward compatible implementations.
type UnimplementedEmbedServer struct {
}

func (UnimplementedEmbedServer) Embed(context.Context, *EmbedRequest) (*EmbedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Embed not implemented")
}
func (UnimplementedEmbedServer) EmbedStream(Embed_EmbedStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method EmbedStream not implemented")
}
func (UnimplementedEmbedServer) EmbedSparse(context.Context, *EmbedSparseRequest) (*EmbedSparseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmbedSparse not implemented")
}
func (UnimplementedEmbedServer) EmbedSparseStream(Embed_EmbedSparseStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method EmbedSparseStream not implemented")
}
func (UnimplementedEmbedServer) EmbedAll(context.Context, *EmbedAllRequest) (*EmbedAllResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmbedAll not implemented")
}
func (UnimplementedEmbedServer) EmbedAllStream(Embed_EmbedAllStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method EmbedAllStream not implemented")
}

// UnsafeEmbedServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to EmbedServer will
// result in compilation errors.
type UnsafeEmbedServer interface {
	mustEmbedUnimplementedEmbedServer()
}

func RegisterEmbedServer(s grpc.ServiceRegistrar, srv EmbedServer) {
	s.RegisterService(&Embed_ServiceDesc, srv)
}

func _Embed_Embed_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmbedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmbedServer).Embed(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Embed_Embed_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmbedServer).Embed(ctx, req.(*EmbedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Embed_EmbedStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EmbedServer).EmbedStream(&embedEmbedStreamServer{stream})
}

type Embed_EmbedStreamServer interface {
	Send(*EmbedResponse) error
	Recv() (*EmbedRequest, error)
	grpc.ServerStream
}

type embedEmbedStreamServer struct {
	grpc.ServerStream
}

func (x *embedEmbedStreamServer) Send(m *EmbedResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *embedEmbedStreamServer) Recv() (*EmbedRequest, error) {
	m := new(EmbedRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Embed_EmbedSparse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmbedSparseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmbedServer).EmbedSparse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Embed_EmbedSparse_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmbedServer).EmbedSparse(ctx, req.(*EmbedSparseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Embed_EmbedSparseStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EmbedServer).EmbedSparseStream(&embedEmbedSparseStreamServer{stream})
}

type Embed_EmbedSparseStreamServer interface {
	Send(*EmbedSparseResponse) error
	Recv() (*EmbedSparseRequest, error)
	grpc.ServerStream
}

type embedEmbedSparseStreamServer struct {
	grpc.ServerStream
}

func (x *embedEmbedSparseStreamServer) Send(m *EmbedSparseResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *embedEmbedSparseStreamServer) Recv() (*EmbedSparseRequest, error) {
	m := new(EmbedSparseRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Embed_EmbedAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmbedAllRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(EmbedServer).EmbedAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Embed_EmbedAll_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(EmbedServer).EmbedAll(ctx, req.(*EmbedAllRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Embed_EmbedAllStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(EmbedServer).EmbedAllStream(&embedEmbedAllStreamServer{stream})
}

type Embed_EmbedAllStreamServer interface {
	Send(*EmbedAllResponse) error
	Recv() (*EmbedAllRequest, error)
	grpc.ServerStream
}

type embedEmbedAllStreamServer struct {
	grpc.ServerStream
}

func (x *embedEmbedAllStreamServer) Send(m *EmbedAllResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *embedEmbedAllStreamServer) Recv() (*EmbedAllRequest, error) {
	m := new(EmbedAllRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Embed_ServiceDesc is the grpc.ServiceDesc for Embed service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Embed_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tei.v1.Embed",
	HandlerType: (*EmbedServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Embed",
			Handler:    _Embed_Embed_Handler,
		},
		{
			MethodName: "EmbedSparse",
			Handler:    _Embed_EmbedSparse_Handler,
		},
		{
			MethodName: "EmbedAll",
			Handler:    _Embed_EmbedAll_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EmbedStream",
			Handler:       _Embed_EmbedStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "EmbedSparseStream",
			Handler:       _Embed_EmbedSparseStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "EmbedAllStream",
			Handler:       _Embed_EmbedAllStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "rag-gateway/api/tei/tei.proto",
}

const (
	Predict_Predict_FullMethodName           = "/tei.v1.Predict/Predict"
	Predict_PredictPair_FullMethodName       = "/tei.v1.Predict/PredictPair"
	Predict_PredictStream_FullMethodName     = "/tei.v1.Predict/PredictStream"
	Predict_PredictPairStream_FullMethodName = "/tei.v1.Predict/PredictPairStream"
)

// PredictClient is the client API for Predict service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PredictClient interface {
	Predict(ctx context.Context, in *PredictRequest, opts ...grpc.CallOption) (*PredictResponse, error)
	PredictPair(ctx context.Context, in *PredictPairRequest, opts ...grpc.CallOption) (*PredictResponse, error)
	PredictStream(ctx context.Context, opts ...grpc.CallOption) (Predict_PredictStreamClient, error)
	PredictPairStream(ctx context.Context, opts ...grpc.CallOption) (Predict_PredictPairStreamClient, error)
}

type predictClient struct {
	cc grpc.ClientConnInterface
}

func NewPredictClient(cc grpc.ClientConnInterface) PredictClient {
	return &predictClient{cc}
}

func (c *predictClient) Predict(ctx context.Context, in *PredictRequest, opts ...grpc.CallOption) (*PredictResponse, error) {
	out := new(PredictResponse)
	err := c.cc.Invoke(ctx, Predict_Predict_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *predictClient) PredictPair(ctx context.Context, in *PredictPairRequest, opts ...grpc.CallOption) (*PredictResponse, error) {
	out := new(PredictResponse)
	err := c.cc.Invoke(ctx, Predict_PredictPair_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *predictClient) PredictStream(ctx context.Context, opts ...grpc.CallOption) (Predict_PredictStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Predict_ServiceDesc.Streams[0], Predict_PredictStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &predictPredictStreamClient{stream}
	return x, nil
}

type Predict_PredictStreamClient interface {
	Send(*PredictRequest) error
	Recv() (*PredictResponse, error)
	grpc.ClientStream
}

type predictPredictStreamClient struct {
	grpc.ClientStream
}

func (x *predictPredictStreamClient) Send(m *PredictRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *predictPredictStreamClient) Recv() (*PredictResponse, error) {
	m := new(PredictResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *predictClient) PredictPairStream(ctx context.Context, opts ...grpc.CallOption) (Predict_PredictPairStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Predict_ServiceDesc.Streams[1], Predict_PredictPairStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &predictPredictPairStreamClient{stream}
	return x, nil
}

type Predict_PredictPairStreamClient interface {
	Send(*PredictPairRequest) error
	Recv() (*PredictResponse, error)
	grpc.ClientStream
}

type predictPredictPairStreamClient struct {
	grpc.ClientStream
}

func (x *predictPredictPairStreamClient) Send(m *PredictPairRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *predictPredictPairStreamClient) Recv() (*PredictResponse, error) {
	m := new(PredictResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// PredictServer is the server API for Predict service.
// All implementations should embed UnimplementedPredictServer
// for forward compatibility
type PredictServer interface {
	Predict(context.Context, *PredictRequest) (*PredictResponse, error)
	PredictPair(context.Context, *PredictPairRequest) (*PredictResponse, error)
	PredictStream(Predict_PredictStreamServer) error
	PredictPairStream(Predict_PredictPairStreamServer) error
}

// UnimplementedPredictServer should be embedded to have forward compatible implementations.
type UnimplementedPredictServer struct {
}

func (UnimplementedPredictServer) Predict(context.Context, *PredictRequest) (*PredictResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Predict not implemented")
}
func (UnimplementedPredictServer) PredictPair(context.Context, *PredictPairRequest) (*PredictResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PredictPair not implemented")
}
func (UnimplementedPredictServer) PredictStream(Predict_PredictStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method PredictStream not implemented")
}
func (UnimplementedPredictServer) PredictPairStream(Predict_PredictPairStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method PredictPairStream not implemented")
}

// UnsafePredictServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PredictServer will
// result in compilation errors.
type UnsafePredictServer interface {
	mustEmbedUnimplementedPredictServer()
}

func RegisterPredictServer(s grpc.ServiceRegistrar, srv PredictServer) {
	s.RegisterService(&Predict_ServiceDesc, srv)
}

func _Predict_Predict_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PredictRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictServer).Predict(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Predict_Predict_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictServer).Predict(ctx, req.(*PredictRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Predict_PredictPair_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PredictPairRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PredictServer).PredictPair(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Predict_PredictPair_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PredictServer).PredictPair(ctx, req.(*PredictPairRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Predict_PredictStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PredictServer).PredictStream(&predictPredictStreamServer{stream})
}

type Predict_PredictStreamServer interface {
	Send(*PredictResponse) error
	Recv() (*PredictRequest, error)
	grpc.ServerStream
}

type predictPredictStreamServer struct {
	grpc.ServerStream
}

func (x *predictPredictStreamServer) Send(m *PredictResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *predictPredictStreamServer) Recv() (*PredictRequest, error) {
	m := new(PredictRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Predict_PredictPairStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(PredictServer).PredictPairStream(&predictPredictPairStreamServer{stream})
}

type Predict_PredictPairStreamServer interface {
	Send(*PredictResponse) error
	Recv() (*PredictPairRequest, error)
	grpc.ServerStream
}

type predictPredictPairStreamServer struct {
	grpc.ServerStream
}

func (x *predictPredictPairStreamServer) Send(m *PredictResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *predictPredictPairStreamServer) Recv() (*PredictPairRequest, error) {
	m := new(PredictPairRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Predict_ServiceDesc is the grpc.ServiceDesc for Predict service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Predict_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tei.v1.Predict",
	HandlerType: (*PredictServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Predict",
			Handler:    _Predict_Predict_Handler,
		},
		{
			MethodName: "PredictPair",
			Handler:    _Predict_PredictPair_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PredictStream",
			Handler:       _Predict_PredictStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "PredictPairStream",
			Handler:       _Predict_PredictPairStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "rag-gateway/api/tei/tei.proto",
}

const (
	Rerank_Rerank_FullMethodName       = "/tei.v1.Rerank/Rerank"
	Rerank_RerankStream_FullMethodName = "/tei.v1.Rerank/RerankStream"
)

// RerankClient is the client API for Rerank service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RerankClient interface {
	Rerank(ctx context.Context, in *RerankRequest, opts ...grpc.CallOption) (*RerankResponse, error)
	RerankStream(ctx context.Context, opts ...grpc.CallOption) (Rerank_RerankStreamClient, error)
}

type rerankClient struct {
	cc grpc.ClientConnInterface
}

func NewRerankClient(cc grpc.ClientConnInterface) RerankClient {
	return &rerankClient{cc}
}

func (c *rerankClient) Rerank(ctx context.Context, in *RerankRequest, opts ...grpc.CallOption) (*RerankResponse, error) {
	out := new(RerankResponse)
	err := c.cc.Invoke(ctx, Rerank_Rerank_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rerankClient) RerankStream(ctx context.Context, opts ...grpc.CallOption) (Rerank_RerankStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Rerank_ServiceDesc.Streams[0], Rerank_RerankStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &rerankRerankStreamClient{stream}
	return x, nil
}

type Rerank_RerankStreamClient interface {
	Send(*RerankStreamRequest) error
	CloseAndRecv() (*RerankResponse, error)
	grpc.ClientStream
}

type rerankRerankStreamClient struct {
	grpc.ClientStream
}

func (x *rerankRerankStreamClient) Send(m *RerankStreamRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *rerankRerankStreamClient) CloseAndRecv() (*RerankResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(RerankResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RerankServer is the server API for Rerank service.
// All implementations should embed UnimplementedRerankServer
// for forward compatibility
type RerankServer interface {
	Rerank(context.Context, *RerankRequest) (*RerankResponse, error)
	RerankStream(Rerank_RerankStreamServer) error
}

// UnimplementedRerankServer should be embedded to have forward compatible implementations.
type UnimplementedRerankServer struct {
}

func (UnimplementedRerankServer) Rerank(context.Context, *RerankRequest) (*RerankResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Rerank not implemented")
}
func (UnimplementedRerankServer) RerankStream(Rerank_RerankStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method RerankStream not implemented")
}

// UnsafeRerankServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RerankServer will
// result in compilation errors.
type UnsafeRerankServer interface {
	mustEmbedUnimplementedRerankServer()
}

func RegisterRerankServer(s grpc.ServiceRegistrar, srv RerankServer) {
	s.RegisterService(&Rerank_ServiceDesc, srv)
}

func _Rerank_Rerank_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RerankRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RerankServer).Rerank(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Rerank_Rerank_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RerankServer).Rerank(ctx, req.(*RerankRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Rerank_RerankStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RerankServer).RerankStream(&rerankRerankStreamServer{stream})
}

type Rerank_RerankStreamServer interface {
	SendAndClose(*RerankResponse) error
	Recv() (*RerankStreamRequest, error)
	grpc.ServerStream
}

type rerankRerankStreamServer struct {
	grpc.ServerStream
}

func (x *rerankRerankStreamServer) SendAndClose(m *RerankResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *rerankRerankStreamServer) Recv() (*RerankStreamRequest, error) {
	m := new(RerankStreamRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Rerank_ServiceDesc is the grpc.ServiceDesc for Rerank service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Rerank_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tei.v1.Rerank",
	HandlerType: (*RerankServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Rerank",
			Handler:    _Rerank_Rerank_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RerankStream",
			Handler:       _Rerank_RerankStream_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "rag-gateway/api/tei/tei.proto",
}

const (
	Tokenize_Tokenize_FullMethodName       = "/tei.v1.Tokenize/Tokenize"
	Tokenize_TokenizeStream_FullMethodName = "/tei.v1.Tokenize/TokenizeStream"
	Tokenize_Decode_FullMethodName         = "/tei.v1.Tokenize/Decode"
	Tokenize_DecodeStream_FullMethodName   = "/tei.v1.Tokenize/DecodeStream"
)

// TokenizeClient is the client API for Tokenize service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TokenizeClient interface {
	Tokenize(ctx context.Context, in *EncodeRequest, opts ...grpc.CallOption) (*EncodeResponse, error)
	TokenizeStream(ctx context.Context, opts ...grpc.CallOption) (Tokenize_TokenizeStreamClient, error)
	Decode(ctx context.Context, in *DecodeRequest, opts ...grpc.CallOption) (*DecodeResponse, error)
	DecodeStream(ctx context.Context, opts ...grpc.CallOption) (Tokenize_DecodeStreamClient, error)
}

type tokenizeClient struct {
	cc grpc.ClientConnInterface
}

func NewTokenizeClient(cc grpc.ClientConnInterface) TokenizeClient {
	return &tokenizeClient{cc}
}

func (c *tokenizeClient) Tokenize(ctx context.Context, in *EncodeRequest, opts ...grpc.CallOption) (*EncodeResponse, error) {
	out := new(EncodeResponse)
	err := c.cc.Invoke(ctx, Tokenize_Tokenize_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tokenizeClient) TokenizeStream(ctx context.Context, opts ...grpc.CallOption) (Tokenize_TokenizeStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Tokenize_ServiceDesc.Streams[0], Tokenize_TokenizeStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &tokenizeTokenizeStreamClient{stream}
	return x, nil
}

type Tokenize_TokenizeStreamClient interface {
	Send(*EncodeRequest) error
	Recv() (*EncodeResponse, error)
	grpc.ClientStream
}

type tokenizeTokenizeStreamClient struct {
	grpc.ClientStream
}

func (x *tokenizeTokenizeStreamClient) Send(m *EncodeRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *tokenizeTokenizeStreamClient) Recv() (*EncodeResponse, error) {
	m := new(EncodeResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *tokenizeClient) Decode(ctx context.Context, in *DecodeRequest, opts ...grpc.CallOption) (*DecodeResponse, error) {
	out := new(DecodeResponse)
	err := c.cc.Invoke(ctx, Tokenize_Decode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tokenizeClient) DecodeStream(ctx context.Context, opts ...grpc.CallOption) (Tokenize_DecodeStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Tokenize_ServiceDesc.Streams[1], Tokenize_DecodeStream_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &tokenizeDecodeStreamClient{stream}
	return x, nil
}

type Tokenize_DecodeStreamClient interface {
	Send(*DecodeRequest) error
	Recv() (*DecodeResponse, error)
	grpc.ClientStream
}

type tokenizeDecodeStreamClient struct {
	grpc.ClientStream
}

func (x *tokenizeDecodeStreamClient) Send(m *DecodeRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *tokenizeDecodeStreamClient) Recv() (*DecodeResponse, error) {
	m := new(DecodeResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TokenizeServer is the server API for Tokenize service.
// All implementations should embed UnimplementedTokenizeServer
// for forward compatibility
type TokenizeServer interface {
	Tokenize(context.Context, *EncodeRequest) (*EncodeResponse, error)
	TokenizeStream(Tokenize_TokenizeStreamServer) error
	Decode(context.Context, *DecodeRequest) (*DecodeResponse, error)
	DecodeStream(Tokenize_DecodeStreamServer) error
}

// UnimplementedTokenizeServer should be embedded to have forward compatible implementations.
type UnimplementedTokenizeServer struct {
}

func (UnimplementedTokenizeServer) Tokenize(context.Context, *EncodeRequest) (*EncodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Tokenize not implemented")
}
func (UnimplementedTokenizeServer) TokenizeStream(Tokenize_TokenizeStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method TokenizeStream not implemented")
}
func (UnimplementedTokenizeServer) Decode(context.Context, *DecodeRequest) (*DecodeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Decode not implemented")
}
func (UnimplementedTokenizeServer) DecodeStream(Tokenize_DecodeStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method DecodeStream not implemented")
}

// UnsafeTokenizeServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TokenizeServer will
// result in compilation errors.
type UnsafeTokenizeServer interface {
	mustEmbedUnimplementedTokenizeServer()
}

func RegisterTokenizeServer(s grpc.ServiceRegistrar, srv TokenizeServer) {
	s.RegisterService(&Tokenize_ServiceDesc, srv)
}

func _Tokenize_Tokenize_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EncodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenizeServer).Tokenize(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tokenize_Tokenize_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenizeServer).Tokenize(ctx, req.(*EncodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tokenize_TokenizeStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TokenizeServer).TokenizeStream(&tokenizeTokenizeStreamServer{stream})
}

type Tokenize_TokenizeStreamServer interface {
	Send(*EncodeResponse) error
	Recv() (*EncodeRequest, error)
	grpc.ServerStream
}

type tokenizeTokenizeStreamServer struct {
	grpc.ServerStream
}

func (x *tokenizeTokenizeStreamServer) Send(m *EncodeResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *tokenizeTokenizeStreamServer) Recv() (*EncodeRequest, error) {
	m := new(EncodeRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Tokenize_Decode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DecodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenizeServer).Decode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Tokenize_Decode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenizeServer).Decode(ctx, req.(*DecodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tokenize_DecodeStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TokenizeServer).DecodeStream(&tokenizeDecodeStreamServer{stream})
}

type Tokenize_DecodeStreamServer interface {
	Send(*DecodeResponse) error
	Recv() (*DecodeRequest, error)
	grpc.ServerStream
}

type tokenizeDecodeStreamServer struct {
	grpc.ServerStream
}

func (x *tokenizeDecodeStreamServer) Send(m *DecodeResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *tokenizeDecodeStreamServer) Recv() (*DecodeRequest, error) {
	m := new(DecodeRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Tokenize_ServiceDesc is the grpc.ServiceDesc for Tokenize service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tokenize_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tei.v1.Tokenize",
	HandlerType: (*TokenizeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Tokenize",
			Handler:    _Tokenize_Tokenize_Handler,
		},
		{
			MethodName: "Decode",
			Handler:    _Tokenize_Decode_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "TokenizeStream",
			Handler:       _Tokenize_TokenizeStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "DecodeStream",
			Handler:       _Tokenize_DecodeStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "rag-gateway/api/tei/tei.proto",
}
