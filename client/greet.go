/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package client

import (
	"context"
	"github.com/imind-lab/greet/server/proto/greet"
	"github.com/imind-lab/micro/grpcx"
	"io"
)

type greetClient struct {
	greet.GreetServiceClient
	closer io.Closer
}

func NewGreetClient(ctx context.Context, name string, tls bool) (*greetClient, error) {
	conn, closer, err := grpcx.ClientConn(ctx, name, tls)
	if err != nil {
		return nil, err
	}
	return &greetClient{
		GreetServiceClient: greet.NewGreetServiceClient(conn),
		closer:             closer,
	}, nil
}

func (tc *greetClient) Close() error {
	return tc.closer.Close()
}
