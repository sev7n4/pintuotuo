package services

// SQLReserveSKUStockForOrder runs inside CreateOrder: subtract line quantity from SKU stock.
// When stock is -1 (unlimited per catalog/admin convention), the value must remain -1 (previous bug: -1 - qty became -2).
const SQLReserveSKUStockForOrder = `UPDATE skus SET stock = CASE WHEN stock = -1 THEN -1 ELSE stock - $1 END WHERE id = $2 AND (stock = -1 OR stock >= $1)`

// SQLRestoreSKUStockFromOrderItems restores aggregated quantities from order_items when a pending order is canceled.
// Unlimited SKUs stay at -1 (otherwise cancel would turn -1 into -1+qty).
const SQLRestoreSKUStockFromOrderItems = `UPDATE skus s
   SET stock = CASE WHEN s.stock = -1 THEN -1 ELSE s.stock + oi.qty END
  FROM (
    SELECT sku_id, SUM(quantity) AS qty
      FROM order_items
     WHERE order_id = $1
     GROUP BY sku_id
  ) oi
 WHERE s.id = oi.sku_id`

// SQLRestoreFlashSaleStockFromOrderItems rolls back flash_sale_products.stock_sold when a pending order is canceled.
const SQLRestoreFlashSaleStockFromOrderItems = `UPDATE flash_sale_products fsp
   SET stock_sold = GREATEST(0, fsp.stock_sold - oi.qty),
       updated_at = CURRENT_TIMESTAMP
  FROM (
    SELECT flash_sale_id, sku_id, SUM(quantity)::bigint AS qty
      FROM order_items
     WHERE order_id = $1 AND flash_sale_id IS NOT NULL
     GROUP BY flash_sale_id, sku_id
  ) oi
 WHERE fsp.flash_sale_id = oi.flash_sale_id AND fsp.sku_id = oi.sku_id`
