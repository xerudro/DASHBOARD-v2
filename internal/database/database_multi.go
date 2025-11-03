package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

// MultiConfig represents multi-database configuration
type MultiConfig struct {
	// Database backend type: postgresql, mysql, mariadb
	Type               string        `mapstructure:"type"`

	// Connection string (alternative to individual fields)
	DSN                string        `mapstructure:"dsn"`

	// Individual connection fields (if DSN not provided)
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	Name               string        `mapstructure:"name"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	SSLMode            string        `mapstructure:"ssl_mode"`

	// Connection pool settings
	MaxConnections     int           `mapstructure:"max_connections"`
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	MaxLifetime        time.Duration `mapstructure:"max_lifetime"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
}

// MultiDB holds multi-database connections
type MultiDB struct {
	// Universal database connection (supports PostgreSQL, MySQL, MariaDB)
	db    DatabaseConnection
	redis *redis.Client
}

// NewMultiDB creates new database connections with multi-database support
func NewMultiDB(dbConfig MultiConfig, redisConfig RedisConfig) (*MultiDB, error) {
	// Build DSN if not provided
	dsn := dbConfig.DSN
	if dsn == "" {
		dsn = buildDSN(dbConfig)
	}

	// Create universal database connection
	db, err := NewDatabaseConnection(dsn, dbConfig.MaxConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info().
		Str("type", string(db.GetType())).
		Int("max_connections", dbConfig.MaxConnections).
		Int("max_idle_connections", dbConfig.MaxIdleConnections).
		Dur("max_lifetime", dbConfig.MaxLifetime).
		Dur("idle_timeout", dbConfig.IdleTimeout).
		Msg("Connected to database with multi-database support")

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
		db.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("host", redisConfig.Host).
		Int("port", redisConfig.Port).
		Int("db", redisConfig.DB).
		Msg("Connected to Redis")

	return &MultiDB{
		db:    db,
		redis: redisClient,
	}, nil
}

// Database returns the universal database connection
func (mdb *MultiDB) Database() DatabaseConnection {
	return mdb.db
}

// PostgreSQL returns the underlying sqlx.DB (for backward compatibility)
// WARNING: This may not work for MySQL/MariaDB - use Database() instead
func (mdb *MultiDB) PostgreSQL() *sqlx.DB {
	return mdb.db.GetDB()
}

// Redis returns the Redis client
func (mdb *MultiDB) Redis() *redis.Client {
	return mdb.redis
}

// GetType returns the database type
func (mdb *MultiDB) GetType() DatabaseType {
	return mdb.db.GetType()
}

// Health checks the health of both databases
func (mdb *MultiDB) Health(ctx context.Context) error {
	// Check database
	if err := mdb.db.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check Redis
	if err := mdb.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Close closes both database connections
func (mdb *MultiDB) Close() error {
	var err error

	// Close database
	if mdb.db != nil {
		if closeErr := mdb.db.Close(); closeErr != nil {
			err = closeErr
			log.Error().Err(closeErr).Msg("Failed to close database connection")
		} else {
			log.Info().Str("type", string(mdb.db.GetType())).Msg("Database connection closed")
		}
	}

	// Close Redis
	if mdb.redis != nil {
		if closeErr := mdb.redis.Close(); closeErr != nil {
			err = closeErr
			log.Error().Err(closeErr).Msg("Failed to close Redis connection")
		} else {
			log.Info().Msg("Redis connection closed")
		}
	}

	return err
}

// Transaction executes a function within a database transaction
func (mdb *MultiDB) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	tx, err := mdb.db.BeginTx(ctx, nil)
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

// Query executes a query with automatic database syntax translation
func (mdb *MultiDB) Query(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return mdb.db.Query(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (mdb *MultiDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return mdb.db.QueryRow(ctx, query, args...)
}

// Exec executes a query without returning rows
func (mdb *MultiDB) Exec(ctx context.Context, query string, args ...interface{}) (interface{}, error) {
	return mdb.db.Exec(ctx, query, args...)
}

// Get executes a query and maps a single row to a struct
func (mdb *MultiDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return mdb.db.Get(ctx, dest, query, args...)
}

// Select executes a query and maps multiple rows to a slice of structs
func (mdb *MultiDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return mdb.db.Select(ctx, dest, query, args...)
}

// buildDSN builds a connection string from individual config fields
func buildDSN(config MultiConfig) string {
	dbType := DatabaseType(config.Type)

	switch dbType {
	case PostgreSQL:
		return fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Name,
			config.SSLMode,
		)

	case MySQL, MariaDB:
		// MySQL DSN format: user:password@tcp(host:port)/database?parseTime=true&charset=utf8mb4
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Name,
		)

	default:
		// Default to PostgreSQL format
		return fmt.Sprintf(
			"postgresql://%s:%s@%s:%d/%s?sslmode=%s",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Name,
			config.SSLMode,
		)
	}
}

// NewQueryBuilder creates a database-specific query builder
func (mdb *MultiDB) NewQueryBuilder() *QueryBuilder {
	return NewQueryBuilder(mdb.db.GetType())
}
