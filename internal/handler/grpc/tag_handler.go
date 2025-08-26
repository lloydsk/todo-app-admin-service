package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/todo-app/services/admin-service/proto/gen/go/todo/v1"
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
	
	// TODO: Implement tag creation
	return nil, status.Error(codes.Unimplemented, "CreateTag not yet implemented")
}

// ListTags lists tags with pagination
func (h *TagHandler) ListTags(ctx context.Context, req *todov1.ListTagsRequest) (*todov1.ListTagsResponse, error) {
	h.logger.Info(ctx, "Listing tags via gRPC", "page_info", req.GetPageInfo())
	
	// TODO: Implement tag listing
	return nil, status.Error(codes.Unimplemented, "ListTags not yet implemented")
}

// UpdateTag updates an existing tag
func (h *TagHandler) UpdateTag(ctx context.Context, req *todov1.UpdateTagRequest) (*todov1.UpdateTagResponse, error) {
	h.logger.Info(ctx, "Updating tag via gRPC", "tag_id", req.GetTagId())
	
	// TODO: Implement tag update
	return nil, status.Error(codes.Unimplemented, "UpdateTag not yet implemented")
}

// DeleteTag deletes a tag
func (h *TagHandler) DeleteTag(ctx context.Context, req *todov1.DeleteTagRequest) (*todov1.DeleteTagResponse, error) {
	h.logger.Info(ctx, "Deleting tag via gRPC", "tag_id", req.GetTagId())
	
	// TODO: Implement tag deletion
	return nil, status.Error(codes.Unimplemented, "DeleteTag not yet implemented")
}