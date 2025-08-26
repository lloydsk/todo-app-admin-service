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
		Page:           int(req.GetPageInfo().GetPage()),
		PageSize:       int(req.GetPageInfo().GetPageSize()),
		IncludeDeleted: req.GetIncludeDeleted(),
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
			Page:       int32(opts.Page),
			PageSize:   int32(opts.PageSize),
			TotalCount: total,
			HasMore:    total > int64((opts.Page+1)*opts.PageSize),
		},
	}

	h.logger.Info(ctx, "Listed categories successfully", "count", len(categories), "total", total)
	return response, nil
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
