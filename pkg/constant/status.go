package constant

import (
	"github.com/imind-lab/micro/status"
)

const (
	GreeterObjectIsEmpty status.Code = iota + 10100
	NameFieldIsEmpty
	CreateGreeterFailed
	FetchGreeterFailed
	FetchGreeterListFailed
	UpdateGreeterFailed
	DeleteGreeterFailed
	StatusIsInvalid
)

var Errors = map[status.Code]string{
	GreeterObjectIsEmpty:   "GreeterObjectIsEmpty",
	NameFieldIsEmpty:       "NameFieldIsEmpty",
	CreateGreeterFailed:    "CreateGreeterFailed",
	FetchGreeterFailed:     "FetchGreeterFailed",
	FetchGreeterListFailed: "FetchGreeterListFailed",
	UpdateGreeterFailed:    "UpdateGreeterFailed",
	DeleteGreeterFailed:    "DeleteGreeterFailed",
	StatusIsInvalid:        "StatusIsInvalid",
}

func init() {
	for k, v := range Errors {
		status.Errors[k] = v
	}
}
