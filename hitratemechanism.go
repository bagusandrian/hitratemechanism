package HitRateMechanism

import (
	"fmt"
	"log"
	"strconv"
	"time"

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
unc (r *redis) getConnection(dbname string) redigo.Conn {
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

func (r *redis) SetexMultiple(dbname string, data map[string]string, expired int) (res map[string]string, err error) {
	res = make(map[string]string)
	expString := fmt.Sprintf("%d", expired)
	conn := r.getConnection(dbname)
	if conn == nil {
		return res, fmt.Errorf("Failed to obtain connection db %s key %+v", dbname, data)
	}
	defer conn.Close()
	for key, val := range data {
		conn.Send("SETEX", key, val, expString)
	}
	conn.Flush()
	for key, _ := range data {
		resultKey, err := redigo.String(conn.Receive())
		if err != nil {
			continue
		}
		res[key] = resultKey
	}
	return
}

func (r *redis) CustomHitRate(dbname, prefix, keyCheck string) (highTraffic bool, err error) {
	keyHitrate := fmt.Sprintf("%s-%s", prefix, keyCheck)
	resultResponse := make(map[int]int64)
	conn := r.getConnection(dbname)
	if conn == nil {
		return false, fmt.Errorf("Failed to obtain connection db %s", dbname)
	}
	defer conn.Close()
	conn.Send("TTL", keyCheck)                     // 1
	conn.Send("HINCRBY", keyHitrate, "count", "1") // 2
	conn.Send("TTL", keyHitrate)                   // 3
	conn.Send("HMGET", keyHitrate, "hightraffic")  // 4 // bisa di hapus karena dari response HINCRBY udah dapet.
	conn.Flush()
	for i := 1; i <= 4; i++ {
		if i == 4 {

			resp, err := redigo.Strings(conn.Receive())
			if err != nil {
				log.Println("err", err)
				continue
			}
			log.Printf("temp: %+v ", resp)
			highTraffic, _ = strconv.ParseBool(resp[0])
			continue
		}
		resultKey, err := redigo.Int64(conn.Receive())
		if err != nil {
			log.Println("err", err)
			continue
		}
		resultResponse[i] = resultKey
	}
	type cmdAddTTl struct {
		command string
		key     string
		expire  int64
	}
	cmds := []cmdAddTTl{}
	if resultResponse[3] > int64(-3) && resultResponse[3] <= int64(30) {
		// if checker key dont have ttl, will set expire for 1 minutes
		fmt.Println("add ttl check hit rate")
		cmds = append(cmds, cmdAddTTl{command: "EXPIRE", key: keyHitrate, expire: 60})
	}
	// calculate base on hit rate
	// simple calculation
	// Request per second (rps) 20
	// Request per minute (rpm) 191.04
	if resultResponse[2] <= int64(190) && len(cmds) == 0 {
		// do nothing if calculation is not pass for add expire on keycheck
		return false, nil
	}
	// expected hit rate > 20 RPS
	// check ttl key check is greater than 300 seconds
	newTTL := 60 + resultResponse[1]
	log.Println("HIGHTRAFFIC:", highTraffic)
	if newTTL > 300 || resultResponse[2] <= int64(190) {
		fmt.Println("no need add ttl because still long")
	} else if resultResponse[2] > int64(190) && !highTraffic {
		highTraffic = true
		err := conn.Send("HMSET", keyHitrate, "hightraffic", strconv.FormatBool(highTraffic))
		if err != nil {
			log.Println("Failed conn.Send", err)
		}
	}
	// add ttl redis key target

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
