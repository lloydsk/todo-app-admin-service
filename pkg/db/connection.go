package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver

	"github.com/todo-app/services/admin-service/internal/config"
)

// Connection represents a database connection
type Connection struct {
	DB *sql.DB
}

// NewConnection creates a new database connection
func NewConnection(cfg config.DatabaseConfig) (*Connection, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Connection{DB: db}, nil
}

// Close closes the database connection
func (c *Connection) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// HealthCheck performs a health check on the database
func (c *Connection) HealthCheck(ctx context.Context) error {
	return c.DB.PingContext(ctx)
}

// SetServiceContext sets the service name for audit trails
func (c *Connection) SetServiceContext(ctx context.Context, serviceName string) error {
	_, err := c.DB.ExecContext(ctx, "SELECT set_config('app.service_name', $1, false)", serviceName)
	return err
}

// BeginTx starts a new transaction
func (c *Connection) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return c.DB.BeginTx(ctx, nil)
}

// Stats returns database connection pool statistics
func (c *Connection) Stats() sql.DBStats {
	return c.DB.Stats()
}
