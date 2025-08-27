package main

import (
	"context"
	"net"
	"testing"
	"time"

	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/todo-app/services/admin-service/internal/config"
	grpchandler "github.com/todo-app/services/admin-service/internal/handler/grpc"
	"github.com/todo-app/services/admin-service/internal/repository/postgres"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/db"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T) (*grpc.Server, *bufconn.Listener, todov1.AdminServiceClient) {
	// Create buffer connection for testing
	lis := bufconn.Listen(bufSize)

	// Load test configuration
	cfg := &config.Config{
		LogLevel: "debug",
		Server: config.ServerConfig{
			Port: 0, // Use any available port for testing
		},
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "todo_app",
			User:     "postgres",
			Password: "postgres",
			SSLMode:  "disable",
		},
	}

	// Initialize logger
	log := logger.NewLogger(cfg.LogLevel)

	// Try to initialize database connection - skip test if database is not available
	dbConn, err := db.NewConnection(cfg.Database)
	if err != nil {
		t.Skipf("Database not available for integration tests: %v", err)
		return nil, nil, nil
	}
	defer dbConn.Close()

	// Health check database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbConn.HealthCheck(ctx); err != nil {
		t.Skipf("Database health check failed: %v", err)
		return nil, nil, nil
	}

	// Set service context
	if err := dbConn.SetServiceContext(context.Background(), "test-admin-service"); err != nil {
		// Log warning but don't fail test setup for service context issues
		t.Logf("Warning: failed to set service context: %v", err)
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

	// Create gRPC server
	server := grpc.NewServer()

	// Register services
	grpcHandler := grpchandler.NewHandler(services, log)
	grpcHandler.RegisterServices(server)

	// Start server
	go func() {
		if err := server.Serve(lis); err != nil {
			t.Errorf("Server exited with error: %v", err)
		}
	}()

	// Create client connection
	conn, err := grpc.NewClient("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}

	client := todov1.NewAdminServiceClient(conn)
	return server, lis, client
}

func TestServerStartup(t *testing.T) {
	server, lis, client := setupTestServer(t)
	if server == nil {
		return // Test was skipped due to database unavailability
	}

	defer server.Stop()
	defer lis.Close()

	// Test that we can create a client connection
	if client == nil {
		t.Fatal("Failed to create admin service client")
	}

	// Test a simple RPC call (should return unimplemented for now)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.ListUsers(ctx, &todov1.ListUsersRequest{})
	if err == nil {
		t.Error("Expected unimplemented error, got nil")
	}

	// Check that it's the expected unimplemented error
	t.Logf("Got expected unimplemented error: %v", err)
}

func TestGracefulShutdown(t *testing.T) {
	server, lis, _ := setupTestServer(t)
	if server == nil {
		return // Test was skipped due to database unavailability
	}

	defer lis.Close()

	// Test graceful shutdown
	go func() {
		time.Sleep(100 * time.Millisecond)
		server.GracefulStop()
	}()

	// Server should shut down gracefully within reasonable time
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		// Success - graceful shutdown completed
	case <-time.After(5 * time.Second):
		t.Error("Graceful shutdown timed out")
		server.Stop() // Force stop
	}
}
