// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.22.1
// source: file_rpc.proto

package file_rpc

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
	Files_FileInfo_FullMethodName = "/user_rpc.files/FileInfo"
)

// FilesClient is the client API for Files service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FilesClient interface {
	FileInfo(ctx context.Context, in *FileInfoRequest, opts ...grpc.CallOption) (*FileInfoResponse, error)
}

type filesClient struct {
	cc grpc.ClientConnInterface
}

func NewFilesClient(cc grpc.ClientConnInterface) FilesClient {
	return &filesClient{cc}
}

func (c *filesClient) FileInfo(ctx context.Context, in *FileInfoRequest, opts ...grpc.CallOption) (*FileInfoResponse, error) {
	out := new(FileInfoResponse)
	err := c.cc.Invoke(ctx, Files_FileInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FilesServer is the server API for Files service.
// All implementations must embed UnimplementedFilesServer
// for forward compatibility
type FilesServer interface {
	FileInfo(context.Context, *FileInfoRequest) (*FileInfoResponse, error)
	mustEmbedUnimplementedFilesServer()
}

// UnimplementedFilesServer must be embedded to have forward compatible implementations.
type UnimplementedFilesServer struct {
}

func (UnimplementedFilesServer) FileInfo(context.Context, *FileInfoRequest) (*FileInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FileInfo not implemented")
}
func (UnimplementedFilesServer) mustEmbedUnimplementedFilesServer() {}

// UnsafeFilesServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FilesServer will
// result in compilation errors.
type UnsafeFilesServer interface {
	mustEmbedUnimplementedFilesServer()
}

func RegisterFilesServer(s grpc.ServiceRegistrar, srv FilesServer) {
	s.RegisterService(&Files_ServiceDesc, srv)
}

func _Files_FileInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FilesServer).FileInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Files_FileInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FilesServer).FileInfo(ctx, req.(*FileInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Files_ServiceDesc is the grpc.ServiceDesc for Files service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Files_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user_rpc.files",
	HandlerType: (*FilesServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FileInfo",
			Handler:    _Files_FileInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "file_rpc.proto",
}
