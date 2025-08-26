package service

import (
	"context"
	"fmt"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type userService struct {
	userRepo repository.UserRepository
	logger   logger.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, log logger.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   log,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	s.logger.Info(ctx, "Creating new user", "email", user.Email, "role", user.Role)
	
	// Business validation
	if err := s.validateUserForCreation(ctx, user); err != nil {
		return nil, err
	}
	
	// Check for duplicate email
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !domain.IsNotFoundError(err) {
		return nil, fmt.Errorf("failed to check for existing user: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrConflict("user with email already exists")
	}
	
	// Create user
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.Error(ctx, "Failed to create user", "error", err, "email", user.Email)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	s.logger.Info(ctx, "User created successfully", "user_id", user.ID, "email", user.Email)
	return user, nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	s.logger.Debug(ctx, "Getting user by ID", "user_id", id)
	
	if id == "" {
		return nil, domain.ErrInvalidInput("user ID is required")
	}
	
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if domain.IsNotFoundError(err) {
			s.logger.Debug(ctx, "User not found", "user_id", id)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to get user by ID", "error", err, "user_id", id)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	s.logger.Debug(ctx, "Getting user by email", "email", email)
	
	if email == "" {
		return nil, domain.ErrInvalidInput("email is required")
	}
	
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if domain.IsNotFoundError(err) {
			s.logger.Debug(ctx, "User not found", "email", email)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to get user by email", "error", err, "email", email)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	s.logger.Info(ctx, "Updating user", "user_id", user.ID, "version", user.Version)
	
	// Business validation
	if err := s.validateUserForUpdate(ctx, user); err != nil {
		return nil, err
	}
	
	// Check if email is being changed and if new email already exists
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing user: %w", err)
	}
	
	if existingUser.Email != user.Email {
		conflictUser, err := s.userRepo.GetByEmail(ctx, user.Email)
		if err != nil && !domain.IsNotFoundError(err) {
			return nil, fmt.Errorf("failed to check for email conflict: %w", err)
		}
		if conflictUser != nil && conflictUser.ID != user.ID {
			return nil, domain.ErrConflict("email already in use by another user")
		}
	}
	
	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "User update version conflict", "user_id", user.ID, "version", user.Version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to update user", "error", err, "user_id", user.ID)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	s.logger.Info(ctx, "User updated successfully", "user_id", user.ID, "new_version", user.Version)
	return user, nil
}

func (s *userService) DeleteUser(ctx context.Context, id string, version int64) error {
	s.logger.Info(ctx, "Soft deleting user", "user_id", id, "version", version)
	
	if id == "" {
		return domain.ErrInvalidInput("user ID is required")
	}
	
	// Business rule: Check if user can be deleted (e.g., admin users, active assignments)
	if err := s.validateUserForDeletion(ctx, id); err != nil {
		return err
	}
	
	if err := s.userRepo.SoftDelete(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "User deletion version conflict", "user_id", id, "version", version)
			return err
		}
		s.logger.Error(ctx, "Failed to delete user", "error", err, "user_id", id)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	s.logger.Info(ctx, "User deleted successfully", "user_id", id)
	return nil
}

func (s *userService) RestoreUser(ctx context.Context, id string, version int64) (*domain.User, error) {
	s.logger.Info(ctx, "Restoring user", "user_id", id, "version", version)
	
	if id == "" {
		return nil, domain.ErrInvalidInput("user ID is required")
	}
	
	if err := s.userRepo.Restore(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "User restoration version conflict", "user_id", id, "version", version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to restore user", "error", err, "user_id", id)
		return nil, fmt.Errorf("failed to restore user: %w", err)
	}
	
	// Get the restored user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored user: %w", err)
	}
	
	s.logger.Info(ctx, "User restored successfully", "user_id", id)
	return user, nil
}

func (s *userService) ListUsers(ctx context.Context, opts repository.ListOptions) ([]*domain.User, int64, error) {
	s.logger.Debug(ctx, "Listing users", "page", opts.Page, "page_size", opts.PageSize)
	
	users, total, err := s.userRepo.List(ctx, opts)
	if err != nil {
		s.logger.Error(ctx, "Failed to list users", "error", err)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	
	s.logger.Debug(ctx, "Listed users successfully", "count", len(users), "total", total)
	return users, total, nil
}

func (s *userService) ChangeUserRole(ctx context.Context, userID string, newRole domain.UserRole, version int64) (*domain.User, error) {
	s.logger.Info(ctx, "Changing user role", "user_id", userID, "new_role", newRole, "version", version)
	
	// Get current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Validate version
	if user.Version != version {
		return nil, domain.ErrVersionConflict("user", version, user.Version)
	}
	
	// Business validation for role change
	if err := s.validateRoleChange(ctx, user, newRole); err != nil {
		return nil, err
	}
	
	// Update role
	user.Role = newRole
	return s.UpdateUser(ctx, user)
}

func (s *userService) ValidateUserPermissions(ctx context.Context, userID string, requiredRole domain.UserRole) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user for permission check: %w", err)
	}
	
	if !s.hasRequiredRole(user.Role, requiredRole) {
		s.logger.Warn(ctx, "User lacks required permissions", 
			"user_id", userID, "user_role", user.Role, "required_role", requiredRole)
		return domain.ErrPermissionDenied("insufficient role privileges")
	}
	
	return nil
}

// Helper methods for business validation

func (s *userService) validateUserForCreation(ctx context.Context, user *domain.User) error {
	if err := user.IsValid(); err != nil {
		return err
	}
	
	// Additional business rules for user creation
	if user.Role == domain.UserRoleUnspecified {
		return domain.ErrInvalidInput("user role must be specified")
	}
	
	return nil
}

func (s *userService) validateUserForUpdate(ctx context.Context, user *domain.User) error {
	if user.ID == "" {
		return domain.ErrInvalidInput("user ID is required for update")
	}
	
	if err := user.IsValid(); err != nil {
		return err
	}
	
	return nil
}

func (s *userService) validateUserForDeletion(ctx context.Context, userID string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user for deletion validation: %w", err)
	}
	
	// Business rule: Prevent deleting the last admin user
	if user.Role == domain.UserRoleAdmin {
		adminCount, err := s.countAdminUsers(ctx)
		if err != nil {
			return fmt.Errorf("failed to count admin users: %w", err)
		}
		
		if adminCount <= 1 {
			return domain.ErrBusinessRule("cannot delete the last admin user")
		}
	}
	
	return nil
}

func (s *userService) validateRoleChange(ctx context.Context, user *domain.User, newRole domain.UserRole) error {
	// Business rule: Prevent removing admin role from the last admin
	if user.Role == domain.UserRoleAdmin && newRole != domain.UserRoleAdmin {
		adminCount, err := s.countAdminUsers(ctx)
		if err != nil {
			return fmt.Errorf("failed to count admin users: %w", err)
		}
		
		if adminCount <= 1 {
			return domain.ErrBusinessRule("cannot remove admin role from the last admin user")
		}
	}
	
	return nil
}

func (s *userService) countAdminUsers(ctx context.Context) (int64, error) {
	// This is a simplified implementation - in a real system you might have a dedicated query
	users, _, err := s.userRepo.List(ctx, repository.ListOptions{
		PageSize:       1000, // Large enough to get all users for counting
		IncludeDeleted: false,
	})
	if err != nil {
		return 0, err
	}
	
	var count int64
	for _, user := range users {
		if user.Role == domain.UserRoleAdmin {
			count++
		}
	}
	
	return count, nil
}

func (s *userService) hasRequiredRole(userRole, requiredRole domain.UserRole) bool {
	roleHierarchy := map[domain.UserRole]int{
		domain.UserRoleUser:  1,
		domain.UserRoleAdmin: 2,
	}
	
	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[requiredRole]
	
	return userLevel >= requiredLevel
}