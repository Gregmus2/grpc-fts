package proto_test

import (
	"context"
	"github.com/bufbuild/protocompile"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"testing"
)

func TestClient_Invoke(t *testing.T) {
	//cfg := &config.Global{
	//	ProtoSources: []string{"test.proto"},
	//	ProtoRoot: "test_data",
	//}
	//testCases := config.TestCases{
	//	{
	//		Steps: []config.Step{
	//			{
	//				Method: "ClientStreamMethod",
	//				Service: config.Service{Service: "TestService"},
	//			},
	//		},
	//	},
	//}
	//descriptorManager, err := proto.NewDescriptorsManager(cfg, testCases)
	//assert.NoError(t, err)

	go func() {
		c := &protocompile.Compiler{
			Resolver: protocompile.WithStandardImports(&protocompile.SourceResolver{
				ImportPaths: []string{"test_data"},
			}),
		}
		compiled, err := c.Compile(context.Background(), "test_data/test.proto")
		assert.NoError(t, err, "error on compile proto")
		serviceDesc := compiled[0].Services().ByName("TestService")

		server := grpc.NewServer()
		server.RegisterService(&grpc.ServiceDesc{
			ServiceName: string(serviceDesc.FullName()),
			HandlerType: nil,
			Methods: []grpc.MethodDesc{
				{
					MethodName: string(serviceDesc.Methods().Get(0).Name()),
					Handler:    nil,
				},
			},
			Streams:  nil,
			Metadata: nil,
		}, services[j])

		ln, err := net.Listen("tcp", ":9000")
		assert.NoError(t, err, "error on listen address :9000")
		err = server.Serve(ln)
		assert.NoError(t, err, "error on serve")
	}()
}
