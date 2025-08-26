package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type tagService struct {
	tagRepo  repository.TagRepository
	taskRepo repository.TaskRepository
	logger   logger.Logger
}

// NewTagService creates a new tag service
func NewTagService(
	tagRepo repository.TagRepository,
	taskRepo repository.TaskRepository,
	log logger.Logger,
) TagService {
	return &tagService{
		tagRepo:  tagRepo,
		taskRepo: taskRepo,
		logger:   log,
	}
}

func (s *tagService) CreateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	s.logger.Info(ctx, "Creating new tag", "name", tag.Name)

	// Business validation
	if err := s.validateTagForCreation(ctx, tag); err != nil {
		return nil, err
	}

	// Normalize tag name
	tag.Name = s.normalizeTagName(tag.Name)

	// Check for duplicate name
	existing, err := s.findTagByName(ctx, tag.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing tag: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrConflict("tag with this name already exists")
	}

	// Create tag
	if err := s.tagRepo.Create(ctx, tag); err != nil {
		s.logger.Error(ctx, "Failed to create tag", "error", err, "name", tag.Name)
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	s.logger.Info(ctx, "Tag created successfully", "tag_id", tag.ID, "name", tag.Name)
	return tag, nil
}

func (s *tagService) GetTagByID(ctx context.Context, id string) (*domain.Tag, error) {
	s.logger.Debug(ctx, "Getting tag by ID", "tag_id", id)

	if id == "" {
		return nil, domain.ErrInvalidInput("tag ID is required")
	}

	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		if domain.IsNotFoundError(err) {
			s.logger.Debug(ctx, "Tag not found", "tag_id", id)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to get tag by ID", "error", err, "tag_id", id)
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

func (s *tagService) UpdateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	s.logger.Info(ctx, "Updating tag", "tag_id", tag.ID, "version", tag.Version)

	// Business validation
	if err := s.validateTagForUpdate(ctx, tag); err != nil {
		return nil, err
	}

	// Normalize tag name
	tag.Name = s.normalizeTagName(tag.Name)

	// Check for duplicate name if name is being changed
	existingTag, err := s.tagRepo.GetByID(ctx, tag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing tag: %w", err)
	}

	if existingTag.Name != tag.Name {
		conflictTag, err := s.findTagByName(ctx, tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check for name conflict: %w", err)
		}
		if conflictTag != nil && conflictTag.ID != tag.ID {
			return nil, domain.ErrConflict("tag name already in use")
		}
	}

	// Update tag
	if err := s.tagRepo.Update(ctx, tag); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Tag update version conflict", "tag_id", tag.ID, "version", tag.Version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to update tag", "error", err, "tag_id", tag.ID)
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	s.logger.Info(ctx, "Tag updated successfully", "tag_id", tag.ID, "new_version", tag.Version)
	return tag, nil
}

func (s *tagService) DeleteTag(ctx context.Context, id string, version int64) error {
	s.logger.Info(ctx, "Soft deleting tag", "tag_id", id, "version", version)

	if id == "" {
		return domain.ErrInvalidInput("tag ID is required")
	}

	// Business rule: Check if tag is in use
	if err := s.ValidateTagUsage(ctx, id); err != nil {
		return err
	}

	if err := s.tagRepo.SoftDelete(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Tag deletion version conflict", "tag_id", id, "version", version)
			return err
		}
		s.logger.Error(ctx, "Failed to delete tag", "error", err, "tag_id", id)
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	s.logger.Info(ctx, "Tag deleted successfully", "tag_id", id)
	return nil
}

func (s *tagService) RestoreTag(ctx context.Context, id string, version int64) (*domain.Tag, error) {
	s.logger.Info(ctx, "Restoring tag", "tag_id", id, "version", version)

	if id == "" {
		return nil, domain.ErrInvalidInput("tag ID is required")
	}

	if err := s.tagRepo.Restore(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Tag restoration version conflict", "tag_id", id, "version", version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to restore tag", "error", err, "tag_id", id)
		return nil, fmt.Errorf("failed to restore tag: %w", err)
	}

	// Get the restored tag
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored tag: %w", err)
	}

	s.logger.Info(ctx, "Tag restored successfully", "tag_id", id)
	return tag, nil
}

func (s *tagService) ListTags(ctx context.Context, opts repository.ListOptions) ([]*domain.Tag, int64, error) {
	s.logger.Debug(ctx, "Listing tags", "page", opts.Page, "page_size", opts.PageSize)

	// Convert to TagListOptions - this will be implemented when we add specific filtering
	tagOpts := repository.TagListOptions{
		ListOptions: opts,
	}

	tags, total, err := s.tagRepo.List(ctx, tagOpts)
	if err != nil {
		s.logger.Error(ctx, "Failed to list tags", "error", err)
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}

	s.logger.Debug(ctx, "Listed tags successfully", "count", len(tags), "total", total)
	return tags, total, nil
}

func (s *tagService) ValidateTagUsage(ctx context.Context, tagID string) error {
	taskCount, err := s.GetTagTaskCount(ctx, tagID)
	if err != nil {
		return fmt.Errorf("failed to check tag usage: %w", err)
	}

	if taskCount > 0 {
		return domain.ErrBusinessRule(fmt.Sprintf("cannot delete tag: it is used by %d tasks", taskCount))
	}

	return nil
}

func (s *tagService) GetTagTaskCount(ctx context.Context, tagID string) (int64, error) {
	// Use task repository to count tasks with this tag
	tasks, total, err := s.taskRepo.List(ctx, repository.TaskListOptions{
		ListOptions: repository.ListOptions{PageSize: 1}, // We only need the count
		TagIDs:      []string{tagID},
	})

	// We don't need the actual tasks, just the total count
	_ = tasks

	if err != nil {
		return 0, fmt.Errorf("failed to count tasks for tag: %w", err)
	}

	return total, nil
}

func (s *tagService) FindOrCreateTag(ctx context.Context, name string) (*domain.Tag, error) {
	s.logger.Debug(ctx, "Finding or creating tag", "name", name)

	// Normalize name
	normalizedName := s.normalizeTagName(name)

	// Try to find existing tag
	existing, err := s.findTagByName(ctx, normalizedName)
	if err != nil {
		return nil, fmt.Errorf("failed to search for existing tag: %w", err)
	}

	if existing != nil {
		s.logger.Debug(ctx, "Found existing tag", "tag_id", existing.ID, "name", normalizedName)
		return existing, nil
	}

	// Create new tag with system creator
	newTag := &domain.Tag{
		Name:      normalizedName,
		Color:     s.generateDefaultTagColor(),
		CreatorID: "system", // Auto-created tags have system as creator
	}

	return s.CreateTag(ctx, newTag)
}

// Helper methods for business validation

func (s *tagService) validateTagForCreation(ctx context.Context, tag *domain.Tag) error {
	if err := tag.IsValid(); err != nil {
		return err
	}

	// Additional business rules
	if tag.Name == "" {
		return domain.ErrInvalidInput("tag name is required")
	}

	// Validate tag name format
	if err := s.validateTagName(tag.Name); err != nil {
		return err
	}

	return nil
}

func (s *tagService) validateTagForUpdate(ctx context.Context, tag *domain.Tag) error {
	if tag.ID == "" {
		return domain.ErrInvalidInput("tag ID is required for update")
	}

	if err := tag.IsValid(); err != nil {
		return err
	}

	// Validate tag name format
	if err := s.validateTagName(tag.Name); err != nil {
		return err
	}

	return nil
}

func (s *tagService) validateTagName(name string) error {
	if len(name) > 50 {
		return domain.ErrInvalidInput("tag name cannot exceed 50 characters")
	}

	// Tag names should not contain special characters except hyphens and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' || char == ' ') {
			return domain.ErrInvalidInput("tag name contains invalid characters")
		}
	}

	return nil
}

func (s *tagService) normalizeTagName(name string) string {
	// Convert to lowercase and trim spaces
	normalized := strings.ToLower(strings.TrimSpace(name))

	// Replace multiple spaces with single space
	words := strings.Fields(normalized)
	return strings.Join(words, "-")
}

func (s *tagService) findTagByName(ctx context.Context, name string) (*domain.Tag, error) {
	// This is a simplified implementation - in a real system you might have a GetByName method
	tags, _, err := s.tagRepo.List(ctx, repository.TagListOptions{
		ListOptions: repository.ListOptions{
			PageSize:    100,
			SearchQuery: name,
		},
	})
	if err != nil {
		return nil, err
	}

	for _, tag := range tags {
		if tag.Name == name {
			return tag, nil
		}
	}

	return nil, nil
}

func (s *tagService) generateDefaultTagColor() string {
	// Default colors for tags
	colors := []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7",
		"#DDA0DD", "#98D8E8", "#F7DC6F", "#BB8FCE", "#85C1E9",
	}

	// In a real implementation, you might rotate through colors or use a hash
	return colors[0] // For now, just return the first color
}
