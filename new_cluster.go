package HitRateMechanism

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/eapache/go-resiliency/breaker"

	"github.com/redis/go-redis/v9"
)

func NewCluster(hostName, network string, IPs []string, customOption ...Options) {
	addrs := []string{}
	opt := Options{
		MaxIdleConn:   10,
		MaxActiveConn: 1000,
		Timeout:       1,
		Wait:          false,
	}
	if len(customOption) > 0 {
		opt = customOption[0]
	}
	// check hostname
	if hostName != "" {
		listIPs, err := checkStringForIpOrHostname(hostName)
		if err != nil {
			log.Printf("failed get list IPs from hostname err: %+v\n", err)
		}
		if len(listIPs) > 0 {
			addrs = listIPs
		}
	}
	if len(addrs) == 0 && len(IPs) > 0 {
		addrs = IPs
	}
	if len(addrs) == 0 {
		log.Fatal("address not found for get connection to redis")
	}
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:      addrs,
		ClientName: "redis-testing",
		Password:   opt.Password,
		Username:   opt.Username,
	})

	Pool.RClusterClient = rdb
	Pool.DBs.Set(hostName, &rdb)
	cfg := config{
		Address: hostName,
		ListIPs: addrs,
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

func (r *hrm) getConnectionCluster(ctx context.Context, dbname string) *redis.ClusterClient {
	var rdsClusterClient *redis.ClusterClient

	circuitbreaker, cbOk := breakerCmap.Get(dbname)
	if !cbOk {
		circuitbreaker = breaker.New(10, 2, 10*time.Second)
		breakerCmap.Set(dbname, circuitbreaker)
	}

	cb := circuitbreaker.(*breaker.Breaker)
	cbResult := cb.Run(func() error {
		rdsClusterClient = Pool.RClusterClient
		resp := rdsClusterClient.Ping(ctx)
		if resp.Err() == nil {
			return nil
		} else {
			log.Println("[redis] ping error:", resp.Err())
			rdsClusterClient.Close() // just in case
		}

		log.Printf("[redis] %s - bad connection, closing and opening new one\n", dbname)
		hosttemp, ok := hosts.Get(dbname)
		if !ok {
			return fmt.Errorf("[redis] %s - failed to retrieve connection info", dbname)
		}

		host := hosttemp.(config)

		Pool.DBs.Remove(dbname)
		// create new connection
		rdb := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:      host.ListIPs,
			ClientName: "redis-testing",
			Password:   host.Option.Password,
			Username:   host.Option.Username,
		})
		Pool.DBs.Set(dbname, &rdb)

		rdsClusterClient = rdb
		resp = rdsClusterClient.Ping(ctx)
		if resp.Err() == nil {
			return nil
		} else {
			log.Println("[redis] ping error:", resp.Err())
			rdsClusterClient.Close() // just in case
		}
		return nil
	})

	if cbResult == breaker.ErrBreakerOpen {
		log.Printf("[redis] redis circuitbreaker open, retrying later.. - %s\n", dbname)
	}

	return rdsClusterClient
}

// GetConnection return redigo.Conn which enable developers to do redis command
// that is not commonly used (special case command, one that don't have wrapper function in this package. e.g: Exists)
func (r *hrm) GetConnectionCluster(ctx context.Context, dbname string) *redis.ClusterClient {
	return r.getConnectionCluster(ctx, dbname)
}

func (r *hrm) ClusterSetex(ctx context.Context, dbname, key, value string, exp int) error {
	conn := r.getConnectionCluster(ctx, dbname)
	check := conn.Ping(ctx)
	if check.Err() != nil {
		return check.Err()
	}
	defer conn.Close()
	resp := conn.SetEx(ctx, key, value, time.Second*time.Duration(exp))
	return resp.Err()
}

// get list of IPs from hostname
func checkStringForIpOrHostname(host string) ([]string, error) {
	addr := net.ParseIP(host)
	if addr == nil {
		fmt.Println("Given String is a Domain Name")

	} else {
		fmt.Println("Given String is a Ip Address")
	}
	listIPs, err := net.LookupHost(host)
	return listIPs, err
}
