package proto

import (
	"context"
	"encoding/json"
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
	conn    Connection
	manager DescriptorsManager
	dec     *protojson.UnmarshalOptions
}

func newClient(conn Connection, manager DescriptorsManager) Client {
	return &client{conn: conn, dec: &protojson.UnmarshalOptions{
		Resolver: nil,
	}, manager: manager}
}

func (c client) Invoke(fullName protoreflect.FullName, msg []byte, md metadata.MD) (*GRPCResponse, error) {
	if !fullName.IsValid() {
		return nil, fmt.Errorf("invalid method name %s", string(fullName))
	}

	descriptor := c.manager.GetDescriptor(fullName)

	switch {
	case descriptor.IsStreamingClient() && descriptor.IsStreamingServer():
		return nil, errors.New("bidirectional streaming is not supported yet")
	case descriptor.IsStreamingClient():
		ctx, err := c.createContext(md)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create context")
		}
		streamDesc := &grpc.StreamDesc{
			StreamName:    string(descriptor.Name()),
			ServerStreams: descriptor.IsStreamingServer(),
			ClientStreams: descriptor.IsStreamingClient(),
		}
		stream, err := c.conn.Stream(ctx, string(descriptor.FullName()), streamDesc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create stream")
		}

		requests := make([]interface{}, 0)
		err = json.Unmarshal(msg, &requests)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal stream requests")
		}
		for _, anyRequest := range requests {
			jsonRequest, err := json.Marshal(anyRequest)
			if err != nil {
				return nil, errors.Wrap(err, "failed to marshal stream request")
			}
			req, err := c.BuildRequest(descriptor.Input(), jsonRequest)
			if err != nil {
				return nil, errors.Wrap(err, "failed to build request")
			}
			if err := stream.SendMsg(req); err != nil {
				return nil, errors.Wrapf(err, "failed to send a RPC to the server stream '%s'", streamDesc.StreamName)
			}
		}

		err = stream.CloseSend()
		if err != nil {
			return nil, errors.Wrap(err, "failed to close the stream")
		}

		res := dynamicpb.NewMessage(descriptor.Output())

		return NewGRPCUnaryResponse(res, err)
	case descriptor.IsStreamingServer():
		ctx, err := c.createContext(md)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create context")
		}
		streamDesc := &grpc.StreamDesc{
			StreamName:    string(descriptor.Name()),
			ServerStreams: descriptor.IsStreamingServer(),
			ClientStreams: descriptor.IsStreamingClient(),
		}
		stream, err := c.conn.Stream(ctx, string(descriptor.FullName()), streamDesc)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create stream")
		}

		req, err := c.BuildRequest(descriptor.Input(), msg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build request")
		}
		if err := stream.SendMsg(req); err != nil {
			return nil, errors.Wrapf(err, "failed to send a RPC to the server stream '%s'", streamDesc.StreamName)
		}

		return NewGRPCStreamResponse(stream, descriptor.Output())
	default:
		req, err := c.BuildRequest(descriptor.Input(), msg)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build request")
		}
		ctx, err := c.createContext(md)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create context")
		}
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
