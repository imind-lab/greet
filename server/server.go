package server

import (
	"fmt"
	"github.com/imind-lab/greeter/application/greeter/event/subscriber"
	"github.com/imind-lab/greeter/application/greeter/proto"
	"github.com/imind-lab/greeter/application/greeter/service"

	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/imind-lab/greeter/pkg/constant"
	"github.com/imind-lab/micro"
	"github.com/imind-lab/micro/broker"
	grpcx "github.com/imind-lab/micro/grpc"
)

func Serve() error {
	svc := micro.NewService()

	// 初始化kafka代理
	endpoint, err := broker.NewBroker(constant.MQName)
	if err != nil {
		return err
	}
	// 设置消息队列事件处理器（可选）
	mqHandler := subscriber.NewGreeter(svc.Options().Context)
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
	greeter.RegisterGreeterServiceServer(grpcSrv, service.NewGreeterService())

	// 注册gRPC-Gateway
	endPoint := fmt.Sprintf(":%d", viper.GetInt("service.port.grpc"))
	fmt.Println(endPoint)

	mux := svc.ServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(grpcCred.ClientCred())}
	err = greeter.RegisterGreeterServiceHandlerFromEndpoint(svc.Options().Context, mux, endPoint, opts)
	if err != nil {
		return err
	}
	return svc.Run()
}
