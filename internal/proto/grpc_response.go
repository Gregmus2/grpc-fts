package proto

import (
	"encoding/json"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type GRPCResponse struct {
	Response           map[string]interface{}
	Status             *status.Status
	Stream             grpc.ClientStream
	IsStream           bool
	responseDescriptor protoreflect.MessageDescriptor
}

func NewGRPCUnaryResponse(response *dynamicpb.Message, err error) (*GRPCResponse, error) {
	result := &GRPCResponse{IsStream: false}

	err = result.UnmarshalResponse(response, err)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal response")
	}

	return result, nil
}

func NewGRPCStreamResponse(stream grpc.ClientStream, descriptor protoreflect.MessageDescriptor) (*GRPCResponse, error) {
	response := &GRPCResponse{IsStream: true, Stream: stream, responseDescriptor: descriptor}

	return response, response.StreamReceive()
}

func (r *GRPCResponse) StreamReceive() error {
	response := dynamicpb.NewMessage(r.responseDescriptor)
	err := r.Stream.RecvMsg(response)

	err = r.UnmarshalResponse(response, err)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal response")
	}

	return nil
}

func (r *GRPCResponse) UnmarshalResponse(response *dynamicpb.Message, err error) error {
	var ok bool
	r.Status, ok = status.FromError(errors.Cause(err))
	if !ok && err != nil {
		return errors.Wrap(err, "failed to send a request")
	}

	b, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(proto.Message(response))
	if err != nil {
		return errors.Wrap(err, "failed to marshal response")
	}
	r.Response = make(map[string]interface{})
	err = json.Unmarshal(b, &r.Response)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal response")
	}

	return nil
}
