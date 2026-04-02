package handlers

import (
	"database/sql"

	"github.com/pintuotuo/backend/models"
)

// applyNullProductID maps nullable groups.product_id (legacy FK) onto Group.ProductID.
func applyNullProductID(g *models.Group, n sql.NullInt64) {
	if n.Valid {
		v := int(n.Int64)
		g.ProductID = &v
	} else {
		g.ProductID = nil
	}
}
