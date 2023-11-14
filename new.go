package hitratemechanism

import (
	"github.com/bagusandrian/hitratemechanism/model"
	hRedis "github.com/bagusandrian/hitratemechanism/redis"
	uRedis "github.com/bagusandrian/hitratemechanism/redis/command"
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
