package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/xerudro/DASHBOARD-v2/internal/avatar"
	"github.com/xerudro/DASHBOARD-v2/internal/cache"
	"github.com/xerudro/DASHBOARD-v2/internal/database"
	"github.com/xerudro/DASHBOARD-v2/internal/handlers"
	"github.com/xerudro/DASHBOARD-v2/internal/middleware"
	"github.com/xerudro/DASHBOARD-v2/internal/monitoring"
	"github.com/xerudro/DASHBOARD-v2/internal/repository"
	"github.com/xerudro/DASHBOARD-v2/internal/services"
	"github.com/xerudro/DASHBOARD-v2/internal/services/sites"
)

// App holds the application dependencies
type App struct {
	config     *Config
	db         *database.DB
	fiber      *fiber.App
	repos      *Repositories
	svcs       *Services
	monitoring *monitoring.MonitoringService
}

// Config holds application configuration
type Config struct {
	Server   ServerConfig         `mapstructure:"server"`
	Database database.Config      `mapstructure:"database"`
	Redis    database.RedisConfig `mapstructure:"redis"`
	JWT      middleware.JWTConfig `mapstructure:"jwt"`
	Log      LogConfig            `mapstructure:"log"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host           string   `mapstructure:"host"`
	Port           int      `mapstructure:"port"`
	Mode           string   `mapstructure:"mode"` // development, production
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json, console
}

// Repositories holds all repository instances
type Repositories struct {
	User   *repository.UserRepository
	Server *repository.ServerRepository
	Tenant *repository.TenantRepository
	Site   *repository.SiteRepository
	// Add other repositories as needed
}

// Services holds all service instances
type Services struct {
	CacheInvalidation *services.CacheInvalidationService
	Permission        *services.PermissionService
	SiteManager       *sites.SiteManager
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Setup logging
	setupLogging(config.Log)

	zlog.Info().Msg("Starting VIP Hosting Panel API Server")

	// Create application
	app := &App{config: config}

	// Initialize database
	if err := app.initDatabase(); err != nil {
		zlog.Fatal().Err(err).Msg("Failed to initialize database")
	}

	// Initialize repositories
	app.initRepositories()

	// Initialize services
	app.initServices()

	// Initialize monitoring
	app.initMonitoring()

	// Initialize Fiber
	app.initFiber()

	// Setup routes
	app.setupRoutes()

	// Start server
	app.start()
}

// loadConfig loads configuration from file and environment variables
func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "development")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_connections", 100)     // Increased from 25 to support 100-200 concurrent users
	viper.SetDefault("database.max_idle_connections", 30) // Increased from 10 to maintain pool efficiency
	viper.SetDefault("database.max_lifetime", "30m")      // Reduced from 1h for better connection recycling
	viper.SetDefault("database.idle_timeout", "5m")       // Added: Close idle connections after 5 minutes
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")

	// Enable environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

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
	// Set log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Set log format
	if config.Format == "console" {
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zlog.Info().
		Str("level", level.String()).
		Str("format", config.Format).
		Msg("Logging configured")
}

// initDatabase initializes database connections
func (app *App) initDatabase() error {
	db, err := database.NewDB(app.config.Database, app.config.Redis)
	if err != nil {
		return err
	}

	app.db = db

	// Test connections
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.db.Health(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	zlog.Info().Msg("Database connections established")
	return nil
}

// initRepositories initializes all repositories
func (app *App) initRepositories() {
	app.repos = &Repositories{
		User:   repository.NewUserRepository(app.db.PostgreSQL()),
		Server: repository.NewServerRepository(app.db.PostgreSQL()),
		Tenant: repository.NewTenantRepository(app.db.PostgreSQL()),
		Site:   repository.NewSiteRepository(app.db.PostgreSQL()),
	}

	zlog.Info().Msg("Repositories initialized")
}

// initServices initializes all services
func (app *App) initServices() {
	// Initialize Redis cache for dashboard
	dashboardCache := cache.NewRedisCache(app.db.Redis(), "dashboard:", 30*time.Second)

	// Initialize site-specific services
	deployer := sites.NewDeployer()
	templateMgr := sites.NewTemplateManager()
	siteManager := sites.NewSiteManager(
		app.repos.Site,
		app.repos.Server,
		app.repos.Tenant,
		deployer,
		templateMgr,
		dashboardCache,
	)

	app.svcs = &Services{
		CacheInvalidation: services.NewCacheInvalidationService(dashboardCache),
		Permission:        services.NewPermissionService(app.repos.Tenant, nil),
		SiteManager:       siteManager,
	}

	zlog.Info().Msg("Services initialized")
}

// initMonitoring initializes the monitoring service
func (app *App) initMonitoring() {
	app.monitoring = monitoring.NewMonitoringService("v2.0.0", app.config.Server.Mode)
	app.monitoring.Start()

	zlog.Info().Msg("Monitoring service initialized")
}

// initFiber initializes the Fiber application
func (app *App) initFiber() {
	// Use optimized network configuration
	networkOptimizer := middleware.NewNetworkOptimizer()
	config := networkOptimizer.OptimizedFiberConfig()

	// Override with app-specific settings
	config.ServerHeader = "VIP-Hosting-Panel"
	config.AppName = "VIP Hosting Panel v2"
	config.ErrorHandler = customErrorHandler

	// Adjust config for production
	if app.config.Server.Mode == "production" {
		config.Prefork = true
	}

	app.fiber = fiber.New(config)

	// Add middleware stack in correct order
	// 1. Security headers (first for maximum coverage)
	app.fiber.Use(middleware.SecurityHeaders())

	// 2. Rate limiting (early to protect against abuse)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute
	app.fiber.Use(rateLimiter.Middleware())

	// 3. Performance optimizations
	responseOptimizer := middleware.NewResponseOptimizer()
	app.fiber.Use(responseOptimizer.OptimizeResponse)

	// 4. Monitoring and metrics
	app.fiber.Use(app.monitoring.Middleware())

	// 5. Request ID for tracing
	app.fiber.Use(requestid.New())

	// 6. Logger with security context
	app.fiber.Use(logger.New(logger.Config{
		Format: "${time} ${status} - ${method} ${path} ${latency} - IP: ${ip} - UA: ${ua}\n",
	}))

	// 7. Panic recovery
	app.fiber.Use(recover.New())

	// 8. CORS with secure configuration
	app.fiber.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(app.config.Server.AllowedOrigins, ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	// 9. Input validation middleware for all routes
	app.fiber.Use(middleware.ValidationMiddleware())

	// 10. SQL injection protection
	app.fiber.Use(middleware.SQLSecurityMiddleware())

	// 11. CSRF protection
	app.fiber.Use(middleware.CSRFProtection())

	// 12. Configuration security
	app.fiber.Use(middleware.ConfigSecurityMiddleware())

	// 13. Secure logging (prevents sensitive data leakage)
	app.fiber.Use(middleware.SecureLoggingMiddleware())

	zlog.Info().Msg("Fiber application initialized")
}

// setupRoutes sets up all application routes
func (app *App) setupRoutes() {
	// Health check endpoint
	app.fiber.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.db.Health(ctx); err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status": "unhealthy",
				"error":  err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	// Static files
	app.fiber.Static("/static", "./web/static")

	// Avatar placeholder
	app.fiber.Get("/api/avatar/:email", avatar.Handler)

	// API routes
	api := app.fiber.Group("/api/v1")

	// JWT middleware for protected routes
	jwtMiddleware := middleware.NewJWT(app.config.JWT)

	// Auth routes (no JWT required)
	authHandler := handlers.NewAuthHandler(app.repos.User)
	authHandler.SetJWT(jwtMiddleware) // Inject JWT middleware
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/register", authHandler.Register)
	auth.Post("/refresh", authHandler.RefreshToken)

	// Protected routes
	protected := api.Group("", jwtMiddleware.Protect())

	// Dashboard routes
	dashboardHandler := handlers.NewDashboardHandler(app.repos.User, app.repos.Server, app.svcs.CacheInvalidation)
	protected.Get("/dashboard", dashboardHandler.GetDashboard)
	protected.Get("/dashboard/stats", dashboardHandler.GetStats)

	// Server routes
	serverHandler := handlers.NewServerHandler(app.repos.Server, app.svcs.CacheInvalidation, app.svcs.Permission)
	servers := protected.Group("/servers")
	servers.Get("/", serverHandler.List)
	servers.Post("/", serverHandler.Create)
	servers.Get("/:id", serverHandler.Get)
	servers.Put("/:id", serverHandler.Update)
	servers.Delete("/:id", serverHandler.Delete)
	servers.Get("/:id/metrics", serverHandler.GetMetrics)

	// User management routes
	userHandler := handlers.NewUserHandler(app.repos.User)
	users := protected.Group("/users")
	users.Get("/", userHandler.List)
	users.Get("/profile", userHandler.GetProfile)
	users.Put("/profile", userHandler.UpdateProfile)

	// Site management routes
	siteHandler := handlers.NewSiteHandler(app.svcs.SiteManager)
	sitesGroup := protected.Group("/sites")
	sitesGroup.Get("/", siteHandler.ListSites)
	sitesGroup.Post("/", siteHandler.CreateSite)
	sitesGroup.Get("/templates", siteHandler.ListTemplates)
	sitesGroup.Get("/templates/:id", siteHandler.GetTemplate)
	sitesGroup.Get("/types", siteHandler.GetSiteTypes)
	sitesGroup.Post("/validate-domain", siteHandler.ValidateDomain)
	sitesGroup.Get("/:id", siteHandler.GetSite)
	sitesGroup.Put("/:id", siteHandler.UpdateSite)
	sitesGroup.Delete("/:id", siteHandler.DeleteSite)
	sitesGroup.Post("/:id/redeploy", siteHandler.RedeploySite)
	sitesGroup.Get("/:id/metrics", siteHandler.GetSiteMetrics)

	// Web routes (HTML responses)
	web := app.fiber.Group("/")

	// Public routes
	web.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/dashboard")
	})
	web.Get("/login", authHandler.LoginPage)
	web.Post("/login", authHandler.LoginForm)
	web.Get("/register", authHandler.RegisterPage)
	web.Post("/register", authHandler.RegisterForm)
	web.Get("/logout", func(c *fiber.Ctx) error {
		jwtMiddleware.ClearTokenCookie(c)
		return c.Redirect("/login")
	})

	// Protected web routes
	webProtected := web.Group("", jwtMiddleware.Protect())
	webProtected.Get("/dashboard", dashboardHandler.DashboardPage)
	webProtected.Get("/servers", serverHandler.ServersPage)
	webProtected.Get("/servers/create", serverHandler.CreateServerPage)
	webProtected.Post("/servers", serverHandler.CreateServerForm)

	// Setup monitoring routes
	monitoring.SetupMonitoringRoutes(app.fiber, app.monitoring)

	zlog.Info().Msg("Routes configured")
}

// start starts the server with graceful shutdown
func (app *App) start() {
	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", app.config.Server.Host, app.config.Server.Port)
		zlog.Info().
			Str("host", app.config.Server.Host).
			Int("port", app.config.Server.Port).
			Str("mode", app.config.Server.Mode).
			Msg("Server starting")

		if err := app.fiber.Listen(addr); err != nil {
			zlog.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	zlog.Info().Msg("Server shutting down gracefully")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown Fiber
	if err := app.fiber.ShutdownWithContext(ctx); err != nil {
		zlog.Error().Err(err).Msg("Failed to shutdown server gracefully")
	}

	// Close database connections
	if err := app.db.Close(); err != nil {
		zlog.Error().Err(err).Msg("Failed to close database connections")
	}

	zlog.Info().Msg("Server stopped")
}

// customErrorHandler handles Fiber errors
func customErrorHandler(ctx *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	zlog.Error().
		Err(err).
		Int("status", code).
		Str("method", ctx.Method()).
		Str("path", ctx.Path()).
		Msg("HTTP error")

	// Return JSON for API routes
	if strings.HasPrefix(ctx.Path(), "/api/") {
		return ctx.Status(code).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Return HTML for web routes
	return ctx.Status(code).SendString(err.Error())
}
