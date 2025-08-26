package domain

import (
	"time"

	pb "github.com/todo-app/todo-app-proto/gen/go/todo/v1"
)

// TaskStatus represents the current state of a task
type TaskStatus string

const (
	TaskStatusUnspecified TaskStatus = "unspecified"
	TaskStatusOpen        TaskStatus = "OPEN"
	TaskStatusInProgress  TaskStatus = "IN_PROGRESS"
	TaskStatusCompleted   TaskStatus = "COMPLETED"
	TaskStatusCancelled   TaskStatus = "CANCELLED"
)

// TaskPriority represents task importance
type TaskPriority string

const (
	TaskPriorityUnspecified TaskPriority = "unspecified"
	TaskPriorityLow         TaskPriority = "LOW"
	TaskPriorityMedium      TaskPriority = "MEDIUM"
	TaskPriorityHigh        TaskPriority = "HIGH"
	TaskPriorityUrgent      TaskPriority = "URGENT"
)

// Task represents a todo task
type Task struct {
	ID          string       `json:"id" db:"id"`
	Title       string       `json:"title" db:"title"`
	Description string       `json:"description" db:"description"`
	AssigneeID  string       `json:"assignee_id" db:"assignee_id"`
	Status      TaskStatus   `json:"status" db:"status"`
	Priority    TaskPriority `json:"priority" db:"priority"`
	DueDate     *time.Time   `json:"due_date,omitempty" db:"due_date"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	Version     int64        `json:"version" db:"version"`
	IsDeleted   bool         `json:"is_deleted" db:"is_deleted"`
	DeletedAt   *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`

	// Related data (loaded separately)
	Categories []Category         `json:"categories,omitempty"`
	Tags       []Tag              `json:"tags,omitempty"`
	History    []TaskHistoryEntry `json:"history,omitempty"`
}

// Category and Tag are defined in separate files to avoid circular dependencies

// TaskHistoryEntry represents a single event in task history
type TaskHistoryEntry struct {
	ID          string    `json:"id" db:"id"`
	TaskID      string    `json:"task_id" db:"task_id"`
	Action      string    `json:"action" db:"action"`
	ActorID     string    `json:"actor_id" db:"actor_id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Details     string    `json:"details" db:"details"` // JSON string
}

// ToProtobuf converts domain Task to protobuf Task
func (t *Task) ToProtobuf() *pb.Task {
	var status pb.TaskStatus
	switch t.Status {
	case TaskStatusOpen:
		status = pb.TaskStatus_TASK_STATUS_OPEN
	case TaskStatusInProgress:
		status = pb.TaskStatus_TASK_STATUS_IN_PROGRESS
	case TaskStatusCompleted:
		status = pb.TaskStatus_TASK_STATUS_COMPLETED
	case TaskStatusCancelled:
		status = pb.TaskStatus_TASK_STATUS_COMPLETED // Use completed for now until proto is updated
	default:
		status = pb.TaskStatus_TASK_STATUS_UNSPECIFIED
	}

	var priority pb.TaskPriority
	switch t.Priority {
	case TaskPriorityLow:
		priority = pb.TaskPriority_TASK_PRIORITY_LOW
	case TaskPriorityMedium:
		priority = pb.TaskPriority_TASK_PRIORITY_MEDIUM
	case TaskPriorityHigh:
		priority = pb.TaskPriority_TASK_PRIORITY_HIGH
	case TaskPriorityUrgent:
		priority = pb.TaskPriority_TASK_PRIORITY_URGENT
	default:
		priority = pb.TaskPriority_TASK_PRIORITY_UNSPECIFIED
	}

	task := &pb.Task{
		Id:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		AssigneeId:  t.AssigneeID,
		Status:      status,
		Priority:    priority,
		CreatedAt:   TimeToProtobuf(t.CreatedAt),
		UpdatedAt:   TimeToProtobuf(t.UpdatedAt),
		Version:     t.Version,
		IsDeleted:   t.IsDeleted,
	}

	if t.DueDate != nil {
		task.DueDate = TimeToProtobuf(*t.DueDate)
	}

	// Convert categories
	for _, cat := range t.Categories {
		task.CategoryIds = append(task.CategoryIds, cat.ID)
	}

	// Convert tags
	for _, tag := range t.Tags {
		task.TagIds = append(task.TagIds, tag.ID)
	}

	// Convert history
	for _, entry := range t.History {
		task.History = append(task.History, entry.ToProtobuf())
	}

	return task
}

// ToProtobuf converts TaskHistoryEntry to protobuf
func (h *TaskHistoryEntry) ToProtobuf() *pb.TaskHistoryEntry {
	return &pb.TaskHistoryEntry{
		Id:        h.ID,
		TaskId:    h.TaskID,
		Action:    h.Action,
		ActorId:   h.ActorID,
		Timestamp: TimeToProtobuf(h.Timestamp),
		Details:   h.Details,
	}
}

// Category and Tag ToProtobuf methods are defined in types.go

// IsValid validates the task data
func (t *Task) IsValid() error {
	if t.Title == "" {
		return ErrInvalidInput("title is required")
	}
	if t.AssigneeID == "" {
		return ErrInvalidInput("assignee is required")
	}
	if t.Status == "" || t.Status == TaskStatusUnspecified {
		return ErrInvalidInput("valid status is required")
	}
	return nil
}
