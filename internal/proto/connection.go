package proto

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	"github.com/res-am/grpc-fts/internal/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"os"
	"strings"
)

type grpcConnection struct {
	conn *grpc.ClientConn
}

func newConnection(address string, userTLS *models.TLS) (connection, error) {
	opts, err := buildTLS(userTLS)
	if err != nil {
		return nil, errors.Wrap(err, "failed to handle TLS")
	}

	conn, err := grpc.DialContext(context.Background(), address, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial to gRPC server")
	}

	return &grpcConnection{conn: conn}, nil
}

func buildTLS(userTLS *models.TLS) ([]grpc.DialOption, error) {
	if userTLS == nil {
		return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, nil
	}

	var tlsCfg tls.Config
	if userTLS.CertFile != nil {
		b, err := os.ReadFile(*userTLS.CertFile)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read the CA certificate")
		}
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM(b) {
			return nil, errors.New("failed to append the client certificate")
		}
		tlsCfg.RootCAs = cp
	}
	if userTLS.CertConfig != nil {
		// Enable mutual authentication
		certificate, err := tls.LoadX509KeyPair(userTLS.CertConfig.Cert, userTLS.CertConfig.Key)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read the client certificate")
		}
		tlsCfg.Certificates = append(tlsCfg.Certificates, certificate)
	}

	creds := credentials.NewTLS(&tlsCfg)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	if userTLS.ServerName != nil {
		opts = append(opts, grpc.WithAuthority(*userTLS.ServerName))
	}

	return opts, nil
}

func (c grpcConnection) Invoke(ctx context.Context, fullName string, req, res interface{}) (header, trailer metadata.MD, err error) {
	endpoint, err := fqrnToEndpoint(fullName)
	if err != nil {
		return nil, nil, err
	}

	wakeUpClientConn(c.conn)
	opts := []grpc.CallOption{grpc.Header(&header), grpc.Trailer(&trailer)}
	err = c.conn.Invoke(ctx, endpoint, req, res, opts...)

	return header, trailer, err
}

func (c grpcConnection) Stream(ctx context.Context, fullName string, streamDesc *grpc.StreamDesc) (grpc.ClientStream, error) {
	endpoint, err := fqrnToEndpoint(fullName)
	if err != nil {
		return nil, err
	}

	wakeUpClientConn(c.conn)
	stream, err := c.conn.NewStream(ctx, streamDesc, endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "error building stream")
	}

	return stream, nil
}

func wakeUpClientConn(conn *grpc.ClientConn) {
	if conn.GetState() == connectivity.TransientFailure {
		conn.ResetConnectBackoff()
	}
}

func fqrnToEndpoint(fqrn string) (string, error) {
	sp := strings.Split(fqrn, ".")
	// FQRN should contain at least service and rpc name.
	if len(sp) < 2 {
		return "", errors.New("invalid FQRN format")
	}

	return fmt.Sprintf("/%s/%s", strings.Join(sp[:len(sp)-1], "."), sp[len(sp)-1]), nil
}
