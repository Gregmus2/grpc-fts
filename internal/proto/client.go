package proto

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"io"
	"time"
)

type client struct {
	conn    connection
	manager DescriptorsManager
	dec     *protojson.UnmarshalOptions
}

func newClient(conn connection, manager DescriptorsManager) Client {
	return &client{conn: conn, dec: &protojson.UnmarshalOptions{
		Resolver: nil,
	}, manager: manager}
}

func (c client) Invoke(fullName protoreflect.FullName, msg []byte, md metadata.MD) (*GRPCResponse, error) {
	if !fullName.IsValid() {
		return nil, fmt.Errorf("invalid method name %s", string(fullName))
	}

	descriptor := c.manager.GetDescriptor(fullName)
	req, err := c.BuildRequest(descriptor.Input(), msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}
	ctx, err := c.createContext(md)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create context")
	}

	switch {
	case descriptor.IsStreamingClient():
		return nil, errors.New("streaming client is not supported yet")
	case descriptor.IsStreamingServer():
		streamDesc := &grpc.StreamDesc{
			StreamName:    string(descriptor.Name()),
			ServerStreams: descriptor.IsStreamingServer(),
			ClientStreams: descriptor.IsStreamingClient(),
		}
		stream, err := c.conn.Stream(ctx, string(descriptor.FullName()), streamDesc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create stream")
		}

		if err := stream.SendMsg(req); err != nil {
			return nil, errors.Wrapf(err, "failed to send a RPC to the server stream '%s'", streamDesc.StreamName)
		}

		return NewGRPCStreamResponse(stream, descriptor.Output())
	default:
		res := dynamicpb.NewMessage(descriptor.Output())
		// todo: handle header and trailer
		_, _, err = c.conn.Invoke(ctx, string(fullName), req, res)

		return NewGRPCUnaryResponse(res, err)
	}
}

func (c client) createContext(md metadata.MD) (context.Context, error) {
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	values := md.Get("timeout")
	if len(values) == 0 {
		return ctx, nil
	}

	timeout, err := time.ParseDuration(values[len(values)-1])
	if err != nil {
		return nil, errors.Wrapf(err, "malformed grpc-timeout header")
	}

	// todo
	ctx, _ = context.WithTimeout(ctx, timeout)

	return ctx, nil
}

func (c client) BuildRequest(desc protoreflect.MessageDescriptor, msg []byte) (*dynamicpb.Message, error) {
	req := dynamicpb.NewMessage(desc)
	err := c.dec.Unmarshal(msg, req)
	if errors.Is(err, io.EOF) {
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	return req, nil
}
