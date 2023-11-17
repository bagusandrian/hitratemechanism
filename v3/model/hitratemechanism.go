package model

import (
	"time"

	"github.com/redis/rueidis"
)

type (
	// config for setup HRM
	HitRateMechanism struct {
		Config      Config
		Redis       rueidis.Client
		RedisConfig rueidis.ClientOption
	}
	Config struct {
		// MaxMemoryUsage u can set max memory usage, by default will set as 128Mb
		MaxMemoryUsage int
		// use for default expired of cache
		DefaultExpiration time.Duration
		// prefix for local cache colecting for calculate RPS
		PrefixKey string
		// limit trend for calculate RPS, suggestion set to > 2.
		LimitTrend int
	}
)
