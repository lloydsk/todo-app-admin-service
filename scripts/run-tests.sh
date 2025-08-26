#!/bin/bash

# Test Runner Script for Todo App Admin Service
# This script provides a unified way to run different types of tests with proper database setup

set -e  # Exit on any error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="$(dirname "$SCRIPT_DIR")"
RESET_SCRIPT="$SCRIPT_DIR/reset-db.sh"

# Function to print colored output
print_header() {
    echo -e "${BOLD}${CYAN}$1${NC}"
    echo -e "${CYAN}$(printf '=%.0s' {1..60})${NC}"
}

print_status() {
    echo -e "${BLUE}🔄 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_test_result() {
    if [ $1 -eq 0 ]; then
        print_success "$2"
    else
        print_error "$2"
        return $1
    fi
}

# Function to run tests with proper setup and cleanup
run_test_suite() {
    local test_name="$1"
    local test_command="$2"
    local cleanup_level="${3:-clean}"  # clean, reset, or none
    
    print_header "$test_name"
    
    # Pre-test database setup
    if [ "$cleanup_level" != "none" ]; then
        print_status "Setting up clean database state..."
        "$RESET_SCRIPT" "$cleanup_level"
        echo ""
    fi
    
    # Run the test
    print_status "Running $test_name..."
    cd "$SERVICE_DIR"
    
    local start_time=$(date +%s)
    if eval "$test_command"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_success "$test_name completed successfully in ${duration}s"
        return 0
    else
        local exit_code=$?
        print_error "$test_name failed with exit code $exit_code"
        return $exit_code
    fi
}

# Function to run all tests in sequence
run_all_tests() {
    local failed_tests=()
    local total_start_time=$(date +%s)
    
    print_header "🧪 Todo App Admin Service - Full Test Suite"
    echo ""
    
    # Test 1: Unit Tests (Core Modules)
    if ! run_test_suite "Unit Tests (Core)" "go test -v -short ./internal/config ./internal/model/domain ./pkg/logger" "none"; then
        failed_tests+=("Unit Tests (Core)")
    fi
    echo ""
    
    # Test 2: Unit Tests (Service Layer)
    if ! run_test_suite "Unit Tests (Service Layer)" "go test -v -short ./internal/service" "none"; then
        failed_tests+=("Unit Tests (Service Layer)")
    fi
    echo ""
    
    # Test 3: Integration Tests (Database)
    if ! run_test_suite "Integration Tests (Database)" "go test -v ./pkg/db ./internal/repository/postgres" "clean"; then
        failed_tests+=("Integration Tests (Database)")
    fi
    echo ""
    
    # Test 6: Build Verification
    if ! run_test_suite "Build Verification" "go build ./..." "none"; then
        failed_tests+=("Build Verification")
    fi
    echo ""
    
    # Summary
    local total_end_time=$(date +%s)
    local total_duration=$((total_end_time - total_start_time))
    
    print_header "🏁 Test Summary"
    
    if [ ${#failed_tests[@]} -eq 0 ]; then
        print_success "All tests passed! 🎉"
        print_status "Total execution time: ${total_duration}s"
        
        # Show final database state
        echo ""
        print_status "Final database state:"
        "$RESET_SCRIPT" status
        
        return 0
    else
        print_error "Some tests failed:"
        for test in "${failed_tests[@]}"; do
            echo -e "  ${RED}• $test${NC}"
        done
        print_status "Total execution time: ${total_duration}s"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_header "🔗 Integration Tests"
    
    # Start with a complete reset for integration tests
    print_status "Performing complete database reset for integration tests..."
    "$RESET_SCRIPT" reset
    echo ""
    
    # Run tests that depend on each other in sequence
    local failed_tests=()
    
    # Foundation -> Repository -> Service chain
    if ! run_test_suite "Foundation → Repository → Service Chain" \
        "go test -v -short ./internal/config ./internal/model/domain ./pkg/logger && go test -v ./internal/repository/postgres && go test -v -short ./internal/service" \
        "none"; then
        failed_tests+=("Integration Chain")
    fi
    
    if [ ${#failed_tests[@]} -eq 0 ]; then
        print_success "Integration tests passed! 🎉"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Function to run performance tests
run_performance_tests() {
    print_header "⚡ Performance Tests"
    
    print_status "Setting up clean database state..."
    "$RESET_SCRIPT" clean
    echo ""
    
    print_status "Running performance tests..."
    cd "$SERVICE_DIR"
    
    # Measure test execution times
    local start_time=$(date +%s)
    
    # Run each test suite and measure performance
    echo "Test Performance Results:"
    
    for test_suite in \
        "Foundation Tests:go run cmd/test/main.go" \
        "Repository Tests:go run cmd/test-repo/main.go" \
        "Service Tests:go run cmd/test-services/main.go"; do
        
        local name=$(echo "$test_suite" | cut -d: -f1)
        local command=$(echo "$test_suite" | cut -d: -f2-)
        
        local test_start=$(date +%s)
        if eval "$command" > /dev/null 2>&1; then
            local test_end=$(date +%s)
            local test_duration=$((test_end - test_start))
            printf "  %-25s %s\n" "$name:" "${test_duration}s ✅"
        else
            printf "  %-25s %s\n" "$name:" "FAILED ❌"
        fi
    done
    
    local total_end_time=$(date +%s)
    local total_duration=$((total_end_time - start_time))
    
    echo ""
    print_success "Performance test completed in ${total_duration}s"
}

# Function to show help
show_help() {
    echo "🧪 Todo App Test Runner"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  all, a           Run all test suites (default)"
    echo "  foundation, f    Run foundation tests only"
    echo "  repository, r    Run repository tests only"
    echo "  services, s      Run service layer tests only"
    echo "  integration, i   Run integration tests"
    echo "  performance, p   Run performance tests"
    echo "  build, b         Run build verification only"
    echo "  utils, u         Run test utilities validation"
    echo "  clean            Clean test data and show status"
    echo "  reset            Reset database completely"
    echo "  status           Show database status"
    echo "  help, h          Show this help message"
    echo ""
    echo "Options:"
    echo "  --no-reset       Skip database reset/cleanup"
    echo "  --reset-level    Set cleanup level: clean, reset, none (default: clean)"
    echo "  --verbose, -v    Enable verbose output"
    echo ""
    echo "Examples:"
    echo "  $0 all              # Run all tests with database cleanup"
    echo "  $0 services         # Run only service layer tests"
    echo "  $0 all --no-reset   # Run all tests without database cleanup"
    echo "  $0 foundation -v    # Run foundation tests with verbose output"
}

# Parse command line options
RESET_LEVEL="clean"
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --no-reset)
            RESET_LEVEL="none"
            shift
            ;;
        --reset-level)
            RESET_LEVEL="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        *)
            break
            ;;
    esac
done

# Main script logic
main() {
    local command="${1:-all}"
    
    # Ensure we're in the service directory
    cd "$SERVICE_DIR"
    
    case "$command" in
        "all"|"a")
            run_all_tests
            ;;
        "foundation"|"f")
            run_test_suite "Foundation Tests" "go test -v -short ./internal/config ./internal/model/domain ./pkg/logger" "none"
            ;;
        "repository"|"r")
            run_test_suite "Repository Tests" "go test -v ./internal/repository/postgres" "$RESET_LEVEL"
            ;;
        "services"|"s")
            run_test_suite "Service Layer Tests" "go test -v -short ./internal/service" "none"
            ;;
        "integration"|"i")
            run_integration_tests
            ;;
        "performance"|"p")
            run_performance_tests
            ;;
        "build"|"b")
            run_test_suite "Build Verification" "go build ./..." "none"
            ;;
        "utils"|"u")
            run_test_suite "Test Utilities Validation" "go run cmd/test-utils/main.go" "none"
            ;;
        "clean")
            "$RESET_SCRIPT" clean
            ;;
        "reset")
            "$RESET_SCRIPT" reset
            ;;
        "status")
            "$RESET_SCRIPT" status
            ;;
        "help"|"h"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            print_status "Use '$0 help' to see available commands"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"