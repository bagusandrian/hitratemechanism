package redis

import (
	"context"

	m "github.com/bagusandrian/hitratemechanism/model"
	"github.com/redis/rueidis"
)

//go:generate mockery --name=Handler --filename=mock_handler.go --inpackage
type Handler interface {
	HgetAll(ctx context.Context, req m.RequestCheck) (resp rueidis.RedisResult, cacheDebug m.Response)
	Get(ctx context.Context, req m.RequestCheck) (resp rueidis.RedisResult, cacheDebug m.Response)
}
