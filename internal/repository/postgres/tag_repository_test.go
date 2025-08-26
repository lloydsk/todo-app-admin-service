package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/todo-app/services/admin-service/internal/config"
	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/db"
)

func setupTagTestDB(t *testing.T) (*db.Connection, string) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Name:     "todo_app",
			User:     "postgres",
			Password: "postgres",
			SSLMode:  "disable",
		},
	}

	dbConn, err := db.NewConnection(cfg.Database)
	if err != nil {
		t.Skipf("Database not available for integration tests: %v", err)
	}

	ctx := context.Background()
	if err := dbConn.HealthCheck(ctx); err != nil {
		t.Skipf("Database health check failed: %v", err)
	}

	dbConn.SetServiceContext(ctx, "tag-integration-test")

	// Create a test user for the foreign key constraint
	userRepo := NewUserRepository(dbConn.DB)
	testUser := &domain.User{
		Name:  "Tag Test User",
		Email: fmt.Sprintf("tag-test-%d@example.com", time.Now().Unix()),
		Role:  domain.UserRoleUser,
	}

	err = userRepo.Create(ctx, testUser)
	if err != nil {
		t.Skipf("Failed to create test user for tags: %v", err)
	}

	return dbConn, testUser.ID
}

func TestTagRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	dbConn, testCreatorID := setupTagTestDB(t)
	defer dbConn.Close()

	ctx := context.Background()
	tagRepo := NewTagRepository(dbConn.DB)

	t.Run("CreateTag", func(t *testing.T) {
		testTag := &domain.Tag{
			Name:      fmt.Sprintf("integration-test-tag-%d", time.Now().Unix()),
			Color:     "#FF5733",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, testTag)
		if err != nil {
			t.Fatalf("Failed to create tag: %v", err)
		}

		if testTag.ID == "" {
			t.Error("Tag ID should be set after creation")
		}
		if testTag.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set after creation")
		}
		if testTag.Version == 0 {
			t.Error("Version should be set after creation")
		}
	})

	t.Run("GetTagByID", func(t *testing.T) {
		// Create a test tag first
		testTag := &domain.Tag{
			Name:      fmt.Sprintf("get-test-tag-%d", time.Now().Unix()),
			Color:     "#33FF57",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, testTag)
		if err != nil {
			t.Fatalf("Failed to create test tag: %v", err)
		}

		retrievedTag, err := tagRepo.GetByID(ctx, testTag.ID)
		if err != nil {
			t.Fatalf("Failed to get tag by ID: %v", err)
		}

		if retrievedTag.ID != testTag.ID {
			t.Errorf("ID mismatch: got %s, want %s", retrievedTag.ID, testTag.ID)
		}
		if retrievedTag.Name != testTag.Name {
			t.Errorf("Name mismatch: got %s, want %s", retrievedTag.Name, testTag.Name)
		}
		if retrievedTag.Color != testTag.Color {
			t.Errorf("Color mismatch: got %s, want %s", retrievedTag.Color, testTag.Color)
		}
	})

	t.Run("UpdateTag", func(t *testing.T) {
		testTag := &domain.Tag{
			Name:      fmt.Sprintf("update-test-tag-%d", time.Now().Unix()),
			Color:     "#5733FF",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, testTag)
		if err != nil {
			t.Fatalf("Failed to create test tag: %v", err)
		}

		originalVersion := testTag.Version
		testTag.Name = "updated-tag-name"
		testTag.Color = "#FF3357"

		err = tagRepo.Update(ctx, testTag)
		if err != nil {
			t.Fatalf("Failed to update tag: %v", err)
		}

		if testTag.Version <= originalVersion {
			t.Error("Version should be incremented after update")
		}

		retrievedTag, err := tagRepo.GetByID(ctx, testTag.ID)
		if err != nil {
			t.Fatalf("Failed to get updated tag: %v", err)
		}

		if retrievedTag.Name != "updated-tag-name" {
			t.Errorf("Name not updated: got %s, want updated-tag-name", retrievedTag.Name)
		}
		if retrievedTag.Color != "#FF3357" {
			t.Errorf("Color not updated: got %s, want #FF3357", retrievedTag.Color)
		}
	})

	t.Run("ListTags", func(t *testing.T) {
		tags, total, err := tagRepo.List(ctx, repository.TagListOptions{
			ListOptions: repository.ListOptions{
				Page:     0, // First page is 0, not 1
				PageSize: 10,
			},
		})
		if err != nil {
			t.Fatalf("Failed to list tags: %v", err)
		}

		if total < int64(len(tags)) {
			t.Errorf("Total count (%d) should be >= returned tags (%d)", total, len(tags))
		}

		// Verify all returned tags have required fields
		for _, tag := range tags {
			if tag.ID == "" {
				t.Error("Tag ID should not be empty")
			}
			if tag.Name == "" {
				t.Error("Tag Name should not be empty")
			}
			if tag.CreatorID == "" {
				t.Error("Tag CreatorID should not be empty")
			}
		}
	})

	t.Run("ListTagsWithSearch", func(t *testing.T) {
		// Create a tag with a very specific name for searching
		uniqueSuffix := fmt.Sprintf("unique-search-%d", time.Now().UnixNano())
		searchTag := &domain.Tag{
			Name:      fmt.Sprintf("searchable-tag-%s", uniqueSuffix),
			Color:     "#FF0011",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, searchTag)
		if err != nil {
			t.Fatalf("Failed to create searchable tag: %v", err)
		}

		// Search for the tag using the unique suffix to be more specific
		tags, total, err := tagRepo.List(ctx, repository.TagListOptions{
			ListOptions: repository.ListOptions{
				Page:        0, // First page is 0, not 1
				PageSize:    10,
				SearchQuery: uniqueSuffix, // Search for the unique part
			},
		})
		if err != nil {
			t.Fatalf("Failed to search tags: %v", err)
		}

		if total == 0 {
			t.Errorf("Should find at least one tag with search query '%s'", uniqueSuffix)
		}

		// Verify the found tag matches our search
		found := false
		for _, tag := range tags {
			if tag.ID == searchTag.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Search should return the created tag. Found %d tags, looking for ID %s", len(tags), searchTag.ID)
			for i, tag := range tags {
				t.Logf("Found tag %d: ID=%s, Name=%s", i, tag.ID, tag.Name)
			}
		}
	})

	t.Run("SoftDeleteTag", func(t *testing.T) {
		testTag := &domain.Tag{
			Name:      fmt.Sprintf("delete-test-tag-%d", time.Now().Unix()),
			Color:     "#FF0022",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, testTag)
		if err != nil {
			t.Fatalf("Failed to create test tag: %v", err)
		}

		err = tagRepo.SoftDelete(ctx, testTag.ID, testTag.Version)
		if err != nil {
			t.Fatalf("Failed to soft delete tag: %v", err)
		}

		// Verify tag is soft deleted (GetByID should return not found error)
		_, err = tagRepo.GetByID(ctx, testTag.ID)
		if err == nil {
			t.Error("Expected error when getting soft deleted tag")
		}
		if !domain.IsNotFoundError(err) {
			t.Errorf("Expected not found error, got: %v", err)
		}
	})

	t.Run("RestoreTag", func(t *testing.T) {
		testTag := &domain.Tag{
			Name:      fmt.Sprintf("restore-test-tag-%d", time.Now().Unix()),
			Color:     "#FF0033",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, testTag)
		if err != nil {
			t.Fatalf("Failed to create test tag: %v", err)
		}

		// Soft delete the tag first
		err = tagRepo.SoftDelete(ctx, testTag.ID, testTag.Version)
		if err != nil {
			t.Fatalf("Failed to soft delete tag: %v", err)
		}

		// Now restore it
		err = tagRepo.Restore(ctx, testTag.ID, testTag.Version+1)
		if err != nil {
			t.Fatalf("Failed to restore tag: %v", err)
		}

		// Verify tag is restored and accessible
		restoredTag, err := tagRepo.GetByID(ctx, testTag.ID)
		if err != nil {
			t.Fatalf("Failed to get restored tag: %v", err)
		}

		if restoredTag.IsDeleted {
			t.Error("Restored tag should not be marked as deleted")
		}
		if restoredTag.Name != testTag.Name {
			t.Errorf("Restored tag name mismatch: got %s, want %s", restoredTag.Name, testTag.Name)
		}
	})

	t.Run("VersionConflict", func(t *testing.T) {
		testTag := &domain.Tag{
			Name:      fmt.Sprintf("version-test-tag-%d", time.Now().Unix()),
			Color:     "#FF0044",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, testTag)
		if err != nil {
			t.Fatalf("Failed to create test tag: %v", err)
		}

		// Try to update with wrong version
		testTag.Name = "updated-name"
		testTag.Version = 999 // Wrong version

		err = tagRepo.Update(ctx, testTag)
		if err == nil {
			t.Error("Expected version conflict error")
		}
		if !domain.IsVersionConflictError(err) {
			t.Errorf("Expected version conflict error, got: %v", err)
		}
	})

	t.Run("DuplicateNameHandling", func(t *testing.T) {
		uniqueName := fmt.Sprintf("duplicate-test-%d", time.Now().Unix())

		// Create first tag
		firstTag := &domain.Tag{
			Name:      uniqueName,
			Color:     "#FF0055",
			CreatorID: testCreatorID,
		}

		err := tagRepo.Create(ctx, firstTag)
		if err != nil {
			t.Fatalf("Failed to create first tag: %v", err)
		}

		// Try to create second tag with same name
		secondTag := &domain.Tag{
			Name:      uniqueName,
			Color:     "#FF0066",
			CreatorID: testCreatorID,
		}

		err = tagRepo.Create(ctx, secondTag)
		if err == nil {
			t.Error("Expected error when creating tag with duplicate name")
		}
		// Note: The actual error type depends on database constraints
		// In a real implementation, this might be a unique constraint violation
	})
}
