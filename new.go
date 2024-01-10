package hitratemechanism

import (
	"github.com/bagusandrian/hitratemechanism/cache"
	uCache "github.com/bagusandrian/hitratemechanism/cache/module"
	"github.com/bagusandrian/hitratemechanism/model"
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
