package connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
	"github.com/todo-app/services/admin-service/internal/auth"
	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

// CategoryHandler implements the ConnectRPC CategoryService
type CategoryHandler struct {
	categoryService service.CategoryService
	logger          logger.Logger
}

// NewCategoryHandler creates a new category ConnectRPC handler
func NewCategoryHandler(categoryService service.CategoryService, logger logger.Logger) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		logger:          logger,
	}
}

// CreateCategory creates a new category
func (h *CategoryHandler) CreateCategory(ctx context.Context, req *connect.Request[todov1.CreateCategoryRequest]) (*connect.Response[todov1.CreateCategoryResponse], error) {
	h.logger.Info(ctx, "Creating category via ConnectRPC", "name", req.Msg.GetName())

	// Validate request
	if req.Msg.GetName() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name is required"))
	}

	// Get user ID from auth context
	creatorID := auth.GetUserIDFromContext(ctx)
	if creatorID == "" {
		h.logger.Warn(ctx, "No user ID found in context, using system")
		creatorID = "system" // Fallback for unauthenticated requests
	}

	// Create domain category with required fields
	category := &domain.Category{
		Name:      req.Msg.GetName(),
		CreatorID: creatorID,
	}

	// Set optional parent ID if provided
	if req.Msg.ParentId != nil {
		parentId := *req.Msg.ParentId
		if parentId != "" {
			category.ParentID = &parentId
		}
	}

	// Set optional description
	if req.Msg.Description != nil {
		category.Description = *req.Msg.Description
	}

	// Create category
	createdCategory, err := h.categoryService.CreateCategory(ctx, category)
	if err != nil {
		h.logger.Error(ctx, "Failed to create category", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.CreateCategoryResponse{
		Category: createdCategory.ToProtobuf(),
	}

	return connect.NewResponse(response), nil
}

// ListCategories lists categories with pagination
func (h *CategoryHandler) ListCategories(ctx context.Context, req *connect.Request[todov1.ListCategoriesRequest]) (*connect.Response[todov1.ListCategoriesResponse], error) {
	h.logger.Info(ctx, "Listing categories via ConnectRPC", "page_info", req.Msg.GetPageInfo())

	// Convert pagination info
	opts := repository.CategoryListOptions{
		ListOptions: repository.ListOptions{
			Page:           0, // Default to first page for now
			PageSize:       req.Msg.GetPageInfo().GetPageSize(),
			IncludeDeleted: req.Msg.GetIncludeDeleted(),
		},
		PublicOnly: req.Msg.GetPublicOnly(),
	}

	if opts.PageSize == 0 {
		opts.PageSize = 50 // Default page size
	}

	// Get categories from service
	categories, total, err := h.categoryService.ListCategories(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list categories", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert domain categories to protobuf categories
	pbCategories := make([]*todov1.Category, len(categories))
	for i, category := range categories {
		pbCategories[i] = category.ToProtobuf()
	}

	response := &todov1.ListCategoriesResponse{
		Categories: pbCategories,
		PageResponse: &todov1.PageResponse{
			TotalCount: int32(total),
		},
	}

	return connect.NewResponse(response), nil
}

// UpdateCategory updates an existing category
func (h *CategoryHandler) UpdateCategory(ctx context.Context, req *connect.Request[todov1.UpdateCategoryRequest]) (*connect.Response[todov1.UpdateCategoryResponse], error) {
	categoryID := req.Msg.GetCategoryId()
	h.logger.Info(ctx, "Updating category via ConnectRPC", "category_id", categoryID)

	if categoryID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("category_id is required"))
	}

	// Get current category to preserve existing fields
	currentCategory, err := h.categoryService.GetCategoryByID(ctx, categoryID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get category for update", "category_id", categoryID, "error", err)
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("category not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Update fields from request (name is required in protobuf)
	currentCategory.Name = req.Msg.GetName()

	if req.Msg.Description != nil {
		currentCategory.Description = *req.Msg.Description
	}

	// Update parent ID with proper optional field semantics
	if req.Msg.ParentId != nil {
		parentId := *req.Msg.ParentId
		if parentId == "" {
			// Empty string means explicitly remove parent
			h.logger.Info(ctx, "Explicitly removing parent from category", "category_id", req.Msg.GetCategoryId())
			currentCategory.ParentID = nil
		} else {
			// Non-empty string means set parent
			h.logger.Info(ctx, "Setting parent for category", "category_id", req.Msg.GetCategoryId(), "parent_id", parentId)
			currentCategory.ParentID = &parentId
		}
	}

	// Update category
	updatedCategory, err := h.categoryService.UpdateCategory(ctx, currentCategory)
	if err != nil {
		h.logger.Error(ctx, "Failed to update category", "category_id", categoryID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.UpdateCategoryResponse{
		Category: updatedCategory.ToProtobuf(),
	}

	return connect.NewResponse(response), nil
}

// DeleteCategory soft-deletes a category
func (h *CategoryHandler) DeleteCategory(ctx context.Context, req *connect.Request[todov1.DeleteCategoryRequest]) (*connect.Response[todov1.DeleteCategoryResponse], error) {
	categoryID := req.Msg.GetCategoryId()
	h.logger.Info(ctx, "Deleting category via ConnectRPC", "category_id", categoryID)

	if categoryID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("category_id is required"))
	}

	err := h.categoryService.DeleteCategory(ctx, categoryID, req.Msg.GetVersion())
	if err != nil {
		h.logger.Error(ctx, "Failed to delete category", "category_id", categoryID, "error", err)
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("category not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.DeleteCategoryResponse{
		Success: true,
	}

	return connect.NewResponse(response), nil
}
