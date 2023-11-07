package cache

import (
	"context"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
)

//go:generate mockery --name=Handler --filename=mock_handler.go --inpackage
type Handler interface {
	CacheValidateTrend(ctx context.Context, req m.RequestCheck) (resp m.Response)
	// cacheGetDataTrend(ctx context.Context, key string) (result m.DataTimeTrend, err error)
	// cacheSetDataTrend(ctx context.Context, req m.RequestCheck, value m.DataTimeTrend)
	// calculateRPS(timeTrend map[int64]int64) int64
	// generateKey(ctx context.Context, key string) string
}
