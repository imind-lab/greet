/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package subscriber

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/imind-lab/micro/broker"
	"go.uber.org/zap"
)

type Greeter struct {
	ctx context.Context
}

func NewGreeter(ctx context.Context) *Greeter {
	svc := &Greeter{ctx}
	return svc
}

func (svc *Greeter) CreateHandle(msg *broker.Message) error {
	logger := ctxzap.Extract(svc.ctx).With(zap.String("layer", "greeterSubscriber"), zap.String("func", "CreateHandle"))
	logger.Debug("greeter_create", zap.String("key", msg.Key), zap.String("body", string(msg.Body)))
	return nil
}

func (svc *Greeter) UpdateCountHandle(msg *broker.Message) error {
	logger := ctxzap.Extract(svc.ctx).With(zap.String("layer", "greeterSubscriber"), zap.String("func", "CreateHandle"))
	logger.Debug("greeter_update_count", zap.String("key", msg.Key), zap.String("body", string(msg.Body)))
	return nil
}
