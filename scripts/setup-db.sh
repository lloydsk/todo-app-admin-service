#!/bin/bash

# Database Setup Script for Todo App Admin Service
# This script initializes the database with schema and seed data

set -e  # Exit on any error

# Default configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-todo_app}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

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

# Function to test database connection
test_connection() {
    print_status "Testing database connection..."
    
    if ! PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; then
        print_error "Cannot connect to database"
        print_status "Connection details:"
        echo "  Host: $DB_HOST"
        echo "  Port: $DB_PORT" 
        echo "  Database: $DB_NAME"
        echo "  User: $DB_USER"
        echo ""
        print_status "Please ensure:"
        echo "  - PostgreSQL is running"
        echo "  - Database '$DB_NAME' exists"
        echo "  - User '$DB_USER' has access"
        echo "  - Environment variables are set correctly"
        exit 1
    fi
    
    print_success "Database connection successful"
}

# Function to run SQL file
run_sql_file() {
    local file_path="$1"
    local description="$2"
    
    if [ ! -f "$file_path" ]; then
        print_error "SQL file not found: $file_path"
        exit 1
    fi
    
    print_status "$description"
    
    if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file_path" > /dev/null 2>&1; then
        print_success "$(basename "$file_path") applied successfully"
    else
        print_error "Failed to apply $(basename "$file_path")"
        print_status "Running with verbose output for debugging..."
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$file_path"
        exit 1
    fi
}

# Function to check if tables exist
check_tables_exist() {
    local tables_exist=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.tables 
        WHERE table_name IN ('users', 'tasks', 'categories', 'tags') 
        AND table_schema = 'public'
    " 2>/dev/null | xargs)
    
    echo "$tables_exist"
}

# Function to show database status
show_database_status() {
    print_status "Database Status:"
    echo ""
    
    # Table counts
    echo "Table record counts:"
    local tables=("users" "tasks" "categories" "tags" "task_categories" "task_tags" "task_history" "task_reminders" "user_sessions")
    
    for table in "${tables[@]}"; do
        local count=$(PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM $table;" 2>/dev/null | xargs || echo "0")
        printf "  %-20s %s\n" "${table}:" "${count}"
    done
    
    echo ""
    
    # Show sample users
    print_status "Sample users in database:"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT name, email, role, created_at 
        FROM users 
        WHERE NOT is_deleted 
        ORDER BY created_at 
        LIMIT 5;
    " 2>/dev/null || echo "  No users found"
}

# Main function
main() {
    echo "🗄️  Todo App Database Setup Script"
    echo "==================================="
    echo ""
    echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
    echo "User: $DB_USER"
    echo ""
    
    # Test database connection
    test_connection
    
    # Check if schema already exists
    local existing_tables=$(check_tables_exist)
    
    if [ "$existing_tables" -gt 0 ]; then
        print_warning "Found $existing_tables core tables already exist"
        if [ "$1" != "--force" ]; then
            print_status "Use --force to recreate the schema or --seed-only to only add seed data"
            echo ""
            echo "Options:"
            echo "  --force      Drop and recreate all tables (destructive)"
            echo "  --seed-only  Only run seed data (preserve existing data)"  
            echo "  --status     Show current database status"
            exit 1
        else
            print_warning "Forcing schema recreation (this will delete all data)"
            print_status "Dropping existing tables..."
            PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
                DROP SCHEMA public CASCADE;
                CREATE SCHEMA public;
                GRANT ALL ON SCHEMA public TO $DB_USER;
                GRANT ALL ON SCHEMA public TO public;
            " > /dev/null 2>&1
            print_success "Schema dropped and recreated"
        fi
    fi
    
    # Handle different modes
    case "${1:-setup}" in
        "--status")
            show_database_status
            exit 0
            ;;
        "--seed-only")
            print_status "Running seed data only..."
            if [ "$existing_tables" -eq 0 ]; then
                print_error "No tables found. Run without --seed-only first to create schema."
                exit 1
            fi
            ;;
        "--force"|"setup"|"")
            print_status "Setting up database schema and seed data..."
            
            # Apply migrations
            print_status "Applying database migrations..."
            for migration_file in database/migrations/*.sql; do
                if [ -f "$migration_file" ]; then
                    run_sql_file "$migration_file" "Applying migration: $(basename "$migration_file")"
                fi
            done
            
            print_success "All migrations applied"
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Usage: $0 [--force|--seed-only|--status]"
            exit 1
            ;;
    esac
    
    # Apply seed data (unless it's status check)
    if [ "$1" != "--status" ]; then
        print_status "Applying seed data..."
        for seed_file in database/seeds/*.sql; do
            if [ -f "$seed_file" ]; then
                run_sql_file "$seed_file" "Applying seed data: $(basename "$seed_file")"
            fi
        done
        
        print_success "All seed data applied"
    fi
    
    echo ""
    print_success "Database setup complete!"
    echo ""
    
    # Show final status
    show_database_status
}

# Run main function with all arguments
main "$@"