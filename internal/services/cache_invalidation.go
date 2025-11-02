package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xerudro/DASHBOARD-v2/internal/cache"
)

// CacheInvalidationService handles cache invalidation across the application
type CacheInvalidationService struct {
	cache *cache.RedisCache
}

// Get retrieves a value from cache (implements CacheService interface)
func (s *CacheInvalidationService) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	if s.cache == nil {
		return false, nil
	}
	return s.cache.Get(ctx, key, dest)
}

// Set stores a value in cache (implements CacheService interface)
func (s *CacheInvalidationService) Set(ctx context.Context, key string, value interface{}, opts cache.CacheOptions) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Set(ctx, key, value, opts)
}

// NewCacheInvalidationService creates a new cache invalidation service
func NewCacheInvalidationService(cache *cache.RedisCache) *CacheInvalidationService {
	return &CacheInvalidationService{
		cache: cache,
	}
}

// InvalidateDashboardStats invalidates dashboard statistics cache for a tenant
func (s *CacheInvalidationService) InvalidateDashboardStats(ctx context.Context, tenantID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}

	// Invalidate by tenant tag to clear all dashboard-related cache entries
	err := s.cache.DeleteByTag(ctx, tenantID.String())
	if err != nil {
		log.Error().
			Err(err).
			Str("tenant_id", tenantID.String()).
			Msg("Failed to invalidate dashboard cache")
		return err
	}

	log.Debug().
		Str("tenant_id", tenantID.String()).
		Msg("Dashboard cache invalidated for tenant")

	return nil
}

// InvalidateServerCache invalidates server-related cache entries
func (s *CacheInvalidationService) InvalidateServerCache(ctx context.Context, tenantID uuid.UUID, serverID *uuid.UUID) error {
	if s.cache == nil {
		return nil
	}

	// Invalidate dashboard stats (contains server counts)
	if err := s.InvalidateDashboardStats(ctx, tenantID); err != nil {
		return err
	}

	// If serverID is provided, invalidate specific server cache entries
	if serverID != nil {
		serverCacheKey := "server:" + serverID.String()
		if err := s.cache.Delete(ctx, serverCacheKey); err != nil {
			log.Error().
				Err(err).
				Str("server_id", serverID.String()).
				Msg("Failed to invalidate server cache")
			return err
		}

		log.Debug().
			Str("server_id", serverID.String()).
			Msg("Server cache invalidated")
	}

	return nil
}

// InvalidateUserCache invalidates user-related cache entries
func (s *CacheInvalidationService) InvalidateUserCache(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	if s.cache == nil {
		return nil
	}

	// Invalidate dashboard stats (contains user counts for admins)
	if err := s.InvalidateDashboardStats(ctx, tenantID); err != nil {
		return err
	}

	// If userID is provided, invalidate specific user cache entries
	if userID != nil {
		userCacheKey := "user:" + userID.String()
		if err := s.cache.Delete(ctx, userCacheKey); err != nil {
			log.Error().
				Err(err).
				Str("user_id", userID.String()).
				Msg("Failed to invalidate user cache")
			return err
		}

		log.Debug().
			Str("user_id", userID.String()).
			Msg("User cache invalidated")
	}

	return nil
}

// InvalidateAllDashboardCaches invalidates all dashboard caches across all tenants
func (s *CacheInvalidationService) InvalidateAllDashboardCaches(ctx context.Context) error {
	if s.cache == nil {
		return nil
	}

	// Delete all entries with the dashboard tag
	err := s.cache.DeleteByTag(ctx, "dashboard")
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to invalidate all dashboard caches")
		return err
	}

	log.Info().Msg("All dashboard caches invalidated")
	return nil
}

// WarmupDashboardCache pre-loads dashboard cache for a tenant
func (s *CacheInvalidationService) WarmupDashboardCache(ctx context.Context, tenantID uuid.UUID, fetchFunc func() (interface{}, error)) error {
	if s.cache == nil {
		return nil
	}

	// Use cache warmup functionality
	warmup := cache.NewCacheWarmup(s.cache)
	warmup.AddTask(cache.WarmupTask{
		Key:       "dashboard:stats:" + tenantID.String(),
		FetchFunc: fetchFunc,
		TTL:       30 * time.Second,
		Tags:      []string{"dashboard", "servers", tenantID.String()},
	})

	return warmup.Execute(ctx)
}
