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

type categoryRepository struct {
	db *sql.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *sql.DB) repository.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	if category.ID == "" {
		category.ID = uuid.New().String()
	}

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now
	category.Version = 1

	if err := category.IsValid(); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	query := `
		INSERT INTO categories (id, name, description, color, parent_id, is_public, creator_id, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description, category.Color,
		category.ParentID, category.IsPublic, category.CreatorID,
		category.CreatedAt, category.UpdatedAt, category.Version)

	return err
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	query := `
		SELECT id, name, description, color, parent_id, is_public, creator_id,
		       created_at, updated_at, version, is_deleted, deleted_at
		FROM categories
		WHERE id = $1 AND is_deleted = false`

	category := &domain.Category{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.Color,
		&category.ParentID, &category.IsPublic, &category.CreatorID,
		&category.CreatedAt, &category.UpdatedAt, &category.Version,
		&category.IsDeleted, &category.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound("category")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

func (r *categoryRepository) List(ctx context.Context, opts repository.CategoryListOptions) ([]*domain.Category, int64, error) {
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
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+opts.SearchQuery+"%")
	}

	// Add parent filter
	if opts.ParentID != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("parent_id = $%d", argIndex))
		args = append(args, *opts.ParentID)
	}

	// Add creator filter
	if opts.CreatorID != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("creator_id = $%d", argIndex))
		args = append(args, opts.CreatorID)
	}

	// Add public filter
	if opts.PublicOnly {
		conditions = append(conditions, "is_public = true")
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM categories %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count categories: %w", err)
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
		SELECT id, name, description, color, parent_id, is_public, creator_id,
		       created_at, updated_at, version, is_deleted, deleted_at
		FROM categories
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIndex+1, argIndex+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		category := &domain.Category{}

		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Color,
			&category.ParentID, &category.IsPublic, &category.CreatorID,
			&category.CreatedAt, &category.UpdatedAt, &category.Version,
			&category.IsDeleted, &category.DeletedAt)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan category: %w", err)
		}

		categories = append(categories, category)
	}

	return categories, total, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	if err := category.IsValid(); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	query := `
		UPDATE categories 
		SET name = $2, description = $3, color = $4, parent_id = $5, is_public = $6, updated_at = NOW()
		WHERE id = $1 AND version = $7 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query,
		category.ID, category.Name, category.Description, category.Color,
		category.ParentID, category.IsPublic, category.Version)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("category", category.Version, category.Version+1)
	}

	// Update version in memory
	category.Version++
	category.UpdatedAt = time.Now()

	return nil
}

func (r *categoryRepository) SoftDelete(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE categories 
		SET is_deleted = true, deleted_at = NOW()
		WHERE id = $1 AND version = $2 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to soft delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("category", version, version+1)
	}

	return nil
}

func (r *categoryRepository) Restore(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE categories 
		SET is_deleted = false, deleted_at = NULL
		WHERE id = $1 AND version = $2 AND is_deleted = true`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to restore category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("category", version, version+1)
	}

	return nil
}
