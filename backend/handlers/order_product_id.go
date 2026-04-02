package handlers

import (
	"database/sql"

	"github.com/pintuotuo/backend/models"
)

func applyNullOrderProductID(o *models.Order, n sql.NullInt64) {
	if n.Valid {
		v := int(n.Int64)
		o.ProductID = &v
	} else {
		o.ProductID = nil
	}
}
