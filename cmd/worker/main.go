package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/xerudro/DASHBOARD-v2/internal/audit"
	"github.com/xerudro/DASHBOARD-v2/internal/cache"
	"github.com/xerudro/DASHBOARD-v2/internal/database"
	"github.com/xerudro/DASHBOARD-v2/internal/jobs"
	"github.com/xerudro/DASHBOARD-v2/internal/repository"
	"github.com/xerudro/DASHBOARD-v2/internal/services/providers"
)

// Config holds worker configuration
type Config struct {
	Redis    RedisConfig    `mapstructure:"redis"`
	Database DatabaseConfig `mapstructure:"database"`
	Worker   WorkerConfig   `mapstructure:"worker"`
	Hetzner  HetznerConfig  `mapstructure:"hetzner"`
	Log      LogConfig      `mapstructure:"log"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	Concurrency int `mapstructure:"concurrency"`
}

// HetznerConfig holds Hetzner API configuration
type HetznerConfig struct {
	APIToken string `mapstructure:"api_token"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Setup logging
	setupLogging(config.Log)

	zlog.Info().Msg("Starting VIP Hosting Panel Background Worker")

	// Initialize database
	db, err := initDatabase(config.Database, config.Redis)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize Redis cache
	redisCache := cache.NewRedisCache(db.Redis(), "worker:", 5*time.Minute)

	// Initialize audit logger
	auditLogger := audit.NewAuditLogger(db.PostgreSQL(), db.Redis(), true)
	defer auditLogger.Close()

	// Initialize Hetzner provider
	hetznerProvider, err := providers.NewHetznerProvider(config.Hetzner.APIToken, redisCache)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to initialize Hetzner provider")
	}

	// Initialize repositories
	serverRepo := repository.NewServerRepository(db.PostgreSQL())

	// Initialize job handlers
	serverProvisioningJob := jobs.NewServerProvisioningJob(
		hetznerProvider,
		serverRepo,
		auditLogger,
	)

	// Create Asynq server
	redisAddr := fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: config.Redis.Password,
			DB:       config.Redis.DB,
		},
		asynq.Config{
			Concurrency: config.Worker.Concurrency,
			Queues: map[string]int{
				"critical": 6, // High priority for critical operations
				"default":  3, // Normal priority
				"low":      1, // Low priority
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				zlog.Error().
					Err(err).
					Str("type", task.Type()).
					Str("payload", string(task.Payload())).
					Msg("Task processing failed")
			}),
			Logger: &asynqLogger{},
		},
	)

	// Create task multiplexer
	mux := asynq.NewServeMux()

	// Register job handlers
	mux.HandleFunc(jobs.TypeServerProvisioning, serverProvisioningJob.ProcessServerProvisioning)
	mux.HandleFunc(jobs.TypeServerDeletion, serverProvisioningJob.ProcessServerDeletion)
	mux.HandleFunc(jobs.TypeServerResize, serverProvisioningJob.ProcessServerResize)

	zlog.Info().
		Int("concurrency", config.Worker.Concurrency).
		Msg("Background worker configured")

	// Start server in goroutine
	go func() {
		if err := srv.Run(mux); err != nil {
			zlog.Fatal().Err(err).Msg("Worker server failed")
		}
	}()

	zlog.Info().Msg("Background worker started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zlog.Info().Msg("Shutting down worker gracefully")

	// Shutdown worker
	srv.Shutdown()

	zlog.Info().Msg("Worker stopped")
}

// loadConfig loads configuration from file and environment variables
func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("worker.concurrency", 10)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")

	// Enable environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvPrefix("VIP")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		zlog.Warn().Msg("No config file found, using defaults and environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setupLogging configures the logger
func setupLogging(config LogConfig) {
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	if config.Format == "console" {
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zlog.Info().
		Str("level", level.String()).
		Str("format", config.Format).
		Msg("Logging configured")
}

// initDatabase initializes database connections
func initDatabase(dbConfig DatabaseConfig, redisConfig RedisConfig) (*database.DB, error) {
	pgConfig := database.Config{
		Host:               dbConfig.Host,
		Port:               dbConfig.Port,
		Name:               dbConfig.Name,
		User:               dbConfig.User,
		Password:           dbConfig.Password,
		SSLMode:            dbConfig.SSLMode,
		MaxConnections:     25,
		MaxIdleConnections: 10,
		MaxLifetime:        time.Hour,
	}

	rConfig := database.RedisConfig{
		Host:         redisConfig.Host,
		Port:         redisConfig.Port,
		Password:     redisConfig.Password,
		DB:           redisConfig.DB,
		PoolSize:     10,
		MinIdleConns: 5,
	}

	db, err := database.NewDB(pgConfig, rConfig)
	if err != nil {
		return nil, err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Health(ctx); err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	zlog.Info().Msg("Database connections established")
	return db, nil
}

// asynqLogger implements asynq.Logger interface
type asynqLogger struct{}

func (l *asynqLogger) Debug(args ...interface{}) {
	zlog.Debug().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Info(args ...interface{}) {
	zlog.Info().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Warn(args ...interface{}) {
	zlog.Warn().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Error(args ...interface{}) {
	zlog.Error().Msg(fmt.Sprint(args...))
}

func (l *asynqLogger) Fatal(args ...interface{}) {
	zlog.Fatal().Msg(fmt.Sprint(args...))
}
