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

        "github.com/imind-lab/greet/pkg/constant"
        "github.com/imind-lab/greet/server/domain"
        "github.com/imind-lab/greet/server/proto/greet"
        "github.com/imind-lab/micro/broker"
        "github.com/imind-lab/micro/util"
)

type GreetService struct {
        greet.UnimplementedGreetServiceServer

        vd *validator.Validate

        dm domain.GreetDomain
}

func NewGreetService() *GreetService {
        dm := domain.NewGreetDomain()
        svc := &GreetService{
                dm: dm,
                vd: validator.New(),
        }

        return svc
}

// CreateGreet 创建Greet
func (svc *GreetService) CreateGreet(ctx context.Context, req *greet.CreateGreetRequest) (*greet.CreateGreetResponse, error) {
        logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "CreateGreet"))
        logger.Debug("Receive CreateGreet request")

        rsp := &greet.CreateGreetResponse{}

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
                logger.Error("Greet不能为空", zap.Any("params", m), zap.Error(err))

                err := &greet.Error{}
                err.Message = "Greet不能为空"
                rsp.Error = err
                return rsp, nil
        }

        err = svc.vd.Var(m.Name, "required,email")
        if err != nil {
                logger.Error("Name不能为空", zap.Any("name", m.Name), zap.Error(err))

                err := &greet.Error{}
                err.Message = "Name不能为空"
                rsp.Error = err
                return rsp, nil
        }
        m.CreateTime = util.GetNowWithMillisecond()
        m.CreateDatetime = time.Now().Format(util.DateTimeFmt)
        m.UpdateDatetime = time.Now().Format(util.DateTimeFmt)
        err = svc.dm.CreateGreet(ctx, m)
        if err != nil {
                logger.Error("创建Greet失败", zap.Any("greet", m), zap.Error(err))

                err := &greet.Error{}
                err.Message = "创建Greet失败"
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
                Topic: endpoint.Options().Topics["creategreet"],
                Body:  []byte(fmt.Sprintf("Greet %s Created", m.Name)),
        })

        return rsp, nil
}

// GetGreetById 根据Id获取Greet
func (svc *GreetService) GetGreetById(ctx context.Context, req *greet.GetGreetByIdRequest) (*greet.GetGreetByIdResponse, error) {
        logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "GetGreetById"))
        logger.Debug("Receive GetGreetById request")

        rsp := &greet.GetGreetByIdResponse{}
        m, err := svc.dm.GetGreetById(ctx, req.Id)
        if err != nil {
                logger.Error("获取Greet失败", zap.Any("greet", m), zap.Error(err))

                err := &greet.Error{}
                err.Message = "获取Greet失败"
                rsp.Error = err
                return rsp, nil
        }

        rsp.Success = true
        rsp.Dto = m
        return rsp, nil
}

func (svc *GreetService) GetGreetList(ctx context.Context, req *greet.GetGreetListRequest) (*greet.GetGreetListResponse, error) {
        logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "GetGreetList"))
        logger.Debug("Receive GetGreetList request")
        rsp := &greet.GetGreetListResponse{}

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

                err := &greet.Error{}
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

        list, err := svc.dm.GetGreetList(ctx, req.Status, req.Lastid, req.Pagesize, req.Page)
        if err != nil {
                logger.Error("获取Greet失败", zap.Any("list", list), zap.Error(err))

                err := &greet.Error{}
                err.Message = "获取GreetList失败"
                rsp.Error = err
                return rsp, nil
        }
        rsp.Success = true
        rsp.Data = list
        return rsp, nil
}

func (svc *GreetService) UpdateGreetStatus(ctx context.Context, req *greet.UpdateGreetStatusRequest) (*greet.UpdateGreetStatusResponse, error) {
        logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "UpdateGreetStatus"))
        logger.Debug("Receive UpdateGreetStatus request")

        rsp := &greet.UpdateGreetStatusResponse{}
        affected, err := svc.dm.UpdateGreetStatus(ctx, req.Id, req.Status)
        if err != nil || affected <= 0 {
                logger.Error("更新Greet失败", zap.Int64("affected", affected), zap.Error(err))

                err := &greet.Error{}
                err.Message = "更新Greet失败"
                rsp.Error = err
                return rsp, nil
        }
        rsp.Success = true
        return rsp, nil
}

func (svc *GreetService) UpdateGreetCount(ctx context.Context, req *greet.UpdateGreetCountRequest) (*greet.UpdateGreetCountResponse, error) {
        logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "UpdateGreetCount"))
        logger.Debug("Receive UpdateGreetCount request")

        rsp := &greet.UpdateGreetCountResponse{}
        affected, err := svc.dm.UpdateGreetCount(ctx, req.Id, req.Num, req.Column)
        if err != nil || affected <= 0 {
                logger.Error("更新Greet失败", zap.Int64("affected", affected), zap.Error(err))

                err := &greet.Error{}
                err.Message = "更新Greet失败"
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
                Topic: endpoint.Options().Topics["updategreetcount"],
                Body:  nil,
        })

        return rsp, nil
}

func (svc *GreetService) DeleteGreetById(ctx context.Context, req *greet.DeleteGreetByIdRequest) (*greet.DeleteGreetByIdResponse, error) {
        logger := ctxzap.Extract(ctx).With(zap.String("layer", "GreetService"), zap.String("func", "DeleteGreetById"))
        logger.Debug("Receive DeleteGreetById request")

        rsp := &greet.DeleteGreetByIdResponse{}
        affected, err := svc.dm.DeleteGreetById(ctx, req.Id)
        if err != nil || affected <= 0 {
                logger.Error("更新Greet失败", zap.Int64("affected", affected), zap.Error(err))

                err := &greet.Error{}
                err.Message = "删除Greet失败"
                rsp.Error = err
                return rsp, nil
        }
        rsp.Success = true
        return rsp, nil
}

func (svc *GreetService) GetGreetListByStream(stream greet.GreetService_GetGreetListByStreamServer) error {
        logger := ctxzap.Extract(stream.Context()).With(zap.String("layer", "GreetService"), zap.String("func", "GetGreetListByStream"))
        logger.Debug("Receive GetGreetListByStream request")

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
                        m, err := svc.dm.GetGreetById(stream.Context(), r.Id)
                        if err != nil {
                                logger.Error("GetGreetById error", zap.Any("greet", m), zap.Error(err))
                                return err
                        }

                        err = stream.Send(&greet.GetGreetListByStreamResponse{
                                Index:  r.Index,
                                Result: m,
                        })
                        if err != nil {
                                logger.Error("Send Stream error", zap.Error(err))
                                return err
                        }
                } else {
                        _ = stream.Send(&greet.GetGreetListByStreamResponse{
                                Index:  r.Index,
                                Result: nil,
                        })
                }

        }
}
