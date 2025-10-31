package shutdown

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// GracefulShutdown manages graceful application shutdown
type GracefulShutdown struct {
	timeout       time.Duration
	shutdownFuncs []ShutdownFunc
	mu            sync.Mutex
	shutdownOnce  sync.Once
	signals       []os.Signal
}

// ShutdownFunc represents a function to be called during shutdown
type ShutdownFunc struct {
	Name     string
	Priority int // Lower priority runs first
	Timeout  time.Duration
	Func     func(context.Context) error
}

// NewGracefulShutdown creates a new graceful shutdown manager
func NewGracefulShutdown(timeout time.Duration) *GracefulShutdown {
	return &GracefulShutdown{
		timeout:       timeout,
		shutdownFuncs: make([]ShutdownFunc, 0),
		signals:       []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT},
	}
}

// RegisterShutdownFunc registers a function to be called during shutdown
func (gs *GracefulShutdown) RegisterShutdownFunc(name string, priority int, timeout time.Duration, fn func(context.Context) error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	shutdownFunc := ShutdownFunc{
		Name:     name,
		Priority: priority,
		Timeout:  timeout,
		Func:     fn,
	}

	gs.shutdownFuncs = append(gs.shutdownFuncs, shutdownFunc)

	log.Debug().
		Str("name", name).
		Int("priority", priority).
		Dur("timeout", timeout).
		Msg("Shutdown function registered")
}

// Start starts listening for shutdown signals
func (gs *GracefulShutdown) Start() <-chan struct{} {
	done := make(chan struct{})

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, gs.signals...)

		sig := <-sigChan
		log.Info().
			Str("signal", sig.String()).
			Msg("Shutdown signal received")

		gs.shutdownOnce.Do(func() {
			gs.executeShutdown()
			close(done)
		})
	}()

	return done
}

// Shutdown triggers shutdown manually
func (gs *GracefulShutdown) Shutdown() {
	gs.shutdownOnce.Do(func() {
		log.Info().Msg("Manual shutdown triggered")
		gs.executeShutdown()
	})
}

// executeShutdown executes all registered shutdown functions
func (gs *GracefulShutdown) executeShutdown() {
	log.Info().
		Int("functions", len(gs.shutdownFuncs)).
		Dur("timeout", gs.timeout).
		Msg("Starting graceful shutdown")

	startTime := time.Now()

	// Sort shutdown functions by priority
	gs.sortShutdownFuncs()

	// Execute shutdown functions
	ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
	defer cancel()

	var wg sync.WaitGroup
	errors := make(chan error, len(gs.shutdownFuncs))

	for _, shutdownFunc := range gs.shutdownFuncs {
		wg.Add(1)

		go func(sf ShutdownFunc) {
			defer wg.Done()

			log.Info().
				Str("name", sf.Name).
				Int("priority", sf.Priority).
				Msg("Executing shutdown function")

			// Create context with timeout for this specific function
			funcCtx, funcCancel := context.WithTimeout(ctx, sf.Timeout)
			defer funcCancel()

			funcStart := time.Now()
			err := sf.Func(funcCtx)
			funcDuration := time.Since(funcStart)

			if err != nil {
				log.Error().
					Err(err).
					Str("name", sf.Name).
					Dur("duration", funcDuration).
					Msg("Shutdown function failed")
				errors <- fmt.Errorf("%s: %w", sf.Name, err)
			} else {
				log.Info().
					Str("name", sf.Name).
					Dur("duration", funcDuration).
					Msg("Shutdown function completed successfully")
			}
		}(shutdownFunc)
	}

	// Wait for all shutdown functions to complete or timeout
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		log.Info().Msg("All shutdown functions completed")
	case <-ctx.Done():
		log.Warn().
			Dur("timeout", gs.timeout).
			Msg("Shutdown timeout reached, forcing shutdown")
	}

	close(errors)

	// Log any errors
	errorCount := 0
	for err := range errors {
		errorCount++
		log.Error().Err(err).Msg("Shutdown error")
	}

	totalDuration := time.Since(startTime)
	log.Info().
		Dur("duration", totalDuration).
		Int("errors", errorCount).
		Msg("Graceful shutdown completed")
}

// sortShutdownFuncs sorts shutdown functions by priority
func (gs *GracefulShutdown) sortShutdownFuncs() {
	// Simple bubble sort (sufficient for small number of functions)
	n := len(gs.shutdownFuncs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if gs.shutdownFuncs[j].Priority > gs.shutdownFuncs[j+1].Priority {
				gs.shutdownFuncs[j], gs.shutdownFuncs[j+1] = gs.shutdownFuncs[j+1], gs.shutdownFuncs[j]
			}
		}
	}
}

// FiberShutdown creates a shutdown function for Fiber app
func FiberShutdown(app *fiber.App) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().Msg("Shutting down Fiber application")

		if err := app.ShutdownWithContext(ctx); err != nil {
			return fmt.Errorf("failed to shutdown Fiber: %w", err)
		}

		log.Info().Msg("Fiber application shut down successfully")
		return nil
	}
}

// DatabaseShutdown creates a shutdown function for database connections
func DatabaseShutdown(closeFunc func() error) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().Msg("Closing database connections")

		// Run close function in goroutine with timeout
		done := make(chan error, 1)
		go func() {
			done <- closeFunc()
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("failed to close database: %w", err)
			}
			log.Info().Msg("Database connections closed successfully")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("database close timeout")
		}
	}
}

// GenericShutdown creates a generic shutdown function
func GenericShutdown(name string, closeFunc func() error) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().
			Str("resource", name).
			Msg("Closing resource")

		done := make(chan error, 1)
		go func() {
			done <- closeFunc()
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("failed to close %s: %w", name, err)
			}
			log.Info().
				Str("resource", name).
				Msg("Resource closed successfully")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("%s close timeout", name)
		}
	}
}

// BackgroundWorkerShutdown creates a shutdown function for background workers
func BackgroundWorkerShutdown(stopFunc func() error) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().Msg("Stopping background workers")

		done := make(chan error, 1)
		go func() {
			done <- stopFunc()
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("failed to stop workers: %w", err)
			}
			log.Info().Msg("Background workers stopped successfully")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("worker shutdown timeout")
		}
	}
}

// CacheShutdown creates a shutdown function for cache
func CacheShutdown(name string, closeFunc func() error) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().
			Str("cache", name).
			Msg("Flushing and closing cache")

		done := make(chan error, 1)
		go func() {
			done <- closeFunc()
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("failed to close cache %s: %w", name, err)
			}
			log.Info().
				Str("cache", name).
				Msg("Cache closed successfully")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("cache %s close timeout", name)
		}
	}
}

// MetricsShutdown creates a shutdown function for metrics exporter
func MetricsShutdown(flushFunc func() error) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().Msg("Flushing metrics")

		done := make(chan error, 1)
		go func() {
			done <- flushFunc()
		}()

		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("failed to flush metrics: %w", err)
			}
			log.Info().Msg("Metrics flushed successfully")
			return nil
		case <-ctx.Done():
			return fmt.Errorf("metrics flush timeout")
		}
	}
}

// ConnectionDraining drains active connections before shutdown
func ConnectionDraining(activeConnections *int64, checkInterval time.Duration, maxWait time.Duration) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().
			Int64("active_connections", *activeConnections).
			Msg("Draining active connections")

		startTime := time.Now()
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if *activeConnections == 0 {
					log.Info().
						Dur("duration", time.Since(startTime)).
						Msg("All connections drained")
					return nil
				}

				if time.Since(startTime) >= maxWait {
					log.Warn().
						Int64("remaining_connections", *activeConnections).
						Dur("duration", time.Since(startTime)).
						Msg("Connection draining timeout, forcing shutdown")
					return nil
				}

				log.Debug().
					Int64("active_connections", *activeConnections).
					Dur("elapsed", time.Since(startTime)).
					Msg("Waiting for connections to drain")

			case <-ctx.Done():
				return fmt.Errorf("connection draining cancelled")
			}
		}
	}
}

// HealthCheckDisable disables health checks during shutdown
func HealthCheckDisable(disableFunc func()) func(context.Context) error {
	return func(ctx context.Context) error {
		log.Info().Msg("Disabling health checks")
		disableFunc()
		log.Info().Msg("Health checks disabled")
		return nil
	}
}

// SetTimeout sets the global shutdown timeout
func (gs *GracefulShutdown) SetTimeout(timeout time.Duration) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.timeout = timeout
	log.Info().
		Dur("timeout", timeout).
		Msg("Shutdown timeout updated")
}

// SetSignals sets the signals to listen for
func (gs *GracefulShutdown) SetSignals(signals ...os.Signal) {
	gs.mu.Lock()
	defer gs.mu.Unlock()
	gs.signals = signals
}

// GetRegisteredFunctions returns the list of registered shutdown functions
func (gs *GracefulShutdown) GetRegisteredFunctions() []string {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	names := make([]string, len(gs.shutdownFuncs))
	for i, sf := range gs.shutdownFuncs {
		names[i] = sf.Name
	}
	return names
}
