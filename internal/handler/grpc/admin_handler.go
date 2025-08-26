package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	
	// TODO: Implement user listing
	return nil, status.Error(codes.Unimplemented, "ListUsers not yet implemented")
}

// GetUser retrieves a user by ID
func (h *AdminHandler) GetUser(ctx context.Context, req *todov1.GetUserRequest) (*todov1.GetUserResponse, error) {
	h.logger.Info(ctx, "Getting user via gRPC", "user_id", req.GetUserId())
	
	// TODO: Implement user retrieval
	return nil, status.Error(codes.Unimplemented, "GetUser not yet implemented")
}

// Task management methods

// CreateTask creates a new task
func (h *AdminHandler) CreateTask(ctx context.Context, req *todov1.CreateTaskRequest) (*todov1.CreateTaskResponse, error) {
	h.logger.Info(ctx, "Creating task via gRPC", "title", req.GetTitle())
	
	// TODO: Implement task creation
	return nil, status.Error(codes.Unimplemented, "CreateTask not yet implemented")
}

// ListTasks lists tasks with filtering and pagination
func (h *AdminHandler) ListTasks(ctx context.Context, req *todov1.ListTasksRequest) (*todov1.ListTasksResponse, error) {
	h.logger.Info(ctx, "Listing tasks via gRPC", "assignee_id", req.GetAssigneeId())
	
	// TODO: Implement task listing
	return nil, status.Error(codes.Unimplemented, "ListTasks not yet implemented")
}

// GetTask retrieves a task by ID
func (h *AdminHandler) GetTask(ctx context.Context, req *todov1.GetTaskRequest) (*todov1.GetTaskResponse, error) {
	h.logger.Info(ctx, "Getting task via gRPC", "task_id", req.GetTaskId())
	
	// TODO: Implement task retrieval
	return nil, status.Error(codes.Unimplemented, "GetTask not yet implemented")
}

// UpdateTask updates an existing task
func (h *AdminHandler) UpdateTask(ctx context.Context, req *todov1.UpdateTaskRequest) (*todov1.UpdateTaskResponse, error) {
	h.logger.Info(ctx, "Updating task via gRPC", "task_id", req.GetTaskId())
	
	// TODO: Implement task update
	return nil, status.Error(codes.Unimplemented, "UpdateTask not yet implemented")
}

// GetTaskHistory retrieves task history
func (h *AdminHandler) GetTaskHistory(ctx context.Context, req *todov1.GetTaskHistoryRequest) (*todov1.GetTaskHistoryResponse, error) {
	h.logger.Info(ctx, "Getting task history via gRPC", "task_id", req.GetTaskId())
	
	// TODO: Implement task history retrieval
	return nil, status.Error(codes.Unimplemented, "GetTaskHistory not yet implemented")
}