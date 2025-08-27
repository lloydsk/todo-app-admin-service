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
	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
)

// TagHandler implements the gRPC TagService
type TagHandler struct {
	todov1.UnimplementedTagServiceServer
	tagService service.TagService
	logger     logger.Logger
}

// NewTagHandler creates a new tag gRPC handler
func NewTagHandler(tagService service.TagService, logger logger.Logger) *TagHandler {
	return &TagHandler{
		tagService: tagService,
		logger:     logger,
	}
}

// CreateTag creates a new tag
func (h *TagHandler) CreateTag(ctx context.Context, req *todov1.CreateTagRequest) (*todov1.CreateTagResponse, error) {
	h.logger.Info(ctx, "Creating tag via gRPC", "name", req.GetName())

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

	// Create domain tag with required fields
	tag := &domain.Tag{
		Name:      req.GetName(),
		CreatorID: creatorID,
	}

	// Set optional color if provided, otherwise use default
	if req.Color != nil {
		tag.Color = *req.Color
	} else {
		tag.Color = "#6B7280" // Default gray
	}

	// Create tag via service
	createdTag, err := h.tagService.CreateTag(ctx, tag)
	if err != nil {
		h.logger.Error(ctx, "Failed to create tag", "name", req.GetName(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create tag: %v", err)
	}

	// Build response
	response := &todov1.CreateTagResponse{
		Tag: createdTag.ToProtobuf(),
	}

	h.logger.Info(ctx, "Created tag successfully", "tag_id", createdTag.ID, "name", createdTag.Name)
	return response, nil
}

// ListTags lists tags with pagination
func (h *TagHandler) ListTags(ctx context.Context, req *todov1.ListTagsRequest) (*todov1.ListTagsResponse, error) {
	h.logger.Info(ctx, "Listing tags via gRPC", "page_info", req.GetPageInfo())

	// Convert pagination info
	opts := repository.ListOptions{
		Page:           0, // Default to first page for now
		PageSize:       req.GetPageInfo().GetPageSize(),
		SearchQuery:    req.GetSearchQuery(),
		IncludeDeleted: req.GetIncludeDeleted(),
	}

	if opts.PageSize == 0 {
		opts.PageSize = 50 // Default page size
	}

	// Get tags from service
	tags, total, err := h.tagService.ListTags(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list tags", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list tags: %v", err)
	}

	// Convert domain tags to protobuf
	pbTags := make([]*todov1.Tag, len(tags))
	for i, tag := range tags {
		pbTags[i] = tag.ToProtobuf()
	}

	// Build response
	response := &todov1.ListTagsResponse{
		Tags: pbTags,
		PageResponse: &todov1.PageResponse{
			NextPageToken: "", // TODO: Implement token-based pagination
			TotalCount:    int32(total),
		},
	}

	h.logger.Info(ctx, "Listed tags successfully", "count", len(tags), "total", total)
	return response, nil
}

// UpdateTag updates an existing tag
func (h *TagHandler) UpdateTag(ctx context.Context, req *todov1.UpdateTagRequest) (*todov1.UpdateTagResponse, error) {
	h.logger.Info(ctx, "Updating tag via gRPC", "tag_id", req.GetTagId())

	// Validate request
	if req.GetTagId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tag_id is required")
	}
	if req.GetVersion() == 0 {
		return nil, status.Error(codes.InvalidArgument, "version is required for optimistic locking")
	}

	// Get current tag to get current data
	currentTag, err := h.tagService.GetTagByID(ctx, req.GetTagId())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "tag not found: %s", req.GetTagId())
		}
		h.logger.Error(ctx, "Failed to get current tag", "tag_id", req.GetTagId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get current tag: %v", err)
	}

	// Check version for optimistic locking
	if currentTag.Version != req.GetVersion() {
		return nil, status.Errorf(codes.FailedPrecondition, "tag version mismatch: expected %d, got %d",
			currentTag.Version, req.GetVersion())
	}

	// Update fields that are provided (using proper optional field semantics)
	currentTag.Name = req.GetName() // Always update name (required field)
	
	// Update optional color if provided
	if req.Color != nil {
		currentTag.Color = *req.Color
	}

	// Update tag via service
	updatedTag, err := h.tagService.UpdateTag(ctx, currentTag)
	if err != nil {
		if domain.IsVersionConflictError(err) {
			return nil, status.Error(codes.FailedPrecondition, "tag was modified by another request")
		}
		h.logger.Error(ctx, "Failed to update tag", "tag_id", req.GetTagId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update tag: %v", err)
	}

	// Build response
	response := &todov1.UpdateTagResponse{
		Tag: updatedTag.ToProtobuf(),
	}

	h.logger.Info(ctx, "Updated tag successfully", "tag_id", req.GetTagId())
	return response, nil
}

// DeleteTag deletes a tag
func (h *TagHandler) DeleteTag(ctx context.Context, req *todov1.DeleteTagRequest) (*todov1.DeleteTagResponse, error) {
	h.logger.Info(ctx, "Deleting tag via gRPC", "tag_id", req.GetTagId())

	// Validate request
	if req.GetTagId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tag_id is required")
	}
	if req.GetVersion() == 0 {
		return nil, status.Error(codes.InvalidArgument, "version is required for optimistic locking")
	}

	// Validate tag usage before deletion
	err := h.tagService.ValidateTagUsage(ctx, req.GetTagId())
	if err != nil {
		h.logger.Error(ctx, "Tag validation failed", "tag_id", req.GetTagId(), "error", err)
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete tag: %v", err)
	}

	// Delete tag via service (soft delete)
	err = h.tagService.DeleteTag(ctx, req.GetTagId(), req.GetVersion())
	if err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "tag not found: %s", req.GetTagId())
		}
		if domain.IsVersionConflictError(err) {
			return nil, status.Error(codes.FailedPrecondition, "tag was modified by another request")
		}
		h.logger.Error(ctx, "Failed to delete tag", "tag_id", req.GetTagId(), "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete tag: %v", err)
	}

	// Build response
	response := &todov1.DeleteTagResponse{
		Success: true,
	}

	h.logger.Info(ctx, "Deleted tag successfully", "tag_id", req.GetTagId())
	return response, nil
}
