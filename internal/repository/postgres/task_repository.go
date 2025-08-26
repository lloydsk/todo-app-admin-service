package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
)

type taskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *sql.DB) repository.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	task.Version = 1

	if err := task.IsValid(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	query := `
		INSERT INTO tasks (id, title, description, assignee_id, status, priority, due_date, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.AssigneeID,
		string(task.Status), string(task.Priority), task.DueDate,
		task.CreatedAt, task.UpdatedAt, task.Version)

	return err
}

func (r *taskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT t.id, t.title, t.description, t.assignee_id, t.status, t.priority, t.due_date,
			   t.created_at, t.updated_at, t.version, t.is_deleted, t.deleted_at
		FROM tasks t
		WHERE t.id = $1 AND t.is_deleted = false`

	task := &domain.Task{}
	var status, priority string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.Title, &task.Description, &task.AssigneeID,
		&status, &priority, &task.DueDate,
		&task.CreatedAt, &task.UpdatedAt, &task.Version,
		&task.IsDeleted, &task.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound("task")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	task.Status = domain.TaskStatus(status)
	task.Priority = domain.TaskPriority(priority)

	// Load related data
	if err := r.loadTaskRelations(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to load task relations: %w", err)
	}

	return task, nil
}

func (r *taskRepository) List(ctx context.Context, opts repository.TaskListOptions) ([]*domain.Task, int64, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 0

	// Base condition for soft deletes
	if !opts.IncludeDeleted {
		conditions = append(conditions, "t.is_deleted = false")
	}

	// Add filters
	if opts.AssigneeID != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("t.assignee_id = $%d", argIndex))
		args = append(args, opts.AssigneeID)
	}

	if opts.Status != "" && opts.Status != domain.TaskStatusUnspecified {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, string(opts.Status))
	}

	if opts.Priority != "" && opts.Priority != domain.TaskPriorityUnspecified {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argIndex))
		args = append(args, string(opts.Priority))
	}

	if opts.SearchQuery != "" {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("(t.title ILIKE $%d OR t.description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+opts.SearchQuery+"%")
	}

	if opts.DueBefore != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("t.due_date <= $%d", argIndex))
		args = append(args, *opts.DueBefore)
	}

	if opts.DueAfter != nil {
		argIndex++
		conditions = append(conditions, fmt.Sprintf("t.due_date >= $%d", argIndex))
		args = append(args, *opts.DueAfter)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Build ORDER BY clause
	orderClause := "ORDER BY t.created_at DESC"
	if opts.SortBy != "" {
		direction := "ASC"
		if opts.SortDesc {
			direction = "DESC"
		}
		orderClause = fmt.Sprintf("ORDER BY t.%s %s", opts.SortBy, direction)
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks t %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Calculate pagination
	pageSize := opts.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	offset := opts.Page * pageSize

	// Build main query
	query := fmt.Sprintf(`
		SELECT t.id, t.title, t.description, t.assignee_id, t.status, t.priority, t.due_date,
			   t.created_at, t.updated_at, t.version, t.is_deleted, t.deleted_at
		FROM tasks t
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, argIndex+1, argIndex+2)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		var status, priority string

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.AssigneeID,
			&status, &priority, &task.DueDate,
			&task.CreatedAt, &task.UpdatedAt, &task.Version,
			&task.IsDeleted, &task.DeletedAt)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan task: %w", err)
		}

		task.Status = domain.TaskStatus(status)
		task.Priority = domain.TaskPriority(priority)

		// Load related data for each task
		if err := r.loadTaskRelations(ctx, task); err != nil {
			return nil, 0, fmt.Errorf("failed to load task relations: %w", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *domain.Task) error {
	if err := task.IsValid(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	query := `
		UPDATE tasks 
		SET title = $2, description = $3, assignee_id = $4, status = $5, priority = $6, 
			due_date = $7, updated_at = NOW()
		WHERE id = $1 AND version = $8 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.AssigneeID,
		string(task.Status), string(task.Priority), task.DueDate, task.Version)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("task", task.Version, task.Version+1)
	}

	// The database trigger handles version increment, so we need to fetch the updated version
	// or increment in memory to stay in sync with the database
	task.Version++
	task.UpdatedAt = time.Now()

	return nil
}

func (r *taskRepository) SoftDelete(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE tasks 
		SET is_deleted = true, deleted_at = NOW()
		WHERE id = $1 AND version = $2 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to soft delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("task", version, version+1)
	}

	return nil
}

func (r *taskRepository) Restore(ctx context.Context, id string, version int64) error {
	query := `
		UPDATE tasks 
		SET is_deleted = false, deleted_at = NULL, version = version + 1
		WHERE id = $1 AND version = $2 AND is_deleted = true`

	result, err := r.db.ExecContext(ctx, query, id, version)
	if err != nil {
		return fmt.Errorf("failed to restore task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrVersionConflict("task", version, version+1)
	}

	return nil
}

func (r *taskRepository) AddCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error {
	if len(categoryIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify task exists and version matches
	var currentVersion int64
	err = tx.QueryRowContext(ctx, "SELECT version FROM tasks WHERE id = $1 AND is_deleted = false", taskID).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to verify task: %w", err)
	}

	if currentVersion != version {
		return domain.ErrVersionConflict("task", version, currentVersion)
	}

	// Use ON CONFLICT to handle duplicates
	query := `
		INSERT INTO task_categories (task_id, category_id)
		VALUES ($1, unnest($2::uuid[]))
		ON CONFLICT (task_id, category_id) DO NOTHING`

	_, err = tx.ExecContext(ctx, query, taskID, pq.Array(categoryIDs))
	if err != nil {
		return fmt.Errorf("failed to add categories: %w", err)
	}

	// Update task version (trigger will handle this)
	_, err = tx.ExecContext(ctx, "UPDATE tasks SET updated_at = NOW() WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to update task timestamp: %w", err)
	}

	return tx.Commit()
}

func (r *taskRepository) RemoveCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error {
	if len(categoryIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify task exists and version matches
	var currentVersion int64
	err = tx.QueryRowContext(ctx, "SELECT version FROM tasks WHERE id = $1 AND is_deleted = false", taskID).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to verify task: %w", err)
	}

	if currentVersion != version {
		return domain.ErrVersionConflict("task", version, currentVersion)
	}

	// Remove category associations
	query := `DELETE FROM task_categories WHERE task_id = $1 AND category_id = ANY($2)`
	_, err = tx.ExecContext(ctx, query, taskID, pq.Array(categoryIDs))
	if err != nil {
		return fmt.Errorf("failed to remove categories: %w", err)
	}

	// Update task version (trigger will handle this)
	_, err = tx.ExecContext(ctx, "UPDATE tasks SET updated_at = NOW() WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to update task timestamp: %w", err)
	}

	return tx.Commit()
}

func (r *taskRepository) AssignTags(ctx context.Context, taskID string, tagIDs []string) error {
	if len(tagIDs) == 0 {
		return nil
	}

	query := `
		INSERT INTO task_tags (task_id, tag_id)
		VALUES ($1, unnest($2::uuid[]))
		ON CONFLICT (task_id, tag_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, taskID, pq.Array(tagIDs))
	return err
}

func (r *taskRepository) RemoveTags(ctx context.Context, taskID string, tagIDs []string, version int64) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify task exists and version matches
	var currentVersion int64
	err = tx.QueryRowContext(ctx, "SELECT version FROM tasks WHERE id = $1 AND is_deleted = false", taskID).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to verify task: %w", err)
	}

	if currentVersion != version {
		return domain.ErrVersionConflict("task", version, currentVersion)
	}

	query := `DELETE FROM task_tags WHERE task_id = $1 AND tag_id = ANY($2)`
	_, err = tx.ExecContext(ctx, query, taskID, pq.Array(tagIDs))
	if err != nil {
		return fmt.Errorf("failed to remove tags: %w", err)
	}

	// Update task version (trigger will handle this)
	_, err = tx.ExecContext(ctx, "UPDATE tasks SET updated_at = NOW() WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to update task timestamp: %w", err)
	}

	return tx.Commit()
}

// AddTags adds tags to a task - using AssignTags as AddTags
func (r *taskRepository) AddTags(ctx context.Context, taskID string, tagIDs []string, version int64) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Verify task exists and version matches
	var currentVersion int64
	err = tx.QueryRowContext(ctx, "SELECT version FROM tasks WHERE id = $1 AND is_deleted = false", taskID).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to verify task: %w", err)
	}

	if currentVersion != version {
		return domain.ErrVersionConflict("task", version, currentVersion)
	}

	// Use ON CONFLICT to handle duplicates
	query := `
		INSERT INTO task_tags (task_id, tag_id)
		VALUES ($1, unnest($2::uuid[]))
		ON CONFLICT (task_id, tag_id) DO NOTHING`

	_, err = tx.ExecContext(ctx, query, taskID, pq.Array(tagIDs))
	if err != nil {
		return fmt.Errorf("failed to add tags: %w", err)
	}

	// Update task version (trigger will handle this)
	_, err = tx.ExecContext(ctx, "UPDATE tasks SET updated_at = NOW() WHERE id = $1", taskID)
	if err != nil {
		return fmt.Errorf("failed to update task timestamp: %w", err)
	}

	return tx.Commit()
}

// GetHistory gets the history of a task
func (r *taskRepository) GetHistory(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	query := `
		SELECT th.id, th.task_id, th.action, th.actor_id, th.timestamp, th.details,
		       u.id, u.name, u.email, u.role
		FROM task_history th
		LEFT JOIN users u ON th.actor_id = u.id
		WHERE th.task_id = $1
		ORDER BY th.timestamp DESC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query task history: %w", err)
	}
	defer rows.Close()

	var history []*domain.TaskHistory
	for rows.Next() {
		h := &domain.TaskHistory{}
		var userID, userName, userEmail string
		var userRole string

		err := rows.Scan(
			&h.ID, &h.TaskID, &h.Action, &h.ActorID, &h.Timestamp, &h.Details,
			&userID, &userName, &userEmail, &userRole)

		if err != nil {
			return nil, fmt.Errorf("failed to scan task history: %w", err)
		}

		// Create actor if data exists
		if userID != "" {
			h.Actor = &domain.User{
				ID:    userID,
				Name:  userName,
				Email: userEmail,
				Role:  domain.UserRole(userRole),
			}
		}

		history = append(history, h)
	}

	return history, nil
}

// loadTaskRelations loads categories, tags, history, and reminders for a task
func (r *taskRepository) loadTaskRelations(ctx context.Context, task *domain.Task) error {
	// Load categories
	categoryQuery := `
		SELECT c.id, c.name, c.description, c.color, c.parent_id, c.is_public, c.creator_id,
			   c.created_at, c.updated_at, c.version, c.is_deleted, c.deleted_at
		FROM categories c
		INNER JOIN task_categories tc ON c.id = tc.category_id
		WHERE tc.task_id = $1 AND c.is_deleted = false`

	categoryRows, err := r.db.QueryContext(ctx, categoryQuery, task.ID)
	if err != nil {
		return fmt.Errorf("failed to load categories: %w", err)
	}
	defer categoryRows.Close()

	for categoryRows.Next() {
		category := domain.Category{}
		err := categoryRows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Color,
			&category.ParentID, &category.IsPublic, &category.CreatorID,
			&category.CreatedAt, &category.UpdatedAt, &category.Version,
			&category.IsDeleted, &category.DeletedAt)
		if err != nil {
			return fmt.Errorf("failed to scan category: %w", err)
		}
		task.Categories = append(task.Categories, category)
	}

	// Load tags
	tagQuery := `
		SELECT t.id, t.name, t.color, t.creator_id, t.created_at, t.updated_at, 
			   t.version, t.is_deleted, t.deleted_at
		FROM tags t
		INNER JOIN task_tags tt ON t.id = tt.tag_id
		WHERE tt.task_id = $1 AND t.is_deleted = false`

	tagRows, err := r.db.QueryContext(ctx, tagQuery, task.ID)
	if err != nil {
		return fmt.Errorf("failed to load tags: %w", err)
	}
	defer tagRows.Close()

	for tagRows.Next() {
		tag := domain.Tag{}
		err := tagRows.Scan(
			&tag.ID, &tag.Name, &tag.Color, &tag.CreatorID,
			&tag.CreatedAt, &tag.UpdatedAt, &tag.Version,
			&tag.IsDeleted, &tag.DeletedAt)
		if err != nil {
			return fmt.Errorf("failed to scan tag: %w", err)
		}
		task.Tags = append(task.Tags, tag)
	}

	// Load history
	historyQuery := `
		SELECT id, task_id, action, actor_id, service_name, timestamp, details
		FROM task_history
		WHERE task_id = $1
		ORDER BY timestamp DESC`

	historyRows, err := r.db.QueryContext(ctx, historyQuery, task.ID)
	if err != nil {
		return fmt.Errorf("failed to load history: %w", err)
	}
	defer historyRows.Close()

	for historyRows.Next() {
		entry := domain.TaskHistoryEntry{}
		err := historyRows.Scan(
			&entry.ID, &entry.TaskID, &entry.Action, &entry.ActorID,
			&entry.ServiceName, &entry.Timestamp, &entry.Details)
		if err != nil {
			return fmt.Errorf("failed to scan history entry: %w", err)
		}
		task.History = append(task.History, entry)
	}

	return nil
}
