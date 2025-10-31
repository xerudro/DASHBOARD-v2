package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// RedisCache provides Redis-based caching for query results
type RedisCache struct {
	client      *redis.Client
	defaultTTL  time.Duration
	keyPrefix   string
	metrics     *CacheMetrics
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Hits       int64
	Misses     int64
	Errors     int64
	Sets       int64
	Deletes    int64
	TotalSize  int64
}

// CacheOptions configures cache behavior
type CacheOptions struct {
	TTL          time.Duration
	Compress     bool
	Tags         []string
	RefreshOnHit bool
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(client *redis.Client, keyPrefix string, defaultTTL time.Duration) *RedisCache {
	return &RedisCache{
		client:     client,
		defaultTTL: defaultTTL,
		keyPrefix:  keyPrefix,
		metrics:    &CacheMetrics{},
	}
}

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	fullKey := rc.keyPrefix + key

	// Add timeout to context
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	data, err := rc.client.Get(timeoutCtx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			rc.metrics.Misses++
			return false, nil // Cache miss
		}
		rc.metrics.Errors++
		log.Error().
			Err(err).
			Str("key", key).
			Msg("Redis GET error")
		return false, err
	}

	// Deserialize data
	if err := json.Unmarshal(data, dest); err != nil {
		rc.metrics.Errors++
		log.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to unmarshal cache data")
		return false, err
	}

	rc.metrics.Hits++
	return true, nil
}

// Set stores a value in cache
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, opts ...CacheOptions) error {
	fullKey := rc.keyPrefix + key

	// Merge options
	options := CacheOptions{
		TTL: rc.defaultTTL,
	}
	if len(opts) > 0 {
		options = opts[0]
	}

	// Serialize data
	data, err := json.Marshal(value)
	if err != nil {
		rc.metrics.Errors++
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Add timeout to context
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Store in Redis
	err = rc.client.Set(timeoutCtx, fullKey, data, options.TTL).Err()
	if err != nil {
		rc.metrics.Errors++
		log.Error().
			Err(err).
			Str("key", key).
			Msg("Redis SET error")
		return err
	}

	// Add tags if provided
	if len(options.Tags) > 0 {
		for _, tag := range options.Tags {
			tagKey := rc.keyPrefix + "tag:" + tag
			rc.client.SAdd(timeoutCtx, tagKey, fullKey)
			rc.client.Expire(timeoutCtx, tagKey, options.TTL*2) // Tags live longer
		}
	}

	rc.metrics.Sets++
	rc.metrics.TotalSize += int64(len(data))

	return nil
}

// Delete removes a value from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := rc.keyPrefix + key

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := rc.client.Del(timeoutCtx, fullKey).Err()
	if err != nil {
		rc.metrics.Errors++
		return err
	}

	rc.metrics.Deletes++
	return nil
}

// DeleteByTag deletes all cache entries with a specific tag
func (rc *RedisCache) DeleteByTag(ctx context.Context, tag string) error {
	tagKey := rc.keyPrefix + "tag:" + tag

	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get all keys with this tag
	keys, err := rc.client.SMembers(timeoutCtx, tagKey).Result()
	if err != nil {
		rc.metrics.Errors++
		return err
	}

	// Delete all keys
	if len(keys) > 0 {
		pipe := rc.client.Pipeline()
		for _, key := range keys {
			pipe.Del(timeoutCtx, key)
		}
		pipe.Del(timeoutCtx, tagKey) // Delete tag set itself
		_, err = pipe.Exec(timeoutCtx)
		if err != nil {
			rc.metrics.Errors++
			return err
		}

		rc.metrics.Deletes += int64(len(keys))
		log.Info().
			Str("tag", tag).
			Int("count", len(keys)).
			Msg("Deleted cache entries by tag")
	}

	return nil
}

// Exists checks if a key exists in cache
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := rc.keyPrefix + key

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	count, err := rc.client.Exists(timeoutCtx, fullKey).Result()
	if err != nil {
		rc.metrics.Errors++
		return false, err
	}

	return count > 0, nil
}

// GetTTL returns the remaining TTL for a key
func (rc *RedisCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := rc.keyPrefix + key

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	ttl, err := rc.client.TTL(timeoutCtx, fullKey).Result()
	if err != nil {
		rc.metrics.Errors++
		return 0, err
	}

	return ttl, nil
}

// Refresh extends the TTL of a key
func (rc *RedisCache) Refresh(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := rc.keyPrefix + key

	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := rc.client.Expire(timeoutCtx, fullKey, ttl).Err()
	if err != nil {
		rc.metrics.Errors++
		return err
	}

	return nil
}

// Clear removes all cache entries with the configured prefix
func (rc *RedisCache) Clear(ctx context.Context) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var cursor uint64
	var deletedCount int

	for {
		keys, nextCursor, err := rc.client.Scan(timeoutCtx, cursor, rc.keyPrefix+"*", 100).Result()
		if err != nil {
			rc.metrics.Errors++
			return err
		}

		if len(keys) > 0 {
			pipe := rc.client.Pipeline()
			for _, key := range keys {
				pipe.Del(timeoutCtx, key)
			}
			_, err = pipe.Exec(timeoutCtx)
			if err != nil {
				rc.metrics.Errors++
				return err
			}

			deletedCount += len(keys)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	rc.metrics.Deletes += int64(deletedCount)
	log.Info().
		Int("count", deletedCount).
		Msg("Cache cleared")

	return nil
}

// GetOrSet retrieves a value from cache or sets it if not found
func (rc *RedisCache) GetOrSet(ctx context.Context, key string, dest interface{}, fetchFunc func() (interface{}, error), opts ...CacheOptions) error {
	// Try to get from cache first
	found, err := rc.Get(ctx, key, dest)
	if err != nil {
		return err
	}

	if found {
		return nil
	}

	// Cache miss - fetch data
	value, err := fetchFunc()
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	// Store in cache
	if err := rc.Set(ctx, key, value, opts...); err != nil {
		log.Error().
			Err(err).
			Str("key", key).
			Msg("Failed to set cache after fetch")
		// Don't return error - we have the data
	}

	// Copy value to dest
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// GetMetrics returns cache metrics
func (rc *RedisCache) GetMetrics() CacheMetrics {
	hitRate := float64(0)
	if rc.metrics.Hits+rc.metrics.Misses > 0 {
		hitRate = float64(rc.metrics.Hits) / float64(rc.metrics.Hits+rc.metrics.Misses) * 100
	}

	log.Debug().
		Int64("hits", rc.metrics.Hits).
		Int64("misses", rc.metrics.Misses).
		Int64("errors", rc.metrics.Errors).
		Float64("hit_rate", hitRate).
		Msg("Cache metrics")

	return *rc.metrics
}

// ResetMetrics resets cache metrics
func (rc *RedisCache) ResetMetrics() {
	rc.metrics = &CacheMetrics{}
}

// CacheWarmup pre-populates cache with common queries
type CacheWarmup struct {
	cache *RedisCache
	tasks []WarmupTask
}

// WarmupTask represents a cache warmup task
type WarmupTask struct {
	Key       string
	FetchFunc func() (interface{}, error)
	TTL       time.Duration
	Tags      []string
}

// NewCacheWarmup creates a new cache warmup instance
func NewCacheWarmup(cache *RedisCache) *CacheWarmup {
	return &CacheWarmup{
		cache: cache,
		tasks: make([]WarmupTask, 0),
	}
}

// AddTask adds a warmup task
func (cw *CacheWarmup) AddTask(task WarmupTask) {
	cw.tasks = append(cw.tasks, task)
}

// Execute runs all warmup tasks
func (cw *CacheWarmup) Execute(ctx context.Context) error {
	log.Info().
		Int("tasks", len(cw.tasks)).
		Msg("Starting cache warmup")

	successCount := 0
	errorCount := 0

	for _, task := range cw.tasks {
		value, err := task.FetchFunc()
		if err != nil {
			errorCount++
			log.Error().
				Err(err).
				Str("key", task.Key).
				Msg("Cache warmup task failed")
			continue
		}

		opts := CacheOptions{
			TTL:  task.TTL,
			Tags: task.Tags,
		}

		if err := cw.cache.Set(ctx, task.Key, value, opts); err != nil {
			errorCount++
			log.Error().
				Err(err).
				Str("key", task.Key).
				Msg("Failed to set cache during warmup")
			continue
		}

		successCount++
	}

	log.Info().
		Int("success", successCount).
		Int("errors", errorCount).
		Msg("Cache warmup completed")

	return nil
}

// MultiGetSet retrieves multiple keys or sets them if not found
func (rc *RedisCache) MultiGetSet(ctx context.Context, keys []string, fetchFunc func([]string) (map[string]interface{}, error), opts ...CacheOptions) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	missingKeys := make([]string, 0)

	// Try to get all keys from cache
	for _, key := range keys {
		var value interface{}
		found, err := rc.Get(ctx, key, &value)
		if err != nil {
			return nil, err
		}

		if found {
			result[key] = value
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	// If all keys found, return
	if len(missingKeys) == 0 {
		return result, nil
	}

	// Fetch missing keys
	fetchedData, err := fetchFunc(missingKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch missing keys: %w", err)
	}

	// Store fetched data in cache
	for key, value := range fetchedData {
		if err := rc.Set(ctx, key, value, opts...); err != nil {
			log.Error().
				Err(err).
				Str("key", key).
				Msg("Failed to cache fetched data")
		}
		result[key] = value
	}

	return result, nil
}
