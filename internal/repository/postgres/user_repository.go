package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.Version = 1

	if err := user.IsValid(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	query := `
		INSERT INTO users (id, name, email, role, password_hash, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Name, user.Email, string(user.Role),
		"dummy-hash", // In real implementation, this would be properly hashed
		user.CreatedAt, user.UpdatedAt, user.Version)

	return err
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, name, email, role, created_at, updated_at, version, is_deleted, deleted_at
		FROM users
		WHERE id = $1 AND is_deleted = false`

	user := &domain.User{}
	var role string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &role,
		&user.CreatedAt, &user.UpdatedAt, &user.Version,
		&user.IsDeleted, &user.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound("user")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user.Role = domain.UserRole(role)
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, name, email, role, created_at, updated_at, version, is_deleted, deleted_at
		FROM users
		WHERE email = $1 AND is_deleted = false`

	user := &domain.User{}
	var role string

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &role,
		&user.CreatedAt, &user.UpdatedAt, &user.Version,
		&user.IsDeleted, &user.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound("user")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	user.Role = domain.UserRole(role)
	return user, nil
}

func (r *userRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.User, int64, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 0

	// Base condition for soft deletes
	if !opts.IncludeDeleted {
		conditions = append(conditions, "is_deleted = false")
	}

	// Add search filter
	if opts.SearchQuery != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+opts.SearchQuery+"%")
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Build ORDER BY clause
	orderClause := "ORDER BY created_at DESC"
	if opts.SortBy != "" {
		direction := "ASC"
		if opts.SortDesc {
			direction = "DESC"
		}
		orderClause = fmt.Sprintf("ORDER BY %s %s", opts.SortBy, direction)
	}

	// Calculate pagination
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := opts.Page * pageSize

	// Build main query
	query := fmt.Sprintf(`
		SELECT id, name, email, role, created_at, updated_at, version, is_deleted, deleted_at
		FROM users
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIndex+1, argIndex+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		var role string

		err := rows.Scan(
			&user.ID, &user.Name, &user.Email, &role,
			&user.CreatedAt, &user.UpdatedAt, &user.Version,
			&user.IsDeleted, &user.DeletedAt)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		user.Role = domain.UserRole(role)
		users = append(users, user)
	}

	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := user.IsValid(); err != nil {
		return fmt.Errorf("invalid user: %w", err)
	}

	query := `
		UPDATE users 
		SET name = $2, email = $3, role = $4, updated_at = NOW()
		WHERE id = $1 AND version = $5 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Name, user.Email, string(user.Role), user.Version)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("user", user.Version, user.Version+1)
	}

	// Update version in memory
	user.Version++
	user.UpdatedAt = time.Now()

	return nil
}

func (r *userRepository) SoftDelete(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE users 
		SET is_deleted = true, deleted_at = NOW()
		WHERE id = $1 AND version = $2 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("user", version, version+1)
	}

	return nil
}

func (r *userRepository) Restore(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE users 
		SET is_deleted = false, deleted_at = NULL, version = version + 1
		WHERE id = $1 AND version = $2 AND is_deleted = true`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to restore user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("user", version, version+1)
	}

	return nil
}
