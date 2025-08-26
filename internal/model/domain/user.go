package domain

import (
	"time"

	pb "github.com/todo-app/todo-app-proto/gen/go/todo/v1"
)

// UserRole represents user access levels
type UserRole string

const (
	UserRoleUnspecified UserRole = "unspecified"
	UserRoleUser        UserRole = "user"
	UserRoleAdmin       UserRole = "admin"
)

// User represents a user in the system
type User struct {
	ID        string     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	Role      UserRole   `json:"role" db:"role"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Version   int64      `json:"version" db:"version"`
	IsDeleted bool       `json:"is_deleted" db:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ToProtobuf converts domain User to protobuf User
func (u *User) ToProtobuf() *pb.User {
	var role pb.UserRole
	switch u.Role {
	case UserRoleAdmin:
		role = pb.UserRole_USER_ROLE_ADMIN
	case UserRoleUser:
		role = pb.UserRole_USER_ROLE_USER
	default:
		role = pb.UserRole_USER_ROLE_UNSPECIFIED
	}

	return &pb.User{
		Id:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      role,
		CreatedAt: TimeToProtobuf(u.CreatedAt),
		UpdatedAt: TimeToProtobuf(u.UpdatedAt),
		Version:   u.Version,
		IsDeleted: u.IsDeleted,
	}
}

// UserFromProtobuf converts protobuf User to domain User
func UserFromProtobuf(pbUser *pb.User) *User {
	var role UserRole
	switch pbUser.Role {
	case pb.UserRole_USER_ROLE_ADMIN:
		role = UserRoleAdmin
	case pb.UserRole_USER_ROLE_USER:
		role = UserRoleUser
	default:
		role = UserRoleUnspecified
	}

	user := &User{
		ID:        pbUser.Id,
		Name:      pbUser.Name,
		Email:     pbUser.Email,
		Role:      role,
		Version:   pbUser.Version,
		IsDeleted: pbUser.IsDeleted,
	}

	if pbUser.CreatedAt != nil {
		user.CreatedAt = pbUser.CreatedAt.AsTime()
	}
	if pbUser.UpdatedAt != nil {
		user.UpdatedAt = pbUser.UpdatedAt.AsTime()
	}

	return user
}

// IsValid validates the user data
func (u *User) IsValid() error {
	if u.Name == "" {
		return ErrInvalidInput("name is required")
	}
	if u.Email == "" {
		return ErrInvalidInput("email is required")
	}
	if u.Role == "" || u.Role == UserRoleUnspecified {
		return ErrInvalidInput("valid role is required")
	}
	return nil
}
