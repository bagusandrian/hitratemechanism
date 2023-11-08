package command

import (
	"context"
	"log"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
	"github.com/redis/rueidis"
)

func (u *usecase) HgetAll(ctx context.Context, req m.RequestCheck) (resp rueidis.RedisResult, cacheDebug m.Response) {
	cacheData := u.HandlerCache.CacheValidateTrend(ctx, req)
	log.Printf("return cacheData: %+v\n", cacheData)
	if cacheData.DataTimeTrend.ReachThresholdRPS {
		resp = u.Redis.DoCache(ctx, u.Redis.B().Hgetall().Key(req.Key).Cache(), req.TTLCache)
	} else {
		resp = u.Redis.Do(ctx, u.Redis.B().Hgetall().Key(req.Key).Build())
	}
	cacheData.DataTimeTrend.HasCache = resp.IsCacheHit()
	return resp, cacheData
}
