package connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
)

// AdminHandler implements the ConnectRPC AdminService
type AdminHandler struct {
	services *service.Services
	logger   logger.Logger
}

// NewAdminHandler creates a new admin ConnectRPC handler
func NewAdminHandler(services *service.Services, logger logger.Logger) *AdminHandler {
	return &AdminHandler{
		services: services,
		logger:   logger,
	}
}

// User management methods

// ListUsers lists users with pagination
func (h *AdminHandler) ListUsers(ctx context.Context, req *connect.Request[todov1.ListUsersRequest]) (*connect.Response[todov1.ListUsersResponse], error) {
	h.logger.Info(ctx, "Listing users via ConnectRPC", "page_info", req.Msg.GetPageInfo())

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

	// Get users from service
	users, total, err := h.services.User.ListUsers(ctx, opts)
	if err != nil {
		h.logger.Error(ctx, "Failed to list users", "error", err)
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert domain users to protobuf users
	pbUsers := make([]*todov1.User, len(users))
	for i, user := range users {
		pbUsers[i] = user.ToProtobuf()
	}

	response := &todov1.ListUsersResponse{
		Users: pbUsers,
		PageResponse: &todov1.PageResponse{
			TotalCount: int32(total),
		},
	}

	return connect.NewResponse(response), nil
}

// GetUser gets a single user by ID
func (h *AdminHandler) GetUser(ctx context.Context, req *connect.Request[todov1.GetUserRequest]) (*connect.Response[todov1.GetUserResponse], error) {
	userID := req.Msg.GetUserId()
	h.logger.Info(ctx, "Getting user via ConnectRPC", "user_id", userID)

	if userID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, 
			fmt.Errorf("user_id is required"))
	}

	user, err := h.services.User.GetUserByID(ctx, userID)
	if err != nil {
		h.logger.Error(ctx, "Failed to get user", "user_id", userID, "error", err)
		if domain.IsNotFoundError(err) {
			return nil, connect.NewError(connect.CodeNotFound, 
				fmt.Errorf("user not found"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	response := &todov1.GetUserResponse{
		User: user.ToProtobuf(),
	}

	return connect.NewResponse(response), nil
}