package model

import (
	"time"

	memoryCache "github.com/patrickmn/go-cache"
	"github.com/redis/rueidis"
)

type (
	HitRateMechanism struct {
		Config      Config
		MemoryCache MemoryCache
		Redis       rueidis.Client
		RedisConfig rueidis.ClientOption
	}
	Config struct {
		MaxMemoryUsage    int
		DefaultExpiration time.Duration
		CleanupInterval   time.Duration
		PrefixKey         string
		LimitTrend        int
	}
	MemoryCache struct {
		Cache *memoryCache.Cache
	}
)
