package proto

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type DescriptorsManager interface {
	GetDescriptor(name protoreflect.FullName) protoreflect.MethodDescriptor
}

type ClientsManager interface {
	GetClient(serviceName string) Client
}

type Client interface {
	Invoke(fullName protoreflect.FullName, msg []byte, metadata metadata.MD) (*GRPCResponse, error)
	BuildRequest(desc protoreflect.MessageDescriptor, msg []byte) (*dynamicpb.Message, error)
}

type Connection interface {
	Invoke(ctx context.Context, fullName string, req, res interface{}) (header, trailer metadata.MD, err error)
	Stream(ctx context.Context, fullName string, streamDesc *grpc.StreamDesc) (grpc.ClientStream, error)
}
