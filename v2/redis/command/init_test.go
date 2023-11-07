package command

import (
	"testing"
	"time"

	"github.com/bagusandrian/hitratemechanism/v2/model"
	m "github.com/bagusandrian/hitratemechanism/v2/model"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/mock"
	"go.uber.org/mock/gomock"
)

func TestNew(t *testing.T) {
	db, _ := redismock.NewClientMock()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	client := mock.NewClient(ctrl)

	type args struct {
		hrm *m.HitRateMechanism
	}
	tests := []struct {
		name      string
		args      args
		wantPanic bool
	}{
		// TODO: Add test cases.
		{
			name: "success",
			args: args{
				hrm: &model.HitRateMechanism{
					Config: model.Config{
						MaxMemoryUsage:    1000,
						DefaultExpiration: 60 * time.Second,
						PrefixKey:         "testing",
						LimitTrend:        5,
					},
					Redis: client,
					RedisConfig: rueidis.ClientOption{
						InitAddress: []string{db.Options().Addr},
					},
				},
			},
		},
		{
			name: "failed connection redis",
			args: args{
				hrm: &model.HitRateMechanism{
					Config: model.Config{
						MaxMemoryUsage:    1000,
						DefaultExpiration: 60 * time.Second,
						PrefixKey:         "testing",
						LimitTrend:        5,
					},
					Redis: client,
					RedisConfig: rueidis.ClientOption{
						InitAddress: []string{"failed"},
					},
				},
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// New(tt.args.hrm)
			if tt.wantPanic {
				assertPanic(t, func() { New(tt.args.hrm) })
			} else {
				New(tt.args.hrm)
			}
		})
	}
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}
