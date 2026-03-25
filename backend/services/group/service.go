package group

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/pintuotuo/backend/cache"
)

// Service defines the group service interface
type Service interface {
	// Read operations
	ListGroups(ctx context.Context, params *ListGroupsParams) (*ListGroupsResult, error)
	GetGroupByID(ctx context.Context, groupID int) (*Group, error)
	GetGroupProgress(ctx context.Context, groupID int) (*GroupProgress, error)

	// Write operations
	CreateGroup(ctx context.Context, creatorID int, req *CreateGroupRequest) (*Group, error)
	JoinGroup(ctx context.Context, userID int, groupID int) (*JoinGroupResult, error)
	CancelGroup(ctx context.Context, creatorID int, groupID int) error
}

// service implements the Service interface
type service struct {
	db  *sql.DB
	log *log.Logger
}

// NewService creates a new group service
func NewService(db *sql.DB, logger *log.Logger) Service {
	if logger == nil {
		logger = log.New(os.Stderr, "[GroupService] ", log.LstdFlags)
	}

	return &service{
		db:  db,
		log: logger,
	}
}

// ListGroups retrieves groups with pagination
func (s *service) ListGroups(ctx context.Context, params *ListGroupsParams) (*ListGroupsResult, error) {
	// Validate parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}
	if params.Status == "" {
		params.Status = "active"
	}

	// Try cache first
	cacheKey := cache.GroupListKey(params.Page, params.PerPage, params.Status)
	if cachedData, err := cache.Get(ctx, cacheKey); err == nil {
		var result ListGroupsResult
		if err := json.Unmarshal([]byte(cachedData), &result); err == nil {
			return &result, nil
		}
	}

	// Query database
	offset := (params.Page - 1) * params.PerPage

	var rows *sql.Rows
	var err error

	if params.Status == "all" {
		rows, err = s.db.QueryContext(
			ctx,
			"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			params.PerPage, offset,
		)
	} else {
		rows, err = s.db.QueryContext(
			ctx,
			"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
			params.Status, params.PerPage, offset,
		)
	}

	if err != nil {
		return nil, wrapError("ListGroups", "query", err)
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		err := rows.Scan(&g.ID, &g.ProductID, &g.CreatorID, &g.TargetCount, &g.CurrentCount, &g.Status, &g.Deadline, &g.CreatedAt, &g.UpdatedAt)
		if err != nil {
			return nil, wrapError("ListGroups", "scan", err)
		}
		groups = append(groups, g)
	}

	// Get total count
	var total int
	if params.Status == "all" {
		s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM groups").Scan(&total)
	} else {
		s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM groups WHERE status = $1", params.Status).Scan(&total)
	}

	result := &ListGroupsResult{
		Total:   total,
		Page:    params.Page,
		PerPage: params.PerPage,
		Data:    groups,
	}

	// Cache result
	if resultJSON, err := json.Marshal(result); err == nil {
		_ = cache.Set(ctx, cacheKey, string(resultJSON), cache.GroupCacheTTL)
	}

	s.log.Printf("Listed groups: page=%d, per_page=%d, status=%s, total=%d", params.Page, params.PerPage, params.Status, total)
	return result, nil
}

// GetGroupByID retrieves a single group by ID
func (s *service) GetGroupByID(ctx context.Context, groupID int) (*Group, error) {
	var group Group
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE id = $1",
		groupID,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrGroupNotFound
		}
		return nil, wrapError("GetGroupByID", "query", err)
	}

	return &group, nil
}

// GetGroupProgress retrieves group progress information
func (s *service) GetGroupProgress(ctx context.Context, groupID int) (*GroupProgress, error) {
	var progress GroupProgress
	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at FROM groups WHERE id = $1",
		groupID,
	).Scan(&progress.GroupID, &progress.ProductID, &progress.CreatorID, &progress.TargetCount, &progress.CurrentCount, &progress.Status, &progress.Deadline, &progress.CreatedAt, &progress.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrGroupNotFound
		}
		return nil, wrapError("GetGroupProgress", "query", err)
	}

	// Calculate progress
	if progress.TargetCount > 0 {
		progress.PercentFilled = float64(progress.CurrentCount) / float64(progress.TargetCount) * 100
	}

	// Calculate time remaining
	now := time.Now()
	if progress.Deadline.Before(now) {
		progress.TimeRemaining = "0s"
	} else {
		duration := progress.Deadline.Sub(now)
		progress.TimeRemaining = duration.String()
	}

	return &progress, nil
}

// CreateGroup creates a new group
func (s *service) CreateGroup(ctx context.Context, creatorID int, req *CreateGroupRequest) (*Group, error) {
	// Validate input
	if req.TargetCount <= 0 {
		return nil, ErrInvalidTargetCount
	}
	if req.Deadline.Before(time.Now()) {
		return nil, ErrInvalidDeadline
	}

	// Verify product exists
	var productID int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM products WHERE id = $1", req.ProductID).Scan(&productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, wrapError("CreateGroup", "verifyProduct", err)
	}

	// Create group (starts with creator as first member)
	var group Group
	err = s.db.QueryRowContext(
		ctx,
		"INSERT INTO groups (product_id, creator_id, target_count, current_count, status, deadline) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at",
		req.ProductID, creatorID, req.TargetCount, 1, "active", req.Deadline,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		return nil, wrapError("CreateGroup", "insert", err)
	}

	// Invalidate list cache
	_ = cache.InvalidatePatterns(ctx, "groups:list:*")

	s.log.Printf("Group created: id=%d, creator_id=%d, product_id=%d, target_count=%d", group.ID, creatorID, req.ProductID, req.TargetCount)
	return &group, nil
}

// JoinGroup adds a user to an existing group
func (s *service) JoinGroup(ctx context.Context, userID int, groupID int) (*JoinGroupResult, error) {
	// Get group info
	var group Group
	var productPrice float64

	err := s.db.QueryRowContext(
		ctx,
		"SELECT id, product_id, target_count, current_count, status, deadline FROM groups WHERE id = $1",
		groupID,
	).Scan(&group.ID, &group.ProductID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrGroupNotFound
		}
		return nil, wrapError("JoinGroup", "getGroup", err)
	}

	// Check group status
	if group.Status != "active" {
		return nil, ErrGroupInactive
	}

	// Check if group is full
	if group.CurrentCount >= group.TargetCount {
		return nil, ErrGroupFull
	}

	// Check if deadline has passed
	if time.Now().After(group.Deadline) {
		return nil, ErrGroupExpired
	}

	// Get product price
	err = s.db.QueryRowContext(
		ctx,
		"SELECT price FROM products WHERE id = $1",
		group.ProductID,
	).Scan(&productPrice)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, wrapError("JoinGroup", "getProduct", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, wrapError("JoinGroup", "beginTx", err)
	}
	defer tx.Rollback()

	// Create order
	var orderID int
	err = tx.QueryRowContext(
		ctx,
		"INSERT INTO orders (user_id, product_id, group_id, quantity, unit_price, total_price, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		userID, group.ProductID, group.ID, 1, productPrice, productPrice, "pending",
	).Scan(&orderID)

	if err != nil {
		return nil, wrapError("JoinGroup", "createOrder", err)
	}

	// Add to group members
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO group_members (group_id, user_id, order_id) VALUES ($1, $2, $3)",
		group.ID, userID, orderID,
	)

	if err != nil {
		return nil, ErrAlreadyInGroup
	}

	// Update group count and status if target reached
	newCount := group.CurrentCount + 1
	newStatus := group.Status
	if newCount >= group.TargetCount {
		newStatus = "completed"
	}

	err = tx.QueryRowContext(
		ctx,
		"UPDATE groups SET current_count = $1, status = $2 WHERE id = $3 RETURNING id, product_id, creator_id, target_count, current_count, status, deadline, created_at, updated_at",
		newCount, newStatus, group.ID,
	).Scan(&group.ID, &group.ProductID, &group.CreatorID, &group.TargetCount, &group.CurrentCount, &group.Status, &group.Deadline, &group.CreatedAt, &group.UpdatedAt)

	if err != nil {
		return nil, wrapError("JoinGroup", "updateGroup", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, wrapError("JoinGroup", "commitTx", err)
	}

	// Invalidate cache
	_ = cache.InvalidatePatterns(ctx, "groups:list:*")

	result := &JoinGroupResult{
		Group:   &group,
		OrderID: orderID,
	}

	s.log.Printf("User joined group: user_id=%d, group_id=%d, order_id=%d", userID, groupID, orderID)
	return result, nil
}

// CancelGroup cancels a group (creator only)
func (s *service) CancelGroup(ctx context.Context, creatorID int, groupID int) error {
	// Get group info
	var storedCreatorID int
	var status string

	err := s.db.QueryRowContext(
		ctx,
		"SELECT creator_id, status FROM groups WHERE id = $1",
		groupID,
	).Scan(&storedCreatorID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			return ErrGroupNotFound
		}
		return wrapError("CancelGroup", "getGroup", err)
	}

	// Verify creator
	if storedCreatorID != creatorID {
		return ErrNotGroupCreator
	}

	// Check if group can be cancelled
	if status == "completed" || status == "cancelled" {
		return ErrCannotCancelGroup
	}

	// Cancel group
	_, err = s.db.ExecContext(
		ctx,
		"UPDATE groups SET status = 'cancelled' WHERE id = $1",
		groupID,
	)

	if err != nil {
		return wrapError("CancelGroup", "update", err)
	}

	// Invalidate cache
	_ = cache.InvalidatePatterns(ctx, "groups:list:*")

	s.log.Printf("Group cancelled: group_id=%d, creator_id=%d", groupID, creatorID)
	return nil
}
