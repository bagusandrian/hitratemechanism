package HitRateMechanism

import (
	"testing"
	"time"
)

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
