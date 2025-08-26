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

	// Validate request
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	// Create domain category
	category := &domain.Category{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Color:       req.GetColor(),
		IsPublic:    req.GetIsPublic(),
		CreatorID:   "system", // TODO: Get from auth context
	}

	// Set parent ID if provided
	if req.GetParentId() != "" {
		category.ParentID = &req.ParentId
	}

	// Set default color if not provided
	if category.Color == "" {
		category.Color = "#6B7280" // Default gray
	}

	// Create category via service
	createdCategory, err := h.categoryService.CreateCategory(ctx, category)
	if err != nil {
		h.logger.Error(ctx, "Failed to create category", "name", req.GetName(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}

	// Build response
	response := &todov1.CreateCategoryResponse{
		Category: createdCategory.ToProtobuf(),
	}

	h.logger.Info(ctx, "Created category successfully", "category_id", createdCategory.ID, "name", createdCategory.Name)
	return response, nil
}

// ListCategories lists categories with pagination
func (h *CategoryHandler) ListCategories(ctx context.Context, req *todov1.ListCategoriesRequest) (*todov1.ListCategoriesResponse, error) {
	h.logger.Info(ctx, "Listing categories via gRPC", "page_info", req.GetPageInfo())

	// Convert pagination info
	opts := repository.ListOptions{
		Page:           0, // Default to first page for now
		PageSize:       req.GetPageInfo().GetPageSize(),
		IncludeDeleted: req.GetIncludeDeleted(),
	}

	if opts.PageSize == 0 {
		opts.PageSize = 50 // Default page size
	}

	// Get categories from service
	categories, total, err := h.categoryService.ListCategories(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list categories", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
	}

	// Filter by public only if requested
	if req.GetPublicOnly() {
		filteredCategories := make([]*domain.Category, 0, len(categories))
		for _, category := range categories {
			if category.IsPublic {
				filteredCategories = append(filteredCategories, category)
			}
		}
		categories = filteredCategories
		total = int64(len(categories)) // Recalculate total for filtered results
	}

	// Convert domain categories to protobuf
	pbCategories := make([]*todov1.Category, len(categories))
	for i, category := range categories {
		pbCategories[i] = category.ToProtobuf()
	}

	// Build response
	response := &todov1.ListCategoriesResponse{
		Categories: pbCategories,
		PageResponse: &todov1.PageResponse{
			NextPageToken: "", // TODO: Implement token-based pagination
			TotalCount:    int32(total),
		},
	}

	h.logger.Info(ctx, "Listed categories successfully", "count", len(categories), "total", total)
	return response, nil
}

// UpdateCategory updates an existing category
func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *todov1.UpdateCategoryRequest) (*todov1.UpdateCategoryResponse, error) {
	h.logger.Info(ctx, "Updating category via gRPC", "category_id", req.GetCategoryId())

	// Validate request
	if req.GetCategoryId() == "" {
		return nil, status.Error(codes.InvalidArgument, "category_id is required")
	}
	if req.GetVersion() == 0 {
		return nil, status.Error(codes.InvalidArgument, "version is required for optimistic locking")
	}

	// Get current category to get current data
	currentCategory, err := h.categoryService.GetCategoryByID(ctx, req.GetCategoryId())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "category not found: %s", req.GetCategoryId())
		}
		h.logger.Error(ctx, "Failed to get current category", "category_id", req.GetCategoryId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get current category: %v", err)
	}

	// Check version for optimistic locking
	if currentCategory.Version != req.GetVersion() {
		return nil, status.Errorf(codes.FailedPrecondition, "category version mismatch: expected %d, got %d",
			currentCategory.Version, req.GetVersion())
	}

	// Update fields that are provided
	if req.GetName() != "" {
		currentCategory.Name = req.GetName()
	}
	if req.GetDescription() != "" {
		currentCategory.Description = req.GetDescription()
	}
	if req.GetColor() != "" {
		currentCategory.Color = req.GetColor()
	}

	// Update parent ID (can be set to empty to remove parent)
	if req.GetParentId() != "" {
		currentCategory.ParentID = &req.ParentId
	} else {
		currentCategory.ParentID = nil
	}

	// Update public flag
	currentCategory.IsPublic = req.GetIsPublic()

	// Update category via service
	updatedCategory, err := h.categoryService.UpdateCategory(ctx, currentCategory)
	if err != nil {
		if domain.IsVersionConflictError(err) {
			return nil, status.Error(codes.FailedPrecondition, "category was modified by another request")
		}
		h.logger.Error(ctx, "Failed to update category", "category_id", req.GetCategoryId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}

	// Build response
	response := &todov1.UpdateCategoryResponse{
		Category: updatedCategory.ToProtobuf(),
	}

	h.logger.Info(ctx, "Updated category successfully", "category_id", req.GetCategoryId())
	return response, nil
}

// DeleteCategory deletes a category
func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *todov1.DeleteCategoryRequest) (*todov1.DeleteCategoryResponse, error) {
	h.logger.Info(ctx, "Deleting category via gRPC", "category_id", req.GetCategoryId())

	// Validate request
	if req.GetCategoryId() == "" {
		return nil, status.Error(codes.InvalidArgument, "category_id is required")
	}
	if req.GetVersion() == 0 {
		return nil, status.Error(codes.InvalidArgument, "version is required for optimistic locking")
	}

	// Validate category usage before deletion
	err := h.categoryService.ValidateCategoryUsage(ctx, req.GetCategoryId())
	if err != nil {
		h.logger.Error(ctx, "Category validation failed", "category_id", req.GetCategoryId(), "error", err)
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete category: %v", err)
	}

	// Delete category via service (soft delete)
	err = h.categoryService.DeleteCategory(ctx, req.GetCategoryId(), req.GetVersion())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "category not found: %s", req.GetCategoryId())
		}
		if domain.IsVersionConflictError(err) {
			return nil, status.Error(codes.FailedPrecondition, "category was modified by another request")
		}
		h.logger.Error(ctx, "Failed to delete category", "category_id", req.GetCategoryId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}

	// Build response
	response := &todov1.DeleteCategoryResponse{
		Success: true,
	}

	h.logger.Info(ctx, "Deleted category successfully", "category_id", req.GetCategoryId())
	return response, nil
}
