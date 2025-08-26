package service

import (
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

// ServiceDependencies contains all the dependencies needed to create services
type ServiceDependencies struct {
	UserRepo     repository.UserRepository
	TaskRepo     repository.TaskRepository
	CategoryRepo repository.CategoryRepository
	TagRepo      repository.TagRepository
	Logger       logger.Logger
}

// NewServices creates a new Services instance with all service implementations
func NewServices(deps ServiceDependencies) *Services {
	userService := NewUserService(deps.UserRepo, deps.Logger)
	
	taskService := NewTaskService(
		deps.TaskRepo,
		deps.UserRepo,
		deps.CategoryRepo,
		deps.TagRepo,
		deps.Logger,
	)
	
	categoryService := NewCategoryService(
		deps.CategoryRepo,
		deps.TaskRepo,
		deps.Logger,
	)
	
	tagService := NewTagService(
		deps.TagRepo,
		deps.TaskRepo,
		deps.Logger,
	)
	
	return &Services{
		User:     userService,
		Task:     taskService,
		Category: categoryService,
		Tag:      tagService,
	}
}