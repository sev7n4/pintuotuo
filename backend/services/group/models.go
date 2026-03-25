package group

import "time"

// CreateGroupRequest represents group creation request
type CreateGroupRequest struct {
	ProductID   int       `json:"product_id" binding:"required"`
	TargetCount int       `json:"target_count" binding:"required,gt=0"`
	Deadline    time.Time `json:"deadline" binding:"required"`
}

// ListGroupsParams represents list parameters
type ListGroupsParams struct {
	Page    int
	PerPage int
	Status  string
}

// GroupProgress represents group progress information
type GroupProgress struct {
	GroupID       int       `json:"group_id"`
	ProductID     int       `json:"product_id"`
	CreatorID     int       `json:"creator_id"`
	TargetCount   int       `json:"target_count"`
	CurrentCount  int       `json:"current_count"`
	Status        string    `json:"status"`
	Deadline      time.Time `json:"deadline"`
	PercentFilled float64   `json:"percent_filled"`
	TimeRemaining string    `json:"time_remaining"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ListGroupsResult represents paginated group list result
type ListGroupsResult struct {
	Total   int     `json:"total"`
	Page    int     `json:"page"`
	PerPage int     `json:"per_page"`
	Data    []Group `json:"data"`
}

// Group represents a group purchase
type Group struct {
	ID          int       `json:"id"`
	ProductID   int       `json:"product_id"`
	CreatorID   int       `json:"creator_id"`
	TargetCount int       `json:"target_count"`
	CurrentCount int      `json:"current_count"`
	Status      string    `json:"status"`
	Deadline    time.Time `json:"deadline"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// JoinGroupResult represents the result of joining a group
type JoinGroupResult struct {
	Group   *Group `json:"group"`
	OrderID int    `json:"order_id"`
}
