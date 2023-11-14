# Hitrate Mechanism Library with Client-Side Cache and Real-Time RPS Estimation

## Overview

Welcome to the Hitrate Mechanism Library, a powerful tool for optimizing Redis performance by combining client-side caching and real-time Requests Per Second (RPS) estimation. This library intelligently decides when to cache data locally, preventing memory bloat and improving overall system efficiency.

## Features

- **Client-Side Caching:** Store frequently accessed data in local memory.
- **Real-Time RPS Estimation:** Dynamically calculate RPS to make informed caching decisions.
- **Configurable Parameters:** Easily adjust parameters such as time trend length, threshold RPS, and TTL for key memory checking.

## Getting Started

Follow these steps to integrate the library into your Redis-based application:

1. Clone the repository:
	```bash
	git clone https://github.com/bagusandrian/hitratemechanism.git
	```

2. Write code on your golang code: 
	```golang
	package main

	import (
		hrm "github.com/bagusandrian/hitratemechanism"
		hrmModel "github.com/bagusandrian/hitratemechanism/model"
	)

	var (
		client     rueidis.Client
		HRMUsecase *hrm.Usecase
	)
	const (
		// change base on your connection
		HostRedis = "127.0.0.1:6379" 
		PasswordRedis = ""
		UserNameRedis = "default"
	)
	func init() {
		var err error
		ctx := context.Background()
		hrmConfig := &hrmModel.HitRateMechanism{
			Config: hrmModel.Config{
				DefaultExpiration: 1 * time.Minute,
				PrefixKey:         "testing",
				LimitTrend:        5,
			},
			RedisConfig: rueidis.ClientOption{
				InitAddress: []string{HostRedis},
				Username:    UserNameRedis,
				Password:    PasswordRedis,
			},
		}
		HRMUsecase = hrm.New(hrmConfig)
	}

	func main() {
		req := hrmModel.RequestCheck{
				Key:          "key_redi_that_u_want_to_get",
				ThresholdRPS: 2,
				TTLCache:     1 * time.Minute,
			}
		redisResult, cacheDebug := HRMUsecase.HandlerRedis.Get(ctx, req)
		if redisResult.Error() {
				// handling error redis
				log.Panic(err)
			}
		result := redisResult.String()
		// will print the result of data from redis
		log.Printf("%+v\n", result)
		// will print result of cache debug
		// {
		//	"response_time": "20ms"
		//	"data_time_trend": {
		//		"reach_threshold_rps": TRUE,
		//		"time_trend": {
		//			0: 1234567,
		//			1: 1234580,
		//			2: 1234600,
		//			3: 1234620,
		//			4: 1234640,
		//		}
		//		"estimate_rps": 100
		//		"threshold_rps": 2
		//		"limit_trend": 5
		//	}
		//	"error": nil
		//	"success_message": ""
		// }
		log.Printf("%+v\n", cacheDebug)
	}

	```

## Benchmark
Benchmark data i compare with common implementation vs reuidis vs hitratemechanism: 
```
goos: darwin
goarch: arm64
pkg: github.com/bagusandrian/benchmark_hrm
BenchmarkNormal-10    	    2931	   3931797 ns/op	   72433 B/op	    1122 allocs/op
BenchmarkNormal-10    	    3037	   3928153 ns/op	   72413 B/op	    1122 allocs/op
BenchmarkHrm-10       	  212164	     56128 ns/op	    7747 B/op	     374 allocs/op
BenchmarkHrm-10       	  213416	     56106 ns/op	    7746 B/op	     374 allocs/op
BenchmarkCache-10     	  486037	     24577 ns/op	    1936 B/op	     132 allocs/op
BenchmarkCache-10     	  484736	     24540 ns/op	    1936 B/op	     132 allocs/op
PASS
ok  	github.com/bagusandrian/benchmark_hrm	74.287s
```
