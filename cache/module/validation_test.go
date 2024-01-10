package module

import (
	"context"
	"fmt"
	"testing"
	"time"

	goCache "github.com/TwiN/gocache/v2"
	m "github.com/bagusandrian/hitratemechanism/model"
	"github.com/go-test/deep"
	jsoniter "github.com/json-iterator/go"
)

func Test_usecase_CacheValidateTrend(t *testing.T) {
	goCacheMock := goCache.NewCache().WithMaxMemoryUsage(1000).WithEvictionPolicy(goCache.LeastRecentlyUsed)
	type fields struct {
		Conf    m.Config
		jsoni   jsoniter.API
		GoCache *goCache.Cache
	}
	type args struct {
		ctx context.Context
		req m.RequestCheck
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		mockFunc func(*usecase, m.Config, m.RequestCheck)
		wantResp m.Response
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name: "happy flow",
			fields: fields{
				Conf: m.Config{
					PrefixKey:  "testing",
					LimitTrend: 5,
				},
				jsoni:   jsoniter.ConfigCompatibleWithStandardLibrary,
				GoCache: goCacheMock,
			},
			args: args{
				ctx: context.Background(),
				req: m.RequestCheck{
					Keys: []string{"testing-1", "testing-2"},
				},
			},
			mockFunc: func(u *usecase, conf m.Config, req m.RequestCheck) {
				for _, key := range req.Keys {
					val := m.DataTimeTrend{
						TimeTrend: map[int64]int64{
							0: 12345,
							1: 12345,
							2: 12345,
							3: 12345,
							4: 12345,
						},
						LimitTrend:        5,
						ReachThresholdRPS: true,
					}
					u.GoCache.Set(fmt.Sprintf("%s:%s", conf.PrefixKey, key), val)
				}
			},
			wantResp: m.Response{
				DataKeys: map[string]m.DataTimeTrend{
					"testing-1": m.DataTimeTrend{
						TimeTrend: map[int64]int64{
							0: 12345,
							1: 12345,
							2: 12345,
							3: 12345,
							4: 12345,
						},
						LimitTrend:        5,
						ReachThresholdRPS: true,
					},
					"testing-2": m.DataTimeTrend{
						TimeTrend: map[int64]int64{
							0: 12345,
							1: 12345,
							2: 12345,
							3: 12345,
							4: 12345,
						},
						LimitTrend:        5,
						ReachThresholdRPS: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usecase{
				Conf:    tt.fields.Conf,
				GoCache: tt.fields.GoCache,
			}
			if tt.mockFunc != nil {
				tt.mockFunc(u, tt.fields.Conf, tt.args.req)
			}
			gotResp := u.CacheValidateTrend(tt.args.ctx, tt.args.req)
			if !tt.wantErr && gotResp.Error != nil {
				t.Errorf("expected no error but got error %+v\n", gotResp.Error)
			}
			if diff := deep.Equal(gotResp.DataKeys, tt.wantResp.DataKeys); diff != nil {
				t.Error(diff)
			}
		})
	}
}

func Test_usecase_cacheSetDataTrend(t *testing.T) {
	goCacheMock := goCache.NewCache().WithMaxMemoryUsage(1000).WithEvictionPolicy(goCache.LeastRecentlyUsed)
	type fields struct {
		Conf    m.Config
		GoCache *goCache.Cache
	}
	type args struct {
		ctx      context.Context
		key      string
		ttlCache time.Duration
		value    m.DataTimeTrend
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "TTL set",
			fields: fields{
				Conf: m.Config{
					DefaultExpiration: 5 * time.Second,
				},
				GoCache: goCacheMock,
			},
			args: args{
				ctx:      context.Background(),
				key:      "testing",
				ttlCache: 10 * time.Second,
				value:    m.DataTimeTrend{},
			},
		},
		{
			name: "TTL 0",
			fields: fields{
				Conf: m.Config{
					DefaultExpiration: 5 * time.Second,
				},
				GoCache: goCacheMock,
			},
			args: args{
				ctx:      context.Background(),
				key:      "testing",
				ttlCache: 0 * time.Second,
				value:    m.DataTimeTrend{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usecase{
				Conf:    tt.fields.Conf,
				GoCache: tt.fields.GoCache,
			}
			u.cacheSetDataTrend(tt.args.ctx, tt.args.key, tt.args.ttlCache, tt.args.value)
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
		{
			name: "no data",
			args: args{
				timeTrend: map[int64]int64{},
			},
			want: 0,
		},
		{
			name: "only 1 data",
			args: args{
				timeTrend: map[int64]int64{
					int64(0): 12345,
					int64(1): 0,
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
