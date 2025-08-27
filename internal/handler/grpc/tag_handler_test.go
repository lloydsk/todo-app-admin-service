package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	todov1 "github.com/lloydsk/todo-app-proto/gen/go/todo/v1"

	"github.com/todo-app/services/admin-service/internal/auth"
	"github.com/todo-app/services/admin-service/internal/model/domain"
	"github.com/todo-app/services/admin-service/internal/repository"
	"github.com/todo-app/services/admin-service/pkg/logger"
)

type mockTagService struct {
	tags map[string]*domain.Tag
}

func newMockTagService() *mockTagService {
	return &mockTagService{
		tags: make(map[string]*domain.Tag),
	}
}

func (m *mockTagService) CreateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	tag.ID = "tag-123"
	tag.Version = 1
	m.tags[tag.ID] = tag
	return tag, nil
}

func (m *mockTagService) GetTagByID(ctx context.Context, id string) (*domain.Tag, error) {
	tag, exists := m.tags[id]
	if !exists {
		return nil, domain.ErrNotFound("tag")
	}
	return tag, nil
}

func (m *mockTagService) UpdateTag(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	existing, exists := m.tags[tag.ID]
	if !exists {
		return nil, domain.ErrNotFound("tag")
	}
	if existing.Version != tag.Version {
		return nil, domain.ErrVersionConflict("tag", tag.Version, existing.Version)
	}
	tag.Version++
	m.tags[tag.ID] = tag
	return tag, nil
}

func (m *mockTagService) DeleteTag(ctx context.Context, id string, version int64) error {
	tag, exists := m.tags[id]
	if !exists {
		return domain.ErrNotFound("tag")
	}
	if tag.Version != version {
		return domain.ErrVersionConflict("tag", version, tag.Version)
	}
	delete(m.tags, id)
	return nil
}

func (m *mockTagService) RestoreTag(ctx context.Context, id string, version int64) (*domain.Tag, error) {
	return nil, nil
}

func (m *mockTagService) ListTags(ctx context.Context, opts repository.ListOptions) ([]*domain.Tag, int64, error) {
	tags := make([]*domain.Tag, 0, len(m.tags))
	for _, tag := range m.tags {
		tags = append(tags, tag)
	}
	return tags, int64(len(tags)), nil
}

func (m *mockTagService) ValidateTagUsage(ctx context.Context, tagID string) error {
	return nil
}

func (m *mockTagService) GetTagTaskCount(ctx context.Context, tagID string) (int64, error) {
	return 0, nil
}

func (m *mockTagService) FindOrCreateTag(ctx context.Context, name string) (*domain.Tag, error) {
	return nil, nil
}

func setupTagHandler(t *testing.T) (*TagHandler, *mockTagService) {
	tagService := newMockTagService()
	logger := logger.NewLogger("debug")
	handler := NewTagHandler(tagService, logger)

	return handler, tagService
}

func TestTagHandler_CreateTag(t *testing.T) {
	handler, tagService := setupTagHandler(t)
	ctx := auth.WithUserID(context.Background(), "test-user-123")

	tests := []struct {
		name       string
		request    *todov1.CreateTagRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful creation",
			request: &todov1.CreateTagRequest{
				Name:  "urgent",
				Color: stringPtr("#FF0000"),
			},
			wantErr: false,
		},
		{
			name: "successful creation with default color",
			request: &todov1.CreateTagRequest{
				Name: "low-priority",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			request: &todov1.CreateTagRequest{
				Name: "",
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.CreateTag(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status: %v", err)
					return
				}

				if st.Code() != tt.wantStatus {
					t.Errorf("expected status %v, got %v", tt.wantStatus, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if resp.Tag.Name != tt.request.Name {
				t.Errorf("expected tag name %s, got %s", tt.request.Name, resp.Tag.Name)
			}

			// Check default color was set if not provided
			if tt.request.Color == nil && resp.Tag.Color != "#6B7280" {
				t.Errorf("expected default color #6B7280, got %s", resp.Tag.Color)
			}

			// Verify tag was added to mock
			if _, exists := tagService.tags["tag-123"]; !exists {
				t.Error("tag was not added to service")
			}
		})
	}
}

func TestTagHandler_ListTags(t *testing.T) {
	handler, tagService := setupTagHandler(t)
	ctx := auth.WithUserID(context.Background(), "test-user-123")

	// Add test tags
	tag1 := &domain.Tag{
		ID:   "tag-1",
		Name: "urgent",
	}
	tag2 := &domain.Tag{
		ID:   "tag-2",
		Name: "low-priority",
	}
	tagService.tags["tag-1"] = tag1
	tagService.tags["tag-2"] = tag2

	tests := []struct {
		name          string
		request       *todov1.ListTagsRequest
		wantErr       bool
		expectedCount int
	}{
		{
			name: "list all tags",
			request: &todov1.ListTagsRequest{
				PageInfo: &todov1.PageInfo{
					PageSize: 10,
				},
			},
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name: "list with search query",
			request: &todov1.ListTagsRequest{
				PageInfo: &todov1.PageInfo{
					PageSize: 10,
				},
				SearchQuery: "urgent",
			},
			wantErr:       false,
			expectedCount: 2, // Mock returns all tags regardless of search
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.ListTags(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if len(resp.Tags) != tt.expectedCount {
				t.Errorf("expected %d tags, got %d", tt.expectedCount, len(resp.Tags))
			}
		})
	}
}

func TestTagHandler_UpdateTag(t *testing.T) {
	handler, tagService := setupTagHandler(t)
	ctx := auth.WithUserID(context.Background(), "test-user-123")

	// Add a test tag
	testTag := &domain.Tag{
		ID:      "tag-1",
		Name:    "original-name",
		Version: 1,
	}
	tagService.tags["tag-1"] = testTag

	tests := []struct {
		name       string
		request    *todov1.UpdateTagRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful update",
			request: &todov1.UpdateTagRequest{
				TagId:   "tag-1",
				Name:    "updated-name",
				Color:   stringPtr("#00FF00"),
				Version: 1,
			},
			wantErr: false,
		},
		{
			name: "tag not found",
			request: &todov1.UpdateTagRequest{
				TagId:   "nonexistent",
				Version: 1,
			},
			wantErr:    true,
			wantStatus: codes.NotFound,
		},
		{
			name: "empty tag ID",
			request: &todov1.UpdateTagRequest{
				TagId:   "",
				Version: 1,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
		{
			name: "missing version",
			request: &todov1.UpdateTagRequest{
				TagId:   "tag-1",
				Version: 0,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
		{
			name: "version mismatch",
			request: &todov1.UpdateTagRequest{
				TagId:   "tag-1",
				Version: 999,
			},
			wantErr:    true,
			wantStatus: codes.FailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.UpdateTag(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status: %v", err)
					return
				}

				if st.Code() != tt.wantStatus {
					t.Errorf("expected status %v, got %v", tt.wantStatus, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if tt.request.Name != "" && resp.Tag.Name != tt.request.Name {
				t.Errorf("expected tag name %s, got %s", tt.request.Name, resp.Tag.Name)
			}
		})
	}
}

func TestTagHandler_DeleteTag(t *testing.T) {
	handler, tagService := setupTagHandler(t)
	ctx := auth.WithUserID(context.Background(), "test-user-123")

	// Add a test tag
	testTag := &domain.Tag{
		ID:      "tag-1",
		Name:    "test-tag",
		Version: 1,
	}
	tagService.tags["tag-1"] = testTag

	tests := []struct {
		name       string
		request    *todov1.DeleteTagRequest
		wantErr    bool
		wantStatus codes.Code
	}{
		{
			name: "successful deletion",
			request: &todov1.DeleteTagRequest{
				TagId:   "tag-1",
				Version: 1,
			},
			wantErr: false,
		},
		{
			name: "empty tag ID",
			request: &todov1.DeleteTagRequest{
				TagId:   "",
				Version: 1,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
		{
			name: "missing version",
			request: &todov1.DeleteTagRequest{
				TagId:   "tag-1",
				Version: 0,
			},
			wantErr:    true,
			wantStatus: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handler.DeleteTag(ctx, tt.request)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("error is not a gRPC status: %v", err)
					return
				}

				if st.Code() != tt.wantStatus {
					t.Errorf("expected status %v, got %v", tt.wantStatus, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp == nil {
				t.Error("response is nil")
				return
			}

			if !resp.Success {
				t.Error("expected success to be true")
			}
		})
	}
}
