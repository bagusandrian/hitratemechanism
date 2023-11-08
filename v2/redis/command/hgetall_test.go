package command

import (
	"context"
	"testing"

	hCache "github.com/bagusandrian/hitratemechanism/v2/cache"
	m "github.com/bagusandrian/hitratemechanism/v2/model"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/mock"
	"go.uber.org/mock/gomock"
)

func Test_usecase_HgetAll(t *testing.T) {
	db, _ := redismock.NewClientMock()
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock.NewClient(ctrl)
	cacheMock := *new(hCache.MockHandler)
	clientMock, _ := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{db.Options().Addr},
	})
	type fields struct {
		Redis        rueidis.Client
		Conf         m.Config
		HandlerCache hCache.MockHandler
	}
	type args struct {
		ctx context.Context
		req m.RequestCheck
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		mockFunc       func()
		wantResp       rueidis.RedisResult
		wantCacheDebug m.Response
	}{
		// TODO: Add test cases.
		{
			name: "reach threshold RPS",
			fields: fields{
				Redis:        clientMock,
				Conf:         m.Config{},
				HandlerCache: cacheMock,
			},
			args: args{
				ctx: ctx,
				req: m.RequestCheck{},
			},
			mockFunc: func() {
				req := m.RequestCheck{}
				resp := m.Response{
					DataTimeTrend: m.DataTimeTrend{
						ReachThresholdRPS: true,
					},
				}
				cacheMock.On("CacheValidateTrend", ctx, req).Return(resp).Once()
			},
		},
		{
			name: "not reach threshold RPS",
			fields: fields{
				Redis:        clientMock,
				Conf:         m.Config{},
				HandlerCache: cacheMock,
			},
			args: args{
				ctx: ctx,
				req: m.RequestCheck{},
			},
			mockFunc: func() {
				req := m.RequestCheck{}
				resp := m.Response{
					DataTimeTrend: m.DataTimeTrend{},
				}
				cacheMock.On("CacheValidateTrend", ctx, req).Return(resp).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := usecase{
				Redis:        tt.fields.Redis,
				Conf:         tt.fields.Conf,
				HandlerCache: &cacheMock,
			}
			tt.mockFunc()
			u.HgetAll(tt.args.ctx, tt.args.req)
		})
	}
}
