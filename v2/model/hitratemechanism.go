package model

import (
	"time"

	memoryCache "github.com/patrickmn/go-cache"
)

type (
	HitRateMechanism struct {
		Config      Config
		MemoryCache MemoryCache
	}
	Config struct {
		DefaultExpiration time.Duration
		CleanupInterval   time.Duration
		PrefixKey         string
		LimitTrend        int
	}
	MemoryCache struct {
		Cache *memoryCache.Cache
	}
)
