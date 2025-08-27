package connect

import (
	"net/http"

	"github.com/lloydsk/todo-app-proto/gen/go/todo/v1/v1connect"
	"github.com/todo-app/services/admin-service/internal/service"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

// Handler holds the ConnectRPC handlers and their dependencies
type Handler struct {
	services *service.Services
	logger   logger.Logger
}

// NewHandler creates a new ConnectRPC handler
func NewHandler(services *service.Services, logger logger.Logger) *Handler {
	return &Handler{
		services: services,
		logger:   logger,
	}
}

// RegisterServices registers all ConnectRPC services with the HTTP mux
func (h *Handler) RegisterServices(mux *http.ServeMux) {
	// Register full admin service with all task management methods
	adminHandler := NewAdminHandler(h.services, h.logger)
	path, handler := v1connect.NewAdminServiceHandler(adminHandler)
	mux.Handle(path, handler)

	// Register category service
	categoryHandler := NewCategoryHandler(h.services.Category, h.logger)
	path, handler = v1connect.NewCategoryServiceHandler(categoryHandler)
	mux.Handle(path, handler)

	// Register tag service
	tagHandler := NewTagHandler(h.services.Tag, h.logger)
	path, handler = v1connect.NewTagServiceHandler(tagHandler)
	mux.Handle(path, handler)

	// Note: UserService is for mobile interface - implement separately if needed
}
