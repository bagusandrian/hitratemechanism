package model

import (
	"context"
	"log"
	"time"

	"github.com/redis/rueidis"
)

type (
	RequestCheck struct {
		Key          string
		ThresholdRPS int64
		TTLCache     time.Duration
	}
	Response struct {
		ResponseTime   string `json:"response_time"`
		DataTimeTrend  `json:"data_time_trend"`
		Error          error  `json:"error"`
		SuccessMessage string `json:"success_message"`
		Ctx            context.Context
		ClientRedis    rueidis.Client
		Req            RequestCheck
	}
	DataTimeTrend struct {
		ReachThresholdRPS bool            `json:"reach_threshold_rps"`
		TimeTrend         map[int64]int64 `json:"time_trend"`
		EstimateRPS       int64           `json:"estimate_rps"`
		ThresholdRPS      int64           `json:"threshold_rps"`
		LimitTrend        int             `json:"limit_trend"`
	}
)

func (r Response) GetDataRedis() Response {
	if r.ClientRedis == nil {
		log.Panic("redis client not exist")
	}
	return r
}
func (r Response) HgetAll(key string) (resp rueidis.RedisResult) {
	if r.ReachThresholdRPS {
		resp = r.ClientRedis.DoCache(r.Ctx, r.ClientRedis.B().Get().Key(key).Cache(), time.Minute)
	} else {
		resp = r.ClientRedis.Do(r.Ctx, r.ClientRedis.B().Get().Key(key).Build())
	}
	return resp
}
