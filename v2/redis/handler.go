package redis

import (
	"context"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
	"github.com/redis/rueidis"
)

type Handler interface {
	HgetAll(ctx context.Context, req m.RequestCheck) (resp rueidis.RedisResult, cacheDebug m.Response)
}
