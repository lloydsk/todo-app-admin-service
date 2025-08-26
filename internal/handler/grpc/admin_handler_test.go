package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/todo-app/services/admin-service/proto/gen/go/todo/v1"
)

type mockUserService struct {
	users map[string]*domain.User
}

func newMockUserService() *mockUserService {
	return &mockUserService{
		users: make(map[string]*domain.User),
	}
}

func (m *mockUserService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.ID = "user-123"
	user.Version = 1
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrNotFound("user")
	}
	return user, nil
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, domain.ErrNotFound("user")
}

func (m *mockUserService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	existing, exists := m.users[user.ID]
	if !exists {
		return nil, domain.ErrNotFound("user")
	}
	if existing.Version != user.Version {
		return nil, domain.ErrVersionConflict("user", user.Version, existing.Version)
	}
	user.Version++
	m.users[user.ID] = user
	return user, nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id string, version int64) error {
	return nil
}

func (m *mockUserService) RestoreUser(ctx context.Context, id string, version int64) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserService) ListUsers(ctx context.Context, opts repository.ListOptions) ([]*domain.User, int64, error) {
	users := make([]*domain.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, int64(len(users)), nil
}

func (m *mockUserService) ChangeUserRole(ctx context.Context, userID string, newRole domain.UserRole, version int64) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserService) ValidateUserPermissions(ctx context.Context, userID string, requiredRole domain.UserRole) error {
	return nil
}

type mockTaskService struct {
	tasks map[string]*domain.Task
}

func newMockTaskService() *mockTaskService {
	return &mockTaskService{
		tasks: make(map[string]*domain.Task),
	}
}

func (m *mockTaskService) CreateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	task.ID = "task-123"
	task.Version = 1
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskService) GetTaskByID(ctx context.Context, id string) (*domain.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, domain.ErrNotFound("task")
	}
	return task, nil
}

func (m *mockTaskService) UpdateTask(ctx context.Context, task *domain.Task) (*domain.Task, error) {
	existing, exists := m.tasks[task.ID]
	if !exists {
		return nil, domain.ErrNotFound("task")
	}
	if existing.Version != task.Version {
		return nil, domain.ErrVersionConflict("task", task.Version, existing.Version)
	}
	task.Version++
	m.tasks[task.ID] = task
	return task, nil
}

func (m *mockTaskService) DeleteTask(ctx context.Context, id string, version int64) error {
	return nil
}

func (m *mockTaskService) RestoreTask(ctx context.Context, id string, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) ListTasks(ctx context.Context, opts repository.TaskListOptions) ([]*domain.Task, int64, error) {
	tasks := make([]*domain.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks, int64(len(tasks)), nil
}

func (m *mockTaskService) AssignTask(ctx context.Context, taskID, assigneeID string, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) ChangeTaskStatus(ctx context.Context, taskID string, status domain.TaskStatus, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) ChangeTaskPriority(ctx context.Context, taskID string, priority domain.TaskPriority, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) AddTaskCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) RemoveTaskCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) AddTaskTags(ctx context.Context, taskID string, tagIDs []string, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) RemoveTaskTags(ctx context.Context, taskID string, tagIDs []string, version int64) (*domain.Task, error) {
	return nil, nil
}

func (m *mockTaskService) GetTaskHistory(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	history := []*domain.TaskHistory{
		{
			ID:     "history-1",
			TaskID: taskID,
			Action: "created",
		},
	}
	return history, nil
}

func TestAdminHandler_ListUsers(t *testing.T) {
	userService := newMockUserService()
	taskService := newMockTaskService()

	services := &service.Services{
		User: userService,
		Task: taskService,
	}

	logger := logger.NewLogger("debug")
	handler := NewAdminHandler(services, logger)
	ctx := context.Background()

	// Add a test user
	testUser := &domain.User{
		ID:    "user-1",
		Email: "test@example.com",
		Name:  "Test User",
		Role:  domain.UserRoleUser,
	}
	userService.users["user-1"] = testUser

	tests := []struct {
		name    string
		request *todov1.ListUsersRequest
		wantErr bool
	}{
		{
			name: "successful listing",
			request: &todov1.ListUsersRequest{
				PageInfo: &todov1.PageInfo{
					PageSize: 10,
				},
			},
			wantErr: false,
		},
		{
			name: "with search query",
			request: &todov1.ListUsersRequest{
				PageInfo: &todov1.PageInfo{
					PageSize: 10,
				},
				SearchQuery: "test",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.ListUsers(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if len(resp.Users) != 1 {
				t.Errorf("expected 1 user, got %d", len(resp.Users))
			}
		})
	}
}

func TestAdminHandler_GetUser(t *testing.T) {
	userService := newMockUserService()
	taskService := newMockTaskService()

	services := &service.Services{
		User: userService,
		Task: taskService,
	}

	logger := logger.NewLogger("debug")
	handler := NewAdminHandler(services, logger)
	ctx := context.Background()

	// Add a test user
	testUser := &domain.User{
		ID:    "user-1",
		Email: "test@example.com",
		Name:  "Test User",
		Role:  domain.UserRoleUser,
	}
	userService.users["user-1"] = testUser

	tests := []struct {
		name       string
		request    *todov1.GetUserRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful get",
			request: &todov1.GetUserRequest{
				UserId: "user-1",
			},
			wantErr: false,
		},
		{
			name: "user not found",
			request: &todov1.GetUserRequest{
				UserId: "nonexistent",
			},
			wantErr:    true,
			wantStatus: codes.NotFound,
		},
		{
			name: "empty user ID",
			request: &todov1.GetUserRequest{
				UserId: "",
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.GetUser(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status: %v", err)
					return
				}

				if st.Code() != tt.wantStatus {
					t.Errorf("expected status %v, got %v", tt.wantStatus, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if resp.User.Id != "user-1" {
				t.Errorf("expected user ID user-1, got %s", resp.User.Id)
			}
		})
	}
}

func TestAdminHandler_CreateTask(t *testing.T) {
	userService := newMockUserService()
	taskService := newMockTaskService()

	services := &service.Services{
		User: userService,
		Task: taskService,
	}

	logger := logger.NewLogger("debug")
	handler := NewAdminHandler(services, logger)
	ctx := context.Background()

	tests := []struct {
		name       string
		request    *todov1.CreateTaskRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful creation",
			request: &todov1.CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				AssigneeId:  "user-1",
			},
			wantErr: false,
		},
		{
			name: "empty title",
			request: &todov1.CreateTaskRequest{
				Title:      "",
				AssigneeId: "user-1",
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
		{
			name: "empty assignee ID",
			request: &todov1.CreateTaskRequest{
				Title:      "Test Task",
				AssigneeId: "",
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.CreateTask(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status: %v", err)
					return
				}

				if st.Code() != tt.wantStatus {
					t.Errorf("expected status %v, got %v", tt.wantStatus, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if resp.Task.Title != tt.request.Title {
				t.Errorf("expected task title %s, got %s", tt.request.Title, resp.Task.Title)
			}

			// Verify task was added to mock
			if _, exists := taskService.tasks["task-123"]; !exists {
				t.Error("task was not added to service")
			}
		})
	}
}

func TestAdminHandler_GetTaskHistory(t *testing.T) {
	userService := newMockUserService()
	taskService := newMockTaskService()

	services := &service.Services{
		User: userService,
		Task: taskService,
	}

	logger := logger.NewLogger("debug")
	handler := NewAdminHandler(services, logger)
	ctx := context.Background()

	tests := []struct {
		name       string
		request    *todov1.GetTaskHistoryRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful get history",
			request: &todov1.GetTaskHistoryRequest{
				TaskId: "task-1",
			},
			wantErr: false,
		},
		{
			name: "empty task ID",
			request: &todov1.GetTaskHistoryRequest{
				TaskId: "",
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.GetTaskHistory(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status: %v", err)
					return
				}

				if st.Code() != tt.wantStatus {
					t.Errorf("expected status %v, got %v", tt.wantStatus, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if len(resp.History) != 1 {
				t.Errorf("expected 1 history entry, got %d", len(resp.History))
			}
		})
	}
}
