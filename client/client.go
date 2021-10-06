/**
 *  MindLab
 *
 *  Create by songli on 2021/09/30
 *  Copyright © 2021 imind.tech All rights reserved.
 */

package client

import (
	"context"
	"strconv"
)

var greeters map[string]*greeterClient

var opts = Options{Name: "greeter", Tls: true}

func init() {
	greeters = make(map[string]*greeterClient)
}

type Options struct {
	Name string
	Tls  bool
}

func New(ctx context.Context, opt ...Option) (*greeterClient, error) {
	for _, o := range opt {
		o(&opts)
	}
	key := opts.Name + strconv.FormatBool(opts.Tls)
	greeterClient, ok := greeters[key]
	if !ok {
		greeterClient, err := NewGreeterClient(ctx, opts.Name, opts.Tls)
		if err == nil {
			greeters[key] = greeterClient
		}
		return greeterClient, err
	}
	return greeterClient, nil
}

type Option func(*Options)

func Name(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

func Tls(tls bool) Option {
	return func(o *Options) {
		o.Tls = tls
	}
}

func Close() {
	for _, client := range greeters {
		client.Close()
	}
}
