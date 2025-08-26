package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/todo-app/services/admin-service/internal/auth"
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

	// Get user ID from auth context
	creatorID := auth.GetUserIDFromContext(ctx)
	if creatorID == "" {
		h.logger.Warn(ctx, "No user ID found in context, using system")
		creatorID = "system" // Fallback for unauthenticated requests
	}

	// Create domain category with required fields
	category := &domain.Category{
		Name:      req.GetName(),
		IsPublic:  req.GetIsPublic(),
		CreatorID: creatorID,
	}

	// Set optional description if provided
	if req.Description != nil {
		category.Description = *req.Description
	}

	// Set optional color if provided, otherwise use default
	if req.Color != nil {
		category.Color = *req.Color
	} else {
		category.Color = "#6B7280" // Default gray
	}

	// Set optional parent ID if provided
	if req.ParentId != nil {
		parentId := *req.ParentId
		if parentId != "" {
			category.ParentID = &parentId
		}
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
	opts := repository.CategoryListOptions{
		ListOptions: repository.ListOptions{
			Page:           0, // Default to first page for now
			PageSize:       req.GetPageInfo().GetPageSize(),
			IncludeDeleted: req.GetIncludeDeleted(),
		},
		PublicOnly: req.GetPublicOnly(),
	}

	if opts.PageSize == 0 {
		opts.PageSize = 50 // Default page size
	}

	// Get categories from service (filtering is now handled at the database level)
	categories, total, err := h.categoryService.ListCategories(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list categories", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list categories: %v", err)
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

	// Update fields that are provided (using proper optional field semantics)
	currentCategory.Name = req.GetName() // Always update name (required field)
	
	// Update optional description if provided
	if req.Description != nil {
		currentCategory.Description = *req.Description
	}
	
	// Update optional color if provided  
	if req.Color != nil {
		currentCategory.Color = *req.Color
	}

	// Update parent ID with proper optional field semantics
	if req.ParentId != nil {
		parentId := *req.ParentId
		if parentId == "" {
			// Empty string means explicitly remove parent
			h.logger.Info(ctx, "Explicitly removing parent from category", "category_id", req.GetCategoryId())
			currentCategory.ParentID = nil
		} else {
			// Non-empty string means set parent
			h.logger.Info(ctx, "Setting parent for category", "category_id", req.GetCategoryId(), "parent_id", parentId)
			currentCategory.ParentID = &parentId
		}
	}
	// If ParentId field is nil, we don't change the current parent (leave it as-is)

	// Update public flag (always update - required field)
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
