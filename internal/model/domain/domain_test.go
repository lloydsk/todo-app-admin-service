package domain

import (
	"testing"
)

func TestUser_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name: "valid admin user",
			user: &User{
				ID:    "user-123",
				Name:  "Test Admin",
				Email: "admin@example.com",
				Role:  UserRoleAdmin,
			},
			wantErr: false,
		},
		{
			name: "valid regular user",
			user: &User{
				ID:    "user-456",
				Name:  "Test User",
				Email: "user@example.com",
				Role:  UserRoleUser,
			},
			wantErr: false,
		},
		{
			name: "invalid user - missing name",
			user: &User{
				ID:    "user-789",
				Name:  "",
				Email: "test@example.com",
				Role:  UserRoleUser,
			},
			wantErr: true,
		},
		{
			name: "invalid user - missing email",
			user: &User{
				ID:    "user-101",
				Name:  "Test User",
				Email: "",
				Role:  UserRoleUser,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_ToProtobuf(t *testing.T) {
	user := &User{
		ID:    "user-123",
		Name:  "Test User",
		Email: "test@example.com",
		Role:  UserRoleAdmin,
	}

	pb := user.ToProtobuf()
	if pb.Id != user.ID {
		t.Errorf("ToProtobuf() ID = %v, want %v", pb.Id, user.ID)
	}
	if pb.Name != user.Name {
		t.Errorf("ToProtobuf() Name = %v, want %v", pb.Name, user.Name)
	}
	if pb.Email != user.Email {
		t.Errorf("ToProtobuf() Email = %v, want %v", pb.Email, user.Email)
	}
}

func TestTask_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: &Task{
				ID:          "task-123",
				Title:       "Test Task",
				Description: "This is a test task",
				AssigneeID:  "user-123",
				Status:      TaskStatusOpen,
				Priority:    TaskPriorityMedium,
			},
			wantErr: false,
		},
		{
			name: "invalid task - missing title",
			task: &Task{
				ID:          "task-456",
				Title:       "",
				Description: "This is a test task",
				AssigneeID:  "user-123",
				Status:      TaskStatusOpen,
				Priority:    TaskPriorityMedium,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTask_ToProtobuf(t *testing.T) {
	task := &Task{
		ID:          "task-123",
		Title:       "Test Task",
		Description: "This is a test task",
		AssigneeID:  "user-123",
		Status:      TaskStatusOpen,
		Priority:    TaskPriorityMedium,
	}

	pb := task.ToProtobuf()
	if pb.Id != task.ID {
		t.Errorf("ToProtobuf() ID = %v, want %v", pb.Id, task.ID)
	}
	if pb.Title != task.Title {
		t.Errorf("ToProtobuf() Title = %v, want %v", pb.Title, task.Title)
	}
	if pb.Description != task.Description {
		t.Errorf("ToProtobuf() Description = %v, want %v", pb.Description, task.Description)
	}
	if pb.AssigneeId != task.AssigneeID {
		t.Errorf("ToProtobuf() AssigneeID = %v, want %v", pb.AssigneeId, task.AssigneeID)
	}
}

func TestDomainErrors(t *testing.T) {
	tests := []struct {
		name     string
		errFunc  func(string) error
		entity   string
		expected string
	}{
		{
			name:     "not found error",
			errFunc:  ErrNotFound,
			entity:   "user",
			expected: "NOT_FOUND: user not found",
		},
		{
			name:     "conflict error",
			errFunc:  ErrConflict,
			entity:   "email already exists",
			expected: "CONFLICT: email already exists",
		},
		{
			name:     "invalid input error",
			errFunc:  ErrInvalidInput,
			entity:   "name is required",
			expected: "INVALID_INPUT: name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc(tt.entity)
			if err.Error() != tt.expected {
				t.Errorf("Error message = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}
