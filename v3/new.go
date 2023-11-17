package hitratemechanism

import (
	"github.com/bagusandrian/hitratemechanism/v3/cache"
	uCache "github.com/bagusandrian/hitratemechanism/v3/cache/module"
	"github.com/bagusandrian/hitratemechanism/v3/model"
)

type Usecase struct {
	// handler
	HandlerCache cache.Handler
}

func New(hrm *model.HitRateMechanism) *Usecase {
	return &Usecase{
		HandlerCache: uCache.New(hrm),
	}
}
