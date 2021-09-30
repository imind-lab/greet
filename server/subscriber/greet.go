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

type Greet struct {
	ctx context.Context
}

func NewGreet(ctx context.Context) *Greet {
	svc := &Greet{ctx}
	return svc
}

func (svc *Greet) CreateHandle(msg *broker.Message) error {
	logger := ctxzap.Extract(svc.ctx).With(zap.String("layer", "greetSubscriber"), zap.String("func", "CreateHandle"))
	logger.Debug("greet_create", zap.String("key", msg.Key), zap.String("body", string(msg.Body)))
	return nil
}

func (svc *Greet) UpdateCountHandle(msg *broker.Message) error {
	logger := ctxzap.Extract(svc.ctx).With(zap.String("layer", "greetSubscriber"), zap.String("func", "CreateHandle"))
	logger.Debug("greet_update_count", zap.String("key", msg.Key), zap.String("body", string(msg.Body)))
	return nil
}
