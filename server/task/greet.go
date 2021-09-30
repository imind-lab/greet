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
	"github.com/imind-lab/greet/pkg/constant"
	"github.com/imind-lab/greet/server/domain"
	"github.com/imind-lab/micro/broker"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type SendGreet struct {
	Ctx  context.Context
	Data interface{}
}

var GreetQueue = make(chan SendGreet, constant.GreetQueueLen)

type GreetTask struct {
	ctx context.Context
	dm  domain.GreetDomain
}

func NewGreetTask(ctx context.Context) *GreetTask {
	dm := domain.NewGreetDomain()
	svc := &GreetTask{ctx: ctx, dm: dm}
	return svc
}

func (t *GreetTask) GreetTaskHandle() error {
	logger := ctxzap.Extract(t.ctx).With(zap.String("layer", "task"), zap.String("func", "AliyunCallbackHandle"))
	logger.Debug("AliyunCallbackHandle Begin...")

	endpoint, err := broker.NewBroker(constant.MQName)
	if err != nil {
		ctxzap.Error(t.ctx, "broker.NewBroker error", zap.Error(err))
		return err
	}

	for {
		select {
		case data, ok := <-GreetQueue:
			fmt.Println("receive data:", data, ok)
			if ok {
				t.greetProcessor(data, endpoint)
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

func (t *GreetTask) greetProcessor(data SendGreet, endpoint broker.Broker) error {
	fmt.Println("greetProcessor", data)

	return nil
}

func (t *GreetTask) Close() error {
	logger := ctxzap.Extract(t.ctx).With(zap.String("layer", "task"), zap.String("func", "Close"))
	logger.Debug("GreetTask Close")

	close(GreetQueue)

	return nil
}
