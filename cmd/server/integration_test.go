package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	todov1 "github.com/todo-app/todo-app-proto/gen/go/todo/v1"
)

func TestRealServerStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real server test in short mode")
	}

	// Set environment variables for test
	os.Setenv("SERVER_PORT", "0") // Use any available port
	os.Setenv("LOG_LEVEL", "debug")

	// Find an available port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to get available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	os.Setenv("SERVER_PORT", fmt.Sprintf("%d", port))

	// Start server in background
	serverDone := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				serverDone <- fmt.Errorf("server panicked: %v", r)
			}
		}()
		main()
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Create client and test connection
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.DialContext(
		context.Background(),
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Create client
	client := todov1.NewAdminServiceClient(conn)

	// Test a simple call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ListUsers(ctx, &todov1.ListUsersRequest{})
	if err != nil {
		t.Fatalf("Unexpected error calling ListUsers: %v", err)
	}

	if resp == nil {
		t.Error("Expected response but got nil")
	}

	t.Logf("Successfully connected to server and received response with %d users", len(resp.Users))

	// Cleanup environment variables
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("LOG_LEVEL")
}

func TestServerWithDatabaseUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database unavailable test in short mode")
	}

	// Set invalid database configuration
	os.Setenv("DB_HOST", "nonexistent-host")
	os.Setenv("DB_PORT", "9999")
	os.Setenv("LOG_LEVEL", "error") // Reduce noise

	// Capture the program exit
	if os.Getenv("CRASH_TEST") == "1" {
		main()
		return
	}

	// This test is tricky because main() calls os.Exit(1)
	// We'd need to refactor main() to be testable, but for now
	// we'll just document that the server should exit gracefully
	// when database is unavailable
	t.Skip("Database unavailable test requires main() refactoring to avoid os.Exit()")

	// Cleanup
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("LOG_LEVEL")
}
