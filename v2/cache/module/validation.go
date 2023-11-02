package module

import (
	"context"
	"fmt"
	"time"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
)

func (u *usecase) CacheValidateTrend(ctx context.Context, req m.RequestCheck) (resp m.Response) {
	now := time.Now()
	data, err := u.cacheGetDataTrend(ctx, req.Key)
	if err != nil {
		return m.Response{
			ResponseTime: time.Since(now).String(),
			Error:        err,
		}
	}
	if len(data.TimeTrend) > 1 {
		data.EstimateRPS = u.calculateRPS(data.TimeTrend)
	}
	var successMessage string
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
		u.cacheSetDataTrend(ctx, req.Key, data)
	} else if !data.HasCache {
		for i := int64(0); i <= int64(4); i++ {
			if i <= 3 {
				data.TimeTrend[i] = data.TimeTrend[i+1]
			} else {
				data.TimeTrend[i] = time.Now().UnixMilli()
			}
		}
		if data.EstimateRPS > req.ThresholdRPS {
			data.HasCache = true
			successMessage = fmt.Sprintf("no need set again! data.HasCache: %t\n", data.HasCache)
		}
		u.cacheSetDataTrend(ctx, req.Key, data)

	} else if data.HasCache {
		successMessage = fmt.Sprintf("no need set again! data.HasCache: %t\n", data.HasCache)
	}

	// log.Printf("no need set again! data.HasCache: %t\n", data.HasCache)
	return m.Response{
		ResponseTime:   time.Since(now).String(),
		SuccessMessage: successMessage,
		Error:          nil,
		DataTimeTrend:  data,
	}
}

func (u *usecase) cacheGetDataTrend(ctx context.Context, key string) (result m.DataTimeTrend, err error) {
	result = m.DataTimeTrend{}
	result.TimeTrend = make(map[int64]int64)
	item, found := u.Cache.Get(u.generateKey(ctx, key))
	if !found {
		return result, nil
	}
	err = u.jsoni.Unmarshal(item.([]byte), &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (u *usecase) cacheSetDataTrend(ctx context.Context, key string, value m.DataTimeTrend) {
	v, _ := u.jsoni.Marshal(value)
	u.Cache.Set(u.generateKey(ctx, key), v, u.Conf.DefaultExpiration)
	return
}

func (u *usecase) calculateRPS(timeTrend map[int64]int64) int64 {
	len := int64(len(timeTrend))
	result := 1000 / ((timeTrend[(len-1)] - timeTrend[0]) / len)
	return result
}

func (u *usecase) generateKey(ctx context.Context, key string) string {
	return fmt.Sprintf("%s:%s", u.Conf.PrefixKey, key)
}
