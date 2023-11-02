package v2

import (
	hCache "github.com/bagusandrian/hitratemechanism/v2/cache"
	uCache "github.com/bagusandrian/hitratemechanism/v2/cache/module"
	"github.com/bagusandrian/hitratemechanism/v2/model"
)

var usecaseHRM usecase

type usecase struct {
	// handler
	handlerCache hCache.Handler
}

func New(hrm *model.HitRateMechanism) {
	usecaseHRM = usecase{
		handlerCache: uCache.New(hrm),
	}
}
