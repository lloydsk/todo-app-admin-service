package repository

import (
	"context"
	"database/sql"

	"github.com/todo-app/services/admin-service/internal/model/domain"
)

// UserRepository defines user data access operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, opts ListOptions) ([]*domain.User, int64, error)
	Update(ctx context.Context, user *domain.User) error
	SoftDelete(ctx context.Context, id string, version int64) error
	Restore(ctx context.Context, id string, version int64) error
}

// TaskRepository defines task data access operations
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id string) (*domain.Task, error)
	List(ctx context.Context, opts TaskListOptions) ([]*domain.Task, int64, error)
	Update(ctx context.Context, task *domain.Task) error
	SoftDelete(ctx context.Context, id string, version int64) error
	Restore(ctx context.Context, id string, version int64) error
	
	// Category associations
	AddCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error
	RemoveCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error
	
	// Tag associations
	AddTags(ctx context.Context, taskID string, tagIDs []string, version int64) error
	RemoveTags(ctx context.Context, taskID string, tagIDs []string, version int64) error
	
	// History
	GetHistory(ctx context.Context, taskID string) ([]*domain.TaskHistory, error)
}

// CategoryRepository defines category data access operations
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context, opts CategoryListOptions) ([]*domain.Category, int64, error)
	Update(ctx context.Context, category *domain.Category) error
	SoftDelete(ctx context.Context, id string, version int64) error
	Restore(ctx context.Context, id string, version int64) error
}

// TagRepository defines tag data access operations
type TagRepository interface {
	Create(ctx context.Context, tag *domain.Tag) error
	GetByID(ctx context.Context, id string) (*domain.Tag, error)
	List(ctx context.Context, opts TagListOptions) ([]*domain.Tag, int64, error)
	Update(ctx context.Context, tag *domain.Tag) error
	SoftDelete(ctx context.Context, id string, version int64) error
	Restore(ctx context.Context, id string, version int64) error
}

// TaskHistoryRepository defines task history operations
type TaskHistoryRepository interface {
	GetByTaskID(ctx context.Context, taskID string) ([]*domain.TaskHistoryEntry, error)
}

// TransactionManager defines transaction operations
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) error
}

// ListOptions defines common list query options
type ListOptions struct {
	Page         int32  `json:"page"`
	PageSize     int32  `json:"page_size"`
	SearchQuery  string `json:"search_query"`
	IncludeDeleted bool  `json:"include_deleted"`
	SortBy       string `json:"sort_by"`
	SortDesc     bool   `json:"sort_desc"`
}

// TaskListOptions defines task-specific list options
type TaskListOptions struct {
	ListOptions
	AssigneeID   string              `json:"assignee_id"`
	Status       domain.TaskStatus   `json:"status"`
	Priority     domain.TaskPriority `json:"priority"`
	CategoryIDs  []string            `json:"category_ids"`
	TagIDs       []string            `json:"tag_ids"`
	DueBefore    *string             `json:"due_before"` // ISO timestamp
	DueAfter     *string             `json:"due_after"`  // ISO timestamp
}

// CategoryListOptions defines category-specific list options
type CategoryListOptions struct {
	ListOptions
	ParentID   *string `json:"parent_id"`
	PublicOnly bool    `json:"public_only"`
	CreatorID  string  `json:"creator_id"`
}

// TagListOptions defines tag-specific list options
type TagListOptions struct {
	ListOptions
	CreatorID string `json:"creator_id"`
}

// Repositories aggregates all repository interfaces
type Repositories struct {
	Users       UserRepository
	Tasks       TaskRepository
	Categories  CategoryRepository
	Tags        TagRepository
	TaskHistory TaskHistoryRepository
	Transaction TransactionManager
}