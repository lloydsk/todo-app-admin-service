package service

import (
	"context"
	"fmt"

	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type categoryService struct {
	categoryRepo repository.CategoryRepository
	taskRepo     repository.TaskRepository
	logger       logger.Logger
}

// NewCategoryService creates a new category service
func NewCategoryService(
	categoryRepo repository.CategoryRepository,
	taskRepo repository.TaskRepository,
	log logger.Logger,
) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		taskRepo:     taskRepo,
		logger:       log,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	s.logger.Info(ctx, "Creating new category", "name", category.Name)
	
	// Business validation
	if err := s.validateCategoryForCreation(ctx, category); err != nil {
		return nil, err
	}
	
	// Check for duplicate name
	existing, err := s.findCategoryByName(ctx, category.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check for existing category: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrConflict("category with this name already exists")
	}
	
	// Create category
	if err := s.categoryRepo.Create(ctx, category); err != nil {
		s.logger.Error(ctx, "Failed to create category", "error", err, "name", category.Name)
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	
	s.logger.Info(ctx, "Category created successfully", "category_id", category.ID, "name", category.Name)
	return category, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id string) (*domain.Category, error) {
	s.logger.Debug(ctx, "Getting category by ID", "category_id", id)
	
	if id == "" {
		return nil, domain.ErrInvalidInput("category ID is required")
	}
	
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if domain.IsNotFoundError(err) {
			s.logger.Debug(ctx, "Category not found", "category_id", id)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to get category by ID", "error", err, "category_id", id)
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	
	return category, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, category *domain.Category) (*domain.Category, error) {
	s.logger.Info(ctx, "Updating category", "category_id", category.ID, "version", category.Version)
	
	// Business validation
	if err := s.validateCategoryForUpdate(ctx, category); err != nil {
		return nil, err
	}
	
	// Check for duplicate name if name is being changed
	existingCategory, err := s.categoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing category: %w", err)
	}
	
	if existingCategory.Name != category.Name {
		conflictCategory, err := s.findCategoryByName(ctx, category.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check for name conflict: %w", err)
		}
		if conflictCategory != nil && conflictCategory.ID != category.ID {
			return nil, domain.ErrConflict("category name already in use")
		}
	}
	
	// Update category
	if err := s.categoryRepo.Update(ctx, category); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Category update version conflict", "category_id", category.ID, "version", category.Version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to update category", "error", err, "category_id", category.ID)
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	
	s.logger.Info(ctx, "Category updated successfully", "category_id", category.ID, "new_version", category.Version)
	return category, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string, version int64) error {
	s.logger.Info(ctx, "Soft deleting category", "category_id", id, "version", version)
	
	if id == "" {
		return domain.ErrInvalidInput("category ID is required")
	}
	
	// Business rule: Check if category is in use
	if err := s.ValidateCategoryUsage(ctx, id); err != nil {
		return err
	}
	
	if err := s.categoryRepo.SoftDelete(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Category deletion version conflict", "category_id", id, "version", version)
			return err
		}
		s.logger.Error(ctx, "Failed to delete category", "error", err, "category_id", id)
		return fmt.Errorf("failed to delete category: %w", err)
	}
	
	s.logger.Info(ctx, "Category deleted successfully", "category_id", id)
	return nil
}

func (s *categoryService) RestoreCategory(ctx context.Context, id string, version int64) (*domain.Category, error) {
	s.logger.Info(ctx, "Restoring category", "category_id", id, "version", version)
	
	if id == "" {
		return nil, domain.ErrInvalidInput("category ID is required")
	}
	
	if err := s.categoryRepo.Restore(ctx, id, version); err != nil {
		if domain.IsVersionConflictError(err) {
			s.logger.Warn(ctx, "Category restoration version conflict", "category_id", id, "version", version)
			return nil, err
		}
		s.logger.Error(ctx, "Failed to restore category", "error", err, "category_id", id)
		return nil, fmt.Errorf("failed to restore category: %w", err)
	}
	
	// Get the restored category
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get restored category: %w", err)
	}
	
	s.logger.Info(ctx, "Category restored successfully", "category_id", id)
	return category, nil
}

func (s *categoryService) ListCategories(ctx context.Context, opts repository.ListOptions) ([]*domain.Category, int64, error) {
	s.logger.Debug(ctx, "Listing categories", "page", opts.Page, "page_size", opts.PageSize)
	
	// Convert to CategoryListOptions - this will be implemented when we add specific filtering
	categoryOpts := repository.CategoryListOptions{
		ListOptions: opts,
	}
	
	categories, total, err := s.categoryRepo.List(ctx, categoryOpts)
	if err != nil {
		s.logger.Error(ctx, "Failed to list categories", "error", err)
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}
	
	s.logger.Debug(ctx, "Listed categories successfully", "count", len(categories), "total", total)
	return categories, total, nil
}

func (s *categoryService) ValidateCategoryUsage(ctx context.Context, categoryID string) error {
	taskCount, err := s.GetCategoryTaskCount(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}
	
	if taskCount > 0 {
		return domain.ErrBusinessRule(fmt.Sprintf("cannot delete category: it is used by %d tasks", taskCount))
	}
	
	return nil
}

func (s *categoryService) GetCategoryTaskCount(ctx context.Context, categoryID string) (int64, error) {
	// Use task repository to count tasks with this category
	tasks, total, err := s.taskRepo.List(ctx, repository.TaskListOptions{
		ListOptions:  repository.ListOptions{PageSize: 1}, // We only need the count
		CategoryIDs:  []string{categoryID},
	})
	
	// We don't need the actual tasks, just the total count
	_ = tasks
	
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks for category: %w", err)
	}
	
	return total, nil
}

// Helper methods for business validation

func (s *categoryService) validateCategoryForCreation(ctx context.Context, category *domain.Category) error {
	if err := category.IsValid(); err != nil {
		return err
	}
	
	// Additional business rules
	if category.Name == "" {
		return domain.ErrInvalidInput("category name is required")
	}
	
	return nil
}

func (s *categoryService) validateCategoryForUpdate(ctx context.Context, category *domain.Category) error {
	if category.ID == "" {
		return domain.ErrInvalidInput("category ID is required for update")
	}
	
	if err := category.IsValid(); err != nil {
		return err
	}
	
	return nil
}

func (s *categoryService) findCategoryByName(ctx context.Context, name string) (*domain.Category, error) {
	// This is a simplified implementation - in a real system you might have a GetByName method
	categories, _, err := s.categoryRepo.List(ctx, repository.CategoryListOptions{
		ListOptions: repository.ListOptions{
			PageSize:    100,
			SearchQuery: name,
		},
	})
	if err != nil {
		return nil, err
	}
	
	for _, category := range categories {
		if category.Name == name {
			return category, nil
		}
	}
	
	return nil, nil
}