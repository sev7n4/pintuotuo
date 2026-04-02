package handlers

import (
	"database/sql"
	"testing"

	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyNullProductID(t *testing.T) {
	var g models.Group
	applyNullProductID(&g, sql.NullInt64{Valid: false})
	assert.Nil(t, g.ProductID)

	applyNullProductID(&g, sql.NullInt64{Int64: 42, Valid: true})
	assert.NotNil(t, g.ProductID)
	assert.Equal(t, 42, *g.ProductID)
}
