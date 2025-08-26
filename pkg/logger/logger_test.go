package logger

import (
	"context"
	"testing"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{
			name:     "debug level",
			logLevel: "debug",
		},
		{
			name:     "info level",
			logLevel: "info",
		},
		{
			name:     "warn level",
			logLevel: "warn",
		},
		{
			name:     "error level",
			logLevel: "error",
		},
		{
			name:     "invalid level defaults to info",
			logLevel: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.logLevel)
			if logger == nil {
				t.Error("Expected logger to be non-nil")
			}
		})
	}
}

func TestLogger_LogMethods(t *testing.T) {
	logger := NewLogger("debug")
	ctx := context.Background()

	// Test that log methods don't panic
	t.Run("Debug", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Debug() panicked: %v", r)
			}
		}()
		logger.Debug(ctx, "test debug message", "key", "value")
	})

	t.Run("Info", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Info() panicked: %v", r)
			}
		}()
		logger.Info(ctx, "test info message", "key", "value")
	})

	t.Run("Warn", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Warn() panicked: %v", r)
			}
		}()
		logger.Warn(ctx, "test warn message", "key", "value")
	})

	t.Run("Error", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Error() panicked: %v", r)
			}
		}()
		logger.Error(ctx, "test error message", "error", "test error")
	})
}

func TestLogger_WithKeyValuePairs(t *testing.T) {
	logger := NewLogger("debug")
	ctx := context.Background()

	// Test logging with various key-value pairs
	tests := []struct {
		name     string
		keyvals  []interface{}
	}{
		{
			name:     "even number of arguments",
			keyvals:  []interface{}{"key1", "value1", "key2", "value2"},
		},
		{
			name:     "odd number of arguments - should handle gracefully",
			keyvals:  []interface{}{"key1", "value1", "orphan"},
		},
		{
			name:     "no key-value pairs",
			keyvals:  []interface{}{},
		},
		{
			name:     "mixed types",
			keyvals:  []interface{}{"string_key", "string_value", "int_key", 42, "bool_key", true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Logging with keyvals %v panicked: %v", tt.keyvals, r)
				}
			}()
			logger.Info(ctx, "test message", tt.keyvals...)
		})
	}
}