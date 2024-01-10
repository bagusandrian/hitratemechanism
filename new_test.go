package hitratemechanism

import (
	"reflect"
	"testing"

	"github.com/bagusandrian/hitratemechanism/model"
)

func TestNew(t *testing.T) {
	type args struct {
		hrm *model.HitRateMechanism
	}
	tests := []struct {
		name string
		args args
		want *Usecase
	}{
		// TODO: Add test cases.
		{
			name: "success",
			args: args{
				hrm: &model.HitRateMechanism{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.hrm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
