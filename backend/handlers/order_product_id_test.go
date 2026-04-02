package handlers

import (
	"database/sql"
	"testing"

	"github.com/pintuotuo/backend/models"
	"github.com/stretchr/testify/assert"
)

func TestApplyNullOrderProductID(t *testing.T) {
	var o models.Order
	applyNullOrderProductID(&o, sql.NullInt64{Valid: false})
	assert.Nil(t, o.ProductID)

	applyNullOrderProductID(&o, sql.NullInt64{Int64: 7, Valid: true})
	assert.NotNil(t, o.ProductID)
	assert.Equal(t, 7, *o.ProductID)
}
