package HitRateMechanism

import (
	"fmt"
	"log"
	"time"

	"github.com/eapache/go-resiliency/breaker"
	redigo "github.com/garyburd/redigo/redis"
	cmap "github.com/orcaman/concurrent-map"
)

var (
	Pool        hrm
	hosts       cmap.ConcurrentMap
	breakerCmap cmap.ConcurrentMap
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
		RedisDBName       string
		ExtendTTLKey      int64
		ExtendTTLKeyCheck int64
		ParseLayoutTime   string
	}
	ThresholdCustomHitrate struct {
		LimitMaxTTL         int64
		MaxRPS              int64
		LimitExtendTTLCheck int64
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

func init() {
	Pool.DBs = cmap.New()
	hosts = cmap.New()
	breakerCmap = cmap.New()
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

	Pool.DBs.Set(hostName, &redigo.Pool{
		MaxIdle:     opt.MaxIdleConn,
		MaxActive:   opt.MaxActiveConn,
		IdleTimeout: time.Duration(opt.Timeout) * time.Second,
		Dial: func() (redigo.Conn, error) {
			password := redigo.DialPassword(opt.Password)
			c, err := redigo.Dial(network, hostAddress, password)
			if err != nil {
				log.Println("[redis][New] error dial host:", err)
				return nil, err
			}
			return c, nil
		},
	})

	// save the connection address and options for later ue
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
func (r *hrm) getConnection(dbname string) redigo.Conn {
	var rdsConn redigo.Conn

	circuitbreaker, cbOk := breakerCmap.Get(dbname)
	if !cbOk {
		circuitbreaker = breaker.New(10, 2, 10*time.Second)
		breakerCmap.Set(dbname, circuitbreaker)
	}

	cb := circuitbreaker.(*breaker.Breaker)
	cbResult := cb.Run(func() error {
		placeholderPool, ok := r.DBs.Get(dbname)
		if !ok {
			return fmt.Errorf("[redis] %s - failed to retrieve redis connection map", dbname)
		}

		redisPool := placeholderPool.(*redigo.Pool)
		rdsConn = redisPool.Get()
		if _, err := rdsConn.Do("PING"); err == nil {
			return nil
		} else {
			log.Println("[redis] ping error:", err)
			rdsConn.Close() // just in case
		}

		log.Printf("[redis] %s - bad connection, closing and opening new one\n", dbname)
		hosttemp, ok := hosts.Get(dbname)
		if !ok {
			return fmt.Errorf("[redis] %s - failed to retrieve connection info", dbname)
		}

		host := hosttemp.(config)

		if err := redisPool.Close(); err != nil {
			log.Printf("[redis] %s - failed to close connection: %+v\n", dbname, err)
			return err
		}

		Pool.DBs.Remove(dbname)
		Pool.DBs.Set(dbname, &redigo.Pool{
			MaxIdle:     host.Option.MaxIdleConn,
			MaxActive:   host.Option.MaxActiveConn,
			IdleTimeout: time.Duration(host.Option.Timeout) * time.Second,
			Dial: func() (redigo.Conn, error) {
				c, err := redigo.Dial(host.Network, host.Address)
				if err != nil {
					log.Println("[redis][getConnection] error dial host:", err)
					return nil, err
				}
				return c, nil
			},
		})

		// return the asked datatype
		rdsTempConn, ok := r.DBs.Get(dbname)
		if !ok {
			return fmt.Errorf("[redis] %s - failed to retrieve newly created redis connection map", dbname)
		}

		redisPool = rdsTempConn.(*redigo.Pool)
		// if the newly open connection is having error than it need human intervention
		rdsConn = redisPool.Get()
		return nil
	})

	if cbResult == breaker.ErrBreakerOpen {
		log.Printf("[redis] redis circuitbreaker open, retrying later.. - %s\n", dbname)
	}

	return rdsConn
}

// GetConnection return redigo.Conn which enable developers to do redis command
// that is not commonly used (special case command, one that don't have wrapper function in this package. e.g: Exists)
func (r *hrm) GetConnection(dbname string) redigo.Conn {
	return r.getConnection(dbname)
}
func (r *hrm) HgetAll(dbname, key string) (map[string]string, error) {
	conn := r.getConnection(dbname)
	if conn == nil {
		return nil, fmt.Errorf("failed get connection")
	}
	defer conn.Close()
	result, err := redigo.StringMap(conn.Do("HGETALL", key))
	return result, err
}
func (r *hrm) HmsetWithExpMultiple(dbname string, data map[string]map[string]interface{}, expired int) (err error) {
	expString := fmt.Sprintf("%d", expired)
	conn := r.getConnection(dbname)
	if conn == nil {
		return fmt.Errorf("Failed to obtain connection db %s key %+v", dbname, data)
	}
	defer conn.Close()
	for key, hashmap := range data {
		for f, v := range hashmap {
			conn.Send("HMSET", key, f, v)
		}
		conn.Send("EXPIRE", key, expString)
	}
	conn.Flush()
	return
}

func (r *hrm) CustomHitRate(req ReqCustomHitRate) RespCustomHitRate {
	// func (r *hrm) CustomHitRate(dbname, prefix, keyCheck string) (highTraffic, haveMaxDateTTL bool, err error) {
	req, err := validateReqCustomHitRate(req)
	if err != nil {
		return RespCustomHitRate{
			Err: err,
		}
	}
	keyHitrate := fmt.Sprintf("%s-%s", req.AttributeKey.Prefix, req.AttributeKey.KeyCheck)
	conn := r.getConnection(req.Config.RedisDBName)
	if conn == nil {
		return RespCustomHitRate{
			Err: fmt.Errorf("failed to obtain connection db %s", req.Config.RedisDBName),
		}
	}
	defer conn.Close()
	hitRateData, err := r.hitRateGetData(conn, req.AttributeKey.KeyCheck, keyHitrate, req.Config.ParseLayoutTime)
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
	cmds := []cmdAddTTl{}
	// if checker key hitrate dont have ttl, will set expire for 1 minute
	// or key hitrate under 30 seconds, will set expire for 1 minute
	if hitRateData.TTLKeyHitRate > int64(-3) && hitRateData.TTLKeyHitRate <= req.Threshold.LimitExtendTTLCheck {
		cmds = append(cmds, cmdAddTTl{command: "EXPIRE", key: keyHitrate, expire: req.Config.ExtendTTLKeyCheck})
	}
	newTTL := calculateNewTTL(hitRateData.TTLKeyCheck, req.Config.ExtendTTLKey, req.Threshold.LimitMaxTTL, hitRateData.MaxDateTTL)
	resp := RespCustomHitRate{}
	if hitRateData.RPS > int64(20) {
		resp.HighTraffic = true
		if newTTL > 0 {
			// add ttl redis key target
			cmds = append(cmds, cmdAddTTl{command: "EXPIRE", key: req.AttributeKey.KeyCheck, expire: newTTL})
			resp.ExtendTTL = true
		}
	}

	if len(cmds) > 0 {
		for _, data := range cmds {

			err := conn.Send(data.command, data.key, data.expire)
			if err != nil {
				return RespCustomHitRate{
					Err: fmt.Errorf("failed conn.Send err:%+v", err),
				}
			}
		}
	}
	err = conn.Flush()
	if err != nil {
		return RespCustomHitRate{
			Err: fmt.Errorf("failed conn.Flush err:%+v", err),
		}
	}
	if !hitRateData.MaxDateTTL.IsZero() {
		resp.HaveMaxDateTTL = true
		resp.MaxDateTTL = hitRateData.MaxDateTTL
	}
	resp.RPS = hitRateData.RPS
	return resp
}

func (r *hrm) hitRateGetData(conn redigo.Conn, keyCheck, keyHitrate, parseLayoutTime string) (result hiteRateData, err error) {
	conn.Send("TTL", keyCheck)                     // 1
	conn.Send("HINCRBY", keyHitrate, "count", "1") // 2
	conn.Send("TTL", keyHitrate)                   // 3
	conn.Send("HMGET", keyHitrate, "end_time")     // 4
	conn.Flush()
	resultResponse := make(map[int]int64)
	for i := 1; i <= 4; i++ {
		if i == 4 {
			resultKey, err := redigo.Strings(conn.Receive())
			if err != nil {
				return result, err
			}
			if resultKey[0] == "" {
				continue
			}
			endTime, err := time.Parse(parseLayoutTime, resultKey[0])
			if err != nil {
				return result, err
			}
			result.MaxDateTTL = endTime
			result.HaveMaxDateTTL = true
			continue
		}
		resultKey, err := redigo.Int64(conn.Receive())
		if err != nil {
			return result, err
		}
		resultResponse[i] = resultKey
	}
	result.TTLKeyCheck = resultResponse[1]
	result.countHitRate = resultResponse[2]
	result.TTLKeyHitRate = resultResponse[3]
	result.RPS = calculateRPS(result.countHitRate)
	return
}

func (r *hrm) SetMaxTTLChecker(dbname, prefix, keyCheck string, endTime time.Time) error {
	conn := r.getConnection(dbname)
	if conn == nil {
		return fmt.Errorf("failed to obtain connection db %s", dbname)
	}
	defer conn.Close()
	conn.Send("HMSET", fmt.Sprintf("%s-%s", prefix, keyCheck), "end_time", endTime.Format("2006-01-02 15:04:05 Z0700 MST"))
	err := conn.Flush()
	return err
}
func calculateRPS(countHit int64) (rps int64) {
	rps = int64(countHit / 60)
	return
}

func calculateNewTTL(TTLKeyCHeck, extendTTL, limitTTL int64, dateMax time.Time) (newTTL int64) {
	maxTTL := int64(limitTTL)
	if !dateMax.IsZero() {
		maxTTL = int64(dateMax.Sub(time.Now()) / time.Second)
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

func validateReqCustomHitRate(req ReqCustomHitRate) (ReqCustomHitRate, error) {
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
	if req.Config.ExtendTTLKeyCheck == 0 {
		req.Config.ExtendTTLKeyCheck = 60
	}
	if req.Threshold.LimitExtendTTLCheck == 0 {
		req.Threshold.LimitExtendTTLCheck = 30
	}
	if req.Threshold.LimitMaxTTL == 0 {
		req.Threshold.LimitMaxTTL = 300
	}
	return req, nil
}
