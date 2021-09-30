/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package repository

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

	"github.com/imind-lab/greet/pkg/constant"
	utilx "github.com/imind-lab/greet/pkg/util"
	"github.com/imind-lab/greet/server/model"
	"github.com/imind-lab/micro/dao"
	"github.com/imind-lab/micro/redisx"
	"github.com/imind-lab/micro/tracing"
	"github.com/imind-lab/micro/util"
)

type GreetRepository interface {
	CreateGreet(ctx context.Context, m model.Greet) (model.Greet, error)

	GetGreetById(ctx context.Context, id int32, opt ...GreetByIdOption) (model.Greet, error)
	FindGreetById(ctx context.Context, id int32) (model.Greet, error)
	GetGreetList(ctx context.Context, status, lastId, pageSize, page int32) ([]model.Greet, int, error)

	UpdateGreetStatus(ctx context.Context, id, status int32) (int64, error)
	UpdateGreetCount(ctx context.Context, id, num int32, column string) (int64, error)

	DeleteGreetById(ctx context.Context, id int32) (int64, error)
}

type greetRepository struct {
	dao.Dao
}

//NewGreetRepository 创建用户仓库实例
func NewGreetRepository() GreetRepository {
	rep := dao.NewRepository(constant.DBName, constant.Realtime)
	repo := greetRepository{
		Dao: rep,
	}
	return repo
}

func (repo greetRepository) CreateGreet(ctx context.Context, m model.Greet) (model.Greet, error) {
	span, ctx := tracing.StartSpan(ctx, "greetRepository.CreateGreet")
	defer span.Finish()

	if err := repo.WriteDB(ctx).Create(&m).Error; err != nil {
		return m, errorsx.Wrap(err, "greetRepository.CreateGreet")
	}
	repo.CacheGreet(ctx, m)
	return m, nil
}

func (repo greetRepository) CacheGreet(ctx context.Context, m model.Greet) error {
	span, ctx := tracing.StartSpan(ctx, "greetRepository.CacheGreet")
	defer span.Finish()

	key := utilx.CacheKey("greet_", strconv.Itoa(int(m.Id)))
	expire := constant.CacheMinute5
	redisx.SetHashTable(ctx, repo.Redis(), key, m, expire)
	return nil
}

func (repo greetRepository) GetGreetById(ctx context.Context, id int32, opt ...GreetByIdOption) (model.Greet, error) {
	span, ctx := tracing.StartSpan(ctx, "greetRepository.GetGreetById")
	defer span.Finish()

	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greetRepository"), zap.String("func", "GetGreetById"))

	opts := GreetByIdOptions{
		randExpire: util.RandDuration(120),
	}
	for _, o := range opt {
		o(&opts)
	}

	var m model.Greet
	key := utilx.CacheKey("greet_", strconv.Itoa(int(id)))
	err := redisx.HGet(ctx, repo.Redis(), key, &m)
	logger.Debug("redis.HGetAll", zap.Any("greet", m), zap.String("key", key), zap.Error(err))
	if err == nil {
		return m, nil
	}

	m, err = repo.FindGreetById(ctx, id)
	if err != nil {
		return m, errorsx.WithMessage(err, "greetRepository.GetGreetById")
	}

	expire := constant.CacheMinute5 + opts.randExpire
	if m.IsEmpty() {
		expire = constant.CacheMinute1
	}
	redisx.SetHashTable(ctx, repo.Redis(), key, m, expire)
	return m, nil
}

func (repo greetRepository) FindGreetById(ctx context.Context, id int32) (model.Greet, error) {
	span, ctx := tracing.StartSpan(ctx, "greetRepository.FindGreetById")
	defer span.Finish()

	var m model.Greet
	err := repo.ReadDB(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return m, nil
		}
		return m, errorsx.Wrap(err, "greetRepository.FindGreetById")
	}
	return m, nil
}

func (repo greetRepository) GetGreetsCount(ctx context.Context, status int32) (int64, error) {
	span, ctx := tracing.StartSpan(ctx, "greetRepository.GetGreetsCount")
	defer span.Finish()

	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greetRepository"), zap.String("func", "GetGreetsCount"))

	key := utilx.CacheKey("greet_cnt_", strconv.Itoa(int(status)))
	cnt, err := redisx.GetNumber(ctx, repo.Redis(), key)
	if err == nil {
		return cnt, nil
	}
	cnt, err = repo.FindGreetsCount(ctx, status)
	if err != nil {
		return 0, errorsx.WithMessage(err, "greetRepository.GetGreetsCount")
	}
	err = repo.Redis().Set(ctx, key, cnt, constant.CacheMinute5).Err()
	if err != nil {
		logger.Error("redis.Set", zap.String("key", key), zap.Error(err))
	}
	return cnt, nil
}

func (repo greetRepository) FindGreetsCount(ctx context.Context, status int32) (int64, error) {
	var count int64
	tx := repo.ReadDB(ctx).Model(model.Greet{}).Select("count(id)")
	tx = tx.Where("status=?", status)
	if err := tx.Count(&count).Error; err != nil {
		return 0, errorsx.Wrap(err, "greetRepository.FindGreetsCount")
	}
	return count, nil
}

func (repo greetRepository) GetGreetList(ctx context.Context, status, lastId, pageSize, page int32) ([]model.Greet, int, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greetRepository"), zap.String("func", "GetGreetList"))

	ids, cnt, err := repo.GetGreetListIds(ctx, status, lastId, pageSize, page)
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "greetRepository.GetGreetList.GetGreetListIds")
	}

	ctx1, cancel := context.WithTimeout(ctx, constant.CRequestTimeout)
	defer cancel()

	greets, err := repo.GetGreetList4Concurrent(ctx1, ids, repo.GetGreetById)
	logger.Debug("GetGreetList4Concurrent", zap.Any("greets", greets), zap.Error(err))
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "greetRepository.GetGreetList.GetGreetList4Concurrent")
	}
	return greets, cnt, nil
}

func (repo greetRepository) GetGreetListIds(ctx context.Context, status, lastId, pageSize, page int32) ([]int32, int, error) {
	key := utilx.CacheKey("greet_ids_", strconv.Itoa(int(status)))

	ids, cnt, err := redisx.ZRevRangeWithCard(ctx, repo.Redis(), key, lastId, pageSize, page)
	if err == nil {
		return ids, cnt, nil
	}

	ids, args, err := repo.FindGreetListIds(ctx, status, lastId, pageSize)
	if err != nil {
		return nil, 0, errorsx.WithMessage(err, "greetRepository.GetGreetList")
	}
	expire := constant.CacheMinute5 + util.RandDuration(120)
	redisx.SetSortedSet(ctx, repo.Redis(), key, args, expire)
	return ids, len(args), nil
}

func (repo greetRepository) FindGreetListIds(ctx context.Context, status, lastId, pageSize int32) ([]int32, []*redis.Z, error) {

	tx := repo.ReadDB(ctx).Model(model.Greet{}).Select("id")
	tx = tx.Where("status=?", status)
	tx = tx.Order("id DESC")
	rows, err := tx.Rows()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []int32{}, []*redis.Z{}, nil
		}
		return nil, nil, errorsx.Wrap(err, "greetRepository.FindGreetListIds.Rows")
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
			return nil, nil, errorsx.Wrap(err, "greetRepository.FindGreetListIds.Scan")
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
		return nil, nil, errorsx.Wrap(err, "greetRepository.FindGreetListIds.Err")
	}
	return ids, args, nil
}

func (repo greetRepository) GetGreetList4Concurrent(ctx context.Context, ids []int32, fn func(context.Context, int32, ...GreetByIdOption) (model.Greet, error)) ([]model.Greet, error) {
	var wg sync.WaitGroup

	count := len(ids)
	outputs := make([]*concurrentGreetOutput, count)
	wg.Add(count)

	for idx, id := range ids {
		go func(idx int, id int32, wg *sync.WaitGroup) {
			defer wg.Done()
			greet, err := fn(ctx, id)
			outputs[idx] = &concurrentGreetOutput{
				object: greet,
				err:    err,
			}
		}(idx, id, &wg)
	}
	wg.Wait()

	greets := make([]model.Greet, 0, count)
	for _, output := range outputs {
		if output.err == nil {
			greets = append(greets, output.object)
		}
	}
	return greets, nil
}

type concurrentGreetOutput struct {
	object model.Greet
	err    error
}

func (repo greetRepository) UpdateGreetStatus(ctx context.Context, id, status int32) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greetRepository"), zap.String("func", "UpdateGreetStatus"))

	logger.Debug("invoke info", zap.Int32("id", id), zap.Int32("status", status))
	tx := repo.WriteDB(ctx).Model(model.Greet{}).Where("id = ?", id)
	tx = tx.Update("status", status)
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "greetRepository.UpdateGreetStatus")
	}
	key := utilx.CacheKey("greet_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	if err != nil {
		logger.Warn("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))
	}
	return tx.RowsAffected, nil
}

func (repo greetRepository) UpdateGreetCount(ctx context.Context, id, num int32, column string) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greetRepository"), zap.String("func", "UpdateGreetCount"))

	logger.Debug("invoke info", zap.Int32("id", id), zap.Int32("num", num), zap.String("column", column))
	tx := repo.WriteDB(ctx).Model(model.Greet{}).Where("id = ?", id)
	tx = tx.Updates(map[string]interface{}{column: gorm.Expr(column+" + ?", num)})
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "greetRepository.UpdateGreetCount")
	}
	key := utilx.CacheKey("greet_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	if err != nil {
		logger.Warn("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))
	}
	return tx.RowsAffected, nil
}

func (repo greetRepository) DeleteGreetById(ctx context.Context, id int32) (int64, error) {
	logger := ctxzap.Extract(ctx).With(zap.String("layer", "greetRepository"), zap.String("func", "DeleteGreetById"))

	logger.Debug("invoke info", zap.Int32("id", id))
	tx := repo.WriteDB(ctx).Delete(&model.Greet{}, id)
	if tx.Error != nil {
		return 0, errorsx.Wrap(tx.Error, "greetRepository.DeleteGreetById")
	}
	key := utilx.CacheKey("greet_", strconv.Itoa(int(id)))
	reply, err := repo.Redis().Del(ctx, key).Result()
	logger.Debug("Del Cache", zap.String("key", key), zap.Int64("reply", reply), zap.Error(err))

	status := []int{0, 1}
	for _, s := range status {
		key := utilx.CacheKey("greet_ids_", strconv.Itoa(s))
		err := repo.Redis().ZRem(ctx, key, id).Err()
		if err != nil {
			logger.Warn("redis.ZRem", zap.String("key", key), zap.Int32("id", id), zap.Error(err))
		}
	}

	return tx.RowsAffected, nil
}
