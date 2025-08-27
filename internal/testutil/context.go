package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/todo-app/services/admin-service/internal/auth"
)

// Context key types to avoid collisions
type contextKey string

const (
	userIDKey      contextKey = "user_id"
	serviceNameKey contextKey = "service_name"
)

// TestContext creates a test context with timeout
func TestContext(t *testing.T, timeout ...time.Duration) context.Context {
	t.Helper()

	defaultTimeout := 30 * time.Second
	if len(timeout) > 0 {
		defaultTimeout = timeout[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)

	// Ensure context is canceled when test completes
	t.Cleanup(func() {
		cancel()
	})

	return ctx
}

// TestContextWithValues creates a test context with predefined values
func TestContextWithValues(t *testing.T, values map[string]interface{}, timeout ...time.Duration) context.Context {
	t.Helper()

	ctx := TestContext(t, timeout...)

	for key, value := range values {
		ctx = context.WithValue(ctx, contextKey(key), value)
	}

	return ctx
}

// TestContextWithUserID creates a test context with a user ID
func TestContextWithUserID(t *testing.T, userID string, timeout ...time.Duration) context.Context {
	t.Helper()

	ctx := TestContext(t, timeout...)
	return auth.WithUserID(ctx, userID)
}

// TestContextWithServiceName creates a test context with service name
func TestContextWithServiceName(t *testing.T, serviceName string, timeout ...time.Duration) context.Context {
	t.Helper()

	ctx := TestContext(t, timeout...)
	return context.WithValue(ctx, serviceNameKey, serviceName)
}
