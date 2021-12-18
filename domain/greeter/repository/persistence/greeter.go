/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package persistence

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	errorsx "github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/imind-lab/greeter/domain/greeter/repository"
	"github.com/imind-lab/greeter/domain/greeter/repository/model"
	"github.com/imind-lab/greeter/pkg/constant"
	utilx "github.com/imind-lab/greeter/pkg/util"
	"github.com/imind-lab/micro/dao"
	redisx "github.com/imind-lab/micro/redis"
	"github.com/imind-lab/micro/tracing"
	"github.com/imind-lab/micro/util"
)

type greeterRepository struct {
	dao.Dao
}

//NewGreeterRepository 创建用户仓库实例
func NewGreeterRepository() repository.GreeterRepository {
	rep := dao.NewDao(constant.DBName)
	repo := greeterRepository{
		Dao: rep,
	}
	return repo
}

func (repo greeterRepository) CreateGreeter(ctx context.Context, m model.Greeter) (model.Greeter, error) {
	span, ctx := tracing.StartSpan(ctx, "greeterRepository.CreateGreeter")
	defer span.Finish()

	if err := repo.DB(ctx).Create(&m).Error; err != nil {
		return m, errorsx.Wrap(err, "greeterRepository.CreateGreeter")
	}
	repo.CacheGreeter(ctx, m)
	return m, nil
}

func (repo greeterRepository) CacheGreeter(ctx context.Context, m model.Greeter) error {
	span, ctx := tracing.StartSpan(ctx, "greeterRepository.CacheGreeter")
	defer span.Finish()

	key := utilx.CacheKey("greeter_", strconv.Itoa(int(m.Id)))
	expire := constant.CacheMinute5
	redisx.SetHashTable(ctx, repo.Redis(), key, m, expire)
	return nil
}

func (repo greeterRepository) GetGreeterById(ctx context.Context, id int32, opt ...repository.GreeterByIdOption) (model.Greeter, error) {
	span, ctx := tracing.StartSpan(ctx, "greeterRepository.GetGreeterById")
	defer span.Finish()

	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greeterRepository"), zap.String("func", "GetGreeterById"))

	opts := repository.NewGreeterByIdOptions(util.RandDuration(120))
	for _, o := range opt {
		o(opts)
	}

	var m model.Greeter
	key := utilx.CacheKey("greeter_", strconv.Itoa(int(id)))
	err := redisx.HGet(ctx, repo.Redis(), key, &m)
	logger.Debug("redis.HGetAll", zap.Any("greeter", m), zap.String("key", key), zap.Error(err))
	if err == nil {
		return m, nil
	}

	m, err = repo.FindGreeterById(ctx, id)
	if err != nil {
		return m, errorsx.WithMessage(err, "greeterRepository.GetGreeterById")
	}

	expire := constant.CacheMinute5 + opts.RandExpire
	if m.IsEmpty() {
		expire = constant.CacheMinute1
	}
	redisx.SetHashTable(ctx, repo.Redis(), key, m, expire)
	return m, nil
}

func (repo greeterRepository) FindGreeterById(ctx context.Context, id int32) (model.Greeter, error) {
	span, ctx := tracing.StartSpan(ctx, "greeterRepository.FindGreeterById")
	defer span.Finish()

	var m model.Greeter
	err := repo.DB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return m, nil
		}
		return m, errorsx.Wrap(err, "greeterRepository.FindGreeterById")
	}
	return m, nil
}

func (repo greeterRepository) GetGreetersCount(ctx context.Context, status int32) (int64, error) {
	span, ctx := tracing.StartSpan(ctx, "greeterRepository.GetGreetersCount")
	defer span.Finish()

	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greeterRepository"), zap.String("func", "GetGreetersCount"))

	key := utilx.CacheKey("greeter_cnt_", strconv.Itoa(int(status)))
	cnt, err := redisx.GetNumber(ctx, repo.Redis(), key)
	if err == nil {
		return cnt, nil
	}
	cnt, err = repo.FindGreetersCount(ctx, status)
	if err != nil {
		return 0, errorsx.WithMessage(err, "greeterRepository.GetGreetersCount")
	}
	err = repo.Redis().Set(ctx, key, cnt, constant.CacheMinute5).Err()
	if err != nil {
		logger.Error("redis.Set", zap.String("key", key), zap.Error(err))
	}
	return cnt, nil
}

func (repo greeterRepository) FindGreetersCount(ctx context.Context, status int32) (int64, error) {
	var count int64
	tx := repo.DB(ctx).Model(model.Greeter{}).Select("count(id)")
	tx = tx.Where("status=?", status)
	if err := tx.Count(&count).Error; err != nil {
		return 0, errorsx.Wrap(err, "greeterRepository.FindGreetersCount")
	}
	return count, nil
}

func (repo greeterRepository) GetGreeterList(ctx context.Context, status, lastId, pageSize, page int32) ([]model.Greeter, int, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greeterRepository"), zap.String("func", "GetGreeterList"))

	ids, cnt, err := repo.GetGreeterListIds(ctx, status, lastId, pageSize, page)
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "greeterRepository.GetGreeterList.GetGreeterListIds")
	}

	ctx1, cancel := context.WithTimeout(ctx, constant.CRequestTimeout)
	defer cancel()

	greeters, err := repo.GetGreeterList4Concurrent(ctx1, ids, repo.GetGreeterById)
	logger.Debug("GetGreeterList4Concurrent", zap.Any("greeters", greeters), zap.Error(err))
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "greeterRepository.GetGreeterList.GetGreeterList4Concurrent")
	}
	return greeters, cnt, nil
}

func (repo greeterRepository) GetGreeterListIds(ctx context.Context, status, lastId, pageSize, page int32) ([]int32, int, error) {
	key := utilx.CacheKey("greeter_ids_", strconv.Itoa(int(status)))

	ids, cnt, err := redisx.ZRevRangeWithCard(ctx, repo.Redis(), key, lastId, pageSize, page)
	if err == nil {
		return ids, cnt, nil
	}

	ids, args, err := repo.FindGreeterListIds(ctx, status, lastId, pageSize)
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "greeterRepository.GetGreeterList")
	}
	expire := constant.CacheMinute5 + util.RandDuration(120)
	redisx.SetSortedSet(ctx, repo.Redis(), key, args, expire)
	return ids, len(args), nil
}

func (repo greeterRepository) FindGreeterListIds(ctx context.Context, status, lastId, pageSize int32) ([]int32, []*redis.Z, error) {

	tx := repo.DB(ctx).Model(model.Greeter{}).Select("id")
	tx = tx.Where("status=?", status)
	tx = tx.Order("id DESC")
	rows, err := tx.Rows()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []int32{}, []*redis.Z{}, nil
		}
		return nil, nil, errorsx.Wrap(err, "greeterRepository.FindGreeterListIds.Rows")
	}
	defer rows.Close()

	var ids []int32
	var args []*redis.Z
	for rows.Next() {
		var (
			id int32
		)
		err = rows.Scan(&id)
		if err != nil {
			return nil, nil, errorsx.Wrap(err, "greeterRepository.FindGreeterListIds.Scan")
		}

		check := false
		if lastId == 0 {
			check = true
		} else if lastId > id {
			check = true
		}
		if check {
			if len(ids) < int(pageSize) {
				ids = append(ids, id)
			}
		}
		args = append(args, &redis.Z{Score: float64(id), Member: id})
	}
	if err = rows.Err(); err != nil {
		return nil, nil, errorsx.Wrap(err, "greeterRepository.FindGreeterListIds.Err")
	}
	return ids, args, nil
}

func (repo greeterRepository) GetGreeterList4Concurrent(ctx context.Context, ids []int32, fn func(context.Context, int32, ...repository.GreeterByIdOption) (model.Greeter, error)) ([]model.Greeter, error) {
	var wg sync.WaitGroup

	count := len(ids)
	outputs := make([]*concurrentGreeterOutput, count)
	wg.Add(count)

	for idx, id := range ids {
		go func(idx int, id int32, wg *sync.WaitGroup) {
			defer wg.Done()
			greeter, err := fn(ctx, id)
			outputs[idx] = &concurrentGreeterOutput{
				object: greeter,
				err:    err,
			}
		}(idx, id, &wg)
	}
	wg.Wait()

	greeters := make([]model.Greeter, 0, count)
	for _, output := range outputs {
		if output.err == nil {
			greeters = append(greeters, output.object)
		}
	}
	return greeters, nil
}

type concurrentGreeterOutput struct {
	object model.Greeter
	err    error
}

func (repo greeterRepository) UpdateGreeterStatus(ctx context.Context, id, status int32) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greeterRepository"), zap.String("func", "UpdateGreeterStatus"))

	logger.Debug("invoke info", zap.Int32("id", id), zap.Int32("status", status))
	tx := repo.DB(ctx).Model(model.Greeter{}).Where("id = ?", id)
	tx = tx.Update("status", status)
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "greeterRepository.UpdateGreeterStatus")
	}
	key := utilx.CacheKey("greeter_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	if err != nil {
		logger.Warn("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))
	}
	return tx.RowsAffected, nil
}

func (repo greeterRepository) UpdateGreeterCount(ctx context.Context, id, num int32, column string) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greeterRepository"), zap.String("func", "UpdateGreeterCount"))

	logger.Debug("invoke info", zap.Int32("id", id), zap.Int32("num", num), zap.String("column", column))
	tx := repo.DB(ctx).Model(model.Greeter{}).Where("id = ?", id)
	tx = tx.Updates(map[string]interface{}{column: gorm.Expr(column+" + ?", num)})
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "greeterRepository.UpdateGreeterCount")
	}
	key := utilx.CacheKey("greeter_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	if err != nil {
		logger.Warn("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))
	}
	return tx.RowsAffected, nil
}

func (repo greeterRepository) DeleteGreeterById(ctx context.Context, id int32) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greeterRepository"), zap.String("func", "DeleteGreeterById"))

	logger.Debug("invoke info", zap.Int32("id", id))
	tx := repo.DB(ctx).Delete(&model.Greeter{}, id)
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "greeterRepository.DeleteGreeterById")
	}
	key := utilx.CacheKey("greeter_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	logger.Debug("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))

	status := []int{0, 1}
	for _, s := range status {
		key := utilx.CacheKey("greeter_ids_", strconv.Itoa(s))
		err := repo.Redis().ZRem(ctx, key, id).Err()
		if err != nil {
			logger.Warn("redis.ZRem", zap.String("key", key), zap.Int32("id", id), zap.Error(err))
		}
	}

	return tx.RowsAffected, nil
}
