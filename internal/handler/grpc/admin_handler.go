package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/todo-app/services/admin-service/proto/gen/go/todo/v1"
)

// AdminHandler implements the gRPC AdminService
type AdminHandler struct {
	todov1.UnimplementedAdminServiceServer
	services *service.Services
	logger   logger.Logger
}

// NewAdminHandler creates a new admin gRPC handler
func NewAdminHandler(services *service.Services, logger logger.Logger) *AdminHandler {
	return &AdminHandler{
		services: services,
		logger:   logger,
	}
}

// User management methods

// ListUsers lists users with pagination
func (h *AdminHandler) ListUsers(ctx context.Context, req *todov1.ListUsersRequest) (*todov1.ListUsersResponse, error) {
	h.logger.Info(ctx, "Listing users via gRPC", "page_info", req.GetPageInfo())

	// Convert pagination info
	opts := repository.ListOptions{
		Page:           0, // Default to first page for now
		PageSize:       req.GetPageInfo().GetPageSize(),
		SearchQuery:    req.GetSearchQuery(),
		IncludeDeleted: req.GetIncludeDeleted(),
	}

	if opts.PageSize == 0 {
		opts.PageSize = 50 // Default page size
	}

	// Get users from service
	users, total, err := h.services.User.ListUsers(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list users", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// Convert domain users to protobuf
	pbUsers := make([]*todov1.User, len(users))
	for i, user := range users {
		pbUsers[i] = user.ToProtobuf()
	}

	// Build response
	response := &todov1.ListUsersResponse{
		Users: pbUsers,
		PageResponse: &todov1.PageResponse{
			NextPageToken: "", // TODO: Implement token-based pagination
			TotalCount:    int32(total),
		},
	}

	h.logger.Info(ctx, "Listed users successfully", "count", len(users), "total", total)
	return response, nil
}

// GetUser retrieves a user by ID
func (h *AdminHandler) GetUser(ctx context.Context, req *todov1.GetUserRequest) (*todov1.GetUserResponse, error) {
	h.logger.Info(ctx, "Getting user via gRPC", "user_id", req.GetUserId())

	// Validate request
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Get user from service
	user, err := h.services.User.GetUserByID(ctx, req.GetUserId())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", req.GetUserId())
		}
		h.logger.Error(ctx, "Failed to get user", "user_id", req.GetUserId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// Build response
	response := &todov1.GetUserResponse{
		User: user.ToProtobuf(),
	}

	h.logger.Info(ctx, "Retrieved user successfully", "user_id", req.GetUserId())
	return response, nil
}

// Task management methods

// CreateTask creates a new task
func (h *AdminHandler) CreateTask(ctx context.Context, req *todov1.CreateTaskRequest) (*todov1.CreateTaskResponse, error) {
	h.logger.Info(ctx, "Creating task via gRPC", "title", req.GetTitle())

	// Validate request
	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}
	if req.GetAssigneeId() == "" {
		return nil, status.Error(codes.InvalidArgument, "assignee_id is required")
	}

	// Create domain task
	task := &domain.Task{
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		AssigneeID:  req.GetAssigneeId(),
		Status:      domain.TaskStatusOpen,     // Default status
		Priority:    domain.TaskPriorityMedium, // Default priority
	}

	// Create task via service
	createdTask, err := h.services.Task.CreateTask(ctx, task)
	if err != nil {
		h.logger.Error(ctx, "Failed to create task", "title", req.GetTitle(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create task: %v", err)
	}

	// Build response
	response := &todov1.CreateTaskResponse{
		Task: createdTask.ToProtobuf(),
	}

	h.logger.Info(ctx, "Created task successfully", "task_id", createdTask.ID, "title", createdTask.Title)
	return response, nil
}

// ListTasks lists tasks with filtering and pagination
func (h *AdminHandler) ListTasks(ctx context.Context, req *todov1.ListTasksRequest) (*todov1.ListTasksResponse, error) {
	h.logger.Info(ctx, "Listing tasks via gRPC", "assignee_id", req.GetAssigneeId())

	// Build task list options
	opts := repository.TaskListOptions{
		ListOptions: repository.ListOptions{
			Page:     0,  // Default first page
			PageSize: 50, // Default page size
		},
		AssigneeID: req.GetAssigneeId(),
	}

	// Convert protobuf status to domain status
	if req.GetStatus() != todov1.TaskStatus_TASK_STATUS_UNSPECIFIED {
		switch req.GetStatus() {
		case todov1.TaskStatus_TASK_STATUS_OPEN:
			opts.Status = domain.TaskStatusOpen
		case todov1.TaskStatus_TASK_STATUS_IN_PROGRESS:
			opts.Status = domain.TaskStatusInProgress
		case todov1.TaskStatus_TASK_STATUS_COMPLETED:
			opts.Status = domain.TaskStatusCompleted
		}
	}

	// Get tasks from service
	tasks, _, err := h.services.Task.ListTasks(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list tasks", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list tasks: %v", err)
	}

	// Convert domain tasks to protobuf
	pbTasks := make([]*todov1.Task, len(tasks))
	for i, task := range tasks {
		pbTasks[i] = task.ToProtobuf()
	}

	// Build response
	response := &todov1.ListTasksResponse{
		Tasks: pbTasks,
	}

	h.logger.Info(ctx, "Listed tasks successfully", "count", len(tasks))
	return response, nil
}

// GetTask retrieves a task by ID
func (h *AdminHandler) GetTask(ctx context.Context, req *todov1.GetTaskRequest) (*todov1.GetTaskResponse, error) {
	h.logger.Info(ctx, "Getting task via gRPC", "task_id", req.GetTaskId())

	// Validate request
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	// Get task from service
	task, err := h.services.Task.GetTaskByID(ctx, req.GetTaskId())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "task not found: %s", req.GetTaskId())
		}
		h.logger.Error(ctx, "Failed to get task", "task_id", req.GetTaskId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get task: %v", err)
	}

	// Build response
	response := &todov1.GetTaskResponse{
		Task: task.ToProtobuf(),
	}

	h.logger.Info(ctx, "Retrieved task successfully", "task_id", req.GetTaskId())
	return response, nil
}

// UpdateTask updates an existing task
func (h *AdminHandler) UpdateTask(ctx context.Context, req *todov1.UpdateTaskRequest) (*todov1.UpdateTaskResponse, error) {
	h.logger.Info(ctx, "Updating task via gRPC", "task_id", req.GetTaskId())

	// Validate request
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	// Get current task to get version and current data
	currentTask, err := h.services.Task.GetTaskByID(ctx, req.GetTaskId())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "task not found: %s", req.GetTaskId())
		}
		h.logger.Error(ctx, "Failed to get current task", "task_id", req.GetTaskId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get current task: %v", err)
	}

	// Update fields that are provided
	if req.GetTitle() != "" {
		currentTask.Title = req.GetTitle()
	}
	if req.GetDescription() != "" {
		currentTask.Description = req.GetDescription()
	}
	if req.GetAssigneeId() != "" {
		currentTask.AssigneeID = req.GetAssigneeId()
	}

	// Convert protobuf status to domain status if provided
	if req.GetStatus() != todov1.TaskStatus_TASK_STATUS_UNSPECIFIED {
		switch req.GetStatus() {
		case todov1.TaskStatus_TASK_STATUS_OPEN:
			currentTask.Status = domain.TaskStatusOpen
		case todov1.TaskStatus_TASK_STATUS_IN_PROGRESS:
			currentTask.Status = domain.TaskStatusInProgress
		case todov1.TaskStatus_TASK_STATUS_COMPLETED:
			currentTask.Status = domain.TaskStatusCompleted
		}
	}

	// Update task via service
	updatedTask, err := h.services.Task.UpdateTask(ctx, currentTask)
	if err != nil {
		if domain.IsVersionConflictError(err) {
			return nil, status.Error(codes.FailedPrecondition, "task was modified by another request")
		}
		h.logger.Error(ctx, "Failed to update task", "task_id", req.GetTaskId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update task: %v", err)
	}

	// Build response
	response := &todov1.UpdateTaskResponse{
		Task: updatedTask.ToProtobuf(),
	}

	h.logger.Info(ctx, "Updated task successfully", "task_id", req.GetTaskId())
	return response, nil
}

// GetTaskHistory retrieves task history
func (h *AdminHandler) GetTaskHistory(ctx context.Context, req *todov1.GetTaskHistoryRequest) (*todov1.GetTaskHistoryResponse, error) {
	h.logger.Info(ctx, "Getting task history via gRPC", "task_id", req.GetTaskId())

	// Validate request
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	// Get task history from service
	history, err := h.services.Task.GetTaskHistory(ctx, req.GetTaskId())
	if err != nil {
		h.logger.Error(ctx, "Failed to get task history", "task_id", req.GetTaskId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get task history: %v", err)
	}

	// Convert domain history to protobuf
	pbHistory := make([]*todov1.TaskHistoryEntry, len(history))
	for i, entry := range history {
		pbHistory[i] = entry.ToProtobuf()
	}

	// Build response
	response := &todov1.GetTaskHistoryResponse{
		History: pbHistory,
	}

	h.logger.Info(ctx, "Retrieved task history successfully", "task_id", req.GetTaskId(), "count", len(history))
	return response, nil
}
