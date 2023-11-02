package module

import (
	"context"
	"testing"
	"time"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
	jsoniterpackage "github.com/json-iterator/go"
	memoryCache "github.com/patrickmn/go-cache"
)

func Test_usecase_CacheValidateTrend(t *testing.T) {
	jsoni := jsoniterpackage.ConfigCompatibleWithStandardLibrary
	conf := m.Config{
		DefaultExpiration: 1 * time.Second,
		CleanupInterval:   1 * time.Second,
		PrefixKey:         "testing",
		LimitTrend:        2,
	}
	cache := memoryCache.New(conf.DefaultExpiration, conf.CleanupInterval)
	type fields struct {
		Cache *memoryCache.Cache
		Conf  m.Config
		jsoni jsoniterpackage.API
	}
	type args struct {
		ctx context.Context
		req m.RequestCheck
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp m.Response
	}{
		// TODO: Add test cases.
		{
			name: "success cache validation trend",
			fields: fields{
				Cache: cache,
				Conf:  conf,
				jsoni: jsoni,
			},
			args: args{
				ctx: context.Background(),
				req: m.RequestCheck{
					Key:          "1",
					ThresholdRPS: 2,
					TTLCache:     10,
				},
			},
			wantResp: m.Response{
				DataTimeTrend: m.DataTimeTrend{
					HasCache: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usecase{
				Cache: tt.fields.Cache,
				Conf:  tt.fields.Conf,
				jsoni: tt.fields.jsoni,
			}
			if gotResp := u.CacheValidateTrend(tt.args.ctx, tt.args.req); gotResp.DataTimeTrend.HasCache {
				t.Errorf("expected HasCache is false, but result is true")
			}
		})
	}
}
