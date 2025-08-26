package service

import (
	"context"
	"fmt"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type taskService struct {
	taskRepo     repository.TaskRepository
	userRepo     repository.UserRepository
	categoryRepo repository.CategoryRepository
	tagRepo      repository.TagRepository
	logger       logger.Logger
}

// NewTaskService creates a new task service
func NewTaskService(
	taskRepo repository.TaskRepository,
	userRepo repository.UserRepository,
	categoryRepo repository.CategoryRepository,
	tagRepo repository.TagRepository,
	log logger.Logger,
) TaskService {
	return &taskService{
		taskRepo:     taskRepo,
		userRepo:     userRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
		logger:       log,
	}
}

func (s *taskService) CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	s.logger.Info(ctx, "Creating new task", "title", task.Title, "assignee_id", task.AssigneeID)

	// Business validation
	if err := s.validateTaskForCreation(ctx, task); err != nil {
		return nil, err
	}

	// Validate assignee exists
	if task.AssigneeID != "" {
		if _, err := s.userRepo.GetByID(ctx, task.AssigneeID); err != nil {
			if domain.IsNotFoundError(err) {
				return nil, domain.ErrInvalidInput("assignee does not exist")
			}
			return nil, fmt.Errorf("failed to validate assignee: %w", err)
		}
	}

	// Create task
	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.logger.Error(ctx, "Failed to create task", "error", err, "title", task.Title)
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	s.logger.Info(ctx, "Task created successfully", "task_id", task.ID, "title", task.Title)
	return task, nil
}

func (s *taskService) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	s.logger.Debug(ctx, "Getting task by ID", "task_id", id)

	if id == "" {
		return nil, domain.ErrInvalidInput("task ID is required")
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if domain.IsNotFoundError(err) {
			s.logger.Debug(ctx, "Task not found", "task_id", id)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to get task by ID", "error", err, "task_id", id)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

func (s *taskService) UpdateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	s.logger.Info(ctx, "Updating task", "task_id", task.ID, "version", task.Version)

	// Business validation
	if err := s.validateTaskForUpdate(ctx, task); err != nil {
		return nil, err
	}

	// Validate assignee exists if changed
	if task.AssigneeID != "" {
		if _, err := s.userRepo.GetByID(ctx, task.AssigneeID); err != nil {
			if domain.IsNotFoundError(err) {
				return nil, domain.ErrInvalidInput("assignee does not exist")
			}
			return nil, fmt.Errorf("failed to validate assignee: %w", err)
		}
	}

	// Update task
	if err := s.taskRepo.Update(ctx, task); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Task update version conflict", "task_id", task.ID, "version", task.Version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to update task", "error", err, "task_id", task.ID)
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	s.logger.Info(ctx, "Task updated successfully", "task_id", task.ID, "new_version", task.Version)
	return task, nil
}

func (s *taskService) DeleteTask(ctx context.Context, id string, version int64) error {
	s.logger.Info(ctx, "Soft deleting task", "task_id", id, "version", version)

	if id == "" {
		return domain.ErrInvalidInput("task ID is required")
	}

	if err := s.taskRepo.SoftDelete(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Task deletion version conflict", "task_id", id, "version", version)
			return err
		}
		s.logger.Error(ctx, "Failed to delete task", "error", err, "task_id", id)
		return fmt.Errorf("failed to delete task: %w", err)
	}

	s.logger.Info(ctx, "Task deleted successfully", "task_id", id)
	return nil
}

func (s *taskService) RestoreTask(ctx context.Context, id string, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Restoring task", "task_id", id, "version", version)

	if id == "" {
		return nil, domain.ErrInvalidInput("task ID is required")
	}

	if err := s.taskRepo.Restore(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Task restoration version conflict", "task_id", id, "version", version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to restore task", "error", err, "task_id", id)
		return nil, fmt.Errorf("failed to restore task: %w", err)
	}

	// Get the restored task
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored task: %w", err)
	}

	s.logger.Info(ctx, "Task restored successfully", "task_id", id)
	return task, nil
}

func (s *taskService) ListTasks(ctx context.Context, opts repository.TaskListOptions) ([]*domain.Task, int64, error) {
	s.logger.Debug(ctx, "Listing tasks", "page", opts.Page, "page_size", opts.PageSize)

	tasks, total, err := s.taskRepo.List(ctx, opts)
	if err != nil {
		s.logger.Error(ctx, "Failed to list tasks", "error", err)
		return nil, 0, fmt.Errorf("failed to list tasks: %w", err)
	}

	s.logger.Debug(ctx, "Listed tasks successfully", "count", len(tasks), "total", total)
	return tasks, total, nil
}

func (s *taskService) AssignTask(ctx context.Context, taskID, assigneeID string, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Assigning task", "task_id", taskID, "assignee_id", assigneeID, "version", version)

	// Get current task
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Validate version
	if task.Version != version {
		return nil, domain.ErrVersionConflict("task", version, task.Version)
	}

	// Validate assignee exists
	if assigneeID != "" {
		if _, err := s.userRepo.GetByID(ctx, assigneeID); err != nil {
			if domain.IsNotFoundError(err) {
				return nil, domain.ErrInvalidInput("assignee does not exist")
			}
			return nil, fmt.Errorf("failed to validate assignee: %w", err)
		}
	}

	// Update assignment
	task.AssigneeID = assigneeID
	return s.UpdateTask(ctx, task)
}

func (s *taskService) ChangeTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Changing task status", "task_id", taskID, "status", status, "version", version)

	// Get current task
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Validate version
	if task.Version != version {
		return nil, domain.ErrVersionConflict("task", version, task.Version)
	}

	// Business validation for status change
	if err := s.validateStatusTransition(task.Status, status); err != nil {
		return nil, err
	}

	// Update status
	task.Status = status
	return s.UpdateTask(ctx, task)
}

func (s *taskService) ChangeTaskPriority(ctx context.Context, taskID string, priority domain.TaskPriority, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Changing task priority", "task_id", taskID, "priority", priority, "version", version)

	// Get current task
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Validate version
	if task.Version != version {
		return nil, domain.ErrVersionConflict("task", version, task.Version)
	}

	// Update priority
	task.Priority = priority
	return s.UpdateTask(ctx, task)
}

func (s *taskService) AddTaskCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Adding categories to task", "task_id", taskID, "categories", categoryIDs, "version", version)

	// Get current task
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Validate version
	if task.Version != version {
		return nil, domain.ErrVersionConflict("task", version, task.Version)
	}

	// Validate categories exist
	for _, categoryID := range categoryIDs {
		if _, err := s.categoryRepo.GetByID(ctx, categoryID); err != nil {
			if domain.IsNotFoundError(err) {
				return nil, domain.ErrInvalidInput(fmt.Sprintf("category %s does not exist", categoryID))
			}
			return nil, fmt.Errorf("failed to validate category %s: %w", categoryID, err)
		}
	}

	// Add categories via repository
	if err := s.taskRepo.AddCategories(ctx, taskID, categoryIDs, version); err != nil {
		s.logger.Error(ctx, "Failed to add categories to task", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("failed to add categories: %w", err)
	}

	// Return updated task
	return s.taskRepo.GetByID(ctx, taskID)
}

func (s *taskService) RemoveTaskCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Removing categories from task", "task_id", taskID, "categories", categoryIDs, "version", version)

	// Get current task to validate version
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Validate version
	if task.Version != version {
		return nil, domain.ErrVersionConflict("task", version, task.Version)
	}

	// Remove categories via repository
	if err := s.taskRepo.RemoveCategories(ctx, taskID, categoryIDs, version); err != nil {
		s.logger.Error(ctx, "Failed to remove categories from task", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("failed to remove categories: %w", err)
	}

	// Return updated task
	return s.taskRepo.GetByID(ctx, taskID)
}

func (s *taskService) AddTaskTags(ctx context.Context, taskID string, tagIDs []string, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Adding tags to task", "task_id", taskID, "tags", tagIDs, "version", version)

	// Get current task
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Validate version
	if task.Version != version {
		return nil, domain.ErrVersionConflict("task", version, task.Version)
	}

	// Validate tags exist
	for _, tagID := range tagIDs {
		if _, err := s.tagRepo.GetByID(ctx, tagID); err != nil {
			if domain.IsNotFoundError(err) {
				return nil, domain.ErrInvalidInput(fmt.Sprintf("tag %s does not exist", tagID))
			}
			return nil, fmt.Errorf("failed to validate tag %s: %w", tagID, err)
		}
	}

	// Add tags via repository
	if err := s.taskRepo.AddTags(ctx, taskID, tagIDs, version); err != nil {
		s.logger.Error(ctx, "Failed to add tags to task", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("failed to add tags: %w", err)
	}

	// Return updated task
	return s.taskRepo.GetByID(ctx, taskID)
}

func (s *taskService) RemoveTaskTags(ctx context.Context, taskID string, tagIDs []string, version int64) (*domain.Task, error) {
	s.logger.Info(ctx, "Removing tags from task", "task_id", taskID, "tags", tagIDs, "version", version)

	// Get current task to validate version
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Validate version
	if task.Version != version {
		return nil, domain.ErrVersionConflict("task", version, task.Version)
	}

	// Remove tags via repository
	if err := s.taskRepo.RemoveTags(ctx, taskID, tagIDs, version); err != nil {
		s.logger.Error(ctx, "Failed to remove tags from task", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("failed to remove tags: %w", err)
	}

	// Return updated task
	return s.taskRepo.GetByID(ctx, taskID)
}

func (s *taskService) GetTaskHistory(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	s.logger.Debug(ctx, "Getting task history", "task_id", taskID)

	if taskID == "" {
		return nil, domain.ErrInvalidInput("task ID is required")
	}

	history, err := s.taskRepo.GetHistory(ctx, taskID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get task history", "error", err, "task_id", taskID)
		return nil, fmt.Errorf("failed to get task history: %w", err)
	}

	return history, nil
}

// Helper methods for business validation

func (s *taskService) validateTaskForCreation(ctx context.Context, task *domain.Task) error {
	if err := task.IsValid(); err != nil {
		return err
	}

	// Additional business rules for task creation
	if task.Title == "" {
		return domain.ErrInvalidInput("task title is required")
	}

	return nil
}

func (s *taskService) validateTaskForUpdate(ctx context.Context, task *domain.Task) error {
	if task.ID == "" {
		return domain.ErrInvalidInput("task ID is required for update")
	}

	if err := task.IsValid(); err != nil {
		return err
	}

	return nil
}

func (s *taskService) validateStatusTransition(currentStatus, newStatus domain.TaskStatus) error {
	// Define valid status transitions
	validTransitions := map[domain.TaskStatus][]domain.TaskStatus{
		domain.TaskStatusOpen: {
			domain.TaskStatusInProgress,
			domain.TaskStatusCompleted,
			domain.TaskStatusCancelled,
		},
		domain.TaskStatusInProgress: {
			domain.TaskStatusCompleted,
			domain.TaskStatusOpen,
			domain.TaskStatusCancelled,
		},
		domain.TaskStatusCompleted: {
			domain.TaskStatusOpen,
			domain.TaskStatusInProgress,
		},
		domain.TaskStatusCancelled: {
			domain.TaskStatusOpen,
			domain.TaskStatusInProgress,
		},
	}

	// Check if transition is valid
	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return domain.ErrBusinessRule(fmt.Sprintf("unknown current status: %s", currentStatus))
	}

	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return nil // Valid transition
		}
	}

	return domain.ErrBusinessRule(fmt.Sprintf("invalid status transition from %s to %s", currentStatus, newStatus))
}
