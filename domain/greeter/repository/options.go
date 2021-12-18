package repository

import (
	"time"
)

type GreeterByIdOptions struct {
	RandExpire time.Duration
}

func NewGreeterByIdOptions(randExpire time.Duration) *GreeterByIdOptions {
	return &GreeterByIdOptions{RandExpire: randExpire}
}

type GreeterByIdOption func(*GreeterByIdOptions)

func GreeterByIdRandExpire(expire time.Duration) GreeterByIdOption {
	return func(o *GreeterByIdOptions) {
		o.RandExpire = expire
	}
}
