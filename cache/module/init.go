package module

import (
	goCache "github.com/TwiN/gocache/v2"
	"github.com/bagusandrian/hitratemechanism/cache"
	m "github.com/bagusandrian/hitratemechanism/model"
	jsoniterpackage "github.com/json-iterator/go"
)

type usecase struct {
	Conf    m.Config
	jsoni   jsoniterpackage.API
	GoCache *goCache.Cache
}

func New(hrm *m.HitRateMechanism) cache.Handler {
	jsoni := jsoniterpackage.ConfigCompatibleWithStandardLibrary
	return &usecase{
		Conf:    hrm.Config,
		jsoni:   jsoni,
		GoCache: goCache.NewCache().WithMaxMemoryUsage(hrm.Config.MaxMemoryUsage).WithEvictionPolicy(goCache.LeastRecentlyUsed),
	}
}
