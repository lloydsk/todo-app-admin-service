package service

import (
	"context"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
)

// UserService defines the business logic for user operations
type UserService interface {
	// User CRUD operations
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUser(ctx context.Context, id string, version int64) error
	RestoreUser(ctx context.Context, id string, version int64) (*domain.User, error)
	ListUsers(ctx context.Context, opts repository.ListOptions) ([]*domain.User, int64, error)
	
	// Business logic methods
	ChangeUserRole(ctx context.Context, userID string, newRole domain.UserRole, version int64) (*domain.User, error)
	ValidateUserPermissions(ctx context.Context, userID string, requiredRole domain.UserRole) error
}

// TaskService defines the business logic for task operations
type TaskService interface {
	// Task CRUD operations
	CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error)
	GetTaskByID(ctx context.Context, id string) (*domain.Task, error)
	UpdateTask(ctx context.Context, task *domain.Task) (*domain.Task, error)
	DeleteTask(ctx context.Context, id string, version int64) error
	RestoreTask(ctx context.Context, id string, version int64) (*domain.Task, error)
	ListTasks(ctx context.Context, opts repository.TaskListOptions) ([]*domain.Task, int64, error)
	
	// Business logic methods
	AssignTask(ctx context.Context, taskID, assigneeID string, version int64) (*domain.Task, error)
	ChangeTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus, version int64) (*domain.Task, error)
	ChangeTaskPriority(ctx context.Context, taskID string, priority domain.TaskPriority, version int64) (*domain.Task, error)
	AddTaskCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) (*domain.Task, error)
	RemoveTaskCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) (*domain.Task, error)
	AddTaskTags(ctx context.Context, taskID string, tagIDs []string, version int64) (*domain.Task, error)
	RemoveTaskTags(ctx context.Context, taskID string, tagIDs []string, version int64) (*domain.Task, error)
	GetTaskHistory(ctx context.Context, taskID string) ([]*domain.TaskHistory, error)
}

// CategoryService defines the business logic for category operations
type CategoryService interface {
	// Category CRUD operations
	CreateCategory(ctx context.Context, category *domain.Category) (*domain.Category, error)
	GetCategoryByID(ctx context.Context, id string) (*domain.Category, error)
	UpdateCategory(ctx context.Context, category *domain.Category) (*domain.Category, error)
	DeleteCategory(ctx context.Context, id string, version int64) error
	RestoreCategory(ctx context.Context, id string, version int64) (*domain.Category, error)
	ListCategories(ctx context.Context, opts repository.ListOptions) ([]*domain.Category, int64, error)
	
	// Business logic methods
	ValidateCategoryUsage(ctx context.Context, categoryID string) error
	GetCategoryTaskCount(ctx context.Context, categoryID string) (int64, error)
}

// TagService defines the business logic for tag operations
type TagService interface {
	// Tag CRUD operations
	CreateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error)
	GetTagByID(ctx context.Context, id string) (*domain.Tag, error)
	UpdateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error)
	DeleteTag(ctx context.Context, id string, version int64) error
	RestoreTag(ctx context.Context, id string, version int64) (*domain.Tag, error)
	ListTags(ctx context.Context, opts repository.ListOptions) ([]*domain.Tag, int64, error)
	
	// Business logic methods
	ValidateTagUsage(ctx context.Context, tagID string) error
	GetTagTaskCount(ctx context.Context, tagID string) (int64, error)
	FindOrCreateTag(ctx context.Context, name string) (*domain.Tag, error)
}

// Services aggregates all service interfaces
type Services struct {
	User     UserService
	Task     TaskService
	Category CategoryService
	Tag      TagService
}