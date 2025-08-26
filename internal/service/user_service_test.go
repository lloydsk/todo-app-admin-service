package service

import (
	"context"
	"testing"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/internal/testutil"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type mockUserRepository struct {
	users    map[string]*domain.User
	emailIdx map[string]*domain.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:    make(map[string]*domain.User),
		emailIdx: make(map[string]*domain.User),
	}
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if _, exists := m.emailIdx[user.Email]; exists {
		return domain.ErrConflict("user with email already exists")
	}
	
	user.ID = "mock-user-" + user.Email
	user.Version = 1
	m.users[user.ID] = user
	m.emailIdx[user.Email] = user
	return nil
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, domain.ErrNotFound("user")
	}
	return user, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, exists := m.emailIdx[email]
	if !exists {
		return nil, domain.ErrNotFound("user")
	}
	return user, nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	existing, exists := m.users[user.ID]
	if !exists {
		return domain.ErrNotFound("user")
	}
	if existing.Version != user.Version {
		return domain.ErrVersionConflict("user", user.Version, existing.Version)
	}
	
	user.Version++
	m.users[user.ID] = user
	if existing.Email != user.Email {
		delete(m.emailIdx, existing.Email)
		m.emailIdx[user.Email] = user
	}
	return nil
}

func (m *mockUserRepository) SoftDelete(ctx context.Context, id string, version int64) error {
	existing, exists := m.users[id]
	if !exists {
		return domain.ErrNotFound("user")
	}
	if existing.Version != version {
		return domain.ErrVersionConflict("user", version, existing.Version)
	}
	
	existing.IsDeleted = true
	existing.Version++
	delete(m.users, id)
	delete(m.emailIdx, existing.Email)
	return nil
}

func (m *mockUserRepository) Restore(ctx context.Context, id string, version int64) error {
	return nil // Not implemented for mock
}

func (m *mockUserRepository) List(ctx context.Context, opts repository.ListOptions) ([]*domain.User, int64, error) {
	users := make([]*domain.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, int64(len(users)), nil
}

func TestUserService_CreateUser(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockLogger := logger.NewLogger("debug")
	service := NewUserService(mockRepo, mockLogger)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *domain.User
		wantErr bool
	}{
		{
			name: "valid user creation",
			user: &domain.User{
				Name:  "Test User",
				Email: "test@example.com",
				Role:  domain.UserRoleUser,
			},
			wantErr: false,
		},
		{
			name: "duplicate email",
			user: &domain.User{
				Name:  "Another User",
				Email: "test@example.com", // Same email as above
				Role:  domain.UserRoleUser,
			},
			wantErr: true,
		},
		{
			name: "invalid user - missing name",
			user: &domain.User{
				Name:  "",
				Email: "invalid@example.com",
				Role:  domain.UserRoleUser,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateUser(ctx, tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_ChangeUserRole(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockLogger := logger.NewLogger("debug")
	service := NewUserService(mockRepo, mockLogger)
	ctx := context.Background()

	// Create a test user first
	testUser := testutil.TestUser()
	testUser.Role = domain.UserRoleUser
	mockRepo.Create(ctx, testUser)

	tests := []struct {
		name    string
		userID  string
		newRole domain.UserRole
		version int64
		wantErr bool
	}{
		{
			name:    "valid role change",
			userID:  testUser.ID,
			newRole: domain.UserRoleAdmin,
			version: testUser.Version,
			wantErr: false,
		},
		{
			name:    "invalid role",
			userID:  testUser.ID,
			newRole: "INVALID_ROLE",
			version: testUser.Version + 1,
			wantErr: true,
		},
		{
			name:    "user not found",
			userID:  "non-existent",
			newRole: domain.UserRoleAdmin,
			version: 1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ChangeUserRole(ctx, tt.userID, tt.newRole, tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChangeUserRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserService_ValidateUserPermissions(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockLogger := logger.NewLogger("debug")
	service := NewUserService(mockRepo, mockLogger)
	ctx := context.Background()

	// Create test users
	adminUser := testutil.TestUser()
	adminUser.Role = domain.UserRoleAdmin
	adminUser.Email = "admin@example.com"
	mockRepo.Create(ctx, adminUser)

	regularUser := testutil.TestUser()
	regularUser.Role = domain.UserRoleUser
	regularUser.Email = "user@example.com"
	mockRepo.Create(ctx, regularUser)

	tests := []struct {
		name         string
		userID       string
		requiredRole domain.UserRole
		wantErr      bool
	}{
		{
			name:         "admin can access admin role",
			userID:       adminUser.ID,
			requiredRole: domain.UserRoleAdmin,
			wantErr:      false,
		},
		{
			name:         "admin can access user role",
			userID:       adminUser.ID,
			requiredRole: domain.UserRoleUser,
			wantErr:      false,
		},
		{
			name:         "user can access user role",
			userID:       regularUser.ID,
			requiredRole: domain.UserRoleUser,
			wantErr:      false,
		},
		{
			name:         "user cannot access admin role",
			userID:       regularUser.ID,
			requiredRole: domain.UserRoleAdmin,
			wantErr:      true,
		},
		{
			name:         "non-existent user",
			userID:       "non-existent",
			requiredRole: domain.UserRoleUser,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateUserPermissions(ctx, tt.userID, tt.requiredRole)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUserPermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}