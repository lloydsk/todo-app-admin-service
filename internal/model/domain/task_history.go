package domain

import (
	"encoding/json"
	"time"

	pb "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TaskHistoryAction represents different types of task actions
type TaskHistoryAction string

const (
	TaskHistoryActionCreated   TaskHistoryAction = "CREATED"
	TaskHistoryActionUpdated   TaskHistoryAction = "UPDATED"
	TaskHistoryActionCompleted TaskHistoryAction = "COMPLETED"
	TaskHistoryActionDeleted   TaskHistoryAction = "DELETED"
	TaskHistoryActionRestored  TaskHistoryAction = "RESTORED"
	TaskHistoryActionAssigned  TaskHistoryAction = "ASSIGNED"
)

// TaskHistory represents an audit trail entry for task changes
type TaskHistory struct {
	ID        string            `json:"id" db:"id"`
	TaskID    string            `json:"task_id" db:"task_id"`
	Action    TaskHistoryAction `json:"action" db:"action"`
	ActorID   string            `json:"actor_id" db:"actor_id"`
	Timestamp time.Time         `json:"timestamp" db:"timestamp"`
	Details   json.RawMessage   `json:"details,omitempty" db:"details"`

	// Related entities
	Actor *User `json:"actor,omitempty"`
	Task  *Task `json:"task,omitempty"`
}

// TaskHistoryDetails represents the structure of the details field
type TaskHistoryDetails struct {
	OldValues map[string]interface{} `json:"old_values,omitempty"`
	NewValues map[string]interface{} `json:"new_values,omitempty"`
	Changes   []string               `json:"changes,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// GetDetails unmarshals the details field into TaskHistoryDetails
func (th *TaskHistory) GetDetails() (*TaskHistoryDetails, error) {
	if th.Details == nil {
		return nil, nil
	}

	var details TaskHistoryDetails
	if err := json.Unmarshal(th.Details, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

// SetDetails marshals TaskHistoryDetails into the details field
func (th *TaskHistory) SetDetails(details *TaskHistoryDetails) error {
	if details == nil {
		th.Details = nil
		return nil
	}

	data, err := json.Marshal(details)
	if err != nil {
		return err
	}

	th.Details = data
	return nil
}

// ToProtobuf converts TaskHistory to protobuf TaskHistoryEntry
func (th *TaskHistory) ToProtobuf() *pb.TaskHistoryEntry {
	return &pb.TaskHistoryEntry{
		Id:        th.ID,
		TaskId:    th.TaskID,
		Action:    string(th.Action),
		ActorId:   th.ActorID,
		Timestamp: timestamppb.New(th.Timestamp),
		Details:   string(th.Details),
	}
}

// Validate validates task history data
func (th *TaskHistory) Validate() error {
	if th.TaskID == "" {
		return ErrInvalidInput("task history must have a task ID")
	}
	if th.ActorID == "" {
		return ErrInvalidInput("task history must have an actor ID")
	}
	if th.Action == "" {
		return ErrInvalidInput("task history must have an action")
	}
	return nil
}
