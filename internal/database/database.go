package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

// Config represents database configuration
type Config struct {
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	Name               string        `mapstructure:"name"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	SSLMode            string        `mapstructure:"ssl_mode"`
	MaxConnections     int           `mapstructure:"max_connections"`
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	MaxLifetime        time.Duration `mapstructure:"max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

// DB holds database connections
type DB struct {
	postgres *sqlx.DB
	redis    *redis.Client
}

// NewDB creates new database connections
func NewDB(pgConfig Config, redisConfig RedisConfig) (*DB, error) {
	// Connect to PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pgConfig.Host,
		pgConfig.Port,
		pgConfig.User,
		pgConfig.Password,
		pgConfig.Name,
		pgConfig.SSLMode,
	)

	postgres, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Configure connection pool
	postgres.SetMaxOpenConns(pgConfig.MaxConnections)
	postgres.SetMaxIdleConns(pgConfig.MaxIdleConnections)
	postgres.SetConnMaxLifetime(pgConfig.MaxLifetime)

	log.Info().
		Str("host", pgConfig.Host).
		Int("port", pgConfig.Port).
		Str("database", pgConfig.Name).
		Msg("Connected to PostgreSQL")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     redisConfig.PoolSize,
		MinIdleConns: redisConfig.MinIdleConns,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		postgres.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("host", redisConfig.Host).
		Int("port", redisConfig.Port).
		Int("db", redisConfig.DB).
		Msg("Connected to Redis")

	return &DB{
		postgres: postgres,
		redis:    redisClient,
	}, nil
}

// PostgreSQL returns the PostgreSQL connection
func (db *DB) PostgreSQL() *sqlx.DB {
	return db.postgres
}

// Redis returns the Redis client
func (db *DB) Redis() *redis.Client {
	return db.redis
}

// Health checks the health of both databases
func (db *DB) Health(ctx context.Context) error {
	// Check PostgreSQL
	if err := db.postgres.PingContext(ctx); err != nil {
		return fmt.Errorf("PostgreSQL health check failed: %w", err)
	}

	// Check Redis
	if err := db.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Close closes both database connections
func (db *DB) Close() error {
	var err error

	// Close PostgreSQL
	if db.postgres != nil {
		if closeErr := db.postgres.Close(); closeErr != nil {
			err = closeErr
			log.Error().Err(closeErr).Msg("Failed to close PostgreSQL connection")
		} else {
			log.Info().Msg("PostgreSQL connection closed")
		}
	}

	// Close Redis
	if db.redis != nil {
		if closeErr := db.redis.Close(); closeErr != nil {
			err = closeErr
			log.Error().Err(closeErr).Msg("Failed to close Redis connection")
		} else {
			log.Info().Msg("Redis connection closed")
		}
	}

	return err
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := db.postgres.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// Migrate runs database migrations
func (db *DB) Migrate(migrationPath string) error {
	// This will be implemented with golang-migrate or custom migration runner
	// For now, we'll assume migrations are run via make migrate
	return nil
}