package module

import (
	goCache "github.com/TwiN/gocache/v2"
	"github.com/bagusandrian/hitratemechanism/v2/cache"
	m "github.com/bagusandrian/hitratemechanism/v2/model"
	jsoniterpackage "github.com/json-iterator/go"
	memoryCache "github.com/patrickmn/go-cache"
)

type usecase struct {
	Cache   *memoryCache.Cache
	Conf    m.Config
	jsoni   jsoniterpackage.API
	GoCache *goCache.Cache
}

func New(hrm *m.HitRateMechanism) cache.Handler {
	jsoni := jsoniterpackage.ConfigCompatibleWithStandardLibrary
	return &usecase{
		Cache:   memoryCache.New(hrm.Config.DefaultExpiration, hrm.Config.CleanupInterval),
		Conf:    hrm.Config,
		jsoni:   jsoni,
		GoCache: goCache.NewCache().WithMaxMemoryUsage(hrm.Config.MaxMemoryUsage).WithEvictionPolicy(goCache.LeastRecentlyUsed),
	}
}
