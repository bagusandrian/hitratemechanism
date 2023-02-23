package HitRateMechanism

import (
	"time"

	cmap "github.com/orcaman/concurrent-map"
	"github.com/redis/go-redis/v9"
)

var (
	Pool                   hrm
	hosts                  cmap.ConcurrentMap
	breakerCmap            cmap.ConcurrentMap
	TTLKeyHitRate          = int64(60)
	ThresholdTTLKeyHitRate = int64(30)
)

// Options configuration options for redis connection
type Options struct {
	MaxIdleConn   int
	MaxActiveConn int
	Timeout       int
	Wait          bool
	Password      string
	Username      string
}

// config used when we need to open new connection automatically
type config struct {
	Address string
	Network string
	Option  Options
}

type hrm struct {
	DBs cmap.ConcurrentMap
	RDb *redis.Conn
}

type hiteRateData struct {
	TTLKeyCheck    int64
	countHitRate   int64
	TTLKeyHitRate  int64
	MaxDateTTL     time.Time
	HaveMaxDateTTL bool
	HighTraffic    bool
	RPS            int64
}

type ReqCustomHitRate struct {
	Config       ConfigCustomHitRate
	Threshold    ThresholdCustomHitrate
	AttributeKey AttributeKeyhitrate
}
type (
	ConfigCustomHitRate struct {
		RedisDBName     string
		ExtendTTLKey    int64
		ParseLayoutTime string
	}
	ThresholdCustomHitrate struct {
		LimitMaxTTL int64
		MaxRPS      int64
	}
	AttributeKeyhitrate struct {
		KeyCheck string
		Prefix   string
	}
)
type RespCustomHitRate struct {
	HighTraffic    bool
	HaveMaxDateTTL bool
	ExtendTTL      bool
	MaxDateTTL     time.Time
	RPS            int64
	Err            error
}
