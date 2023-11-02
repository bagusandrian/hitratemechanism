package cache

import (
	"context"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
)

type Handler interface {
	CacheValidateTrend(ctx context.Context, req m.RequestCheck) (resp m.Response)
}
