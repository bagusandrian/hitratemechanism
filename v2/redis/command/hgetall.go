package command

import (
	"context"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
	"github.com/redis/rueidis"
)

func (u *usecase) HgetAll(ctx context.Context, req m.RequestCheck) (resp rueidis.RedisResult) {
	cacheData := u.HandlerCache.CacheValidateTrend(ctx, req)
	if cacheData.DataTimeTrend.HasCache {
		resp = u.Redis.DoCache(ctx, u.Redis.B().Hgetall().Key(req.Key).Cache(), req.TTLCache)
	} else {
		resp = u.Redis.Do(ctx, u.Redis.B().Hgetall().Key(req.Key).Build())
	}
	return resp
}
