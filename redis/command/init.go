package command

import (
	"log"

	hCache "github.com/bagusandrian/hitratemechanism/cache"
	uCache "github.com/bagusandrian/hitratemechanism/cache/module"
	m "github.com/bagusandrian/hitratemechanism/model"
	"github.com/bagusandrian/hitratemechanism/redis"
	"github.com/redis/rueidis"
)

type usecase struct {
	Redis        rueidis.Client
	Conf         m.Config
	HandlerCache hCache.Handler
}

func New(hrm *m.HitRateMechanism) redis.Handler {
	client, err := rueidis.NewClient(hrm.RedisConfig)
	if err != nil {
		log.Panic(err)
	}
	return &usecase{
		Redis:        client,
		Conf:         hrm.Config,
		HandlerCache: uCache.New(hrm),
	}
}
