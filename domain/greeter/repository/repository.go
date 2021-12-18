package repository

import (
	"context"
	"github.com/imind-lab/greeter/domain/greeter/repository/model"
)

type GreeterRepository interface {
	CreateGreeter(ctx context.Context, m model.Greeter) (model.Greeter, error)

	GetGreeterById(ctx context.Context, id int32, opt ...GreeterByIdOption) (model.Greeter, error)
	FindGreeterById(ctx context.Context, id int32) (model.Greeter, error)
	GetGreeterList(ctx context.Context, status, lastId, pageSize, page int32) ([]model.Greeter, int, error)

	UpdateGreeterStatus(ctx context.Context, id, status int32) (int64, error)
	UpdateGreeterCount(ctx context.Context, id, num int32, column string) (int64, error)

	DeleteGreeterById(ctx context.Context, id int32) (int64, error)
}
