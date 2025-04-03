package proto_test

import (
	"context"
	"github.com/bufbuild/protocompile"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/config"
	"github.com/res-am/grpc-fts/internal/proto"
	grpc_fts "github.com/res-am/grpc-fts/internal/proto/test_data"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net"
	"testing"
)

type TestService struct {
	grpc_fts.UnimplementedTestServiceServer
}

func (t TestService) UnaryMethod(ctx context.Context, request *grpc_fts.TestMessage) (*grpc_fts.TestMessage, error) {
	return &grpc_fts.TestMessage{Data: "ok"}, nil
}

func (t TestService) ClientStreamMethod(stream grpc.ClientStreamingServer[grpc_fts.TestMessage, grpc_fts.TestMessage]) error {
	for {
		_, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&grpc_fts.TestMessage{Data: "ok"})
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error on receive message")
		}
	}
}
func (t TestService) ServerStreamMethod(req *grpc_fts.TestMessage, stream grpc.ServerStreamingServer[grpc_fts.TestMessage]) error {
	stream.SendMsg(&grpc_fts.TestMessage{Data: "ok"})
	stream.SendMsg(&grpc_fts.TestMessage{Data: "ok2"})

	return nil
}
func (t TestService) BidiStreamMethod(stream grpc.BidiStreamingServer[grpc_fts.TestMessage, grpc_fts.TestMessage]) error {
	stream.SendMsg(&grpc_fts.TestMessage{Data: "ok"})
	stream.SendMsg(&grpc_fts.TestMessage{Data: "ok2"})

	for {
		_, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "error on receive message")
		}
	}

	return nil
}

func TestClient_Invoke(t *testing.T) {
	c := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
			ImportPaths: []string{"test_data"},
		}),
	}
	compiled, err := c.Compile(context.Background(), "test.proto")
	assert.NoError(t, err, "error on compile proto")
	serviceDesc := compiled[0].Services().ByName("TestService")

	server := grpc.NewServer()
	server.RegisterService(&grpc_fts.TestService_ServiceDesc, &TestService{})

	ln, err := net.Listen("tcp", ":9005")
	assert.NoError(t, err, "error on listen address :9000")

	defer server.Stop()
	go func() {
		err = server.Serve(ln)
		assert.NoError(t, err, "error on serve")
	}()

	cfg := &config.Global{
		ProtoSources: []string{"test.proto"},
		ProtoRoot:    "test_data",
	}
	testCases := config.TestCases{
		{
			Steps: []config.Step{
				{
					Method:  "UnaryMethod",
					Service: config.Service{Service: "test.TestService"},
				},
				{
					Method:  "BidiStreamMethod",
					Service: config.Service{Service: "test.TestService"},
				},
				{
					Method:  "ServerStreamMethod",
					Service: config.Service{Service: "test.TestService"},
				},
				{
					Method:  "ClientStreamMethod",
					Service: config.Service{Service: "test.TestService"},
				},
			},
		},
	}
	descriptorManager, err := proto.NewDescriptorsManager(cfg, testCases)
	assert.NoError(t, err)
	conn, err := proto.NewConnection(":9005", nil)
	assert.NoError(t, err)

	client := proto.NewClient(conn, descriptorManager)
	res, err := client.Invoke(serviceDesc.Methods().ByName("UnaryMethod").FullName(), []byte(`{"data": "test"}`), nil)
	assert.NoError(t, err, "error on invoke")
	assert.Equal(t, codes.OK, res.Status.Code())

	res, err = client.Invoke(serviceDesc.Methods().ByName("ClientStreamMethod").FullName(), []byte(`[{"data": "test"}, {"data": "test2"}]`), nil)
	assert.NoError(t, err, "error on invoke")
	assert.Equal(t, codes.OK, res.Status.Code())

	res, err = client.Invoke(serviceDesc.Methods().ByName("ServerStreamMethod").FullName(), []byte(`{"data": "test2"}`), nil)
	assert.NoError(t, err, "error on invoke")
	assert.Equal(t, codes.OK, res.Status.Code())

	res, err = client.Invoke(serviceDesc.Methods().ByName("BidiStreamMethod").FullName(), []byte(`[{"data": "test"}, {"data": "test2"}]`), nil)
	assert.NoError(t, err, "error on invoke")
	assert.Equal(t, codes.OK, res.Status.Code())
}
