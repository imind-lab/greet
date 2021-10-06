/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package client

import (
	"context"
	"github.com/imind-lab/greeter/server/proto/greeter"
	"github.com/imind-lab/micro/grpcx"
	"io"
)

type greeterClient struct {
	greeter.GreeterServiceClient
	closer io.Closer
}

func NewGreeterClient(ctx context.Context, name string, tls bool) (*greeterClient, error) {
	conn, closer, err := grpcx.ClientConn(ctx, name, tls)
	if err != nil {
		return nil, err
	}
	return &greeterClient{
		GreeterServiceClient: greeter.NewGreeterServiceClient(conn),
		closer:               closer,
	}, nil
}

func (tc *greeterClient) Close() error {
	return tc.closer.Close()
}
