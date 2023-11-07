package model

import (
	"time"

	"github.com/redis/rueidis"
)

type (
	HitRateMechanism struct {
		Config      Config
		Redis       rueidis.Client
		RedisConfig rueidis.ClientOption
	}
	Config struct {
		MaxMemoryUsage    int
		DefaultExpiration time.Duration
		PrefixKey         string
		LimitTrend        int
	}
)
