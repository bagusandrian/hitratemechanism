package HitRateMechanism

import (
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
)

func BenchmarkCustomHitRate(b *testing.B) {
	for i := 0; i <= 1000; i++ {
		prefix := fmt.Sprintf("prefix+%d", i)
		keyCheck := fmt.Sprintf("keycheck+%d", i)
		miniRds, err := miniredis.Run()
		if err != nil {
			b.Errorf("failed run miniRds err: %+v\n", err)
		}
		defer miniRds.Close()

		redisSrv, err := miniredis.Run()
		if err != nil {
			b.Errorf("failed run redisSrv err: %+v\n", err)
		}
		defer redisSrv.Close()
		New("local", redisSrv.Addr(), "tcp")
		prefixKeycheck := fmt.Sprintf("%s-%s", prefix, keyCheck)
		_, err = redisSrv.HIncr(prefixKeycheck, "count", i*100)
		if err != nil {
			b.Errorf("got errof for incr data mock err:%+v\n", err)
		}
		redisSrv.SetTTL(prefixKeycheck, 10*time.Second)
		redisSrv.HSet(keyCheck, "field", "value")
		redisSrv.SetTTL(keyCheck, 10*time.Second)
		b.Run(fmt.Sprintf("benchmark CustomHitRate %d", i), func(b *testing.B) {
			Pool.CustomHitRate("local", prefix, keyCheck)
		})
	}
}
func TestCustomHitRate(t *testing.T) {
	type args struct {
		dbname, prefix, keyCheck string
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		preparationData func(mockRedis *miniredis.Miniredis)
	}{
		{
			name: "Positive CustomHitRate",
			args: args{
				dbname:   "local",
				prefix:   "prefix",
				keyCheck: "keycheck",
			},
			preparationData: func(mockRedis *miniredis.Miniredis) {},
		},
		{
			name: "Negative CustomHitRate",
			args: args{
				dbname:   "wrongDB",
				prefix:   "prefix",
				keyCheck: "keycheck",
			},
			preparationData: func(mockRedis *miniredis.Miniredis) {},
			wantErr:         true,
		},
		{
			name: "CustomHitRate->hitRateGetData low RPS",
			args: args{
				dbname:   "local",
				prefix:   "prefix",
				keyCheck: "keycheck",
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
			args: args{
				dbname:   "local",
				prefix:   "prefix",
				keyCheck: "keycheck",
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
			args: args{
				dbname:   "local",
				prefix:   "prefix",
				keyCheck: "keycheck",
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
				New(tt.args.dbname, redisSrv.Addr(), "tcp")
				tt.preparationData(redisSrv)
			}
			_, _, err := Pool.CustomHitRate(tt.args.dbname, tt.args.prefix, tt.args.keyCheck)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr: %t got error CustomHitRate err: %+v\n", tt.wantErr, err)
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
