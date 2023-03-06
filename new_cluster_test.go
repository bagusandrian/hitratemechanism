package HitRateMechanism

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
)

func TestClusterSetex(t *testing.T) {
	type args struct {
		dbname, key, value string
		expired            int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Positive ClusterSetex",
			args: args{
				dbname:  "local",
				key:     "njajal",
				value:   "njajal_value",
				expired: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterClient, clusterMock := redismock.NewClusterMock()
			Pool.RClusterClient = clusterClient
			clusterMock.ExpectPing().SetVal("PONG")
			clusterMock.ExpectPing().SetVal("PONG")
			clusterMock.ExpectSetEx(tt.args.key, tt.args.value, time.Second*time.Duration(tt.args.expired)).SetVal("OK")
			ctx := context.Background()
			err := Pool.ClusterSetex(ctx, tt.args.dbname, tt.args.key, tt.args.value, tt.args.expired)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr: %t got error ClusterSetex err: %+v\n", tt.wantErr, err)
			}
			if err := clusterMock.ExpectationsWereMet(); err != nil {
				t.Error(err)
			}
		})
	}
}
