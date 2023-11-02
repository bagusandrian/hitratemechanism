package v2

import (
	"github.com/bagusandrian/hitratemechanism/v2/model"
	hRedis "github.com/bagusandrian/hitratemechanism/v2/redis"
	uRedis "github.com/bagusandrian/hitratemechanism/v2/redis/command"
)

var usecaseHRM Usecase

type Usecase struct {
	// handler
	HandlerRedis hRedis.Handler
}

func New(hrm *model.HitRateMechanism) *Usecase {
	return &Usecase{
		HandlerRedis: uRedis.New(hrm),
	}
}
