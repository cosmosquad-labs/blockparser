// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: example.proto

package example

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

// UserServiceClient is the client API for UserService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserServiceClient interface {
	AddUser(ctx context.Context, in *AddUserRequest, opts ...grpc.CallOption) (*User, error)
	ListUsers(ctx context.Context, in *ListUsersRequest, opts ...grpc.CallOption) (UserService_ListUsersClient, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) AddUser(ctx context.Context, in *AddUserRequest, opts ...grpc.CallOption) (*User, error) {
	out := new(User)
	err := c.cc.Invoke(ctx, "/example.UserService/AddUser", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) ListUsers(ctx context.Context, in *ListUsersRequest, opts ...grpc.CallOption) (UserService_ListUsersClient, error) {
	stream, err := c.cc.NewStream(ctx, &UserService_ServiceDesc.Streams[0], "/example.UserService/ListUsers", opts...)
	if err != nil {
		return nil, err
	}
	x := &userServiceListUsersClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type UserService_ListUsersClient interface {
	Recv() (*User, error)
	grpc.ClientStream
}

type userServiceListUsersClient struct {
	grpc.ClientStream
}

func (x *userServiceListUsersClient) Recv() (*User, error) {
	m := new(User)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// UserServiceServer is the server API for UserService service.
// All implementations should embed UnimplementedUserServiceServer
// for forward compatibility
type UserServiceServer interface {
	AddUser(context.Context, *AddUserRequest) (*User, error)
	ListUsers(*ListUsersRequest, UserService_ListUsersServer) error
}

// UnimplementedUserServiceServer should be embedded to have forward compatible implementations.
type UnimplementedUserServiceServer struct {
}

func (UnimplementedUserServiceServer) AddUser(context.Context, *AddUserRequest) (*User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddUser not implemented")
}
func (UnimplementedUserServiceServer) ListUsers(*ListUsersRequest, UserService_ListUsersServer) error {
	return status.Errorf(codes.Unimplemented, "method ListUsers not implemented")
}

// UnsafeUserServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServiceServer will
// result in compilation errors.
type UnsafeUserServiceServer interface {
	mustEmbedUnimplementedUserServiceServer()
}

func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	s.RegisterService(&UserService_ServiceDesc, srv)
}

func _UserService_AddUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).AddUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.UserService/AddUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).AddUser(ctx, req.(*AddUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_ListUsers_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListUsersRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UserServiceServer).ListUsers(m, &userServiceListUsersServer{stream})
}

type UserService_ListUsersServer interface {
	Send(*User) error
	grpc.ServerStream
}

type userServiceListUsersServer struct {
	grpc.ServerStream
}

func (x *userServiceListUsersServer) Send(m *User) error {
	return x.ServerStream.SendMsg(m)
}

// UserService_ServiceDesc is the grpc.ServiceDesc for UserService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "example.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddUser",
			Handler:    _UserService_AddUser_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListUsers",
			Handler:       _UserService_ListUsers_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "example.proto",
}
