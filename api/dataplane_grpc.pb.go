// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package api

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

// DataPlaneClient is the client API for DataPlane service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataPlaneClient interface {
	ApplySRPolicy(ctx context.Context, in *ApplySRPolicyRequest, opts ...grpc.CallOption) (*ApplySRPolicyResponse, error)
}

type dataPlaneClient struct {
	cc grpc.ClientConnInterface
}

func NewDataPlaneClient(cc grpc.ClientConnInterface) DataPlaneClient {
	return &dataPlaneClient{cc}
}

func (c *dataPlaneClient) ApplySRPolicy(ctx context.Context, in *ApplySRPolicyRequest, opts ...grpc.CallOption) (*ApplySRPolicyResponse, error) {
	out := new(ApplySRPolicyResponse)
	err := c.cc.Invoke(ctx, "/acar.DataPlane/ApplySRPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataPlaneServer is the server API for DataPlane service.
// All implementations must embed UnimplementedDataPlaneServer
// for forward compatibility
type DataPlaneServer interface {
	ApplySRPolicy(context.Context, *ApplySRPolicyRequest) (*ApplySRPolicyResponse, error)
	mustEmbedUnimplementedDataPlaneServer()
}

// UnimplementedDataPlaneServer must be embedded to have forward compatible implementations.
type UnimplementedDataPlaneServer struct {
}

func (UnimplementedDataPlaneServer) ApplySRPolicy(context.Context, *ApplySRPolicyRequest) (*ApplySRPolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplySRPolicy not implemented")
}
func (UnimplementedDataPlaneServer) mustEmbedUnimplementedDataPlaneServer() {}

// UnsafeDataPlaneServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataPlaneServer will
// result in compilation errors.
type UnsafeDataPlaneServer interface {
	mustEmbedUnimplementedDataPlaneServer()
}

func RegisterDataPlaneServer(s grpc.ServiceRegistrar, srv DataPlaneServer) {
	s.RegisterService(&DataPlane_ServiceDesc, srv)
}

func _DataPlane_ApplySRPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApplySRPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataPlaneServer).ApplySRPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/acar.DataPlane/ApplySRPolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataPlaneServer).ApplySRPolicy(ctx, req.(*ApplySRPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DataPlane_ServiceDesc is the grpc.ServiceDesc for DataPlane service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataPlane_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "acar.DataPlane",
	HandlerType: (*DataPlaneServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ApplySRPolicy",
			Handler:    _DataPlane_ApplySRPolicy_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dataplane.proto",
}
