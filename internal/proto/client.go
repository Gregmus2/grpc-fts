package proto

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
	"io"
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

func (c client) Invoke(fullName protoreflect.FullName, msg []byte, md metadata.MD) (map[string]interface{}, *status.Status, error) {
	if !fullName.IsValid() {
		return nil, nil, fmt.Errorf("invalid method name %s", string(fullName))
	}

	descriptor := c.manager.GetDescriptor(fullName)
	req, err := c.buildRequest(descriptor.Input(), msg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to build request")
	}
	res := dynamicpb.NewMessage(descriptor.Output())
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	// todo: handle header and trailer
	_, _, err = c.conn.Invoke(ctx, string(fullName), req, res)
	stat, ok := status.FromError(errors.Cause(err))
	if !ok && err != nil {
		return nil, nil, errors.Wrap(err, "failed to send a request")
	}

	b, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(proto.Message(res))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to marshal response")
	}

	mapResponse := make(map[string]interface{})
	err = json.Unmarshal(b, &mapResponse)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal response")
	}

	return mapResponse, stat, nil
}

func (c client) buildRequest(desc protoreflect.MessageDescriptor, msg []byte) (*dynamicpb.Message, error) {
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
