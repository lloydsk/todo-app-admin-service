package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/todo-app/services/admin-service/proto/gen/go/todo/v1"
)

// CategoryHandler implements the gRPC CategoryService
type CategoryHandler struct {
	todov1.UnimplementedCategoryServiceServer
	categoryService service.CategoryService
	logger          logger.Logger
}

// NewCategoryHandler creates a new category gRPC handler
func NewCategoryHandler(categoryService service.CategoryService, logger logger.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

// CreateCategory creates a new category
func (h *CategoryHandler) CreateCategory(ctx context.Context, req *todov1.CreateCategoryRequest) (*todov1.CreateCategoryResponse, error) {
	h.logger.Info(ctx, "Creating category via gRPC", "name", req.GetName())
	
	// TODO: Implement category creation
	return nil, status.Error(codes.Unimplemented, "CreateCategory not yet implemented")
}

// ListCategories lists categories with pagination
func (h *CategoryHandler) ListCategories(ctx context.Context, req *todov1.ListCategoriesRequest) (*todov1.ListCategoriesResponse, error) {
	h.logger.Info(ctx, "Listing categories via gRPC", "page_info", req.GetPageInfo())
	
	// TODO: Implement category listing
	return nil, status.Error(codes.Unimplemented, "ListCategories not yet implemented")
}

// UpdateCategory updates an existing category
func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *todov1.UpdateCategoryRequest) (*todov1.UpdateCategoryResponse, error) {
	h.logger.Info(ctx, "Updating category via gRPC", "category_id", req.GetCategoryId())
	
	// TODO: Implement category update
	return nil, status.Error(codes.Unimplemented, "UpdateCategory not yet implemented")
}

// DeleteCategory deletes a category
func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *todov1.DeleteCategoryRequest) (*todov1.DeleteCategoryResponse, error) {
	h.logger.Info(ctx, "Deleting category via gRPC", "category_id", req.GetCategoryId())
	
	// TODO: Implement category deletion
	return nil, status.Error(codes.Unimplemented, "DeleteCategory not yet implemented")
}