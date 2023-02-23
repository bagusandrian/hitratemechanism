package HitRateMechanism

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eapache/go-resiliency/breaker"
	cmap "github.com/orcaman/concurrent-map"

	"github.com/redis/go-redis/v9"
)

func init() {
	Pool.RDb = &redis.Conn{}
	Pool.DBs = cmap.New()
	hosts = cmap.New()
	breakerCmap = cmap.New()
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func New(hostName, hostAddress, network string, customOption ...Options) {
	opt := Options{
		MaxIdleConn:   10,
		MaxActiveConn: 1000,
		Timeout:       1,
		Wait:          false,
	}
	if len(customOption) > 0 {
		opt = customOption[0]
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         hostAddress,
		Password:     opt.Password,
		Username:     opt.Username,
		MaxIdleConns: opt.MaxIdleConn,
		DialTimeout:  time.Duration(opt.Timeout) * time.Second,
		Dialer: redis.NewDialer(&redis.Options{Addr: hostAddress,
			Password: opt.Password,
			Username: opt.Username}),
	})
	Pool.RDb = rdb.Conn()
	Pool.DBs.Set(hostName, &rdb)
	cfg := config{
		Address: hostAddress,
		Network: network,
		Option: Options{
			MaxActiveConn: opt.MaxActiveConn,
			MaxIdleConn:   opt.MaxIdleConn,
			Timeout:       opt.Timeout,
			Wait:          true,
			Username:      opt.Username,
			Password:      opt.Password,
		},
	}
	hosts.Set(hostName, cfg)
}

func (r *hrm) getConnection(ctx context.Context, dbname string) redis.Conn {
	var rdsConn redis.Conn

	circuitbreaker, cbOk := breakerCmap.Get(dbname)
	if !cbOk {
		circuitbreaker = breaker.New(10, 2, 10*time.Second)
		breakerCmap.Set(dbname, circuitbreaker)
	}

	cb := circuitbreaker.(*breaker.Breaker)
	cbResult := cb.Run(func() error {
		rdsConn = *Pool.RDb
		resp := rdsConn.Ping(ctx)
		if resp.Err() == nil {
			return nil
		} else {
			log.Println("[redis] ping error:", resp.Err())
			rdsConn.Close() // just in case
		}

		log.Printf("[redis] %s - bad connection, closing and opening new one\n", dbname)
		hosttemp, ok := hosts.Get(dbname)
		if !ok {
			return fmt.Errorf("[redis] %s - failed to retrieve connection info", dbname)
		}

		host := hosttemp.(config)

		Pool.DBs.Remove(dbname)
		// create new connection
		rdb := redis.NewClient(&redis.Options{
			Addr:         host.Address,
			Password:     host.Option.Password,
			Username:     host.Option.Username,
			MaxIdleConns: host.Option.MaxIdleConn,
			DialTimeout:  time.Duration(host.Option.Timeout) * time.Second,
			Dialer: redis.NewDialer(&redis.Options{Addr: host.Address,
				Password: host.Option.Password,
				Username: host.Option.Username}),
		})
		Pool.DBs.Set(dbname, &rdb)

		rdsConn = *rdb.Conn()
		resp = rdsConn.Ping(ctx)
		if resp.Err() == nil {
			return nil
		} else {
			log.Println("[redis] ping error:", resp.Err())
			rdsConn.Close() // just in case
		}
		return nil
	})

	if cbResult == breaker.ErrBreakerOpen {
		log.Printf("[redis] redis circuitbreaker open, retrying later.. - %s\n", dbname)
	}

	return rdsConn
}

// GetConnection return redigo.Conn which enable developers to do redis command
// that is not commonly used (special case command, one that don't have wrapper function in this package. e.g: Exists)
func (r *hrm) GetConnection(ctx context.Context, dbname string) redis.Conn {
	return r.getConnection(ctx, dbname)
}
func (r *hrm) HgetAll(ctx context.Context, dbname, key string) (map[string]string, error) {
	conn := r.getConnection(ctx, dbname)
	check := conn.Ping(ctx)
	if check.Err() != nil {
		return nil, fmt.Errorf("failed get connection")
	}
	defer conn.Close()
	resp := conn.HGetAll(ctx, key)
	if resp.Err() != nil {
		return nil, fmt.Errorf("got error on HGETALL: %+v\n", resp.Err())
	}
	result, err := resp.Result()
	return result, err
}

func (r *hrm) HmsetWithExpMultiple(ctx context.Context, dbname string, data map[string]map[string]interface{}, expired int) (err error) {
	conn := r.getConnection(ctx, dbname)
	check := conn.Ping(ctx)
	if check.Err() != nil {
		return fmt.Errorf("failed get connection")
	}
	defer conn.Close()
	pipeline := conn.Pipeline()
	for key, hashmap := range data {
		for f, v := range hashmap {
			pipeline.HMSet(ctx, key, f, v)
		}
		pipeline.Expire(ctx, key, time.Second*time.Duration(expired))
	}
	_, err = pipeline.Exec(ctx)

	return
}

func (r *hrm) CustomHitRate(ctx context.Context, req ReqCustomHitRate) RespCustomHitRate {
	req, err := validateReqCustomHitRate(ctx, req)
	if err != nil {
		return RespCustomHitRate{
			Err: err,
		}
	}
	keyHitrate := fmt.Sprintf("%s-%s", req.AttributeKey.Prefix, req.AttributeKey.KeyCheck)
	conn := r.getConnection(ctx, req.Config.RedisDBName)
	check := conn.Ping(ctx)
	if check.Err() != nil {
		return RespCustomHitRate{
			Err: err,
		}
	}
	defer conn.Close()
	hitRateData, err := r.hitRateGetData(ctx, conn, req.AttributeKey.KeyCheck, keyHitrate, req.Config.ParseLayoutTime)
	if err != nil {
		return RespCustomHitRate{
			Err: err,
		}
	}
	type cmdAddTTl struct {
		command string
		key     string
		expire  int64
	}
	// cmds := []cmdAddTTl{}
	pipeline := conn.Pipeline()
	// if checker key hitrate dont have ttl, will set expire for 1 minute
	// or key hitrate under 30 seconds, will set expire for 1 minute
	if hitRateData.TTLKeyHitRate > int64(-3) && hitRateData.TTLKeyHitRate <= ThresholdTTLKeyHitRate {
		if hitRateData.HaveMaxDateTTL {
			maxTTL := int64(time.Until(hitRateData.MaxDateTTL) / time.Second)
			if maxTTL < int64(60) {
				pipeline.Expire(ctx, keyHitrate, time.Second*time.Duration(maxTTL))
			}
		} else {
			pipeline.Expire(ctx, keyHitrate, time.Second*time.Duration(TTLKeyHitRate))
			// cmds = append(cmds, cmdAddTTl{command: "EXPIRE", key: keyHitrate, expire: TTLKeyHitRate})
		}
	}
	newTTL := calculateNewTTL(hitRateData.TTLKeyCheck, req.Config.ExtendTTLKey, req.Threshold.LimitMaxTTL, hitRateData.MaxDateTTL)
	resp := RespCustomHitRate{}
	if hitRateData.RPS >= req.Threshold.MaxRPS {
		resp.HighTraffic = true
		if newTTL > 0 {
			// add ttl redis key target
			pipeline.Expire(ctx, req.AttributeKey.KeyCheck, time.Second*time.Duration(newTTL))
			// cmds = append(cmds, cmdAddTTl{command: "EXPIRE", key: req.AttributeKey.KeyCheck, expire: newTTL})
			resp.ExtendTTL = true
		}
	}
	if pipeline.Len() > 0 {
		_, err := pipeline.Exec(ctx)
		if err != nil {
			return RespCustomHitRate{
				Err: fmt.Errorf("pipeline.Exec err:%+v", err),
			}
		}
	}

	if !hitRateData.MaxDateTTL.IsZero() {
		resp.HaveMaxDateTTL = true
		resp.MaxDateTTL = hitRateData.MaxDateTTL
	}
	resp.RPS = hitRateData.RPS
	return resp
}

func (r *hrm) hitRateGetData(ctx context.Context, conn redis.Conn, keyCheck, keyHitrate, parseLayoutTime string) (result hiteRateData, err error) {
	pipeline := conn.Pipeline()
	cmdsDuration := map[string]*redis.DurationCmd{}
	cmdsInt := map[string]*redis.IntCmd{}
	cmdsSlice := map[string]*redis.SliceCmd{}
	cmdsDuration["keycheck"] = pipeline.TTL(ctx, keyCheck)
	cmdsDuration["keyhitrate"] = pipeline.TTL(ctx, keyHitrate)
	cmdsInt["keyhitrate"] = pipeline.HIncrBy(ctx, keyHitrate, "count", 1)
	cmdsSlice["kehitrate_end_time"] = pipeline.HMGet(ctx, keyHitrate, "end_time")
	_, err = pipeline.Exec(ctx)
	if err != nil {
		return result, err
	}
	// mapping TTLKeyCheck
	if cmdsDuration["keycheck"].Err() != nil {
		return result, cmdsDuration["keycheck"].Err()
	}
	result.TTLKeyCheck = int64(cmdsDuration["keycheck"].Val().Seconds())
	// mapping TTLKeyHitRate
	if cmdsDuration["keyhitrate"].Err() != nil {
		return result, cmdsDuration["keyhitrate"].Err()
	}
	result.TTLKeyHitRate = int64(cmdsDuration["keyhitrate"].Val().Seconds())
	// mapping countHitRate
	if cmdsInt["keyhitrate"].Err() != nil {
		return result, cmdsInt["keyhitrate"].Err()
	}
	result.countHitRate = cmdsInt["keyhitrate"].Val()
	// mapping keyhitrate_end_time
	if cmdsSlice["kehitrate_end_time"].Err() != nil {
		return result, cmdsSlice["kehitrate_end_time"].Err()
	}
	temp := cmdsSlice["kehitrate_end_time"].Val()
	if len(temp) > 0 {
		dateFormat, err := time.Parse(parseLayoutTime, fmt.Sprintf("%v", temp[0]))
		if err == nil {
			result.MaxDateTTL = dateFormat
			result.HaveMaxDateTTL = true
		}
	}
	// mapping RPS
	result.RPS = calculateRPS(result.countHitRate)
	return
}

func (r *hrm) SetMaxTTLChecker(ctx context.Context, dbname, prefix, keyCheck string, endTime time.Time) error {
	key := fmt.Sprintf("%s-%s", prefix, keyCheck)
	conn := r.getConnection(ctx, dbname)
	check := conn.Ping(ctx)
	if check.Err() != nil {
		return fmt.Errorf("failed get connection")
	}
	defer conn.Close()
	pipeline := conn.Pipeline()
	maxTTL := int64(time.Until(endTime) / time.Second)
	if maxTTL < int64(60) {
		pipeline.Expire(ctx, key, time.Second*time.Duration(maxTTL))
	}
	pipeline.HMSet(ctx, key, "end_time", endTime.Format("2006-01-02 15:04:05 Z0700 MST"))
	_, err := pipeline.Exec(ctx)
	return err
}

func calculateRPS(countHit int64) (rps int64) {
	rps = int64(countHit / 60)
	return
}

func calculateNewTTL(TTLKeyCHeck, extendTTL, limitTTL int64, dateMax time.Time) (newTTL int64) {
	maxTTL := int64(limitTTL)
	if !dateMax.IsZero() {
		maxTTL = int64(time.Until(dateMax) / time.Second)
	}
	newTTL = extendTTL + TTLKeyCHeck
	if newTTL > limitTTL {
		newTTL = 0
		return
	}
	if newTTL > maxTTL {
		newTTL = maxTTL
		return
	}
	return
}

func validateReqCustomHitRate(ctx context.Context, req ReqCustomHitRate) (ReqCustomHitRate, error) {
	if req.Config.RedisDBName == "" {
		return req, fmt.Errorf("empty redisDBName config")
	}
	if req.Config.ParseLayoutTime == "" {
		return req, fmt.Errorf("empty layout format for parse time")
	}
	if req.Threshold.MaxRPS == 0 {
		return req, fmt.Errorf("empty threshold maxRPS")
	}
	if req.AttributeKey.KeyCheck == "" {
		return req, fmt.Errorf("empty key check")
	}
	if req.AttributeKey.Prefix == "" {
		return req, fmt.Errorf("empty prefix")
	}
	if req.Config.ExtendTTLKey == 0 {
		req.Config.ExtendTTLKey = 60
	}
	if req.Threshold.LimitMaxTTL == 0 {
		req.Threshold.LimitMaxTTL = 300
	}
	return req, nil
}
