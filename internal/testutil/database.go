package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/todo-app/services/admin-service/internal/config"
	"github.com/todo-app/services/admin-service/pkg/db"
)

// TestDBConfig returns a test database configuration
func TestDBConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Name:     "todo_app",
		User:     "postgres",
		Password: "postgres",
		SSLMode:  "disable",
	}
}

// SetupTestDB creates a test database connection and sets service context
func SetupTestDB(t *testing.T, serviceName string) *db.Connection {
	cfg := TestDBConfig()

	dbConn, err := db.NewConnection(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	ctx := context.Background()
	if err := dbConn.HealthCheck(ctx); err != nil {
		t.Fatalf("Test database health check failed: %v", err)
	}

	dbConn.SetServiceContext(ctx, serviceName)
	return dbConn
}

// CleanupTestDB closes the database connection
func CleanupTestDB(t *testing.T, dbConn *db.Connection) {
	if err := dbConn.Close(); err != nil {
		t.Logf("Warning: Failed to close test database connection: %v", err)
	}
}

// CleanupTestData removes test data from tables (useful for integration tests)
func CleanupTestData(t *testing.T, dbConn *db.Connection, tables ...string) {
	ctx := context.Background()

	// Default tables to clean if none specified
	if len(tables) == 0 {
		tables = []string{
			"task_history",
			"task_reminders",
			"task_tags",
			"task_categories",
			"tasks",
			"categories",
			"tags",
			"users",
		}
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE created_at > NOW() - INTERVAL '5 minutes'", table)
		_, err := dbConn.DB.ExecContext(ctx, query)
		if err != nil {
			t.Logf("Warning: Failed to cleanup test data from table %s: %v", table, err)
		}
	}
}

// ExecuteInTransaction executes a function within a database transaction
// Useful for test isolation
func ExecuteInTransaction(t *testing.T, dbConn *db.Connection, fn func(*sql.Tx) error) {
	ctx := context.Background()

	tx, err := dbConn.DB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Logf("Warning: Failed to rollback transaction: %v", err)
		}
	}()

	if err := fn(tx); err != nil {
		t.Fatalf("Transaction function failed: %v", err)
	}
}

// CountRecords counts records in a table with optional WHERE clause
func CountRecords(t *testing.T, dbConn *db.Connection, table string, whereClause string, args ...interface{}) int {
	ctx := context.Background()

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	var count int
	err := dbConn.DB.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count records in table %s: %v", table, err)
	}

	return count
}
