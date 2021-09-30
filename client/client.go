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

var greets map[string]*greetClient

var opts = Options{Name: "greet", Tls: true}

func init() {
	greets = make(map[string]*greetClient)
}

type Options struct {
	Name string
	Tls  bool
}

func New(ctx context.Context, opt ...Option) (*greetClient, error) {
	for _, o := range opt {
		o(&opts)
	}
	key := opts.Name + strconv.FormatBool(opts.Tls)
	greetClient, ok := greets[key]
	if !ok {
		greetClient, err := NewGreetClient(ctx, opts.Name, opts.Tls)
		if err == nil {
			greets[key] = greetClient
		}
		return greetClient, err
	}
	return greetClient, nil
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
	for _, client := range greets {
		client.Close()
	}
}
