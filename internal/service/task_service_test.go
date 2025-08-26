package service

import (
	"context"
	"testing"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/testutil"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

func TestTaskService_CreateTask(t *testing.T) {
	mockUserRepo := newMockUserRepository()
	mockTaskRepo := newMockTaskRepository()
	mockCategoryRepo := newMockCategoryRepository()
	mockTagRepo := newMockTagRepository()
	mockLogger := logger.NewLogger("debug")
	
	service := NewTaskService(mockTaskRepo, mockUserRepo, mockCategoryRepo, mockTagRepo, mockLogger)
	ctx := context.Background()

	// Create a test user for assignment
	testUser := testutil.TestUser()
	testUser.Role = domain.UserRoleUser
	mockUserRepo.Create(ctx, testUser)

	tests := []struct {
		name    string
		task    *domain.Task
		wantErr bool
	}{
		{
			name: "valid task creation",
			task: &domain.Task{
				Title:       "Test Task",
				Description: "Test Description",
				AssigneeID:  testUser.ID,
				Status:      domain.TaskStatusOpen,
				Priority:    domain.TaskPriorityMedium,
			},
			wantErr: false,
		},
		{
			name: "invalid task - missing title",
			task: &domain.Task{
				Title:       "",
				Description: "Test Description",
				AssigneeID:  testUser.ID,
				Status:      domain.TaskStatusOpen,
				Priority:    domain.TaskPriorityMedium,
			},
			wantErr: true,
		},
		{
			name: "invalid task - non-existent assignee",
			task: &domain.Task{
				Title:       "Test Task",
				Description: "Test Description",
				AssigneeID:  "non-existent-user",
				Status:      domain.TaskStatusOpen,
				Priority:    domain.TaskPriorityMedium,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateTask(ctx, tt.task)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskService_ChangeTaskStatus(t *testing.T) {
	mockUserRepo := newMockUserRepository()
	mockTaskRepo := newMockTaskRepository()
	mockCategoryRepo := newMockCategoryRepository()
	mockTagRepo := newMockTagRepository()
	mockLogger := logger.NewLogger("debug")
	
	service := NewTaskService(mockTaskRepo, mockUserRepo, mockCategoryRepo, mockTagRepo, mockLogger)
	ctx := context.Background()

	// Create a test task
	testUser := testutil.TestUser()
	mockUserRepo.Create(ctx, testUser)
	
	testTask := testutil.TestTask(testUser.ID)
	testTask.Status = domain.TaskStatusOpen
	mockTaskRepo.Create(ctx, testTask)

	tests := []struct {
		name      string
		taskID    string
		newStatus domain.TaskStatus
		version   int64
		wantErr   bool
	}{
		{
			name:      "valid status change - open to in progress",
			taskID:    testTask.ID,
			newStatus: domain.TaskStatusInProgress,
			version:   testTask.Version,
			wantErr:   false,
		},
		{
			name:      "valid status change - open to cancelled",
			taskID:    testTask.ID,
			newStatus: domain.TaskStatusCancelled,
			version:   testTask.Version + 1,
			wantErr:   false,
		},
		{
			name:      "task not found",
			taskID:    "non-existent-task",
			newStatus: domain.TaskStatusInProgress,
			version:   1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ChangeTaskStatus(ctx, tt.taskID, tt.newStatus, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeTaskStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskService_AssignTask(t *testing.T) {
	mockUserRepo := newMockUserRepository()
	mockTaskRepo := newMockTaskRepository()
	mockCategoryRepo := newMockCategoryRepository()
	mockTagRepo := newMockTagRepository()
	mockLogger := logger.NewLogger("debug")
	
	service := NewTaskService(mockTaskRepo, mockUserRepo, mockCategoryRepo, mockTagRepo, mockLogger)
	ctx := context.Background()

	// Create test users
	originalUser := testutil.TestUser()
	originalUser.Email = "original@example.com"
	mockUserRepo.Create(ctx, originalUser)

	newUser := testutil.TestUser()
	newUser.Email = "new@example.com"
	mockUserRepo.Create(ctx, newUser)

	// Create a test task
	testTask := testutil.TestTask(originalUser.ID)
	mockTaskRepo.Create(ctx, testTask)

	tests := []struct {
		name       string
		taskID     string
		assigneeID string
		version    int64
		wantErr    bool
	}{
		{
			name:       "valid task assignment",
			taskID:     testTask.ID,
			assigneeID: newUser.ID,
			version:    testTask.Version,
			wantErr:    false,
		},
		{
			name:       "assign to non-existent user",
			taskID:     testTask.ID,
			assigneeID: "non-existent-user",
			version:    testTask.Version + 1,
			wantErr:    true,
		},
		{
			name:       "task not found",
			taskID:     "non-existent-task",
			assigneeID: newUser.ID,
			version:    1,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.AssignTask(ctx, tt.taskID, tt.assigneeID, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssignTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type mockTagRepository struct {
	tags map[string]*domain.Tag
}

func newMockTagRepository() *mockTagRepository {
	return &mockTagRepository{
		tags: make(map[string]*domain.Tag),
	}
}

func (m *mockTagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	tag.ID = "mock-tag-" + tag.Name
	tag.Version = 1
	m.tags[tag.ID] = tag
	return nil
}

func (m *mockTagRepository) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	tag, exists := m.tags[id]
	if !exists {
		return nil, domain.ErrNotFound("tag")
	}
	return tag, nil
}

func (m *mockTagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	existing, exists := m.tags[tag.ID]
	if !exists {
		return domain.ErrNotFound("tag")
	}
	if existing.Version != tag.Version {
		return domain.ErrVersionConflict("tag", tag.Version, existing.Version)
	}
	
	tag.Version++
	m.tags[tag.ID] = tag
	return nil
}

func (m *mockTagRepository) SoftDelete(ctx context.Context, id string, version int64) error {
	existing, exists := m.tags[id]
	if !exists {
		return domain.ErrNotFound("tag")
	}
	if existing.Version != version {
		return domain.ErrVersionConflict("tag", version, existing.Version)
	}
	
	existing.IsDeleted = true
	existing.Version++
	delete(m.tags, id)
	return nil
}

func (m *mockTagRepository) Restore(ctx context.Context, id string, version int64) error {
	return nil
}

func (m *mockTagRepository) List(ctx context.Context, opts repository.TagListOptions) ([]*domain.Tag, int64, error) {
	tags := make([]*domain.Tag, 0, len(m.tags))
	for _, tag := range m.tags {
		tags = append(tags, tag)
	}
	return tags, int64(len(tags)), nil
}