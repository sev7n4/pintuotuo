package group

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pintuotuo/backend/cache"
	"github.com/pintuotuo/backend/config"
)

var testService Service
var userIDCounter int64

func generateUserID() int {
	return int(atomic.AddInt64(&userIDCounter, 1)) + 2000 + int(time.Now().Unix()%1000)
}

func createUser(t *testing.T, email string) int {
	db := config.GetDB()
	var id int
	err := db.QueryRow("INSERT INTO users (email, name, password_hash) VALUES ($1, $2, $3) RETURNING id", email, "Test User", "hash").Scan(&id)
	require.NoError(t, err)

	// Initialize tokens
	_, _ = db.Exec("INSERT INTO tokens (user_id, balance) VALUES ($1, 1000.00) ON CONFLICT DO NOTHING", id)
	return id
}

func init() {
	// Initialize test database
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to init test DB: %v", err)
	}

	// Initialize cache
	if err := cache.Init(); err != nil {
		log.Fatalf("Failed to init cache: %v", err)
	}

	// Clean database and seed for CI environment
	config.TruncateAndSeed()

	logger := log.New(os.Stderr, "[TestGroupService] ", log.LstdFlags)
	testService = NewService(config.GetDB(), logger)
}

// TestCreateGroupValid tests valid group creation
func TestCreateGroupValid(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("groupvalid%d@test.com", generateUserID()))
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}

	group, err := testService.CreateGroup(context.Background(), uid, req)

	assert.NoError(t, err)
	assert.NotNil(t, group)
	assert.Equal(t, req.ProductID, group.ProductID)
	assert.Equal(t, req.TargetCount, group.TargetCount)
	assert.Equal(t, 1, group.CurrentCount) // Creator is first member
	assert.Equal(t, "active", group.Status)
	assert.Equal(t, uid, group.CreatorID)
	assert.True(t, group.ID > 0)
}

// TestCreateGroupInvalidTargetCount tests creation with invalid target count
func TestCreateGroupInvalidTargetCount(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("groupinvtarget%d@test.com", generateUserID()))
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 0,
		Deadline:    time.Now().Add(24 * time.Hour),
	}

	group, err := testService.CreateGroup(context.Background(), uid, req)
	assert.Error(t, err)
	assert.Nil(t, group)
	assert.Equal(t, ErrInvalidTargetCount, err)
}

// TestCreateGroupInvalidDeadline tests creation with invalid deadline
func TestCreateGroupInvalidDeadline(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("groupinvdeadline%d@test.com", generateUserID()))
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(-1 * time.Hour), // Past deadline
	}

	group, err := testService.CreateGroup(context.Background(), uid, req)
	assert.Error(t, err)
	assert.Nil(t, group)
	assert.Equal(t, ErrInvalidDeadline, err)
}

// TestGetGroupByIDValid tests retrieving group by valid ID
func TestGetGroupByIDValid(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("groupget%d@test.com", generateUserID()))
	// Create group first
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(context.Background(), uid, req)
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
	uid := createUser(t, fmt.Sprintf("groupprogress%d@test.com", generateUserID()))

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 10,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid, req)
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
	uid1 := createUser(t, fmt.Sprintf("groupjoin1_%d@test.com", generateUserID()))
	uid2 := createUser(t, fmt.Sprintf("groupjoin2_%d@test.com", generateUserID()))

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid1, req)
	require.NoError(t, err)

	// Join group
	result, err := testService.JoinGroup(ctx, uid2, created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, created.ID, result.Group.ID)
	assert.Equal(t, 2, result.Group.CurrentCount) // Creator + joiner
	assert.True(t, result.OrderID > 0)
}

// TestJoinGroupFull tests joining a full group
func TestJoinGroupFull(t *testing.T) {
	ctx := context.Background()
	uid1 := createUser(t, fmt.Sprintf("groupfull1_%d@test.com", generateUserID()))
	uid2 := createUser(t, fmt.Sprintf("groupfull2_%d@test.com", generateUserID()))

	// Create group with target 1 (already full with creator)
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 1,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid1, req)
	require.NoError(t, err)

	// Try to join full group
	result, err := testService.JoinGroup(ctx, uid2, created.ID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrGroupFull, err)
}

// TestJoinGroupExpired tests joining an expired group
func TestJoinGroupExpired(t *testing.T) {
	ctx := context.Background()
	uid1 := createUser(t, fmt.Sprintf("groupexp1_%d@test.com", generateUserID()))
	uid2 := createUser(t, fmt.Sprintf("groupexp2_%d@test.com", generateUserID()))

	// Create group with future deadline first (to pass validation)
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid1, req)
	require.NoError(t, err)

	// Manually update group deadline to the far past in database
	_, err = config.GetDB().ExecContext(ctx, "UPDATE groups SET deadline = '2000-01-01 00:00:00' WHERE id = $1", created.ID)
	require.NoError(t, err)

	// Try to join expired group
	result, err := testService.JoinGroup(ctx, uid2, created.ID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrGroupExpired, err)
}

// TestJoinGroupInactive tests joining an inactive group
func TestJoinGroupInactive(t *testing.T) {
	ctx := context.Background()
	uid1 := createUser(t, fmt.Sprintf("groupinact1_%d@test.com", generateUserID()))
	uid2 := createUser(t, fmt.Sprintf("groupinact2_%d@test.com", generateUserID()))

	// Create and then cancel group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid1, req)
	require.NoError(t, err)

	// Cancel group
	err = testService.CancelGroup(ctx, uid1, created.ID)
	require.NoError(t, err)

	// Try to join cancelled group
	result, err := testService.JoinGroup(ctx, uid2, created.ID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrGroupInactive, err)
}

// TestCancelGroupValid tests valid group cancellation
func TestCancelGroupValid(t *testing.T) {
	ctx := context.Background()
	uid := createUser(t, fmt.Sprintf("groupcancel%d@test.com", generateUserID()))

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid, req)
	require.NoError(t, err)

	// Cancel group
	err = testService.CancelGroup(ctx, uid, created.ID)
	assert.NoError(t, err)

	// Verify status is cancelled
	group, err := testService.GetGroupByID(ctx, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "cancelled", group.Status)
}

// TestCancelGroupNotCreator tests cancelling group as non-creator
func TestCancelGroupNotCreator(t *testing.T) {
	ctx := context.Background()
	uid1 := createUser(t, fmt.Sprintf("groupnotcr1_%d@test.com", generateUserID()))
	uid2 := createUser(t, fmt.Sprintf("groupnotcr2_%d@test.com", generateUserID()))

	// Create group
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 5,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid1, req)
	require.NoError(t, err)

	// Try to cancel as different user
	err = testService.CancelGroup(ctx, uid2, created.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrNotGroupCreator, err)
}

// TestCancelGroupNotFound tests cancelling non-existent group
func TestCancelGroupNotFound(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("groupcancelnotfound%d@test.com", generateUserID()))
	err := testService.CancelGroup(context.Background(), uid, 99999)
	assert.Error(t, err)
	assert.Equal(t, ErrGroupNotFound, err)
}

// TestGroupCompletionOnJoin tests group auto-completion when target reached
func TestGroupCompletionOnJoin(t *testing.T) {
	ctx := context.Background()
	uid1 := createUser(t, fmt.Sprintf("groupcompl1_%d@test.com", generateUserID()))
	uid2 := createUser(t, fmt.Sprintf("groupcompl2_%d@test.com", generateUserID()))

	// Create group with target 2 (1 for creator + 1 needed)
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 2,
		Deadline:    time.Now().Add(24 * time.Hour),
	}
	created, err := testService.CreateGroup(ctx, uid1, req)
	require.NoError(t, err)
	assert.Equal(t, "active", created.Status)
	assert.Equal(t, 1, created.CurrentCount)

	// Join to complete group
	result, err := testService.JoinGroup(ctx, uid2, created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "completed", result.Group.Status)
	assert.Equal(t, 2, result.Group.CurrentCount)
}

// TestConcurrentGroupCreation tests concurrent group creation
func TestConcurrentGroupCreation(t *testing.T) {
	var count int32
	var wg sync.WaitGroup
	ctx := context.Background()

	// Ensure at least one product and user exist for testing
	// These are typically pre-seeded or created in init/setup
	uid := createUser(t, "concurrent_creator@example.com")
	_, _ = config.GetDB().ExecContext(ctx, "INSERT INTO products (id, merchant_id, name, price, stock) VALUES (1, $1, 'Test Product', 99.99, 100) ON CONFLICT DO NOTHING", uid)

	// Create multiple groups concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := &CreateGroupRequest{
				ProductID:   1,
				TargetCount: 5 + idx,
				Deadline:    time.Now().Add(24 * time.Hour),
			}
			_, err := testService.CreateGroup(context.Background(), uid, req)
			if err == nil {
				atomic.AddInt32(&count, 1)
			}
		}(i)
	}

	// Wait for all goroutines
	wg.Wait()

	// All should succeed
	assert.Equal(t, int32(5), count)
}

// TestGroupFieldsOnCreate tests that all fields are properly set on creation
func TestGroupFieldsOnCreate(t *testing.T) {
	uid := createUser(t, fmt.Sprintf("groupfields%d@test.com", generateUserID()))
	deadline := time.Now().Add(48 * time.Hour)
	req := &CreateGroupRequest{
		ProductID:   1,
		TargetCount: 10,
		Deadline:    deadline,
	}

	group, err := testService.CreateGroup(context.Background(), uid, req)
	require.NoError(t, err)

	assert.Equal(t, 1, group.ProductID)
	assert.Equal(t, uid, group.CreatorID)
	assert.Equal(t, 10, group.TargetCount)
	assert.Equal(t, 1, group.CurrentCount)
	assert.Equal(t, "active", group.Status)
	assert.NotZero(t, group.ID)
	assert.NotZero(t, group.CreatedAt)
	assert.NotZero(t, group.UpdatedAt)
}
