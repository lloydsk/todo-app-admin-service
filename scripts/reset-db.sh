#!/bin/bash

# Database Reset Script for Todo App Admin Service
# This script cleans and resets the database to a known state for testing

set -e  # Exit on any error

# Default configuration
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="todo_app"
DB_USER="postgres"
DB_PASSWORD="postgres"
CONTAINER_NAME="todo-postgres-dev"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to check if Docker container is running
check_container() {
    if ! docker ps --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        print_error "PostgreSQL container '${CONTAINER_NAME}' is not running"
        print_status "Starting PostgreSQL container..."
        docker start ${CONTAINER_NAME} 2>/dev/null || {
            print_error "Failed to start container. Make sure it exists."
            print_status "To create the container, run:"
            echo "docker run --name ${CONTAINER_NAME} -e POSTGRES_PASSWORD=${DB_PASSWORD} -p ${DB_PORT}:5432 -d postgres:15-alpine"
            exit 1
        }
        
        # Wait for PostgreSQL to be ready
        print_status "Waiting for PostgreSQL to be ready..."
        for i in {1..30}; do
            if docker exec ${CONTAINER_NAME} pg_isready -U ${DB_USER} >/dev/null 2>&1; then
                break
            fi
            sleep 1
        done
    fi
    print_success "PostgreSQL container is running"
}

# Function to execute SQL commands
execute_sql() {
    local sql_command="$1"
    local description="$2"
    
    if [ -n "$description" ]; then
        print_status "$description"
    fi
    
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -c "$sql_command" > /dev/null 2>&1
}

# Function to execute SQL file
execute_sql_file() {
    local file_path="$1"
    local description="$2"
    
    if [ -n "$description" ]; then
        print_status "$description"
    fi
    
    if [ ! -f "$file_path" ]; then
        print_error "SQL file not found: $file_path"
        exit 1
    fi
    
    docker exec -i ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} < "$file_path" > /dev/null 2>&1
}

# Function to clean test data (soft approach)
clean_test_data() {
    print_status "Cleaning test data (preserving seed data)..."
    
    # Delete test data based on patterns or recent timestamps
    # Use timestamp column for task_history, and created_at for other tables
    execute_sql "DELETE FROM task_history WHERE timestamp > NOW() - INTERVAL '1 hour';" "Cleaning recent task history"
    execute_sql "DELETE FROM task_reminders WHERE created_at > NOW() - INTERVAL '1 hour';" "Cleaning recent reminders"
    execute_sql "DELETE FROM task_tags WHERE task_id IN (SELECT id FROM tasks WHERE created_at > NOW() - INTERVAL '1 hour');" "Cleaning recent task-tag associations"
    execute_sql "DELETE FROM task_categories WHERE task_id IN (SELECT id FROM tasks WHERE created_at > NOW() - INTERVAL '1 hour');" "Cleaning recent task-category associations"
    execute_sql "DELETE FROM tasks WHERE created_at > NOW() - INTERVAL '1 hour';" "Cleaning recent tasks"
    execute_sql "DELETE FROM categories WHERE created_at > NOW() - INTERVAL '1 hour' AND name LIKE '%Test%';" "Cleaning test categories"
    execute_sql "DELETE FROM tags WHERE created_at > NOW() - INTERVAL '1 hour' AND name LIKE '%test%';" "Cleaning test tags"
    execute_sql "DELETE FROM users WHERE created_at > NOW() - INTERVAL '1 hour' AND email LIKE '%test%';" "Cleaning test users"
    
    print_success "Test data cleaned"
}

# Function to reset database completely (nuclear option)
reset_database() {
    print_warning "Performing complete database reset..."
    
    # Drop and recreate the database
    print_status "Dropping database..."
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -c "DROP DATABASE IF EXISTS ${DB_NAME};" > /dev/null 2>&1
    
    print_status "Creating database..."
    docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -c "CREATE DATABASE ${DB_NAME};" > /dev/null 2>&1
    
    # Apply migrations
    print_status "Applying database migrations..."
    local migration_dir="../database/migrations"
    
    if [ -d "$migration_dir" ]; then
        for migration_file in "$migration_dir"/*.sql; do
            if [ -f "$migration_file" ]; then
                local filename=$(basename "$migration_file")
                print_status "Applying migration: $filename"
                execute_sql_file "$migration_file"
            fi
        done
    else
        print_warning "Migration directory not found: $migration_dir"
        print_status "Looking for migration files in current directory..."
        for migration_file in *.sql; do
            if [ -f "$migration_file" ]; then
                local filename=$(basename "$migration_file")
                print_status "Applying migration: $filename"
                execute_sql_file "$migration_file"
            fi
        done
    fi
    
    # Apply seed data
    print_status "Applying seed data..."
    local seed_file="../database/seed.sql"
    if [ -f "$seed_file" ]; then
        execute_sql_file "$seed_file"
    else
        print_warning "Seed file not found: $seed_file"
        # Create basic seed data
        execute_sql "INSERT INTO users (id, name, email, role, created_at, updated_at, version) VALUES ('11111111-1111-1111-1111-111111111111', 'System Admin', 'admin@todo-app.com', 'admin', NOW(), NOW(), 1) ON CONFLICT (id) DO NOTHING;" "Creating system admin user"
    fi
    
    print_success "Database reset complete"
}

# Function to show database status
show_status() {
    print_status "Database Status:"
    echo ""
    
    # Count records in each table
    echo "Record counts:"
    for table in users tasks categories tags task_categories task_tags task_history task_reminders user_sessions; do
        count=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM ${table};" 2>/dev/null | xargs || echo "0")
        printf "  %-20s %s\n" "${table}:" "${count}"
    done
    
    echo ""
    echo "Recent test data (last hour):"
    recent_users=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '1 hour' AND email LIKE '%test%';" 2>/dev/null | xargs || echo "0")
    recent_tasks=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM tasks WHERE created_at > NOW() - INTERVAL '1 hour';" 2>/dev/null | xargs || echo "0")
    recent_categories=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM categories WHERE created_at > NOW() - INTERVAL '1 hour' AND name LIKE '%Test%';" 2>/dev/null | xargs || echo "0")
    recent_tags=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM tags WHERE created_at > NOW() - INTERVAL '1 hour' AND name LIKE '%test%';" 2>/dev/null | xargs || echo "0")
    recent_history=$(docker exec ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -t -c "SELECT COUNT(*) FROM task_history WHERE timestamp > NOW() - INTERVAL '1 hour';" 2>/dev/null | xargs || echo "0")
    
    printf "  %-20s %s\n" "Test users:" "${recent_users}"
    printf "  %-20s %s\n" "Recent tasks:" "${recent_tasks}"
    printf "  %-20s %s\n" "Test categories:" "${recent_categories}"
    printf "  %-20s %s\n" "Test tags:" "${recent_tags}"
    printf "  %-20s %s\n" "Recent history:" "${recent_history}"
}

# Main script logic
main() {
    echo "🗄️  Todo App Database Reset Script"
    echo "================================="
    echo ""
    
    # Parse command line arguments
    case "${1:-status}" in
        "clean"|"c")
            check_container
            clean_test_data
            show_status
            ;;
        "reset"|"r")
            check_container
            reset_database
            show_status
            ;;
        "status"|"s"|"")
            check_container
            show_status
            ;;
        "help"|"h"|"-h"|"--help")
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  clean, c     Clean test data (preserves seed data)"
            echo "  reset, r     Complete database reset (nuclear option)"
            echo "  status, s    Show database status (default)"
            echo "  help, h      Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  DB_HOST      Database host (default: localhost)"
            echo "  DB_PORT      Database port (default: 5432)"
            echo "  DB_NAME      Database name (default: todo_app)"
            echo "  DB_USER      Database user (default: postgres)"
            echo "  DB_PASSWORD  Database password (default: postgres)"
            echo "  CONTAINER_NAME Docker container name (default: todo-postgres-dev)"
            ;;
        *)
            print_error "Unknown command: $1"
            print_status "Use '$0 help' to see available commands"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"