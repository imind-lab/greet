package server

import (
	"fmt"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/imind-lab/greet/pkg/constant"
	"github.com/imind-lab/greet/server/proto/greet"
	"github.com/imind-lab/greet/server/service"
	"github.com/imind-lab/greet/server/subscriber"
	"github.com/imind-lab/micro"
	"github.com/imind-lab/micro/broker"
	"github.com/imind-lab/micro/grpcx"
)

func Serve() error {
	svc := micro.NewService()

	// 初始化kafka代理
	endpoint, err := broker.NewBroker(constant.MQName)
	if err != nil {
		return err
	}
	// 设置消息队列事件处理器（可选）
	mqHandler := subscriber.NewGreet(svc.Options().Context)
	endpoint.Subscribe(
		broker.Processor{Topic: endpoint.Options().Topics["createuser"], Handler: mqHandler.CreateHandle, Retry: 1},
		broker.Processor{Topic: endpoint.Options().Topics["updateusercount"], Handler: mqHandler.UpdateCountHandle, Retry: 0},
	)

	grpcCred := grpcx.NewGrpcCred()

	svc.Init(
		micro.Broker(endpoint),
		micro.ServerCred(grpcCred.ServerCred()),
		micro.ClientCred(grpcCred.ClientCred()))

	grpcSrv := svc.GrpcServer()
	greet.RegisterGreetServiceServer(grpcSrv, service.NewGreetService())

	// 注册gRPC-Gateway
	endPoint := fmt.Sprintf(":%d", viper.GetInt("service.port.grpc"))
	fmt.Println(endPoint)

	mux := svc.ServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(grpcCred.ClientCred())}
	err = greet.RegisterGreetServiceHandlerFromEndpoint(svc.Options().Context, mux, endPoint, opts)
	if err != nil {
		return err
	}
	return svc.Run()
}
