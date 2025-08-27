package connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
)

// AdminHandler implements the ConnectRPC AdminService
type AdminHandler struct {
	services *service.Services
	logger   logger.Logger
}

// NewAdminHandler creates a new admin ConnectRPC handler
func NewAdminHandler(services *service.Services, logger logger.Logger) *AdminHandler {
	return &AdminHandler{
		services: services,
		logger:   logger,
	}
}

// User management methods

// ListUsers lists users with pagination
func (h *AdminHandler) ListUsers(ctx context.Context, req *connect.Request[todov1.ListUsersRequest]) (*connect.Response[todov1.ListUsersResponse], error) {
	h.logger.Info(ctx, "Listing users via ConnectRPC", "page_info", req.Msg.GetPageInfo())

	// Convert pagination info
	opts := repository.ListOptions{
		Page:           0, // Default to first page for now
		PageSize:       req.Msg.GetPageInfo().GetPageSize(),
		SearchQuery:    req.Msg.GetSearchQuery(),
		IncludeDeleted: req.Msg.GetIncludeDeleted(),
	}

	if opts.PageSize == 0 {
		opts.PageSize = 50 // Default page size
	}

	// Get users from service
	users, total, err := h.services.User.ListUsers(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list users", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert domain users to protobuf users
	pbUsers := make([]*todov1.User, len(users))
	for i, user := range users {
		pbUsers[i] = user.ToProtobuf()
	}

	response := &todov1.ListUsersResponse{
		Users: pbUsers,
		PageResponse: &todov1.PageResponse{
			TotalCount: int32(total),
		},
	}

	return connect.NewResponse(response), nil
}

// GetUser gets a single user by ID
func (h *AdminHandler) GetUser(ctx context.Context, req *connect.Request[todov1.GetUserRequest]) (*connect.Response[todov1.GetUserResponse], error) {
	userID := req.Msg.GetUserId()
	h.logger.Info(ctx, "Getting user via ConnectRPC", "user_id", userID)

	if userID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, 
			fmt.Errorf("user_id is required"))
	}

	user, err := h.services.User.GetUserByID(ctx, userID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get user", "user_id", userID, "error", err)
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, 
				fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.GetUserResponse{
		User: user.ToProtobuf(),
	}

	return connect.NewResponse(response), nil
}

// Task management methods

// CreateTask creates a new task
func (h *AdminHandler) CreateTask(ctx context.Context, req *connect.Request[todov1.CreateTaskRequest]) (*connect.Response[todov1.CreateTaskResponse], error) {
	h.logger.Info(ctx, "Creating task via ConnectRPC", "title", req.Msg.GetTitle())

	// Validate request
	if req.Msg.GetTitle() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("title is required"))
	}
	if req.Msg.GetAssigneeId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("assignee_id is required"))
	}

	// Create domain task
	task := &domain.Task{
		Title:       req.Msg.GetTitle(),
		Description: req.Msg.GetDescription(),
		AssigneeID:  req.Msg.GetAssigneeId(),
		Status:      domain.TaskStatusOpen,     // Default status
		Priority:    domain.TaskPriorityMedium, // Default priority
	}

	// Create task via service
	createdTask, err := h.services.Task.CreateTask(ctx, task)
	if err != nil {
		h.logger.Error(ctx, "Failed to create task", "title", req.Msg.GetTitle(), "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Build response
	response := &todov1.CreateTaskResponse{
		Task: createdTask.ToProtobuf(),
	}

	h.logger.Info(ctx, "Created task successfully", "task_id", createdTask.ID, "title", createdTask.Title)
	return connect.NewResponse(response), nil
}

// ListTasks lists tasks with filtering and pagination
func (h *AdminHandler) ListTasks(ctx context.Context, req *connect.Request[todov1.ListTasksRequest]) (*connect.Response[todov1.ListTasksResponse], error) {
	h.logger.Info(ctx, "Listing tasks via ConnectRPC", "assignee_id", req.Msg.GetAssigneeId())

	// Build task list options
	opts := repository.TaskListOptions{
		ListOptions: repository.ListOptions{
			Page:     0,  // Default first page
			PageSize: 50, // Default page size
		},
		AssigneeID: req.Msg.GetAssigneeId(),
	}

	// Convert protobuf status to domain status
	if req.Msg.GetStatus() != todov1.TaskStatus_TASK_STATUS_UNSPECIFIED {
		switch req.Msg.GetStatus() {
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
		return nil, connect.NewError(connect.CodeInternal, err)
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
	return connect.NewResponse(response), nil
}

// GetTask retrieves a task by ID
func (h *AdminHandler) GetTask(ctx context.Context, req *connect.Request[todov1.GetTaskRequest]) (*connect.Response[todov1.GetTaskResponse], error) {
	h.logger.Info(ctx, "Getting task via ConnectRPC", "task_id", req.Msg.GetTaskId())

	// Validate request
	if req.Msg.GetTaskId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("task_id is required"))
	}

	// Get task from service
	task, err := h.services.Task.GetTaskByID(ctx, req.Msg.GetTaskId())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("task not found: %s", req.Msg.GetTaskId()))
		}
		h.logger.Error(ctx, "Failed to get task", "task_id", req.Msg.GetTaskId(), "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Build response
	response := &todov1.GetTaskResponse{
		Task: task.ToProtobuf(),
	}

	h.logger.Info(ctx, "Retrieved task successfully", "task_id", req.Msg.GetTaskId())
	return connect.NewResponse(response), nil
}

// UpdateTask updates an existing task
func (h *AdminHandler) UpdateTask(ctx context.Context, req *connect.Request[todov1.UpdateTaskRequest]) (*connect.Response[todov1.UpdateTaskResponse], error) {
	h.logger.Info(ctx, "Updating task via ConnectRPC", "task_id", req.Msg.GetTaskId())

	// Validate request
	if req.Msg.GetTaskId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("task_id is required"))
	}

	// Get current task to get version and current data
	currentTask, err := h.services.Task.GetTaskByID(ctx, req.Msg.GetTaskId())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("task not found: %s", req.Msg.GetTaskId()))
		}
		h.logger.Error(ctx, "Failed to get current task", "task_id", req.Msg.GetTaskId(), "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Update fields that are provided
	if req.Msg.GetTitle() != "" {
		currentTask.Title = req.Msg.GetTitle()
	}
	if req.Msg.GetDescription() != "" {
		currentTask.Description = req.Msg.GetDescription()
	}
	if req.Msg.GetAssigneeId() != "" {
		currentTask.AssigneeID = req.Msg.GetAssigneeId()
	}

	// Convert protobuf status to domain status if provided
	if req.Msg.GetStatus() != todov1.TaskStatus_TASK_STATUS_UNSPECIFIED {
		switch req.Msg.GetStatus() {
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
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("task was modified by another request"))
		}
		h.logger.Error(ctx, "Failed to update task", "task_id", req.Msg.GetTaskId(), "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Build response
	response := &todov1.UpdateTaskResponse{
		Task: updatedTask.ToProtobuf(),
	}

	h.logger.Info(ctx, "Updated task successfully", "task_id", req.Msg.GetTaskId())
	return connect.NewResponse(response), nil
}

// GetTaskHistory retrieves task history
func (h *AdminHandler) GetTaskHistory(ctx context.Context, req *connect.Request[todov1.GetTaskHistoryRequest]) (*connect.Response[todov1.GetTaskHistoryResponse], error) {
	h.logger.Info(ctx, "Getting task history via ConnectRPC", "task_id", req.Msg.GetTaskId())

	// Validate request
	if req.Msg.GetTaskId() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("task_id is required"))
	}

	// Get task history from service
	history, err := h.services.Task.GetTaskHistory(ctx, req.Msg.GetTaskId())
	if err != nil {
		h.logger.Error(ctx, "Failed to get task history", "task_id", req.Msg.GetTaskId(), "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
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

	h.logger.Info(ctx, "Retrieved task history successfully", "task_id", req.Msg.GetTaskId(), "count", len(history))
	return connect.NewResponse(response), nil
}