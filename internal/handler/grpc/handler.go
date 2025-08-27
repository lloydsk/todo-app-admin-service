package grpc

import (
	"google.golang.org/grpc"

	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
)

// Handler holds the gRPC handlers and their dependencies
type Handler struct {
	services *service.Services
	logger   logger.Logger
}

// NewHandler creates a new gRPC handler
func NewHandler(services *service.Services, logger logger.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

// RegisterServices registers all gRPC services with the server
func (h *Handler) RegisterServices(server *grpc.Server) {
	// Register admin service (for web interface)
	adminHandler := NewAdminHandler(h.services, h.logger)
	todov1.RegisterAdminServiceServer(server, adminHandler)

	// Register category service
	categoryHandler := NewCategoryHandler(h.services.Category, h.logger)
	todov1.RegisterCategoryServiceServer(server, categoryHandler)

	// Register tag service
	tagHandler := NewTagHandler(h.services.Tag, h.logger)
	todov1.RegisterTagServiceServer(server, tagHandler)

	// Note: UserService is for mobile interface - implement separately if needed
}
