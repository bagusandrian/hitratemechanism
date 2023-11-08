package cache

import (
	"context"

	m "github.com/bagusandrian/hitratemechanism/v2/model"
)

//go:generate mockery --name=Handler --filename=mock_handler.go --inpackage
type Handler interface {
	CacheValidateTrend(ctx context.Context, req m.RequestCheck) (resp m.Response)
}
