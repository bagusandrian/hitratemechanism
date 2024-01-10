package module

import (
	goCache "github.com/TwiN/gocache/v2"
	"github.com/bagusandrian/hitratemechanism/cache"
	m "github.com/bagusandrian/hitratemechanism/model"
)

type usecase struct {
	Conf    m.Config
	GoCache *goCache.Cache
}

func New(hrm *m.HitRateMechanism) cache.Handler {
	var MaxMemoryUsage int
	if hrm.Config.MaxMemoryUsage > 0 {
		MaxMemoryUsage = hrm.Config.MaxMemoryUsage
	} else {
		MaxMemoryUsage = 128 * (1 << 20)
	}

	return &usecase{
		Conf:    hrm.Config,
		GoCache: goCache.NewCache().WithMaxMemoryUsage(MaxMemoryUsage).WithEvictionPolicy(goCache.LeastRecentlyUsed),
	}
}
