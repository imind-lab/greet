package repository

import (
	"time"
)

type GreetByIdOptions struct {
	randExpire time.Duration
}

type GreetByIdOption func(*GreetByIdOptions)

func GreetByIdRandExpire(expire time.Duration) GreetByIdOption {
	return func(o *GreetByIdOptions) {
		o.randExpire = expire
	}
}
