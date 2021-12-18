/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package service

import (
	"context"
	"github.com/imind-lab/greeter/application/greeter/proto"
	"github.com/imind-lab/greeter/domain/greeter/repository"
	"github.com/imind-lab/greeter/domain/greeter/repository/model"
	"github.com/imind-lab/greeter/domain/greeter/repository/persistence"
	"math"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/imind-lab/micro/dao"
)

type GreeterDomain interface {
	CreateGreeter(ctx context.Context, dto *greeter.Greeter) error

	GetGreeterById(ctx context.Context, id int32) (*greeter.Greeter, error)
	GetGreeterList(ctx context.Context, status, lastId, pageSize, page int32) (*greeter.GreeterList, error)

	UpdateGreeterStatus(ctx context.Context, id, status int32) (int64, error)
	UpdateGreeterCount(ctx context.Context, id, num int32, column string) (int64, error)

	DeleteGreeterById(ctx context.Context, id int32) (int64, error)
}

type greeterDomain struct {
	dao.Cache

	repo repository.GreeterRepository
}

func NewGreeterDomain() GreeterDomain {
	repo := persistence.NewGreeterRepository()
	dm := greeterDomain{
		Cache: dao.NewCache(),
		repo:  repo}
	return dm
}

func (dm greeterDomain) CreateGreeter(ctx context.Context, dto *greeter.Greeter) error {
	m := GreeterDto2Model(dto)
	_, err := dm.repo.CreateGreeter(ctx, m)
	return err
}

func (dm greeterDomain) GetGreeterById(ctx context.Context, id int32) (*greeter.Greeter, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greeterDomain"), zap.String("func", "GetGreeterById"))

	logger.Info("greeterDomain.GetGreeterById invoke")
	m, err := dm.repo.GetGreeterById(ctx, id)
	return GreeterModel2Dto(m), errors.WithMessage(err, "greeterDomain.GetGreeterById")
}

func (dm greeterDomain) GetGreeterList(ctx context.Context, status, lastId, pageSize, page int32) (*greeter.GreeterList, error) {
	list, total, err := dm.repo.GetGreeterList(ctx, status, lastId, pageSize, page)
	if err != nil {
		return nil, err
	}
	greeters := GreeterMap(list, GreeterModel2Dto)

	var totalPage int32 = 0
	if total == 0 {
		page = 1
	} else {
		totalPage = int32(math.Ceil(float64(total) / float64(pageSize)))
	}
	greeterList := &greeter.GreeterList{}
	greeterList.Datalist = greeters
	greeterList.Total = int32(total)
	greeterList.TotalPage = totalPage
	greeterList.CurPage = page

	return greeterList, nil
}

func (dm greeterDomain) UpdateGreeterStatus(ctx context.Context, id, status int32) (int64, error) {
	return dm.repo.UpdateGreeterStatus(ctx, id, status)
}

func (dm greeterDomain) UpdateGreeterCount(ctx context.Context, id, num int32, column string) (int64, error) {
	return dm.repo.UpdateGreeterCount(ctx, id, num, column)
}

func (dm greeterDomain) DeleteGreeterById(ctx context.Context, id int32) (int64, error) {
	return dm.repo.DeleteGreeterById(ctx, id)
}

func GreeterMap(pos []model.Greeter, fn func(model.Greeter) *greeter.Greeter) []*greeter.Greeter {
	var dtos []*greeter.Greeter
	for _, po := range pos {
		dtos = append(dtos, fn(po))
	}
	return dtos
}

func GreeterModel2Dto(po model.Greeter) *greeter.Greeter {
	if po.IsEmpty() {
		return nil
	}

	dto := &greeter.Greeter{}
	dto.Id = po.Id
	dto.Name = po.Name
	dto.ViewNum = po.ViewNum
	dto.Status = po.Status
	dto.CreateTime = po.CreateTime
	dto.UpdateDatetime = po.UpdateDatetime
	dto.CreateDatetime = po.CreateDatetime

	return dto
}

func GreeterDto2Model(dto *greeter.Greeter) model.Greeter {
	if dto == nil {
		return model.Greeter{}
	}

	po := model.Greeter{}
	po.Id = dto.Id
	po.Name = dto.Name
	po.ViewNum = dto.ViewNum
	po.Status = dto.Status
	po.CreateTime = dto.CreateTime
	po.UpdateDatetime = dto.UpdateDatetime
	po.CreateDatetime = dto.CreateDatetime

	return po
}
