package hitratemechanism

import (
	"testing"
	"time"

	"github.com/bagusandrian/hitratemechanism/model"
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
		hrm *model.HitRateMechanism
	}
	tests := []struct {
		name string
		args args
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			New(tt.args.hrm)
		})
	}
}
