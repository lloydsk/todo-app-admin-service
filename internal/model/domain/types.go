package domain

import (
	"time"

	pb "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"
)

// Category represents a task category
type Category struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	Color       string     `json:"color" db:"color"`
	ParentID    *string    `json:"parent_id,omitempty" db:"parent_id"`
	IsPublic    bool       `json:"is_public" db:"is_public"`
	CreatorID   string     `json:"creator_id" db:"creator_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	Version     int64      `json:"version" db:"version"`
	IsDeleted   bool       `json:"is_deleted" db:"is_deleted"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Tag represents a task tag
type Tag struct {
	ID        string     `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Color     string     `json:"color" db:"color"`
	CreatorID string     `json:"creator_id" db:"creator_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Version   int64      `json:"version" db:"version"`
	IsDeleted bool       `json:"is_deleted" db:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ToProtobuf converts Category to protobuf
func (c *Category) ToProtobuf() *pb.Category {
	category := &pb.Category{
		Id:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		Color:       c.Color,
		IsPublic:    c.IsPublic,
		CreatorId:   c.CreatorID,
		CreatedAt:   TimeToProtobuf(c.CreatedAt),
		UpdatedAt:   TimeToProtobuf(c.UpdatedAt),
		Version:     c.Version,
		IsDeleted:   c.IsDeleted,
	}

	if c.ParentID != nil {
		category.ParentId = *c.ParentID
	}

	return category
}

// ToProtobuf converts Tag to protobuf
func (t *Tag) ToProtobuf() *pb.Tag {
	return &pb.Tag{
		Id:        t.ID,
		Name:      t.Name,
		Color:     t.Color,
		CreatorId: t.CreatorID,
		CreatedAt: TimeToProtobuf(t.CreatedAt),
		UpdatedAt: TimeToProtobuf(t.UpdatedAt),
		Version:   t.Version,
		IsDeleted: t.IsDeleted,
	}
}

// CategoryFromProtobuf converts protobuf Category to domain Category
func CategoryFromProtobuf(pb *pb.Category) *Category {
	category := &Category{
		ID:          pb.Id,
		Name:        pb.Name,
		Description: pb.Description,
		Color:       pb.Color,
		IsPublic:    pb.IsPublic,
		CreatorID:   pb.CreatorId,
		Version:     pb.Version,
		IsDeleted:   pb.IsDeleted,
	}

	if pb.ParentId != "" {
		category.ParentID = &pb.ParentId
	}

	if pb.CreatedAt != nil {
		category.CreatedAt = pb.CreatedAt.AsTime()
	}
	if pb.UpdatedAt != nil {
		category.UpdatedAt = pb.UpdatedAt.AsTime()
	}

	return category
}

// TagFromProtobuf converts protobuf Tag to domain Tag
func TagFromProtobuf(pb *pb.Tag) *Tag {
	tag := &Tag{
		ID:        pb.Id,
		Name:      pb.Name,
		Color:     pb.Color,
		CreatorID: pb.CreatorId,
		Version:   pb.Version,
		IsDeleted: pb.IsDeleted,
	}

	if pb.CreatedAt != nil {
		tag.CreatedAt = pb.CreatedAt.AsTime()
	}
	if pb.UpdatedAt != nil {
		tag.UpdatedAt = pb.UpdatedAt.AsTime()
	}

	return tag
}
