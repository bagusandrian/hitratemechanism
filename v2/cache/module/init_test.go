package module

import (
	"reflect"
	"testing"
	"time"

	"github.com/bagusandrian/hitratemechanism/v2/cache"
	m "github.com/bagusandrian/hitratemechanism/v2/model"
	jsoniterpackage "github.com/json-iterator/go"
	memoryCache "github.com/patrickmn/go-cache"
)

func TestNew(t *testing.T) {
	jsoni := jsoniterpackage.ConfigCompatibleWithStandardLibrary
	type args struct {
		hrm *m.HitRateMechanism
	}
	tests := []struct {
		name string
		args args
		want cache.Handler
	}{
		// TODO: Add test cases.
		{
			name: "success init",
			args: args{
				hrm: &m.HitRateMechanism{},
			},
			want: &usecase{
				Cache: memoryCache.New(time.Second*0, time.Second*0),
				Conf:  m.Config{},
				jsoni: jsoni,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.hrm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
