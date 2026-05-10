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
