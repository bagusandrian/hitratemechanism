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
	Pool        redis
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

type redis struct {
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
func (r *redis) getConnection(dbname string) redigo.Conn {
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
func (r *redis) GetConnection(dbname string) redigo.Conn {
	return r.getConnection(dbname)
}
func (r *redis) HgetAll(dbname, key string) (map[string]string, error) {
	conn := r.getConnection(dbname)
	defer conn.Close()
	if conn == nil {
		return nil, fmt.Errorf("failed get connection")
	}
	result, err := redigo.StringMap(conn.Do("HGETALL", key))
	return result, err
}
func (r *redis) HmsetWithExpMultiple(dbname string, data map[string]map[string]interface{}, expired int) (err error) {
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
func (r *redis) CustomHitRate(dbname, prefix, keyCheck string) (highTraffic bool, err error) {
	keyHitrate := fmt.Sprintf("%s-%s", prefix, keyCheck)
	conn := r.getConnection(dbname)
	if conn == nil {
		return false, fmt.Errorf("Failed to obtain connection db %s", dbname)
	}
	defer conn.Close()
	hitRateData, _ := r.hitRateGetData(conn, keyCheck, keyHitrate)
	type cmdAddTTl struct {
		command string
		key     string
		expire  int64
	}
	cmds := []cmdAddTTl{}
	// if checker key hitrate dont have ttl, will set expire for 1 minute
	// or key hitrate under 30 seconds, will set expire for 1 minute
	if hitRateData.TTLKeyHitRate > int64(-3) && hitRateData.TTLKeyHitRate <= int64(30) {
		fmt.Println("add ttl check hit rate")
		cmds = append(cmds, cmdAddTTl{command: "EXPIRE", key: keyHitrate, expire: 60})
	}
	newTTL := calculateNewTTL(hitRateData.TTLKeyCheck, int64(60), int64(300), maxTimeTTL)

	if hitRateData.RPS > int64(20) {
		highTraffic = true
		if newTTL > 0 {
			log.Println("add TTL redis")
			cmds = append(cmds, cmdAddTTl{command: "EXPIRE", key: keyCheck, expire: newTTL})
			// add ttl redis key target
		} else {
			log.Println("no need add TTL redis")
		}
	}

	if len(cmds) > 0 {
		log.Println("run cmds")
		for _, data := range cmds {

			err := conn.Send(data.command, data.key, data.expire)
			if err != nil {
				log.Println("Failed conn.Send", err)
				continue
			}
		}
	}
	err = conn.Flush()
	if err != nil {
		log.Println("Failed conn.Flush", err)
	}
	return highTraffic, nil
}

func (r *redis) hitRateGetData(conn redigo.Conn, keyCheck, keyHitrate string) (result hiteRateData, err error) {
	conn.Send("TTL", keyCheck)                     // 1
	conn.Send("HINCRBY", keyHitrate, "count", "1") // 2
	conn.Send("TTL", keyHitrate)                   // 3
	conn.Send("HMGET", keyHitrate, "end_time")     // 4
	conn.Flush()
	resultResponse := make(map[int]int64)
	for i := 1; i <= 4; i++ {
		if i == 4 {
			resultKey, err := redigo.String(conn.Receive())
			if err != nil {
				log.Println("err", err)
				continue
			}
			if resultKey == "" {
				result.HaveMaxDateTTL = false
				continue
			}
			endTime, err := time.Parse("2006-01-02 15:04:05", resultKey)
			if err != nil {
				log.Println("err", err)
				continue
			}
			result.MaxDateTTL = endTime
			result.HaveMaxDateTTL = true
		}
		resultKey, err := redigo.Int64(conn.Receive())
		if err != nil {
			log.Println("err", err)
		}
		resultResponse[i] = resultKey
	}
	result.TTLKeyCheck = resultResponse[1]
	result.countHitRate = resultResponse[2]
	result.TTLKeyHitRate = resultResponse[3]
	result.RPS = calculateRPS(result.countHitRate)
	log.Printf("%+v\n", result)
	return
}

func (r *redis) SetMaxTTLChecker(dbname, prefix, keyCheck string, endTime time.Time) error {
	conn := r.getConnection(dbname)
	if conn == nil {
		return fmt.Errorf("failed to obtain connection db %s", dbname)
	}
	defer conn.Close()
	conn.Send("HMSET", fmt.Sprintf("%s-%s", prefix, keyCheck), "end_time", endTime.Format("2006-01-02 15:04:05"))
	err := conn.Flush()
	return err
}
func calculateRPS(countHit int64) (rps int64) {
	rps = int64(countHit / 60)
	return
}

func calculateNewTTL(TTLKeyCHeck, extendTTL, limitTTL int64, dateMax time.Time) (newTTL int64) {
	maxTTL := int64(dateMax.Sub(time.Now()) / time.Second)
	newTTL = int64(60) + TTLKeyCHeck
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
