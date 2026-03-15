package group

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
)

var testService Service

func init() {
	// Initialize test database
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to init test DB: %v", err)
	}

	// Initialize cache
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to init cache: %v", err)
	}

	logger := log.New(os.Stderr, "[TestGroupService] ", log.LstdFlags)
	testService = NewService(config.GetDB(), logger)
}

// TestCreateGroupValid tests valid group creation
func TestCreateGroupValid(t *testing.T) {
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}

	group, err := testService.CreateGroup(context.Background(), 1, req)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, req.ProductID, group.ProductID)
	assert.Equal(t, req.TargetCount, group.TargetCount)
	assert.Equal(t, 1, group.CurrentCount) // Creator is first member
	assert.Equal(t, "active", group.Status)
	assert.Equal(t, 1, group.CreatorID)
	assert.True(t, group.ID > 0)
}

// TestCreateGroupInvalidTargetCount tests creation with invalid target count
func TestCreateGroupInvalidTargetCount(t *testing.T) {
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 0,
		Deadline:    time.Now().Add(24 * time.Hour),
	}

	group, err := testService.CreateGroup(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, group)
	assert.Equal(t, ErrInvalidTargetCount, err)
}

// TestCreateGroupInvalidDeadline tests creation with invalid deadline
func TestCreateGroupInvalidDeadline(t *testing.T) {
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(-1 * time.Hour), // Past deadline
	}

	group, err := testService.CreateGroup(context.Background(), 1, req)
	assert.Error(t, err)
	assert.Nil(t, group)
	assert.Equal(t, ErrInvalidDeadline, err)
}

// TestGetGroupByIDValid tests retrieving group by valid ID
func TestGetGroupByIDValid(t *testing.T) {
	// Create group first
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(context.Background(), 1, req)
	require.NoError(t, err)

	// Retrieve group
	group, err := testService.GetGroupByID(context.Background(), created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, created.ID, group.ID)
	assert.Equal(t, created.ProductID, group.ProductID)
}

// TestGetGroupByIDNotFound tests retrieving non-existent group
func TestGetGroupByIDNotFound(t *testing.T) {
	group, err := testService.GetGroupByID(context.Background(), 99999)
	assert.Error(t, err)
	assert.Nil(t, group)
	assert.Equal(t, ErrGroupNotFound, err)
}

// TestListGroupsValid tests valid group listing
func TestListGroupsValid(t *testing.T) {
	params := &ListGroupsParams{
		Page:    1,
		PerPage: 20,
		Status:  "active",
	}

	result, err := testService.ListGroups(context.Background(), params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PerPage)
	assert.True(t, result.Total >= 0)
}

// TestListGroupsPagination tests pagination validation
func TestListGroupsPagination(t *testing.T) {
	tests := []struct {
		name            string
		page            int
		perPage         int
		expectedPage    int
		expectedPerPage int
	}{
		{name: "Valid first page", page: 1, perPage: 20, expectedPage: 1, expectedPerPage: 20},
		{name: "Invalid page 0", page: 0, perPage: 20, expectedPage: 1, expectedPerPage: 20},
		{name: "Invalid perPage 0", page: 1, perPage: 0, expectedPage: 1, expectedPerPage: 20},
		{name: "PerPage exceeds max", page: 1, perPage: 150, expectedPage: 1, expectedPerPage: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &ListGroupsParams{
				Page:    tt.page,
				PerPage: tt.perPage,
				Status:  "active",
			}

			result, err := testService.ListGroups(context.Background(), params)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedPage, result.Page)
			assert.Equal(t, tt.expectedPerPage, result.PerPage)
		})
	}
}

// TestGetGroupProgress tests group progress calculation
func TestGetGroupProgress(t *testing.T) {
	ctx := context.Background()

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 10,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)

	// Get progress
	progress, err := testService.GetGroupProgress(ctx, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, progress)
	assert.Equal(t, created.ID, progress.GroupID)
	assert.Equal(t, 10, progress.TargetCount)
	assert.Equal(t, 1, progress.CurrentCount)
	assert.True(t, progress.PercentFilled > 0)
	assert.NotEmpty(t, progress.TimeRemaining)
}

// TestJoinGroupValid tests valid group join
func TestJoinGroupValid(t *testing.T) {
	ctx := context.Background()

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)

	// Join group
	result, err := testService.JoinGroup(ctx, 2, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, created.ID, result.Group.ID)
	assert.Equal(t, 2, result.Group.CurrentCount) // Creator + joiner
	assert.True(t, result.OrderID > 0)
}

// TestJoinGroupFull tests joining a full group
func TestJoinGroupFull(t *testing.T) {
	ctx := context.Background()

	// Create group with target 1 (already full with creator)
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 1,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)

	// Try to join full group
	result, err := testService.JoinGroup(ctx, 2, created.ID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrGroupFull, err)
}

// TestJoinGroupExpired tests joining an expired group
func TestJoinGroupExpired(t *testing.T) {
	ctx := context.Background()

	// Create group with past deadline
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(-1 * time.Hour), // Past deadline
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)

	// Try to join expired group (should not be allowed in real scenario,
	// but for test we'll still try)
	result, err := testService.JoinGroup(ctx, 2, created.ID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrGroupExpired, err)
}

// TestJoinGroupInactive tests joining an inactive group
func TestJoinGroupInactive(t *testing.T) {
	ctx := context.Background()

	// Create and then cancel group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)

	// Cancel group
	err = testService.CancelGroup(ctx, 1, created.ID)
	require.NoError(t, err)

	// Try to join cancelled group
	result, err := testService.JoinGroup(ctx, 2, created.ID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrGroupInactive, err)
}

// TestCancelGroupValid tests valid group cancellation
func TestCancelGroupValid(t *testing.T) {
	ctx := context.Background()

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)

	// Cancel group
	err = testService.CancelGroup(ctx, 1, created.ID)
	assert.NoError(t, err)

	// Verify status is cancelled
	group, err := testService.GetGroupByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", group.Status)
}

// TestCancelGroupNotCreator tests cancelling group as non-creator
func TestCancelGroupNotCreator(t *testing.T) {
	ctx := context.Background()

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)

	// Try to cancel as different user
	err = testService.CancelGroup(ctx, 2, created.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotGroupCreator, err)
}

// TestCancelGroupNotFound tests cancelling non-existent group
func TestCancelGroupNotFound(t *testing.T) {
	err := testService.CancelGroup(context.Background(), 1, 99999)
	assert.Error(t, err)
	assert.Equal(t, ErrGroupNotFound, err)
}

// TestGroupCompletionOnJoin tests group auto-completion when target reached
func TestGroupCompletionOnJoin(t *testing.T) {
	ctx := context.Background()

	// Create group with target 2 (1 for creator + 1 needed)
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 2,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, "active", created.Status)
	assert.Equal(t, 1, created.CurrentCount)

	// Join to complete group
	result, err := testService.JoinGroup(ctx, 2, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", result.Group.Status)
	assert.Equal(t, 2, result.Group.CurrentCount)
}

// TestConcurrentGroupCreation tests concurrent group creation
func TestConcurrentGroupCreation(t *testing.T) {
	done := make(chan bool)
	count := 0

	// Create multiple groups concurrently
	for i := 0; i < 5; i++ {
		go func(idx int) {
			req := &CreateGroupRequest{
				ProductID:   1,
				TargetCount: 5 + idx,
				Deadline:    time.Now().Add(24 * time.Hour),
			}
			_, err := testService.CreateGroup(context.Background(), idx, req)
			if err == nil {
				count++
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// All should succeed
	assert.Equal(t, 5, count)
}

// TestGroupFieldsOnCreate tests that all fields are properly set on creation
func TestGroupFieldsOnCreate(t *testing.T) {
	deadline := time.Now().Add(48 * time.Hour)
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 10,
		Deadline:    deadline,
	}

	group, err := testService.CreateGroup(context.Background(), 42, req)
	require.NoError(t, err)

	assert.Equal(t, 1, group.ProductID)
	assert.Equal(t, 42, group.CreatorID)
	assert.Equal(t, 10, group.TargetCount)
	assert.Equal(t, 1, group.CurrentCount)
	assert.Equal(t, "active", group.Status)
	assert.NotZero(t, group.ID)
	assert.NotZero(t, group.CreatedAt)
	assert.NotZero(t, group.UpdatedAt)
}
