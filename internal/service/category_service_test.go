package service

import (
	"context"
	"testing"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/testutil"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type mockCategoryRepository struct {
	categories map[string]*domain.Category
	nameIndex  map[string]*domain.Category
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{
		categories: make(map[string]*domain.Category),
		nameIndex:  make(map[string]*domain.Category),
	}
}

func (m *mockCategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	if _, exists := m.nameIndex[category.Name]; exists {
		return domain.ErrConflict("category with this name already exists")
	}

	category.ID = "mock-category-" + category.Name
	category.Version = 1
	m.categories[category.ID] = category
	m.nameIndex[category.Name] = category
	return nil
}

func (m *mockCategoryRepository) GetByID(ctx context.Context, id string) (*domain.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, domain.ErrNotFound("category")
	}
	return category, nil
}

func (m *mockCategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	existing, exists := m.categories[category.ID]
	if !exists {
		return domain.ErrNotFound("category")
	}
	if existing.Version != category.Version {
		return domain.ErrVersionConflict("category", category.Version, existing.Version)
	}

	category.Version++
	if existing.Name != category.Name {
		delete(m.nameIndex, existing.Name)
		m.nameIndex[category.Name] = category
	}
	m.categories[category.ID] = category
	return nil
}

func (m *mockCategoryRepository) SoftDelete(ctx context.Context, id string, version int64) error {
	existing, exists := m.categories[id]
	if !exists {
		return domain.ErrNotFound("category")
	}
	if existing.Version != version {
		return domain.ErrVersionConflict("category", version, existing.Version)
	}

	existing.IsDeleted = true
	existing.Version++
	delete(m.categories, id)
	delete(m.nameIndex, existing.Name)
	return nil
}

func (m *mockCategoryRepository) Restore(ctx context.Context, id string, version int64) error {
	return nil // Not implemented for mock
}

func (m *mockCategoryRepository) List(ctx context.Context, opts repository.CategoryListOptions) ([]*domain.Category, int64, error) {
	categories := make([]*domain.Category, 0, len(m.categories))
	for _, category := range m.categories {
		categories = append(categories, category)
	}
	return categories, int64(len(categories)), nil
}

type mockTaskRepository struct {
	tasks map[string]*domain.Task
}

func newMockTaskRepository() *mockTaskRepository {
	return &mockTaskRepository{
		tasks: make(map[string]*domain.Task),
	}
}

func (m *mockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	task.ID = "mock-task-" + task.Title
	task.Version = 1
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id string) (*domain.Task, error) {
	task, exists := m.tasks[id]
	if !exists {
		return nil, domain.ErrNotFound("task")
	}
	return task, nil
}

func (m *mockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
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

func (m *mockTaskRepository) SoftDelete(ctx context.Context, id string, version int64) error {
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

func (m *mockTaskRepository) Restore(ctx context.Context, id string, version int64) error {
	return nil
}

func (m *mockTaskRepository) List(ctx context.Context, opts repository.TaskListOptions) ([]*domain.Task, int64, error) {
	tasks := make([]*domain.Task, 0)
	for _, task := range m.tasks {
		// Filter by category if specified
		if len(opts.CategoryIDs) > 0 {
			hasCategory := false
			for _, categoryID := range opts.CategoryIDs {
				for _, category := range task.Categories {
					if category.ID == categoryID {
						hasCategory = true
						break
					}
				}
				if hasCategory {
					break
				}
			}
			if hasCategory {
				tasks = append(tasks, task)
			}
		} else {
			tasks = append(tasks, task)
		}
	}
	return tasks, int64(len(tasks)), nil
}

func (m *mockTaskRepository) AddCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepository) RemoveCategories(ctx context.Context, taskID string, categoryIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepository) AddTags(ctx context.Context, taskID string, tagIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepository) RemoveTags(ctx context.Context, taskID string, tagIDs []string, version int64) error {
	return nil
}

func (m *mockTaskRepository) GetHistory(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	return nil, nil
}

func TestCategoryService_CreateCategory(t *testing.T) {
	mockCategoryRepo := newMockCategoryRepository()
	mockTaskRepo := newMockTaskRepository()
	mockLogger := logger.NewLogger("debug")
	service := NewCategoryService(mockCategoryRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	tests := []struct {
		name     string
		category *domain.Category
		wantErr  bool
	}{
		{
			name: "valid category creation",
			category: &domain.Category{
				Name:        "Test Category",
				Description: "Test Description",
				CreatorID:   "creator-123",
			},
			wantErr: false,
		},
		{
			name: "duplicate category name",
			category: &domain.Category{
				Name:        "Test Category", // Same as above
				Description: "Another Description",
				CreatorID:   "creator-456",
			},
			wantErr: true,
		},
		{
			name: "invalid category - missing name",
			category: &domain.Category{
				Name:        "",
				Description: "Test Description",
				CreatorID:   "creator-789",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateCategory(ctx, tt.category)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCategory() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryService_ValidateCategoryUsage(t *testing.T) {
	mockCategoryRepo := newMockCategoryRepository()
	mockTaskRepo := newMockTaskRepository()
	mockLogger := logger.NewLogger("debug")
	service := NewCategoryService(mockCategoryRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Create a test category
	testCategory := testutil.TestCategory()
	mockCategoryRepo.Create(ctx, testCategory)

	// Create a task that uses the category
	testTask := testutil.TestTask("assignee-123")
	testTask.Categories = []domain.Category{*testCategory}
	mockTaskRepo.Create(ctx, testTask)

	tests := []struct {
		name       string
		categoryID string
		wantErr    bool
	}{
		{
			name:       "category in use - should fail validation",
			categoryID: testCategory.ID,
			wantErr:    true,
		},
		{
			name:       "category not in use - should pass validation",
			categoryID: "unused-category-id",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateCategoryUsage(ctx, tt.categoryID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCategoryUsage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCategoryService_GetCategoryTaskCount(t *testing.T) {
	mockCategoryRepo := newMockCategoryRepository()
	mockTaskRepo := newMockTaskRepository()
	mockLogger := logger.NewLogger("debug")
	service := NewCategoryService(mockCategoryRepo, mockTaskRepo, mockLogger)
	ctx := context.Background()

	// Create a test category
	testCategory := testutil.TestCategory()
	mockCategoryRepo.Create(ctx, testCategory)

	// Create tasks that use the category
	for i := 0; i < 3; i++ {
		task := testutil.TestTask("assignee-123")
		task.Categories = []domain.Category{*testCategory}
		task.Title = task.Title + string(rune(i))
		mockTaskRepo.Create(ctx, task)
	}

	count, err := service.GetCategoryTaskCount(ctx, testCategory.ID)
	if err != nil {
		t.Fatalf("GetCategoryTaskCount() error = %v", err)
	}

	expectedCount := int64(3)
	if count != expectedCount {
		t.Errorf("GetCategoryTaskCount() = %d, want %d", count, expectedCount)
	}
}
