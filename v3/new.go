package hitratemechanism

import (
	"github.com/bagusandrian/hitratemechanism/v3/cache"
	"github.com/bagusandrian/hitratemechanism/v3/model"
)

type Usecase struct {
	// handler
	HandlerCache cache.Handler
}

func New(hrm *model.HitRateMechanism) *Usecase {
	return &Usecase{}
}
