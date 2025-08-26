package db

import (
	"context"
	"testing"

	"github.com/todo-app/services/admin-service/internal/config"
)

func TestNewConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration tests in short mode")
	}

	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "todo_app",
		User:     "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}

	conn, err := NewConnection(cfg)
	if err != nil {
		t.Skipf("Database not available for testing: %v", err)
	}
	defer conn.Close()

	if conn.DB == nil {
		t.Error("Expected database connection to be non-nil")
	}
}

func TestConnection_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration tests in short mode")
	}

	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "todo_app",
		User:     "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}

	conn, err := NewConnection(cfg)
	if err != nil {
		t.Skipf("Database not available for testing: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	err = conn.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() failed: %v", err)
	}
}

func TestConnection_SetServiceContext(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration tests in short mode")
	}

	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "todo_app",
		User:     "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}

	conn, err := NewConnection(cfg)
	if err != nil {
		t.Skipf("Database not available for testing: %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	err = conn.SetServiceContext(ctx, "test-service")
	if err != nil {
		t.Errorf("SetServiceContext() failed: %v", err)
	}
}

func TestDatabaseConfig_ConnectionString(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	actual := cfg.ConnectionString()

	if actual != expected {
		t.Errorf("ConnectionString() = %v, want %v", actual, expected)
	}
}