package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type mockCategoryService struct {
	categories map[string]*domain.Category
}

func newMockCategoryService() *mockCategoryService {
	return &mockCategoryService{
		categories: make(map[string]*domain.Category),
	}
}

func (m *mockCategoryService) CreateCategory(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	category.ID = "category-123"
	category.Version = 1
	m.categories[category.ID] = category
	return category, nil
}

func (m *mockCategoryService) GetCategoryByID(ctx context.Context, id string) (*domain.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, domain.ErrNotFound("category")
	}
	return category, nil
}

func (m *mockCategoryService) UpdateCategory(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	existing, exists := m.categories[category.ID]
	if !exists {
		return nil, domain.ErrNotFound("category")
	}
	if existing.Version != category.Version {
		return nil, domain.ErrVersionConflict("category", category.Version, existing.Version)
	}
	category.Version++
	m.categories[category.ID] = category
	return category, nil
}

func (m *mockCategoryService) DeleteCategory(ctx context.Context, id string, version int64) error {
	category, exists := m.categories[id]
	if !exists {
		return domain.ErrNotFound("category")
	}
	if category.Version != version {
		return domain.ErrVersionConflict("category", version, category.Version)
	}
	delete(m.categories, id)
	return nil
}

func (m *mockCategoryService) RestoreCategory(ctx context.Context, id string, version int64) (*domain.Category, error) {
	return nil, nil
}

func (m *mockCategoryService) ListCategories(ctx context.Context, opts repository.CategoryListOptions) ([]*domain.Category, int64, error) {
	categories := make([]*domain.Category, 0, len(m.categories))
	for _, category := range m.categories {
		// Apply public-only filter if requested
		if opts.PublicOnly && !category.IsPublic {
			continue
		}
		categories = append(categories, category)
	}
	return categories, int64(len(categories)), nil
}

func (m *mockCategoryService) ValidateCategoryUsage(ctx context.Context, categoryID string) error {
	return nil
}

func (m *mockCategoryService) GetCategoryTaskCount(ctx context.Context, categoryID string) (int64, error) {
	return 0, nil
}

func setupCategoryHandler(t *testing.T) (*CategoryHandler, *mockCategoryService) {
	categoryService := newMockCategoryService()
	logger := logger.NewLogger("debug")
	handler := NewCategoryHandler(categoryService, logger)

	return handler, categoryService
}

func TestCategoryHandler_CreateCategory(t *testing.T) {
	handler, categoryService := setupCategoryHandler(t)
	ctx := context.Background()

	tests := []struct {
		name       string
		request    *todov1.CreateCategoryRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful creation",
			request: &todov1.CreateCategoryRequest{
				Name:        "Work",
				Description: stringPtr("Work related tasks"),
				Color:       stringPtr("#FF0000"),
				IsPublic:    true,
			},
			wantErr: false,
		},
		{
			name: "successful creation with default color",
			request: &todov1.CreateCategoryRequest{
				Name:        "Personal",
				Description: stringPtr("Personal tasks"),
				IsPublic:    false,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			request: &todov1.CreateCategoryRequest{
				Name: "",
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.CreateCategory(ctx, tt.request)

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

			if resp.Category.Name != tt.request.Name {
				t.Errorf("expected category name %s, got %s", tt.request.Name, resp.Category.Name)
			}

			// Check default color was set if not provided
			if tt.request.Color == nil && resp.Category.Color != "#6B7280" {
				t.Errorf("expected default color #6B7280, got %s", resp.Category.Color)
			}

			// Verify category was added to mock
			if _, exists := categoryService.categories["category-123"]; !exists {
				t.Error("category was not added to service")
			}
		})
	}
}

func TestCategoryHandler_ListCategories(t *testing.T) {
	handler, categoryService := setupCategoryHandler(t)
	ctx := context.Background()

	// Add test categories
	publicCategory := &domain.Category{
		ID:       "cat-1",
		Name:     "Public Category",
		IsPublic: true,
	}
	privateCategory := &domain.Category{
		ID:       "cat-2",
		Name:     "Private Category",
		IsPublic: false,
	}
	categoryService.categories["cat-1"] = publicCategory
	categoryService.categories["cat-2"] = privateCategory

	tests := []struct {
		name          string
		request       *todov1.ListCategoriesRequest
		wantErr       bool
		expectedCount int
	}{
		{
			name: "list all categories",
			request: &todov1.ListCategoriesRequest{
				PageInfo: &todov1.PageInfo{
					PageSize: 10,
				},
			},
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name: "list public only",
			request: &todov1.ListCategoriesRequest{
				PageInfo: &todov1.PageInfo{
					PageSize: 10,
				},
				PublicOnly: true,
			},
			wantErr:       false,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.ListCategories(ctx, tt.request)

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

			if len(resp.Categories) != tt.expectedCount {
				t.Errorf("expected %d categories, got %d", tt.expectedCount, len(resp.Categories))
			}
		})
	}
}

func TestCategoryHandler_UpdateCategory(t *testing.T) {
	handler, categoryService := setupCategoryHandler(t)
	ctx := context.Background()

	// Add a test category
	testCategory := &domain.Category{
		ID:      "category-1",
		Name:    "Original Name",
		Version: 1,
	}
	categoryService.categories["category-1"] = testCategory

	tests := []struct {
		name       string
		request    *todov1.UpdateCategoryRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful update",
			request: &todov1.UpdateCategoryRequest{
				CategoryId:  "category-1",
				Name:        "Updated Name",
				Description: stringPtr("Updated Description"),
				Color:       stringPtr("#00FF00"),
				IsPublic:    true,
				Version:     1,
			},
			wantErr: false,
		},
		{
			name: "category not found",
			request: &todov1.UpdateCategoryRequest{
				CategoryId: "nonexistent",
				Version:    1,
			},
			wantErr:    true,
			wantStatus: codes.NotFound,
		},
		{
			name: "empty category ID",
			request: &todov1.UpdateCategoryRequest{
				CategoryId: "",
				Version:    1,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
		{
			name: "missing version",
			request: &todov1.UpdateCategoryRequest{
				CategoryId: "category-1",
				Version:    0,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
		{
			name: "version mismatch",
			request: &todov1.UpdateCategoryRequest{
				CategoryId: "category-1",
				Version:    999,
			},
			wantErr:    true,
			wantStatus: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.UpdateCategory(ctx, tt.request)

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

			if tt.request.Name != "" && resp.Category.Name != tt.request.Name {
				t.Errorf("expected category name %s, got %s", tt.request.Name, resp.Category.Name)
			}
		})
	}
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	handler, categoryService := setupCategoryHandler(t)
	ctx := context.Background()

	// Add a test category
	testCategory := &domain.Category{
		ID:      "category-1",
		Name:    "Test Category",
		Version: 1,
	}
	categoryService.categories["category-1"] = testCategory

	tests := []struct {
		name       string
		request    *todov1.DeleteCategoryRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful deletion",
			request: &todov1.DeleteCategoryRequest{
				CategoryId: "category-1",
				Version:    1,
			},
			wantErr: false,
		},
		{
			name: "empty category ID",
			request: &todov1.DeleteCategoryRequest{
				CategoryId: "",
				Version:    1,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
		{
			name: "missing version",
			request: &todov1.DeleteCategoryRequest{
				CategoryId: "category-1",
				Version:    0,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.DeleteCategory(ctx, tt.request)

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

			if !resp.Success {
				t.Error("expected success to be true")
			}
		})
	}
}
