package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/siddarth/go-app/internal/config"
	httphandler "github.com/siddarth/go-app/internal/handler/http"
	"github.com/siddarth/go-app/internal/middleware"
	"github.com/siddarth/go-app/internal/repository/memory"
	"github.com/siddarth/go-app/internal/service"
	"github.com/siddarth/go-app/pkg/observability"
)

func main() {
	// Run application
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger := observability.NewLogger(cfg.Logging)
	logger.Info("starting pacman game server",
		"service", cfg.Observability.ServiceName,
		"version", cfg.Observability.ServiceVersion,
		"environment", cfg.Observability.Environment,
	)

	// Initialize tracing
	shutdownTracing, err := observability.InitTracing(ctx, cfg.Observability)
	if err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}
	defer func() {
		if err := shutdownTracing(ctx); err != nil {
			logger.Error("failed to shutdown tracing", "error", err)
		}
	}()

	// Initialize dependencies
	gameRepo := memory.NewGameRepository()
	gameService := service.NewGameService(gameRepo, logger)
	gameHandler := httphandler.NewGameHandler(gameService, logger)

	// Setup Gin router
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()

	// Register middleware
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logging(logger))
	r.Use(middleware.CORS())
	r.Use(middleware.Tracing(cfg.Observability.ServiceName))

	// Register routes
	gameHandler.RegisterRoutes(r)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server listening",
			"port", cfg.Server.Port,
			"url", fmt.Sprintf("http://localhost:%s", cfg.Server.Port),
		)
		fmt.Printf("🎮 Starting Pacman Game Server on http://localhost:%s\n", cfg.Server.Port)
		fmt.Printf("Open your browser and navigate to http://localhost:%s\n", cfg.Server.Port)

		serverErrors <- srv.ListenAndServe()
	}()

	// Listen for shutdown signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Block until shutdown signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig.String())

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		logger.Info("shutting down server gracefully")
		if err := srv.Shutdown(shutdownCtx); err != nil {
			// Force close if graceful shutdown fails
			logger.Error("forcing server shutdown", "error", err)
			if err := srv.Close(); err != nil {
				return fmt.Errorf("failed to close server: %w", err)
			}
		}

		logger.Info("server shutdown complete")
	}

	return nil
}
