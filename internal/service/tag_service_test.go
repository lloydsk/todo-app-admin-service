package service

import (
	"context"
	"testing"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/testutil"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type mockTagRepositoryForTagService struct {
	tags      map[string]*domain.Tag
	nameIndex map[string]*domain.Tag
}

func newMockTagRepositoryForTagService() *mockTagRepositoryForTagService {
	return &mockTagRepositoryForTagService{
		tags:      make(map[string]*domain.Tag),
		nameIndex: make(map[string]*domain.Tag),
	}
}

func (m *mockTagRepositoryForTagService) Create(ctx context.Context, tag *domain.Tag) error {
	if _, exists := m.nameIndex[tag.Name]; exists {
		return domain.ErrConflict("tag with this name already exists")
	}

	tag.ID = "mock-tag-" + tag.Name
	tag.Version = 1
	m.tags[tag.ID] = tag
	m.nameIndex[tag.Name] = tag
	return nil
}

func (m *mockTagRepositoryForTagService) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	tag, exists := m.tags[id]
	if !exists {
		return nil, domain.ErrNotFound("tag")
	}
	return tag, nil
}

func (m *mockTagRepositoryForTagService) Update(ctx context.Context, tag *domain.Tag) error {
	existing, exists := m.tags[tag.ID]
	if !exists {
		return domain.ErrNotFound("tag")
	}
	if existing.Version != tag.Version {
		return domain.ErrVersionConflict("tag", tag.Version, existing.Version)
	}

	tag.Version++
	if existing.Name != tag.Name {
		delete(m.nameIndex, existing.Name)
		m.nameIndex[tag.Name] = tag
	}
	m.tags[tag.ID] = tag
	return nil
}

func (m *mockTagRepositoryForTagService) SoftDelete(ctx context.Context, id string, version int64) error {
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
	delete(m.nameIndex, existing.Name)
	return nil
}

func (m *mockTagRepositoryForTagService) Restore(ctx context.Context, id string, version int64) error {
	return nil // Not implemented for mock
}

func (m *mockTagRepositoryForTagService) List(ctx context.Context, opts repository.TagListOptions) ([]*domain.Tag, int64, error) {
	tags := make([]*domain.Tag, 0, len(m.tags))

	// Filter by search query if provided
	if opts.SearchQuery != "" {
		for _, tag := range m.tags {
			if tag.Name == opts.SearchQuery {
				tags = append(tags, tag)
			}
		}
	} else {
		for _, tag := range m.tags {
			tags = append(tags, tag)
		}
	}

	return tags, int64(len(tags)), nil
}

type mockTaskRepositoryForTagService struct {
	tasks map[string]*domain.Task
}

func newMockTaskRepositoryForTagService() *mockTaskRepositoryForTagService {
	return &mockTaskRepositoryForTagService{
		tasks: make(map[string]*domain.Task),
	}
}

func (m *mockTaskRepositoryForTagService) Create(ctx context.Context, task *domain.Task) error {
	task.ID = "mock-task-" + task.Title
	task.Version = 1
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepositoryForTagService) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, domain.ErrNotFound("task")
	}
	return task, nil
}

func (m *mockTaskRepositoryForTagService) Update(ctx context.Context, task *domain.Task) error {
	existing, exists := m.tasks[task.ID]
	if !exists {
		return domain.ErrNotFound("task")
	}
	if existing.Version != task.Version {
		return domain.ErrVersionConflict("task", task.Version, existing.Version)
	}

	task.Version++
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepositoryForTagService) SoftDelete(ctx context.Context, id string, version int64) error {
	existing, exists := m.tasks[id]
	if !exists {
		return domain.ErrNotFound("task")
	}
	if existing.Version != version {
		return domain.ErrVersionConflict("task", version, existing.Version)
	}

	existing.IsDeleted = true
	existing.Version++
	delete(m.tasks, id)
	return nil
}

func (m *mockTaskRepositoryForTagService) Restore(ctx context.Context, id string, version int64) error {
	return nil
}

func (m *mockTaskRepositoryForTagService) List(ctx context.Context, opts repository.TaskListOptions) ([]*domain.Task, int64, error) {
	tasks := make([]*domain.Task, 0)
	for _, task := range m.tasks {
		// Filter by tags if specified
		if len(opts.TagIDs) > 0 {
			hasTag := false
			for _, tagID := range opts.TagIDs {
				for _, tag := range task.Tags {
					if tag.ID == tagID {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if hasTag {
				tasks = append(tasks, task)
			}
		} else {
			tasks = append(tasks, task)
		}
	}
	return tasks, int64(len(tasks)), nil
}

func (m *mockTaskRepositoryForTagService) AddCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepositoryForTagService) RemoveCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepositoryForTagService) AddTags(ctx context.Context, taskID string, tagIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepositoryForTagService) RemoveTags(ctx context.Context, taskID string, tagIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepositoryForTagService) GetHistory(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	return nil, nil
}

func TestTagService_CreateTag(t *testing.T) {
	mockTagRepo := newMockTagRepositoryForTagService()
	mockTaskRepo := newMockTaskRepositoryForTagService()
	mockLogger := logger.NewLogger("debug")
	service := NewTagService(mockTagRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	tests := []struct {
		name    string
		tag     *domain.Tag
		wantErr bool
	}{
		{
			name: "valid tag creation",
			tag: &domain.Tag{
				Name:      "urgent",
				Color:     "#FF0000",
				CreatorID: "creator-123",
			},
			wantErr: false,
		},
		{
			name: "duplicate tag name",
			tag: &domain.Tag{
				Name:      "urgent", // Same as above
				Color:     "#00FF00",
				CreatorID: "creator-456",
			},
			wantErr: true,
		},
		{
			name: "invalid tag - missing name",
			tag: &domain.Tag{
				Name:      "",
				Color:     "#0000FF",
				CreatorID: "creator-789",
			},
			wantErr: true,
		},
		{
			name: "tag name normalization",
			tag: &domain.Tag{
				Name:      "  IMPORTANT  ",
				Color:     "#FFFF00",
				CreatorID: "creator-101",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CreateTag(ctx, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				// Verify name normalization
				if tt.tag.Name == "  IMPORTANT  " {
					if result.Name != "important" {
						t.Errorf("Expected normalized name 'important', got '%s'", result.Name)
					}
				}
			}
		})
	}
}

func TestTagService_FindOrCreateTag(t *testing.T) {
	mockTagRepo := newMockTagRepositoryForTagService()
	mockTaskRepo := newMockTaskRepositoryForTagService()
	mockLogger := logger.NewLogger("debug")
	service := NewTagService(mockTagRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Pre-create a tag
	existingTag := &domain.Tag{
		Name:      "existing",
		Color:     "#000000",
		CreatorID: "system",
	}
	mockTagRepo.Create(ctx, existingTag)

	tests := []struct {
		name        string
		tagName     string
		expectFound bool
		wantErr     bool
	}{
		{
			name:        "find existing tag",
			tagName:     "existing",
			expectFound: true,
			wantErr:     false,
		},
		{
			name:        "create new tag",
			tagName:     "newbie",
			expectFound: false,
			wantErr:     false,
		},
		{
			name:        "normalize and find",
			tagName:     "  EXISTING  ",
			expectFound: true,
			wantErr:     false,
		},
		{
			name:    "invalid name",
			tagName: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FindOrCreateTag(ctx, tt.tagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOrCreateTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("Expected result to be non-nil")
					return
				}

				if tt.expectFound && result.ID != existingTag.ID {
					t.Error("Expected to find existing tag")
				}

				if !tt.expectFound && result.CreatorID != "system" {
					t.Error("Expected new tag to have system creator")
				}
			}
		})
	}
}

func TestTagService_ValidateTagUsage(t *testing.T) {
	mockTagRepo := newMockTagRepositoryForTagService()
	mockTaskRepo := newMockTaskRepositoryForTagService()
	mockLogger := logger.NewLogger("debug")
	service := NewTagService(mockTagRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Create a test tag
	testTag := testutil.TestTag()
	mockTagRepo.Create(ctx, testTag)

	// Create a task that uses the tag
	testTask := testutil.TestTask("assignee-123")
	testTask.Tags = []domain.Tag{*testTag}
	mockTaskRepo.Create(ctx, testTask)

	tests := []struct {
		name    string
		tagID   string
		wantErr bool
	}{
		{
			name:    "tag in use - should fail validation",
			tagID:   testTag.ID,
			wantErr: true,
		},
		{
			name:    "tag not in use - should pass validation",
			tagID:   "unused-tag-id",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateTagUsage(ctx, tt.tagID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTagUsage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagService_GetTagTaskCount(t *testing.T) {
	mockTagRepo := newMockTagRepositoryForTagService()
	mockTaskRepo := newMockTaskRepositoryForTagService()
	mockLogger := logger.NewLogger("debug")
	service := NewTagService(mockTagRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Create a test tag
	testTag := testutil.TestTag()
	mockTagRepo.Create(ctx, testTag)

	// Create tasks that use the tag
	for i := 0; i < 5; i++ {
		task := testutil.TestTask("assignee-123")
		task.Tags = []domain.Tag{*testTag}
		task.Title = task.Title + string(rune(i))
		mockTaskRepo.Create(ctx, task)
	}

	count, err := service.GetTagTaskCount(ctx, testTag.ID)
	if err != nil {
		t.Fatalf("GetTagTaskCount() error = %v", err)
	}

	expectedCount := int64(5)
	if count != expectedCount {
		t.Errorf("GetTagTaskCount() = %d, want %d", count, expectedCount)
	}
}

func TestTagService_UpdateTag(t *testing.T) {
	mockTagRepo := newMockTagRepositoryForTagService()
	mockTaskRepo := newMockTaskRepositoryForTagService()
	mockLogger := logger.NewLogger("debug")
	service := NewTagService(mockTagRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Create a test tag
	testTag := testutil.TestTag()
	mockTagRepo.Create(ctx, testTag)

	tests := []struct {
		name    string
		tag     *domain.Tag
		wantErr bool
	}{
		{
			name: "valid tag update",
			tag: &domain.Tag{
				ID:        testTag.ID,
				Name:      "updated-name",
				Color:     "#FFFFFF",
				CreatorID: testTag.CreatorID,
				Version:   testTag.Version,
			},
			wantErr: false,
		},
		{
			name: "tag not found",
			tag: &domain.Tag{
				ID:        "non-existent",
				Name:      "test",
				Color:     "#000000",
				CreatorID: "creator",
				Version:   1,
			},
			wantErr: true,
		},
		{
			name: "version conflict",
			tag: &domain.Tag{
				ID:        testTag.ID,
				Name:      "conflict",
				Color:     "#000000",
				CreatorID: testTag.CreatorID,
				Version:   999, // Wrong version
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.UpdateTag(ctx, tt.tag)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagService_DeleteTag(t *testing.T) {
	mockTagRepo := newMockTagRepositoryForTagService()
	mockTaskRepo := newMockTaskRepositoryForTagService()
	mockLogger := logger.NewLogger("debug")
	service := NewTagService(mockTagRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Create test tags
	unusedTag := testutil.TestTag()
	unusedTag.Name = "unused"
	mockTagRepo.Create(ctx, unusedTag)

	usedTag := testutil.TestTag()
	usedTag.Name = "used"
	mockTagRepo.Create(ctx, usedTag)

	// Create a task that uses one tag
	testTask := testutil.TestTask("assignee-123")
	testTask.Tags = []domain.Tag{*usedTag}
	mockTaskRepo.Create(ctx, testTask)

	tests := []struct {
		name    string
		tagID   string
		version int64
		wantErr bool
	}{
		{
			name:    "delete unused tag",
			tagID:   unusedTag.ID,
			version: unusedTag.Version,
			wantErr: false,
		},
		{
			name:    "delete used tag - should fail",
			tagID:   usedTag.ID,
			version: usedTag.Version,
			wantErr: true,
		},
		{
			name:    "delete non-existent tag",
			tagID:   "non-existent",
			version: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteTag(ctx, tt.tagID, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteTag() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTagService_ListTags(t *testing.T) {
	mockTagRepo := newMockTagRepositoryForTagService()
	mockTaskRepo := newMockTaskRepositoryForTagService()
	mockLogger := logger.NewLogger("debug")
	service := NewTagService(mockTagRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Create some test tags
	for i := 0; i < 3; i++ {
		tag := testutil.TestTag()
		tag.Name = tag.Name + string(rune(i))
		mockTagRepo.Create(ctx, tag)
	}

	tags, total, err := service.ListTags(ctx, repository.ListOptions{
		Page:     1,
		PageSize: 10,
	})

	if err != nil {
		t.Fatalf("ListTags() error = %v", err)
	}

	if len(tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(tags))
	}

	if total != 3 {
		t.Errorf("Expected total 3, got %d", total)
	}
}
