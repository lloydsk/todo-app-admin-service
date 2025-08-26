# Todo App Admin Service

A Go-based admin service for the distributed todo application, providing administrative operations for users, tasks, categories, and tags.

## Features

- **User Management**: CRUD operations for user accounts with role-based access control
- **Task Management**: Complete task lifecycle management with categories and tags
- **Category Management**: Hierarchical category organization
- **Tag Management**: Flexible tagging system with auto-creation
- **Soft Deletes**: All entities support soft deletion and restoration
- **Version Control**: Optimistic locking for concurrent updates
- **Comprehensive Testing**: Full unit and integration test coverage

## Architecture

```
├── cmd/                    # Application entrypoints
├── internal/
│   ├── config/            # Configuration management
│   ├── model/
│   │   └── domain/        # Domain models and business logic
│   ├── repository/        # Data access layer
│   │   └── postgres/      # PostgreSQL implementations
│   └── service/           # Business logic layer
├── pkg/
│   ├── db/                # Database connection utilities
│   └── logger/            # Logging utilities
├── proto/                 # Protocol buffer definitions
└── scripts/               # Utility scripts
```

## Tech Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL with soft deletes and versioning
- **Testing**: Go testing with integration tests
- **Logging**: Structured logging with slog
- **Architecture**: Clean Architecture with domain-driven design

## Getting Started

### Prerequisites

- Go 1.21 or later
- PostgreSQL 14+
- Docker (optional, for database)

### Database Setup

The service requires a PostgreSQL database. Use the provided scripts to set up the database:

```bash
# Reset and initialize the database
./scripts/reset-db.sh reset
```

### Running Tests

```bash
# Run all tests
./scripts/run-tests.sh all

# Run specific test suites
./scripts/run-tests.sh services      # Service layer tests
./scripts/run-tests.sh repository    # Integration tests
./scripts/run-tests.sh foundation    # Core/domain tests
```

### Test Coverage

- **TagService**: 6 test suites, 17 scenarios ✅
- **TagRepository**: 9 integration scenarios ✅  
- **TaskRepository**: 5 integration scenarios ✅
- **UserRepository**: 6 integration scenarios ✅
- **Domain Models**: Comprehensive validation testing ✅

## Development

### Project Status

- ✅ Domain models and business logic
- ✅ Repository layer with PostgreSQL
- ✅ Service layer with comprehensive business rules
- ✅ Complete test coverage (unit + integration)
- 🚧 gRPC API handlers (planned)
- 🚧 Main server implementation (planned)
- 🚧 Authentication & authorization (planned)

### Running the Service

*Note: Server implementation is in progress*

### API Documentation

*Coming soon: gRPC API documentation*

## Contributing

1. Follow Go conventions and maintain test coverage
2. Run the full test suite before committing
3. Use the provided scripts for database operations

## License

[Your License Here]