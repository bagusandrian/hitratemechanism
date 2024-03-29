package module

import (
	"context"
	"time"

	m "github.com/bagusandrian/hitratemechanism/model"
)

func (u *usecase) CacheValidateTrend(ctx context.Context, req m.RequestCheck) (resp m.Response) {
	now := time.Now()
	dataKeys := make(map[string]m.DataTimeTrend)
	for _, key := range req.Keys {
		data := u.cacheGetDataTrend(ctx, key)
		if data.ReachThresholdRPS {
			dataKeys[key] = data
			continue
		}
		if len(data.TimeTrend) > 2 {
			data.EstimateRPS = u.calculateRPS(data.TimeTrend)
		}
		data.ThresholdRPS = req.ThresholdRPS
		data.LimitTrend = u.Conf.LimitTrend
		if len(data.TimeTrend) < u.Conf.LimitTrend {
			for i := int64(0); i <= int64(u.Conf.LimitTrend-1); i++ {
				if _, exist := data.TimeTrend[i]; exist {
					continue
				} else {
					data.TimeTrend[i] = time.Now().UnixMilli()
					break
				}
			}
			u.cacheSetDataTrend(ctx, key, req.TTLCache, data)
		} else {
			if !data.ReachThresholdRPS {
				for i := int64(0); i <= int64(data.LimitTrend-1); i++ {
					if i <= 3 {
						data.TimeTrend[i] = data.TimeTrend[i+1]
					} else {
						data.TimeTrend[i] = time.Now().UnixMilli()
					}
				}
				if data.EstimateRPS > req.ThresholdRPS {
					data.ReachThresholdRPS = true
				}
				u.cacheSetDataTrend(ctx, key, req.TTLCache, data)

			}
		}
		dataKeys[key] = data
	}

	return m.Response{
		ResponseTime:   time.Since(now).String(),
		SuccessMessage: "",
		Error:          nil,
		DataKeys:       dataKeys,
	}
}

func (u *usecase) cacheGetDataTrend(ctx context.Context, key string) (result m.DataTimeTrend) {
	result = m.DataTimeTrend{}
	result.TimeTrend = make(map[int64]int64)
	item, found := u.GoCache.Get(u.generateKey(ctx, key))
	if !found {
		return result
	}
	result = item.(m.DataTimeTrend)
	return result
}

func (u *usecase) cacheSetDataTrend(ctx context.Context, key string, ttlCache time.Duration, value m.DataTimeTrend) {
	var TTL time.Duration
	if ttlCache > 0 {
		TTL = ttlCache
	} else {
		TTL = u.Conf.DefaultExpiration
	}
	u.GoCache.SetWithTTL(u.generateKey(ctx, key), value, TTL)
}

func (u *usecase) calculateRPS(timeTrend map[int64]int64) int64 {
	var fristTime, lastTime int64
	len := int64(len(timeTrend))
	if len < 2 {
		return 0
	}
	if v, ok := timeTrend[0]; ok {
		fristTime = v
	} else {
		return 0
	}
	if v, ok := timeTrend[len-1]; ok {
		lastTime = v
	} else {
		return 0
	}
	if fristTime == lastTime {
		return 0
	}
	avg := (lastTime - fristTime) / len
	if avg <= 0 {
		return 0
	}
	result := 1000 / avg
	return result
}

func (u *usecase) generateKey(ctx context.Context, key string) string {
	return u.Conf.PrefixKey + ":" + key
}
