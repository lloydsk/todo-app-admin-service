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

type tagRepository struct {
	db *sql.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *sql.DB) repository.TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}

	now := time.Now()
	tag.CreatedAt = now
	tag.UpdatedAt = now
	tag.Version = 1

	if err := tag.IsValid(); err != nil {
		return fmt.Errorf("invalid tag: %w", err)
	}

	query := `
		INSERT INTO tags (id, name, color, creator_id, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		tag.ID, tag.Name, tag.Color, tag.CreatorID,
		tag.CreatedAt, tag.UpdatedAt, tag.Version)

	return err
}

func (r *tagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	query := `
		SELECT id, name, color, creator_id, created_at, updated_at, version, is_deleted, deleted_at
		FROM tags
		WHERE id = $1 AND is_deleted = false`

	tag := &domain.Tag{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tag.ID, &tag.Name, &tag.Color, &tag.CreatorID,
		&tag.CreatedAt, &tag.UpdatedAt, &tag.Version,
		&tag.IsDeleted, &tag.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound("tag")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

func (r *tagRepository) List(ctx context.Context, opts repository.TagListOptions) ([]*domain.Tag, int64, error) {
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
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+opts.SearchQuery+"%")
	}

	// Add creator filter
	if opts.CreatorID != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("creator_id = $%d", argIndex))
		args = append(args, opts.CreatorID)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tags %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tags: %w", err)
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
		SELECT id, name, color, creator_id, created_at, updated_at, version, is_deleted, deleted_at
		FROM tags
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIndex+1, argIndex+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.Tag
	for rows.Next() {
		tag := &domain.Tag{}

		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Color, &tag.CreatorID,
			&tag.CreatedAt, &tag.UpdatedAt, &tag.Version,
			&tag.IsDeleted, &tag.DeletedAt)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tag: %w", err)
		}

		tags = append(tags, tag)
	}

	return tags, total, nil
}

func (r *tagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	if err := tag.IsValid(); err != nil {
		return fmt.Errorf("invalid tag: %w", err)
	}

	query := `
		UPDATE tags 
		SET name = $2, color = $3, updated_at = NOW()
		WHERE id = $1 AND version = $4 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query,
		tag.ID, tag.Name, tag.Color, tag.Version)

	if err != nil {
		return fmt.Errorf("failed to update tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("tag", tag.Version, tag.Version+1)
	}

	// Update version in memory
	tag.Version++
	tag.UpdatedAt = time.Now()

	return nil
}

func (r *tagRepository) SoftDelete(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE tags 
		SET is_deleted = true, deleted_at = NOW()
		WHERE id = $1 AND version = $2 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to soft delete tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("tag", version, version+1)
	}

	return nil
}

func (r *tagRepository) Restore(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE tags 
		SET is_deleted = false, deleted_at = NULL
		WHERE id = $1 AND version = $2 AND is_deleted = true`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to restore tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("tag", version, version+1)
	}

	return nil
}
