package HitRateMechanism

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	// "github.com/garyburd/redigo/redis"
	"github.com/redis/go-redis/v9"
)

func BenchmarkCustomHitRate(b *testing.B) {
	for i := 0; i <= 100; i++ {
		prefix := fmt.Sprintf("prefix+%d", i)
		keyCheck := fmt.Sprintf("keycheck+%d", i)
		rdb := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		ctx := context.Background()
		New("local", "localhost:6379", "tcp")
		prefixKeycheck := fmt.Sprintf("%s-%s", prefix, keyCheck)
		_, err := rdb.HIncrBy(ctx, prefixKeycheck, "count", int64(i*100)).Result()
		if err != nil {
			b.Errorf("got errof for incr data mock err:%+v\n", err)
		}
		rdb.Expire(ctx, prefixKeycheck, 10*time.Second)
		rdb.HSet(ctx, keyCheck, "field", "value")
		rdb.Expire(ctx, keyCheck, 10*time.Second)
		b.Run(fmt.Sprintf("benchmark CustomHitRate %d", i), func(b *testing.B) {
			req := ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "local",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			}
			Pool.CustomHitRate(req)
		})
	}
}
func TestCustomHitRate(t *testing.T) {
	tests := []struct {
		name            string
		args            ReqCustomHitRate
		wantErr         bool
		preparationData func(mockRedis *miniredis.Miniredis)
	}{
		{
			name: "Positive CustomHitRate",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "local",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			preparationData: func(mockRedis *miniredis.Miniredis) {},
		},
		{
			name: "Negative CustomHitRate",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "wrongDB",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			preparationData: func(mockRedis *miniredis.Miniredis) {},
			wantErr:         true,
		},
		{
			name: "CustomHitRate->hitRateGetData low RPS",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "local",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			wantErr: false,
			preparationData: func(mockRedis *miniredis.Miniredis) {
				_, err := mockRedis.HIncr("prefix-keycheck", "count", 1000)
				if err != nil {
					t.Errorf("got errof for incr data mock err:%+v\n", err)
				}
				mockRedis.SetTTL("prefix-keycheck", 100*time.Second)
				mockRedis.HSet("keycheck", "field", "value")
				mockRedis.SetTTL("keycheck", 100*time.Second)
			},
		},
		{
			name: "CustomHitRate->hitRateGetData High RPS",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "local",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			wantErr: false,
			preparationData: func(mockRedis *miniredis.Miniredis) {
				_, err := mockRedis.HIncr("prefix-keycheck", "count", 3000)
				if err != nil {
					t.Errorf("got errof for incr data mock err:%+v\n", err)
				}
				mockRedis.SetTTL("prefix-keycheck", 100*time.Second)
				mockRedis.HSet("keycheck", "field", "value")
				mockRedis.SetTTL("keycheck", 100*time.Second)
			},
		},
		{
			name: "CustomHitRate->hitRateGetData High RPS new TTL < end_time",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "local",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			wantErr: false,
			preparationData: func(mockRedis *miniredis.Miniredis) {
				_, err := mockRedis.HIncr("prefix-keycheck", "count", 3000)
				if err != nil {
					t.Errorf("got errof for incr data mock err:%+v\n", err)
				}

				mockRedis.SetTTL("prefix-keycheck", 100*time.Second)
				mockRedis.HSet("prefix-keycheck", "end_time", time.Now().Add(10*time.Second).Format("2006-01-02 15:04:05 Z0700 MST"))
				mockRedis.HSet("keycheck", "field", "value")
				mockRedis.HSet("keycheck", "end_time", time.Now().Add(10*time.Second).Format("2006-01-02 15:04:05 Z0700 MST"))
				mockRedis.SetTTL("keycheck", 100*time.Second)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				miniRds, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run miniRds err: %+v\n", err)
				}
				defer miniRds.Close()

				redisSrv, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run redisSrv err: %+v\n", err)
				}
				New(tt.args.Config.RedisDBName, redisSrv.Addr(), "tcp")
				tt.preparationData(redisSrv)
			}
			resp := Pool.CustomHitRate(tt.args)
			if (resp.Err != nil) != tt.wantErr {
				t.Errorf("wantErr: %t got error CustomHitRate err: %+v\n", tt.wantErr, resp.Err)
			}
		})
	}
}
func TestHmsetWithExpMultiple(t *testing.T) {
	type args struct {
		dbname  string
		data    map[string]map[string]interface{}
		expired int
	}
	data := make(map[string]map[string]interface{})
	for i := 0; i < 1; i++ {
		key := fmt.Sprintf("%s+%d", "testing_key", i)
		data[key] = make(map[string]interface{})
		data[key]["field"] = i
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive HmsetWithExpMultiple",
			args: args{
				dbname:  "local",
				data:    data,
				expired: 1,
			},
		},
		{
			name: "Negative HmsetWithExpMultiple",
			args: args{
				dbname:  "wrongDB",
				data:    data,
				expired: 1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				miniRds, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run miniRds err: %+v\n", err)
				}
				defer miniRds.Close()

				redisSrv, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run redisSrv err: %+v\n", err)
				}
				New(tt.args.dbname, redisSrv.Addr(), "tcp")
			}
			err := Pool.HmsetWithExpMultiple(tt.args.dbname, tt.args.data, tt.args.expired)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr: %t got error HmsetWithExpMultiple err: %+v\n", tt.wantErr, err)
			}
		})
	}
}
func TestHgetAll(t *testing.T) {
	type args struct {
		dbname, key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive HgetAll",
			args: args{
				dbname: "local",
				key:    "prefix",
			},
		},
		{
			name: "Negative HgetAll",
			args: args{
				dbname: "wrongDB",
				key:    "prefix",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				miniRds, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run miniRds err: %+v\n", err)
				}
				defer miniRds.Close()

				redisSrv, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run redisSrv err: %+v\n", err)
				}
				New(tt.args.dbname, redisSrv.Addr(), "tcp")
			}
			_, err := Pool.HgetAll(tt.args.dbname, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr: %t got error TestHgetAll err: %+v\n", tt.wantErr, err)
			}
		})
	}
}
func TestSetMaxTTLChecker(t *testing.T) {
	type args struct {
		dbname, prefix, keyCheck string
		endTime                  time.Time
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive SetMaxTTLChecker",
			args: args{
				dbname:   "local",
				prefix:   "prefix",
				keyCheck: "keycheck",
				endTime:  time.Now().Add(5 * time.Minute),
			},
		},
		{
			name: "Negative SetMaxTTLChecker",
			args: args{
				dbname:   "wrongDB",
				prefix:   "prefix",
				keyCheck: "keycheck",
				endTime:  time.Now().Add(5 * time.Minute),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				miniRds, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run miniRds err: %+v\n", err)
				}
				defer miniRds.Close()

				redisSrv, err := miniredis.Run()
				if err != nil {
					t.Errorf("failed run redisSrv err: %+v\n", err)
				}
				New(tt.args.dbname, redisSrv.Addr(), "tcp")
			}
			err := Pool.SetMaxTTLChecker(tt.args.dbname, tt.args.prefix, tt.args.keyCheck, tt.args.endTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr: %t got error SetMaxTTLChecker err: %+v\n", tt.wantErr, err)
			}
		})
	}
}
func TestCalculateRPS(t *testing.T) {
	type args struct {
		countHit int64
	}
	tests := []struct {
		name   string
		args   args
		output int64
	}{
		{
			name: "test count hit",
			args: args{
				countHit: 120,
			},
			output: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRPS(tt.args.countHit)
			if result != tt.output {
				t.Errorf("[calculateRPS] want output: %v. but got: %v", tt.output, result)
			}
		})
	}
}
func TestCalculateNewTTL(t *testing.T) {
	type args struct {
		TTLKeyCHeck, extendTTL, limitTTL int64
		dateMax                          time.Time
	}
	tests := []struct {
		name string
		args args
		// mock    func()
		// wantErr bool
		output int64
	}{
		{
			name: "new TTL < max TTL",
			args: args{
				TTLKeyCHeck: 60,
				extendTTL:   60,
				limitTTL:    300,
			},
			output: 120,
		},
		{
			name: "new TTL > limit",
			args: args{
				TTLKeyCHeck: 300,
				extendTTL:   60,
				limitTTL:    300,
			},
			output: 0,
		},
		{
			name: "date max exist",
			args: args{
				TTLKeyCHeck: 1,
				extendTTL:   60,
				limitTTL:    300,
				dateMax:     time.Now().Add(120 * time.Second),
			},
			output: 61,
		},
		{
			name: "newTTL > dateMax",
			args: args{
				TTLKeyCHeck: 120,
				extendTTL:   60,
				limitTTL:    300,
				dateMax:     time.Now().Add(120 * time.Second),
			},
			output: 119,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateNewTTL(tt.args.TTLKeyCHeck, tt.args.extendTTL, tt.args.limitTTL, tt.args.dateMax)
			if result != tt.output {
				t.Errorf("[calculateNewTTL] want output: %v. but got: %v", tt.output, result)
			}
		})
	}

}

func TestValidateReqCustomHitRate(t *testing.T) {
	tests := []struct {
		name    string
		args    ReqCustomHitRate
		output  ReqCustomHitRate
		wantErr bool
		err     error
	}{
		{
			name: "positive case",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "local",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			output: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "local",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			wantErr: false,
			err:     nil,
		},
		{
			name:    "empty db name",
			args:    ReqCustomHitRate{},
			output:  ReqCustomHitRate{},
			wantErr: true,
			err:     fmt.Errorf("empty redisDBName config"),
		},
		{
			name: "empty db ParseLayoutTime",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName: "testing",
				},
			},
			output: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName: "testing",
				},
			},
			wantErr: true,
			err:     fmt.Errorf("empty layout format for parse time"),
		},
		{
			name: "empty maxRPS",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ParseLayoutTime: "parse layout",
				},
			},
			output: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ParseLayoutTime: "parse layout",
				},
			},
			wantErr: true,
			err:     fmt.Errorf("empty threshold maxRPS"),
		},
		{
			name: "empty key check",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ParseLayoutTime: "parse layout",
				},
				Threshold: ThresholdCustomHitrate{
					MaxRPS: 20,
				},
			},
			output: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ParseLayoutTime: "parse layout",
				},
				Threshold: ThresholdCustomHitrate{
					MaxRPS: 20,
				},
			},
			wantErr: true,
			err:     fmt.Errorf("empty key check"),
		},
		{
			name: "empty prefix",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ParseLayoutTime: "parse layout",
				},
				Threshold: ThresholdCustomHitrate{
					MaxRPS: 20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
				},
			},
			output: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ParseLayoutTime: "parse layout",
				},
				Threshold: ThresholdCustomHitrate{
					MaxRPS: 20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
				},
			},
			wantErr: true,
			err:     fmt.Errorf("empty prefix"),
		},
		{
			name: "replacing default value",
			args: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					MaxRPS: 20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			output: ReqCustomHitRate{
				Config: ConfigCustomHitRate{
					RedisDBName:     "testing",
					ExtendTTLKey:    60,
					ParseLayoutTime: "2006-01-02 15:04:05 Z0700 MST",
				},
				Threshold: ThresholdCustomHitrate{
					LimitMaxTTL: 300,
					MaxRPS:      20,
				},
				AttributeKey: AttributeKeyhitrate{
					KeyCheck: "keycheck",
					Prefix:   "prefix",
				},
			},
			wantErr: false,
			err:     nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateReqCustomHitRate(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr: %t got error validateReqCustomHitRate err: %+v\n", tt.wantErr, err)
			}
			if !reflect.DeepEqual(result, tt.output) {
				t.Errorf("result is not same with output. result %+v | output %+v\n", result, tt.output)
			}
		})
	}
}
