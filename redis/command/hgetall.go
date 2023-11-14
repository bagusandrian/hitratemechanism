package command

import (
	"context"

	m "github.com/bagusandrian/hitratemechanism/model"
	"github.com/redis/rueidis"
)

func (u *usecase) HgetAll(ctx context.Context, req m.RequestCheck) (resp rueidis.RedisResult, cacheDebug m.Response) {
	cacheData := u.HandlerCache.CacheValidateTrend(ctx, req)
	if cacheData.DataTimeTrend.ReachThresholdRPS {
		resp = u.Redis.DoCache(ctx, u.Redis.B().Hgetall().Key(req.Key).Cache(), req.TTLCache)
	} else {
		resp = u.Redis.Do(ctx, u.Redis.B().Hgetall().Key(req.Key).Build())
	}
	return resp, cacheData
}