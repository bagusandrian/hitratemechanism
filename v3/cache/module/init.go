package module

import (
	"log"

	goCache "github.com/TwiN/gocache/v2"
	"github.com/bagusandrian/hitratemechanism/v3/cache"
	m "github.com/bagusandrian/hitratemechanism/v3/model"
	"github.com/redis/rueidis"
)

type usecase struct {
	Conf        m.Config
	GoCache     *goCache.Cache
	RedisClient *rueidis.Client
}

func New(hrm *m.HitRateMechanism) cache.Handler {
	var MaxMemoryUsage int
	if hrm.Config.MaxMemoryUsage > 0 {
		MaxMemoryUsage = hrm.Config.MaxMemoryUsage
	} else {
		MaxMemoryUsage = 128 * (1 << 20)
	}
	var redisClient rueidis.Client
	if len(hrm.RedisConfig.InitAddress) > 0 {
		var err error
		redisClient, err = rueidis.NewClient(hrm.RedisConfig)
		if err != nil {
			log.Panic(err)
		}
	}
	return &usecase{
		Conf:        hrm.Config,
		GoCache:     goCache.NewCache().WithMaxMemoryUsage(MaxMemoryUsage).WithEvictionPolicy(goCache.LeastRecentlyUsed),
		RedisClient: &redisClient,
	}
}
