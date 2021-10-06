/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	"github.com/imind-lab/greeter/pkg/constant"
	"github.com/imind-lab/greeter/server/domain"
	"github.com/imind-lab/greeter/server/proto/greeter"
	"github.com/imind-lab/micro/broker"
	"github.com/imind-lab/micro/util"
)

type GreeterService struct {
	greeter.UnimplementedGreeterServiceServer

	vd *validator.Validate

	dm domain.GreeterDomain
}

func NewGreeterService() *GreeterService {
	dm := domain.NewGreeterDomain()
	svc := &GreeterService{
		dm: dm,
		vd: validator.New(),
	}

	return svc
}

// CreateGreeter 创建Greeter
func (svc *GreeterService) CreateGreeter(ctx context.Context, req *greeter.CreateGreeterRequest) (*greeter.CreateGreeterResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "CreateGreeter"))
	logger.Debug("Receive CreateGreeter request")

	rsp := &greeter.CreateGreeterResponse{}

	m := req.Dto
	fmt.Println("Dto", m)
	err := svc.vd.Struct(req)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
		}

		for _, err := range err.(validator.ValidationErrors) {
			fmt.Println(err.Namespace())
			fmt.Println(err.Field())
			fmt.Println(err.StructNamespace())
			fmt.Println(err.StructField())
			fmt.Println(err.Tag())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Type())
			fmt.Println(err.Value())
			fmt.Println(err.Param())
			fmt.Println()
		}

	}
	if m == nil {
		logger.Error("Greeter不能为空", zap.Any("params", m), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "Greeter不能为空"
		rsp.Error = err
		return rsp, nil
	}

	err = svc.vd.Var(m.Name, "required,email")
	if err != nil {
		logger.Error("Name不能为空", zap.Any("name", m.Name), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "Name不能为空"
		rsp.Error = err
		return rsp, nil
	}
	m.CreateTime = util.GetNowWithMillisecond()
	m.CreateDatetime = time.Now().Format(util.DateTimeFmt)
	m.UpdateDatetime = time.Now().Format(util.DateTimeFmt)
	err = svc.dm.CreateGreeter(ctx, m)
	if err != nil {
		logger.Error("创建Greeter失败", zap.Any("greeter", m), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "创建Greeter失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = true

	endpoint, err := broker.NewBroker(constant.MQName)
	if err != nil {
		ctxzap.Error(ctx, "broker.NewBroker error", zap.Error(err))
		return rsp, err
	}
	endpoint.Publish(&broker.Message{
		Topic: endpoint.Options().Topics["creategreeter"],
		Body:  []byte(fmt.Sprintf("Greeter %s Created", m.Name)),
	})

	return rsp, nil
}

// GetGreeterById 根据Id获取Greeter
func (svc *GreeterService) GetGreeterById(ctx context.Context, req *greeter.GetGreeterByIdRequest) (*greeter.GetGreeterByIdResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "GetGreeterById"))
	logger.Debug("Receive GetGreeterById request")

	rsp := &greeter.GetGreeterByIdResponse{}
	m, err := svc.dm.GetGreeterById(ctx, req.Id)
	if err != nil {
		logger.Error("获取Greeter失败", zap.Any("greeter", m), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "获取Greeter失败"
		rsp.Error = err
		return rsp, nil
	}

	rsp.Success = true
	rsp.Dto = m
	return rsp, nil
}

func (svc *GreeterService) GetGreeterList(ctx context.Context, req *greeter.GetGreeterListRequest) (*greeter.GetGreeterListResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "GetGreeterList"))
	logger.Debug("Receive GetGreeterList request")
	rsp := &greeter.GetGreeterListResponse{}

	err := svc.vd.Struct(req)
	if err != nil {

		if _, ok := err.(*validator.InvalidValidationError); ok {
			fmt.Println(err)
		}

		for _, err := range err.(validator.ValidationErrors) {

			fmt.Println(err.Namespace())
			fmt.Println(err.Field())
			fmt.Println(err.StructNamespace())
			fmt.Println(err.StructField())
			fmt.Println(err.Tag())
			fmt.Println(err.ActualTag())
			fmt.Println(err.Kind())
			fmt.Println(err.Type())
			fmt.Println(err.Value())
			fmt.Println(err.Param())
			fmt.Println()
		}

	}
	err = svc.vd.Var(req.Status, "gte=0,lte=3")
	if err != nil {
		logger.Error("请输入有效的Status", zap.Int32("status", req.Status), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "请输入有效的Status"
		rsp.Error = err
		return rsp, nil
	}

	if req.Pagesize <= 0 {
		req.Pagesize = 20
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	list, err := svc.dm.GetGreeterList(ctx, req.Status, req.Lastid, req.Pagesize, req.Page)
	if err != nil {
		logger.Error("获取Greeter失败", zap.Any("list", list), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "获取GreeterList失败"
		rsp.Error = err
		return rsp, nil
	}
	rsp.Success = true
	rsp.Data = list
	return rsp, nil
}

func (svc *GreeterService) UpdateGreeterStatus(ctx context.Context, req *greeter.UpdateGreeterStatusRequest) (*greeter.UpdateGreeterStatusResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "UpdateGreeterStatus"))
	logger.Debug("Receive UpdateGreeterStatus request")

	rsp := &greeter.UpdateGreeterStatusResponse{}
	affected, err := svc.dm.UpdateGreeterStatus(ctx, req.Id, req.Status)
	if err != nil || affected <= 0 {
		logger.Error("更新Greeter失败", zap.Int64("affected", affected), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "更新Greeter失败"
		rsp.Error = err
		return rsp, nil
	}
	rsp.Success = true
	return rsp, nil
}

func (svc *GreeterService) UpdateGreeterCount(ctx context.Context, req *greeter.UpdateGreeterCountRequest) (*greeter.UpdateGreeterCountResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "UpdateGreeterCount"))
	logger.Debug("Receive UpdateGreeterCount request")

	rsp := &greeter.UpdateGreeterCountResponse{}
	affected, err := svc.dm.UpdateGreeterCount(ctx, req.Id, req.Num, req.Column)
	if err != nil || affected <= 0 {
		logger.Error("更新Greeter失败", zap.Int64("affected", affected), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "更新Greeter失败"
		rsp.Error = err
		return rsp, nil
	}
	rsp.Success = true

	endpoint, err := broker.NewBroker(constant.MQName)
	if err != nil {
		ctxzap.Error(ctx, "kafka.New error", zap.Error(err))
		return rsp, err
	}
	endpoint.Publish(&broker.Message{
		Topic: endpoint.Options().Topics["updategreetercount"],
		Body:  nil,
	})

	return rsp, nil
}

func (svc *GreeterService) DeleteGreeterById(ctx context.Context, req *greeter.DeleteGreeterByIdRequest) (*greeter.DeleteGreeterByIdResponse, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreeterService"), zap.String("func", "DeleteGreeterById"))
	logger.Debug("Receive DeleteGreeterById request")

	rsp := &greeter.DeleteGreeterByIdResponse{}
	affected, err := svc.dm.DeleteGreeterById(ctx, req.Id)
	if err != nil || affected <= 0 {
		logger.Error("更新Greeter失败", zap.Int64("affected", affected), zap.Error(err))

		err := &greeter.Error{}
		err.Message = "删除Greeter失败"
		rsp.Error = err
		return rsp, nil
	}
	rsp.Success = true
	return rsp, nil
}

func (svc *GreeterService) GetGreeterListByStream(stream greeter.GreeterService_GetGreeterListByStreamServer) error {
	logger := ctxzap.Extract(stream.Context()).With(zap.String("layer", "GreeterService"), zap.String("func", "GetGreeterListByStream"))
	logger.Debug("Receive GetGreeterListByStream request")

	for {
		r, err := stream.Recv()
		ctxzap.Debug(stream.Context(), "stream.Recv", zap.Any("r", r), zap.Error(err))
		if err == io.EOF {
			return nil
		}
		if err != nil {
			logger.Error("Recv Stream error", zap.Error(err))
			return err
		}

		if r.Id > 0 {
			m, err := svc.dm.GetGreeterById(stream.Context(), r.Id)
			if err != nil {
				logger.Error("GetGreeterById error", zap.Any("greeter", m), zap.Error(err))
				return err
			}

			err = stream.Send(&greeter.GetGreeterListByStreamResponse{
				Index:  r.Index,
				Result: m,
			})
			if err != nil {
				logger.Error("Send Stream error", zap.Error(err))
				return err
			}
		} else {
			_ = stream.Send(&greeter.GetGreeterListByStreamResponse{
				Index:  r.Index,
				Result: nil,
			})
		}

	}
}
