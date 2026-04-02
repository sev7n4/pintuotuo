package handlers

import (
	"database/sql"

	"github.com/pintuotuo/backend/models"
)

// productFromSKU scans one row from listSKUProductsBaseQuery-shaped SELECT into models.Product.
// id is SKU id (C端「商品」以可下单 SKU 为粒度展示).
func productFromSKU(
	rows *sql.Rows,
) (models.Product, error) {
	var p models.Product
	var mid sql.NullInt64
	var orig sql.NullFloat64
	err := rows.Scan(
		&p.ID, &mid, &p.SpuID, &p.Name, &p.Description, &p.Price, &orig,
		&p.Stock, &p.SoldCount, &p.Category, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return p, err
	}
	if mid.Valid {
		p.MerchantID = int(mid.Int64)
	}
	if orig.Valid {
		p.OriginalPrice = orig.Float64
	} else {
		p.OriginalPrice = p.Price
	}
	return p, nil
}

// listSKUProductsBaseQuery selects marketplace rows: one row per active SKU.
const listSKUProductsBaseQuery = `
SELECT s.id,
	COALESCE(s.merchant_id, 0) AS merchant_id,
	s.spu_id,
	sp.name || ' · ' || s.sku_code AS name,
	COALESCE(sp.description, '') AS description,
	s.retail_price AS price,
	s.original_price AS original_price,
	CASE WHEN s.stock = -1 THEN 999999 ELSE s.stock END AS stock,
	COALESCE(s.sales_count, 0) AS sold_count,
	COALESCE(sp.model_tier, '') AS category,
	s.status,
	s.created_at,
	s.updated_at
FROM skus s
JOIN spus sp ON s.spu_id = sp.id
WHERE s.status = 'active' AND sp.status = 'active'
	AND (s.stock > 0 OR s.stock = -1)`

// getProductBySKUID loads one marketplace Product by SKU id (C端详情页 id).
func getProductBySKUID(db *sql.DB, skuID int) (models.Product, error) {
	row := db.QueryRow(
		listSKUProductsBaseQuery+` AND s.id = $1`,
		skuID,
	)
	var p models.Product
	var mid sql.NullInt64
	var orig sql.NullFloat64
	err := row.Scan(
		&p.ID, &mid, &p.SpuID, &p.Name, &p.Description, &p.Price, &orig,
		&p.Stock, &p.SoldCount, &p.Category, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return p, err
	}
	if mid.Valid {
		p.MerchantID = int(mid.Int64)
	}
	if orig.Valid {
		p.OriginalPrice = orig.Float64
	} else {
		p.OriginalPrice = p.Price
	}
	return p, nil
}
