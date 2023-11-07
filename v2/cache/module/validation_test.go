package module

import (
	"context"
	"reflect"
	"testing"
	"time"

	goCache "github.com/TwiN/gocache/v2"
	m "github.com/bagusandrian/hitratemechanism/v2/model"
	jsoniter "github.com/json-iterator/go"
)

func Test_usecase_generateKey(t *testing.T) {
	type fields struct {
		Conf m.Config
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		// TODO: Add test cases.
		{
			name: "success",
			fields: fields{
				Conf: m.Config{
					PrefixKey: "testing",
				},
			},
			args: args{
				ctx: context.Background(),
				key: "1",
			},
			want: "testing:1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usecase{
				Conf: tt.fields.Conf,
			}
			if got := u.generateKey(tt.args.ctx, tt.args.key); got != tt.want {
				t.Errorf("usecase.generateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_usecase_calculateRPS(t *testing.T) {
	type args struct {
		timeTrend map[int64]int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
		{
			name: "low RPS",
			args: args{
				timeTrend: map[int64]int64{
					int64(0): 1699347180000,
					int64(1): 1699347181000,
					int64(2): 1699347182000,
					int64(3): 1699347183000,
					int64(4): 1699347184000,
				},
			},
			want: 1,
		},
		{
			name: "high RPS",
			args: args{
				timeTrend: map[int64]int64{
					int64(0): 1699347180000,
					int64(1): 1699347180100,
					int64(2): 1699347180100,
					int64(3): 1699347180100,
					int64(4): 1699347180100,
				},
			},
			want: 50,
		},
		{
			name: "len < 2",
			args: args{
				timeTrend: map[int64]int64{
					int64(0): 12345,
				},
			},
			want: 0,
		},
		{
			name: "first == last",
			args: args{
				timeTrend: map[int64]int64{
					int64(0): 12345,
					int64(1): 12345,
				},
			},
			want: 0,
		},
		{
			name: "first == last > 2",
			args: args{
				timeTrend: map[int64]int64{
					int64(0): 12345,
					int64(1): 12345,
					int64(2): 12345,
					int64(3): 12345,
					int64(4): 12345,
				},
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usecase{
				Conf: m.Config{},
			}
			if got := u.calculateRPS(tt.args.timeTrend); got != tt.want {
				t.Errorf("usecase.calculateRPS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_usecase_cacheSetDataTrend(t *testing.T) {
	type fields struct {
		Conf    m.Config
		GoCache *goCache.Cache
		jsoni   jsoniter.API
	}
	type args struct {
		ctx   context.Context
		req   m.RequestCheck
		value m.DataTimeTrend
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "ttl req > 0",
			fields: fields{
				Conf: m.Config{
					DefaultExpiration: 60 * time.Second,
				},
				GoCache: goCache.NewCache().WithMaxMemoryUsage(1000).WithEvictionPolicy(goCache.LeastRecentlyUsed),
				jsoni:   jsoniter.ConfigCompatibleWithStandardLibrary,
			},
			args: args{
				ctx: context.Background(),
				req: m.RequestCheck{
					Key:          "testing",
					ThresholdRPS: 10,
					TTLCache:     10 * time.Second,
				},
				value: m.DataTimeTrend{},
			},
		},
		{
			name: "ttl req = 0",
			fields: fields{
				Conf: m.Config{
					DefaultExpiration: 60 * time.Second,
				},
				GoCache: goCache.NewCache().WithMaxMemoryUsage(1000).WithEvictionPolicy(goCache.LeastRecentlyUsed),
				jsoni:   jsoniter.ConfigCompatibleWithStandardLibrary,
			},
			args: args{
				ctx: context.Background(),
				req: m.RequestCheck{
					Key:          "testing",
					ThresholdRPS: 10,
				},
				value: m.DataTimeTrend{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usecase{
				Conf:    tt.fields.Conf,
				GoCache: tt.fields.GoCache,
				jsoni:   tt.fields.jsoni,
			}
			defer u.GoCache.Clear()
			u.cacheSetDataTrend(tt.args.ctx, tt.args.req, tt.args.value)
		})
	}
}

func Test_usecase_cacheGetDataTrend(t *testing.T) {
	type fields struct {
		Conf    m.Config
		jsoni   jsoniter.API
		GoCache *goCache.Cache
	}
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult m.DataTimeTrend
		wantErr    bool
		mock       func(u *usecase)
	}{
		// TODO: Add test cases.
		{
			name: "happy flow",
			fields: fields{
				Conf: m.Config{
					PrefixKey: "testing",
				},
				jsoni:   jsoniter.ConfigCompatibleWithStandardLibrary,
				GoCache: goCache.NewCache().WithMaxMemoryUsage(1000).WithEvictionPolicy(goCache.LeastRecentlyUsed),
			},
			args: args{
				ctx: context.Background(),
				key: "testing",
			},
			wantResult: m.DataTimeTrend{
				TimeTrend: map[int64]int64{
					0: 1000,
					1: 1000,
					2: 1000,
					3: 1000,
					4: 1000,
				},
			},
			mock: func(u *usecase) {
				val := m.DataTimeTrend{
					TimeTrend: map[int64]int64{
						int64(0): 1000,
						int64(1): 1000,
						int64(2): 1000,
						int64(3): 1000,
						int64(4): 1000,
					},
				}
				v, _ := u.jsoni.Marshal(val)
				u.GoCache.Set("testing:testing", v)
			},
		},
		{
			name: "error unmarshal",
			fields: fields{
				Conf: m.Config{
					PrefixKey: "testing",
				},
				jsoni:   jsoniter.ConfigCompatibleWithStandardLibrary,
				GoCache: goCache.NewCache().WithMaxMemoryUsage(1000).WithEvictionPolicy(goCache.LeastRecentlyUsed),
			},
			args: args{
				ctx: context.Background(),
				key: "testing",
			},
			mock: func(u *usecase) {
				val := "error unmarshal"
				v, _ := u.jsoni.Marshal(val)
				u.GoCache.Set("testing:testing", v)
			},
			wantErr: true,
			wantResult: m.DataTimeTrend{
				ReachThresholdRPS: false,
				HasCache:          false,
				TimeTrend:         make(map[int64]int64),
				EstimateRPS:       0,
				ThresholdRPS:      0,
				LimitTrend:        0,
			},
		},
		{
			name: "not found key",
			fields: fields{
				Conf: m.Config{
					PrefixKey: "testing",
				},
				jsoni:   jsoniter.ConfigCompatibleWithStandardLibrary,
				GoCache: goCache.NewCache().WithMaxMemoryUsage(1000).WithEvictionPolicy(goCache.LeastRecentlyUsed),
			},
			args: args{
				ctx: context.Background(),
				key: "testing",
			},
			wantResult: m.DataTimeTrend{
				ReachThresholdRPS: false,
				HasCache:          false,
				TimeTrend:         make(map[int64]int64),
				EstimateRPS:       0,
				ThresholdRPS:      0,
				LimitTrend:        0,
			},
			mock: func(u *usecase) {},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usecase{
				Conf:    tt.fields.Conf,
				jsoni:   tt.fields.jsoni,
				GoCache: tt.fields.GoCache,
			}
			tt.mock(u)
			defer u.GoCache.Clear()
			gotResult, err := u.cacheGetDataTrend(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("usecase.cacheGetDataTrend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("usecase.cacheGetDataTrend() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
