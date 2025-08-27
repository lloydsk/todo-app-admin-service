package connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/todo-app/services/admin-service/internal/auth"
	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
)

// TagHandler implements the ConnectRPC TagService
type TagHandler struct {
	tagService service.TagService
	logger     logger.Logger
}

// NewTagHandler creates a new tag ConnectRPC handler
func NewTagHandler(tagService service.TagService, logger logger.Logger) *TagHandler {
	return &TagHandler{
		tagService: tagService,
		logger:     logger,
	}
}

// CreateTag creates a new tag
func (h *TagHandler) CreateTag(ctx context.Context, req *connect.Request[todov1.CreateTagRequest]) (*connect.Response[todov1.CreateTagResponse], error) {
	h.logger.Info(ctx, "Creating tag via ConnectRPC", "name", req.Msg.GetName())

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

	// Create domain tag with required fields
	tag := &domain.Tag{
		Name:      req.Msg.GetName(),
		CreatorID: creatorID,
	}

	// Set optional color
	if req.Msg.Color != nil {
		tag.Color = *req.Msg.Color
	}

	// Create tag
	createdTag, err := h.tagService.CreateTag(ctx, tag)
	if err != nil {
		h.logger.Error(ctx, "Failed to create tag", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.CreateTagResponse{
		Tag: createdTag.ToProtobuf(),
	}

	return connect.NewResponse(response), nil
}


// ListTags lists tags with pagination
func (h *TagHandler) ListTags(ctx context.Context, req *connect.Request[todov1.ListTagsRequest]) (*connect.Response[todov1.ListTagsResponse], error) {
	h.logger.Info(ctx, "Listing tags via ConnectRPC", "page_info", req.Msg.GetPageInfo())

	// Convert pagination info
	opts := repository.ListOptions{
		Page:           0, // Default to first page for now
		PageSize:       req.Msg.GetPageInfo().GetPageSize(),
		SearchQuery:    req.Msg.GetSearchQuery(),
		IncludeDeleted: req.Msg.GetIncludeDeleted(),
	}

	if opts.PageSize == 0 {
		opts.PageSize = 50 // Default page size
	}

	// Get tags from service
	tags, total, err := h.tagService.ListTags(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list tags", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert domain tags to protobuf tags
	pbTags := make([]*todov1.Tag, len(tags))
	for i, tag := range tags {
		pbTags[i] = tag.ToProtobuf()
	}

	response := &todov1.ListTagsResponse{
		Tags: pbTags,
		PageResponse: &todov1.PageResponse{
			TotalCount: int32(total),
		},
	}

	return connect.NewResponse(response), nil
}

// UpdateTag updates an existing tag
func (h *TagHandler) UpdateTag(ctx context.Context, req *connect.Request[todov1.UpdateTagRequest]) (*connect.Response[todov1.UpdateTagResponse], error) {
	tagID := req.Msg.GetTagId()
	h.logger.Info(ctx, "Updating tag via ConnectRPC", "tag_id", tagID)

	if tagID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("tag_id is required"))
	}

	// Get current tag to preserve existing fields
	currentTag, err := h.tagService.GetTagByID(ctx, tagID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get tag for update", "tag_id", tagID, "error", err)
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("tag not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Update fields from request (name is required in protobuf)
	currentTag.Name = req.Msg.GetName()

	if req.Msg.Color != nil {
		currentTag.Color = *req.Msg.Color
	}

	// Update tag
	updatedTag, err := h.tagService.UpdateTag(ctx, currentTag)
	if err != nil {
		h.logger.Error(ctx, "Failed to update tag", "tag_id", tagID, "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.UpdateTagResponse{
		Tag: updatedTag.ToProtobuf(),
	}

	return connect.NewResponse(response), nil
}

// DeleteTag soft-deletes a tag
func (h *TagHandler) DeleteTag(ctx context.Context, req *connect.Request[todov1.DeleteTagRequest]) (*connect.Response[todov1.DeleteTagResponse], error) {
	tagID := req.Msg.GetTagId()
	h.logger.Info(ctx, "Deleting tag via ConnectRPC", "tag_id", tagID)

	if tagID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("tag_id is required"))
	}

	err := h.tagService.DeleteTag(ctx, tagID, req.Msg.GetVersion())
	if err != nil {
		h.logger.Error(ctx, "Failed to delete tag", "tag_id", tagID, "error", err)
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("tag not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.DeleteTagResponse{
		Success: true,
	}

	return connect.NewResponse(response), nil
}