package services

import (
	"context"
	"time"

	"github.com/xerudro/DASHBOARD-v2/internal/cache"
)

// CacheService defines the interface for caching operations
type CacheService interface {
	Get(ctx context.Context, key string, dest interface{}) (bool, error)
	Set(ctx context.Context, key string, value interface{}, opts cache.CacheOptions) error
	Delete(ctx context.Context, key string) error
	DeleteByTag(ctx context.Context, tag string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Refresh(ctx context.Context, key string, ttl time.Duration) error
	Clear(ctx context.Context) error
	GetOrSet(ctx context.Context, key string, dest interface{}, fetchFunc func() (interface{}, error), opts ...cache.CacheOptions) error
}
