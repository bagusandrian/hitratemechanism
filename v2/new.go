package v2

import (
	"fmt"
	"time"

	jsoniterpackage "github.com/json-iterator/go"
	memoryCache "github.com/patrickmn/go-cache"
)

type (
	HitRateMechanism struct {
		Config struct {
			DefaultExpiration time.Duration
			CleanupInterval   time.Duration
			PrefixKey         string
			LimitTrend        int
		}
		MemoryCache MemoryCache
	}
	DataTimeTrend struct {
		HasCache     bool            `json:"has_cache"`
		TimeTrend    map[int64]int64 `json:"time_trend"`
		EstimateRPS  int64           `json:"estimate_rps"`
		ThresholdRPS int64           `json:"threshold_rps"`
		LimitTrend   int             `json:"limit_trend"`
	}
	MemoryCache struct {
		Cache *memoryCache.Cache
	}
	RequestCheck struct {
		Key          string
		ThresholdRPS int64
	}
	Response struct {
		ResponseTime   string `json:"response_time"`
		DataTimeTrend  `json:"data_time_trend"`
		Error          error  `json:"error"`
		SuccessMessage string `json:"success_message"`
	}
)

var jsoni jsoniterpackage.API

func New(hrm *HitRateMechanism) {
	jsoni = jsoniterpackage.ConfigCompatibleWithStandardLibrary
	hrm.Config.LimitTrend = 5
	hrm.MemoryCache.Cache = memoryCache.New(hrm.Config.DefaultExpiration, hrm.Config.CleanupInterval)
	return
}

func (hrm *HitRateMechanism) CacheValidateTrend(req RequestCheck) (resp Response) {
	now := time.Now()
	data, err := hrm.cacheGetDataTrend(req.Key)
	if err != nil {
		return Response{
			ResponseTime: time.Since(now).String(),
			Error:        err,
		}
	}
	var successMessage string
	data.ThresholdRPS = req.ThresholdRPS
	data.LimitTrend = hrm.Config.LimitTrend
	if len(data.TimeTrend) < hrm.Config.LimitTrend {
		for i := int64(0); i <= int64(hrm.Config.LimitTrend-1); i++ {
			if _, exist := data.TimeTrend[i]; exist {
				continue
			} else {
				data.TimeTrend[i] = time.Now().UnixMilli()
				break
			}
		}
		hrm.cacheSetDataTrend(req.Key, data)
	} else if !data.HasCache {
		for i := int64(0); i <= int64(4); i++ {
			if i <= 3 {
				data.TimeTrend[i] = data.TimeTrend[i+1]
			} else {
				data.TimeTrend[i] = time.Now().UnixMilli()
			}
		}
		timeAvg := (((data.TimeTrend[4] - data.TimeTrend[3]) + (data.TimeTrend[3] - data.TimeTrend[2]) + (data.TimeTrend[2] - data.TimeTrend[1]) + (data.TimeTrend[1] - data.TimeTrend[0])) / 4)
		data.EstimateRPS = 1000 / timeAvg
		if data.EstimateRPS > req.ThresholdRPS {
			data.HasCache = true
			successMessage = fmt.Sprintf("no need set again! data.HasCache: %t\n", data.HasCache)
		}
		hrm.cacheSetDataTrend(req.Key, data)

	} else if data.HasCache {
		timeAvg := (((data.TimeTrend[4] - data.TimeTrend[3]) + (data.TimeTrend[3] - data.TimeTrend[2]) + (data.TimeTrend[2] - data.TimeTrend[1]) + (data.TimeTrend[1] - data.TimeTrend[0])) / 4)
		data.EstimateRPS = 1000 / timeAvg
		successMessage = fmt.Sprintf("no need set again! data.HasCache: %t\n", data.HasCache)
	}

	// log.Printf("no need set again! data.HasCache: %t\n", data.HasCache)
	return Response{
		ResponseTime:   time.Since(now).String(),
		SuccessMessage: successMessage,
		Error:          nil,
		DataTimeTrend:  data,
	}
}

func (hrm *HitRateMechanism) cacheGetDataTrend(key string) (result DataTimeTrend, err error) {
	result = DataTimeTrend{}
	result.TimeTrend = make(map[int64]int64)
	item, found := hrm.MemoryCache.Cache.Get(hrm.generateKey(key))
	if !found {
		return result, nil
	}
	err = jsoni.Unmarshal(item.([]byte), &result)
	if err != nil {
		return result, err
	}
	return result, nil

}

func (hrm *HitRateMechanism) cacheSetDataTrend(key string, value DataTimeTrend) {
	v, _ := jsoni.Marshal(value)
	hrm.MemoryCache.Cache.Set(hrm.generateKey(key), v, hrm.Config.DefaultExpiration)
	return
}

func (hrm *HitRateMechanism) generateKey(key string) string {
	return fmt.Sprintf("%s:%s", hrm.Config.PrefixKey, key)
}
