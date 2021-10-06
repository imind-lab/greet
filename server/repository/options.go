package repository

import (
	"time"
)

type GreeterByIdOptions struct {
	randExpire time.Duration
}

type GreeterByIdOption func(*GreeterByIdOptions)

func GreeterByIdRandExpire(expire time.Duration) GreeterByIdOption {
	return func(o *GreeterByIdOptions) {
		o.randExpire = expire
	}
}
