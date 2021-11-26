package grpc

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func MakeConn(addr string, tls bool, cert string) (*grpc.ClientConn, error) {
	opts, err := MakeClientOpts(tls, cert)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(addr, opts...)
}

func MakeClientOpts(tls bool, cert string) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	if tls {
		if cert == "" {
			return nil, fmt.Errorf("certificate file path is undefined")
		}
		creds, err := credentials.NewClientTLSFromFile(cert, "")
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS credentials: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithBlock())
	return opts, nil
}

func MakeServer(tls bool, cert, key string) (*grpc.Server, error) {
	var opts []grpc.ServerOption

	if tls {
		if cert == "" || key == "" {
			return nil, errors.New("cert file path or key file path is empty")
		}
		creds, err := credentials.NewServerTLSFromFile(cert, key)
		if err != nil {
			return nil, err
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	return grpc.NewServer(opts...), nil
}
