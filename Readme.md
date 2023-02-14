
# HitRateMechanism

A hit rate mechanism is a technique used to track the frequency with which certain events or data are accessed, with the goal of optimizing performance by reducing the amount of time spent accessing or processing less frequently used items. One common use case for hit rate mechanisms to define traffic comming into your apps is high or not, and u can decide if traffic indicate with high traffic, u can handle it. 
## Authors

- [@bagusandrian](https://www.github.com/bagusandrian)


## Pre Requisite
- Golang min version 1.16
- redis min version 6.0
## Features

- Calculation of RPS
- Indication of high hraffic
- Auto add TTL of key redis target


## Installation

Install Hitratemechanism using the "go get" command"

```bash
  go get hightub.com/bagusandrian/hitratemechanism
```
    
## Usage/Examples
This is for example of usage hrm on your implementation. I give u example create 1 endpoint and optimize using hrm: 

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/rueian/rueidis"

	hrm "github.com/bagusandrian/hitratemechanism"
)

const (
	KeyPrimary  = "product_active"
	Prefix      = "hitrate"
	HostRedis   = "localhost:6379"
	RedisDBName = "local"
)

var (
	client rueidis.Client
)

type Fields struct {
	ProductID     int64
	OriginalPrice float64
	StartDate     string
	EndDate       string
}

func init() {
	ctx := context.Background()
	var err error
	client, err = rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{HostRedis},
	})
	if err != nil {
		panic(err)
	}
	opt := hrm.Options{
		MaxActiveConn: 100,
		MaxIdleConn:   10,
		Timeout:       3,
		Wait:          true,
	}
	hrm.New(RedisDBName, HostRedis, "tcp", opt)
	buildRedis(ctx)
}
func main() {
	ctx := context.Background()
	http.HandleFunc("/dummy-api", func(w http.ResponseWriter, r *http.Request) {
		resp, err := DummyAPI(ctx)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Println(resp)
		json.NewEncoder(w).Encode(resp)

	})
	log.Println("running server on port :8080")
	http.ListenAndServe("127.0.0.1:8080", nil)
}

func DummyAPI(ctx context.Context) (response Fields, err error) {
	key := genKey()
	reqHrm := hrm.ReqCustomHitRate{
		Config: hrm.ConfigCustomHitRate{
			RedisDBName:       RedisDBName,
			ExtendTTLKey:      60,
			ExtendTTLKeyCheck: 30,
			ParseLayoutTime:   "2006-01-02 15:04:05 Z0700 MST",
		},
		Threshold: hrm.ThresholdCustomHitrate{
			LimitMaxTTL:         300,
			MaxRPS:              20,
			LimitExtendTTLCheck: 60,
		},
		AttributeKey: hrm.AttributeKeyhitrate{
			KeyCheck: genKey(),
			Prefix:   "hitrate",
		},
	}
	respHrm := hrm.Pool.CustomHitRate(reqHrm)
	if respHrm.Err != nil {
		log.Println("failed jumpin on custom hitrate")
	}
	log.Printf("%+v\n", respHrm)
	// logic for highTraffic
	redisResult := make(map[string]string)
	log.Println("HIGH TRAFFIC indicated:", respHrm.HighTraffic)
	if respHrm.HighTraffic {
		// HGETALL myhash
		resp := client.DoCache(ctx, client.B().Hgetall().Key(key).Cache(), (30 * time.Second))
		// log.Println(resp.IsCacheHit()) // false
		// log.Println(resp.AsStrMap())   // map[f:v]
		redisResult, err = resp.AsStrMap()
		if err != nil {
			log.Println("error bos")
			return
		}
	} else {
		redisResult, err = hrm.Pool.HgetAll(RedisDBName, key)
	}
	if len(redisResult) == 0 {
		buildRedis(ctx)
		return
	}
	response.ProductID, _ = strconv.ParseInt(redisResult["product_id"], 10, 64)
	response.OriginalPrice, _ = strconv.ParseFloat(redisResult["original_price"], 64)
	response.StartDate = redisResult["start_time"]
	response.EndDate = redisResult["end_time"]
	if !respHrm.HaveMaxDateTTL {
		endTime, err := time.Parse("2006-01-02 15:04:05 Z0700 MST", response.EndDate)
		if err != nil {
			log.Println("err", err)
			return response, err
		}
		hrm.Pool.SetMaxTTLChecker(RedisDBName, Prefix, key, endTime)
	}
	return
}

func buildRedis(ctx context.Context) {
	data := make(map[string]map[string]interface{})
	for i := 0; i < 11; i++ {
		key := fmt.Sprintf("%s+%d", KeyPrimary, i)
		data[key] = make(map[string]interface{})
		data[key]["product_id"] = i
		data[key]["original_price"] = float64(12000)
		data[key]["start_time"] = time.Now().Format("2006-01-02 15:04:05 Z0700 MST")
		data[key]["end_time"] = time.Now().Add(6 * time.Minute).Format("2006-01-02 15:04:05 Z0700 MST")

	}
	err := hrm.Pool.HmsetWithExpMultiple(RedisDBName, data, 100)
	if err != nil {
		log.Println("error", err)
	}
}

func genKey() (key string) {
	key = fmt.Sprintf("%s+%d", KeyPrimary, rand.Intn(10))
	return
}

```
On this binary have endpoint `/dummy-api` and processing get information from redis. But I colaborate with rueidis (combination of redis & memcached) if traffic come to high, i will swich using rueidis but if not i will still using common redis implementation: 
```
respHrm := hrm.Pool.CustomHitRate(reqHrm)
	if respHrm.Err != nil {
		log.Println("failed jumpin on custom hitrate")
	}
	log.Printf("%+v\n", respHrm)
	// logic for highTraffic
	redisResult := make(map[string]string)
	log.Println("HIGH TRAFFIC indicated:", respHrm.HighTraffic)
	if respHrm.HighTraffic {
        ...
    } else {
        ...
    }
```
