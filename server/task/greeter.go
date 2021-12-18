/**
 *  MindLab
 *
 *  Create by songli on 2020/10/23
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package task

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/imind-lab/greeter/domain/greeter/service"
	"github.com/imind-lab/greeter/pkg/constant"
	"github.com/imind-lab/micro/broker"
	"go.uber.org/zap"
)

type SendGreeter struct {
	Ctx  context.Context
	Data interface{}
}

var GreeterQueue = make(chan SendGreeter, constant.GreeterQueueLen)

type GreeterTask struct {
	ctx context.Context
	dm  service.GreeterDomain
}

func NewGreeterTask(ctx context.Context) *GreeterTask {
	dm := service.NewGreeterDomain()
	svc := &GreeterTask{ctx: ctx, dm: dm}
	return svc
}

func (t *GreeterTask) GreeterTaskHandle() error {
	logger := ctxzap.Extract(t.ctx).With(zap.String("layer", "task"), zap.String("func", "AliyunCallbackHandle"))
	logger.Debug("AliyunCallbackHandle Begin...")

	endpoint, err := broker.NewBroker(constant.MQName)
	if err != nil {
		ctxzap.Error(t.ctx, "broker.NewBroker error", zap.Error(err))
		return err
	}

	for {
		select {
		case data, ok := <-GreeterQueue:
			fmt.Println("receive data:", data, ok)
			if ok {
				t.greeterProcessor(data, endpoint)
			} else {
				logger.Warn("channel closed")
				return nil
			}
		case <-t.ctx.Done():
			fmt.Println("上下文结束", t.ctx.Err())
			return nil
		}
	}
	return nil
}

func (t *GreeterTask) greeterProcessor(data SendGreeter, endpoint broker.Broker) error {
	fmt.Println("greeterProcessor", data)

	return nil
}

func (t *GreeterTask) Close() error {
	logger := ctxzap.Extract(t.ctx).With(zap.String("layer", "task"), zap.String("func", "Close"))
	logger.Debug("GreeterTask Close")

	close(GreeterQueue)

	return nil
}
