/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package domain

import (
	"context"
	"math"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/imind-lab/greet/server/model"
	"github.com/imind-lab/greet/server/proto/greet"
	"github.com/imind-lab/greet/server/repository"
	"github.com/imind-lab/micro/dao"
)

type GreetDomain interface {
	CreateGreet(ctx context.Context, dto *greet.Greet) error

	GetGreetById(ctx context.Context, id int32) (*greet.Greet, error)
	GetGreetList(ctx context.Context, status, lastId, pageSize, page int32) (*greet.GreetList, error)

	UpdateGreetStatus(ctx context.Context, id, status int32) (int64, error)
	UpdateGreetCount(ctx context.Context, id, num int32, column string) (int64, error)

	DeleteGreetById(ctx context.Context, id int32) (int64, error)
}

type greetDomain struct {
	dao.Cache

	repo repository.GreetRepository
}

func NewGreetDomain() GreetDomain {
	repo := repository.NewGreetRepository()
	dm := greetDomain{
		Cache: dao.NewCache(),
		repo:  repo}
	return dm
}

func (dm greetDomain) CreateGreet(ctx context.Context, dto *greet.Greet) error {
	m := GreetDto2Model(dto)
	_, err := dm.repo.CreateGreet(ctx, m)
	return err
}

func (dm greetDomain) GetGreetById(ctx context.Context, id int32) (*greet.Greet, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greetDomain"), zap.String("func", "GetGreetById"))

	logger.Info("greetDomain.GetGreetById invoke")
	m, err := dm.repo.GetGreetById(ctx, id)
	return GreetModel2Dto(m), errors.WithMessage(err, "greetDomain.GetGreetById")
}

func (dm greetDomain) GetGreetList(ctx context.Context, status, lastId, pageSize, page int32) (*greet.GreetList, error) {
	list, total, err := dm.repo.GetGreetList(ctx, status, lastId, pageSize, page)
	if err != nil {
		return nil, err
	}
	greets := GreetMap(list, GreetModel2Dto)

	var totalPage int32 = 0
	if total == 0 {
		page = 1
	} else {
		totalPage = int32(math.Ceil(float64(total) / float64(pageSize)))
	}
	greetList := &greet.GreetList{}
	greetList.Datalist = greets
	greetList.Total = int32(total)
	greetList.TotalPage = totalPage
	greetList.CurPage = page

	return greetList, nil
}

func (dm greetDomain) UpdateGreetStatus(ctx context.Context, id, status int32) (int64, error) {
	return dm.repo.UpdateGreetStatus(ctx, id, status)
}

func (dm greetDomain) UpdateGreetCount(ctx context.Context, id, num int32, column string) (int64, error) {
	return dm.repo.UpdateGreetCount(ctx, id, num, column)
}

func (dm greetDomain) DeleteGreetById(ctx context.Context, id int32) (int64, error) {
	return dm.repo.DeleteGreetById(ctx, id)
}

func GreetMap(pos []model.Greet, fn func(model.Greet) *greet.Greet) []*greet.Greet {
	var dtos []*greet.Greet
	for _, po := range pos {
		dtos = append(dtos, fn(po))
	}
	return dtos
}

func GreetModel2Dto(po model.Greet) *greet.Greet {
	if po.IsEmpty() {
		return nil
	}

	dto := &greet.Greet{}
	dto.Id = po.Id
	dto.Name = po.Name
	dto.ViewNum = po.ViewNum
	dto.Status = po.Status
	dto.CreateTime = po.CreateTime
	dto.UpdateDatetime = po.UpdateDatetime
	dto.CreateDatetime = po.CreateDatetime

	return dto
}

func GreetDto2Model(dto *greet.Greet) model.Greet {
	if dto == nil {
		return model.Greet{}
	}

	po := model.Greet{}
	po.Id = dto.Id
	po.Name = dto.Name
	po.ViewNum = dto.ViewNum
	po.Status = dto.Status
	po.CreateTime = dto.CreateTime
	po.UpdateDatetime = dto.UpdateDatetime
	po.CreateDatetime = dto.CreateDatetime

	return po
}
