package module

import (
	"testing"
	"time"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
)

func TestNew(t *testing.T) {
	type args struct {
		hrm *m.HitRateMechanism
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "success",
			args: args{
				hrm: &m.HitRateMechanism{
					Config: m.Config{
						MaxMemoryUsage:    1000,
						DefaultExpiration: 60 * time.Second,
						PrefixKey:         "testing",
						LimitTrend:        5,
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
