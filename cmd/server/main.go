package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/todo-app/services/admin-service/internal/config"
	connecthandler "github.com/todo-app/services/admin-service/internal/handler/connect"
	"github.com/todo-app/services/admin-service/internal/repository/postgres"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/db"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger(cfg.LogLevel)
	log.Info(context.Background(), "Starting TODO Admin Service", "version", "1.0.0")

	// Initialize database connection
	dbConn, err := db.NewConnection(cfg.Database)
	if err != nil {
		log.Error(context.Background(), "Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	// Health check database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbConn.HealthCheck(ctx); err != nil {
		log.Error(context.Background(), "Database health check failed", "error", err)
		os.Exit(1)
	}

	// Set service context
	if err := dbConn.SetServiceContext(context.Background(), "admin-service"); err != nil {
		log.Warn(context.Background(), "Failed to set service context", "error", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(dbConn.DB)
	taskRepo := postgres.NewTaskRepository(dbConn.DB)
	categoryRepo := postgres.NewCategoryRepository(dbConn.DB)
	tagRepo := postgres.NewTagRepository(dbConn.DB)

	// Initialize services
	services := &service.Services{
		User:     service.NewUserService(userRepo, log),
		Task:     service.NewTaskService(taskRepo, userRepo, categoryRepo, tagRepo, log),
		Category: service.NewCategoryService(categoryRepo, taskRepo, log),
		Tag:      service.NewTagService(tagRepo, taskRepo, log),
	}

	// Initialize HTTP mux for ConnectRPC (supports both gRPC and HTTP/JSON)
	mux := http.NewServeMux()

	// Add a simple health check endpoint for debugging
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Error(r.Context(), "Failed to write health check response", "error", err)
		}
	})

	// Register ConnectRPC handlers
	connectHandler := connecthandler.NewHandler(services, log)
	connectHandler.RegisterServices(mux)

	// Create HTTP server that can handle both ConnectRPC protocols (gRPC and HTTP/JSON)
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	log.Info(context.Background(), "Starting server with ConnectRPC support (gRPC + HTTP/JSON)", "address", httpServer.Addr)

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Wait for shutdown signal
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Error(context.Background(), "Server error", "error", err)
		os.Exit(1)
	case sig := <-shutdown:
		log.Info(context.Background(), "Received shutdown signal", "signal", sig.String())

		// Graceful shutdown
		log.Info(context.Background(), "Starting graceful shutdown...")

		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Channel to receive shutdown completion
		shutdownComplete := make(chan struct{})

		go func() {
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				log.Error(context.Background(), "Server shutdown error", "error", err)
			}
			close(shutdownComplete)
		}()

		// Wait for graceful shutdown or timeout
		select {
		case <-shutdownComplete:
			log.Info(context.Background(), "Server shutdown completed")
		case <-shutdownCtx.Done():
			log.Warn(context.Background(), "Shutdown timeout exceeded, forcing stop")
			httpServer.Close()
		}
	}

	log.Info(context.Background(), "Server stopped")
}
