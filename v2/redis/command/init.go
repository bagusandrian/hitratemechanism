package command

import (
	"log"

	hCache "github.com/bagusandrian/hitratemechanism/v2/cache"
	uCache "github.com/bagusandrian/hitratemechanism/v2/cache/module"
	m "github.com/bagusandrian/hitratemechanism/v2/model"
	"github.com/bagusandrian/hitratemechanism/v2/redis"
	"github.com/redis/rueidis"
	koredis "github.com/tokopedia/koredis/v8"
)

type usecase struct {
	Redis        rueidis.Client
	Conf         m.Config
	HandlerCache hCache.Handler
	Koredis      *koredis.ClusterClient
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
