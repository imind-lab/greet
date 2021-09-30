package util

import (
	"github.com/imind-lab/greet/pkg/constant"
	"github.com/imind-lab/micro/util"
)

func CacheKey(keys ...string) string {
	return constant.CachePrefix + util.AppendString(keys...)
}
