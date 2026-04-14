package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

type EntitlementPackageItem struct {
	ID              int     `json:"id"`
	SKUID           int     `json:"sku_id"`
	SKUCode         string  `json:"sku_code"`
	SPUName         string  `json:"spu_name"`
	SKUType         string  `json:"sku_type"`
	DefaultQuantity int     `json:"default_quantity"`
	RetailPrice     float64 `json:"retail_price"`
}

type EntitlementPackage struct {
	ID          int                      `json:"id"`
	PackageCode string                   `json:"package_code"`
	Name        string                   `json:"name"`
	Description string                   `json:"description,omitempty"`
	Status      string                   `json:"status"`
	SortOrder   int                      `json:"sort_order"`
	StartAt     *time.Time               `json:"start_at,omitempty"`
	EndAt       *time.Time               `json:"end_at,omitempty"`
	IsFeatured  bool                     `json:"is_featured"`
	BadgeText   string                   `json:"badge_text,omitempty"`
	Items       []EntitlementPackageItem `json:"items"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

type EntitlementPackageUserView struct {
	EntitlementPackage
	CoveredItems int  `json:"covered_items"`
	TotalItems   int  `json:"total_items"`
	IsActive     bool `json:"is_active"`
}

type entitlementPackageUpsertReq struct {
	PackageCode string     `json:"package_code"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	SortOrder   int        `json:"sort_order"`
	StartAt     *time.Time `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	IsFeatured  bool       `json:"is_featured"`
	BadgeText   string     `json:"badge_text"`
	Items       []struct {
		SKUID           int `json:"sku_id"`
		DefaultQuantity int `json:"default_quantity"`
	} `json:"items"`
}

func validateEntitlementPackageReq(req *entitlementPackageUpsertReq, needCode bool) *apperrors.AppError {
	if needCode && strings.TrimSpace(req.PackageCode) == "" {
		return apperrors.NewAppError("INVALID_REQUEST", "package_code 不能为空", http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(req.Name) == "" {
		return apperrors.NewAppError("INVALID_REQUEST", "name 不能为空", http.StatusBadRequest, nil)
	}
	if len(req.Items) == 0 {
		return apperrors.NewAppError("INVALID_REQUEST", "items 至少包含一个 SKU", http.StatusBadRequest, nil)
	}
	seen := map[int]struct{}{}
	for _, it := range req.Items {
		if it.SKUID <= 0 {
			return apperrors.NewAppError("INVALID_REQUEST", "sku_id 非法", http.StatusBadRequest, nil)
		}
		if it.DefaultQuantity <= 0 {
			return apperrors.NewAppError("INVALID_REQUEST", "default_quantity 必须大于 0", http.StatusBadRequest, nil)
		}
		if _, ok := seen[it.SKUID]; ok {
			return apperrors.NewAppError("INVALID_REQUEST", "items 中 sku_id 重复", http.StatusBadRequest, nil)
		}
		seen[it.SKUID] = struct{}{}
	}
	if req.Status != "" && req.Status != merchantStatusActive && req.Status != merchantSKUStatusInactive {
		return apperrors.NewAppError("INVALID_REQUEST", "status 只能是 active/inactive", http.StatusBadRequest, nil)
	}
	if req.StartAt != nil && req.EndAt != nil && !req.EndAt.After(*req.StartAt) {
		return apperrors.NewAppError("INVALID_REQUEST", "end_at 必须晚于 start_at", http.StatusBadRequest, nil)
	}
	return nil
}

func loadEntitlementPackageItems(db *sql.DB, packageID int) ([]EntitlementPackageItem, error) {
	rows, err := db.Query(
		`SELECT epi.id, epi.sku_id, s.sku_code, sp.name, s.sku_type, epi.default_quantity, s.retail_price
		 FROM entitlement_package_items epi
		 JOIN skus s ON epi.sku_id = s.id
		 JOIN spus sp ON s.spu_id = sp.id
		 WHERE epi.package_id = $1
		 ORDER BY epi.id ASC`,
		packageID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EntitlementPackageItem, 0)
	for rows.Next() {
		var it EntitlementPackageItem
		if err = rows.Scan(&it.ID, &it.SKUID, &it.SKUCode, &it.SPUName, &it.SKUType, &it.DefaultQuantity, &it.RetailPrice); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func ListAdminEntitlementPackages(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	db := config.GetDB()
	rows, err := db.Query(
		`SELECT id, package_code, name, COALESCE(description, ''), status, sort_order, start_at, end_at, is_featured, COALESCE(badge_text, ''), created_at, updated_at
		 FROM entitlement_packages
		 ORDER BY is_featured DESC, sort_order ASC, id ASC`,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	pkgs := make([]EntitlementPackage, 0)
	for rows.Next() {
		var p EntitlementPackage
		var startAt, endAt sql.NullTime
		if err = rows.Scan(&p.ID, &p.PackageCode, &p.Name, &p.Description, &p.Status, &p.SortOrder, &startAt, &endAt, &p.IsFeatured, &p.BadgeText, &p.CreatedAt, &p.UpdatedAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if startAt.Valid {
			t := startAt.Time
			p.StartAt = &t
		}
		if endAt.Valid {
			t := endAt.Time
			p.EndAt = &t
		}
		items, iErr := loadEntitlementPackageItems(db, p.ID)
		if iErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		p.Items = items
		pkgs = append(pkgs, p)
	}
	c.JSON(http.StatusOK, gin.H{"data": pkgs})
}

func CreateAdminEntitlementPackage(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	var req entitlementPackageUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if vErr := validateEntitlementPackageReq(&req, true); vErr != nil {
		middleware.RespondWithError(c, vErr)
		return
	}
	db := config.GetDB()
	tx, err := db.Begin()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer tx.Rollback()

	status := req.Status
	if status == "" {
		status = productStatusActive
	}

	var packageID int
	err = tx.QueryRow(
		`INSERT INTO entitlement_packages (package_code, name, description, status, sort_order, start_at, end_at, is_featured, badge_text)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, ''))
		 RETURNING id`,
		strings.ToUpper(strings.TrimSpace(req.PackageCode)),
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.Description),
		status,
		req.SortOrder,
		req.StartAt,
		req.EndAt,
		req.IsFeatured,
		strings.TrimSpace(req.BadgeText),
	).Scan(&packageID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.NewAppError("CREATE_FAILED", "创建权益包失败", http.StatusInternalServerError, err))
		return
	}
	for _, it := range req.Items {
		if _, err = tx.Exec(
			`INSERT INTO entitlement_package_items (package_id, sku_id, default_quantity) VALUES ($1, $2, $3)`,
			packageID, it.SKUID, it.DefaultQuantity,
		); err != nil {
			middleware.RespondWithError(c, apperrors.NewAppError("CREATE_FAILED", "写入权益包明细失败", http.StatusInternalServerError, err))
			return
		}
	}
	if err = tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": packageID})
}

func UpdateAdminEntitlementPackage(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	packageID, err := strconv.Atoi(c.Param("id"))
	if err != nil || packageID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	var req entitlementPackageUpsertReq
	if err = c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	if vErr := validateEntitlementPackageReq(&req, false); vErr != nil {
		middleware.RespondWithError(c, vErr)
		return
	}
	db := config.GetDB()
	tx, err := db.Begin()
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer tx.Rollback()

	status := req.Status
	if status == "" {
		status = productStatusActive
	}
	if _, err = tx.Exec(
		`UPDATE entitlement_packages
		 SET name = $1, description = $2, status = $3, sort_order = $4,
		     start_at = $5, end_at = $6, is_featured = $7, badge_text = NULLIF($8, ''),
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = $9`,
		strings.TrimSpace(req.Name), strings.TrimSpace(req.Description), status, req.SortOrder,
		req.StartAt, req.EndAt, req.IsFeatured, strings.TrimSpace(req.BadgeText), packageID,
	); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if _, err = tx.Exec(`DELETE FROM entitlement_package_items WHERE package_id = $1`, packageID); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	for _, it := range req.Items {
		if _, err = tx.Exec(
			`INSERT INTO entitlement_package_items (package_id, sku_id, default_quantity) VALUES ($1, $2, $3)`,
			packageID, it.SKUID, it.DefaultQuantity,
		); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteAdminEntitlementPackage(c *gin.Context) {
	if !ensureAdmin(c) {
		return
	}
	packageID, err := strconv.Atoi(c.Param("id"))
	if err != nil || packageID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	db := config.GetDB()
	if _, err = db.Exec(`DELETE FROM entitlement_packages WHERE id = $1`, packageID); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func ListPublicEntitlementPackages(c *gin.Context) {
	db := config.GetDB()
	rows, err := db.Query(
		`SELECT id, package_code, name, COALESCE(description, ''), status, sort_order, start_at, end_at, is_featured, COALESCE(badge_text, ''), created_at, updated_at
		 FROM entitlement_packages
		 WHERE status = 'active'
		   AND (start_at IS NULL OR start_at <= NOW())
		   AND (end_at IS NULL OR end_at > NOW())
		 ORDER BY is_featured DESC, sort_order ASC, id ASC`,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	pkgs := make([]EntitlementPackage, 0)
	for rows.Next() {
		var p EntitlementPackage
		var startAt, endAt sql.NullTime
		if err = rows.Scan(&p.ID, &p.PackageCode, &p.Name, &p.Description, &p.Status, &p.SortOrder, &startAt, &endAt, &p.IsFeatured, &p.BadgeText, &p.CreatedAt, &p.UpdatedAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if startAt.Valid {
			t := startAt.Time
			p.StartAt = &t
		}
		if endAt.Valid {
			t := endAt.Time
			p.EndAt = &t
		}
		items, iErr := loadEntitlementPackageItems(db, p.ID)
		if iErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		p.Items = items
		pkgs = append(pkgs, p)
	}
	c.JSON(http.StatusOK, gin.H{"data": pkgs})
}

func GetMyEntitlementPackages(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	userID, ok := userIDRaw.(int)
	if !ok {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	db := config.GetDB()

	coveredRows, err := db.Query(
		`SELECT DISTINCT sku_id
		 FROM (
		   SELECT us.sku_id
		     FROM user_subscriptions us
		    WHERE us.user_id = $1
		      AND us.status = 'active'
		   UNION
		   SELECT oi.sku_id
		     FROM orders o
		     JOIN order_items oi ON oi.order_id = o.id
		    WHERE o.user_id = $1
		      AND o.status IN ('paid', 'completed')
		 ) x`,
		userID,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer coveredRows.Close()
	covered := map[int]struct{}{}
	for coveredRows.Next() {
		var skuID int
		if err = coveredRows.Scan(&skuID); err == nil {
			covered[skuID] = struct{}{}
		}
	}

	rows, err := db.Query(
		`SELECT id, package_code, name, COALESCE(description, ''), status, sort_order, start_at, end_at, is_featured, COALESCE(badge_text, ''), created_at, updated_at
		 FROM entitlement_packages
		 WHERE status = 'active'
		   AND (start_at IS NULL OR start_at <= NOW())
		   AND (end_at IS NULL OR end_at > NOW())
		 ORDER BY is_featured DESC, sort_order ASC, id ASC`,
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	out := make([]EntitlementPackageUserView, 0)
	for rows.Next() {
		var p EntitlementPackageUserView
		var startAt, endAt sql.NullTime
		if err = rows.Scan(&p.ID, &p.PackageCode, &p.Name, &p.Description, &p.Status, &p.SortOrder, &startAt, &endAt, &p.IsFeatured, &p.BadgeText, &p.CreatedAt, &p.UpdatedAt); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		if startAt.Valid {
			t := startAt.Time
			p.StartAt = &t
		}
		if endAt.Valid {
			t := endAt.Time
			p.EndAt = &t
		}
		items, iErr := loadEntitlementPackageItems(db, p.ID)
		if iErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		p.Items = items
		p.TotalItems = len(items)
		for _, it := range items {
			if _, ok = covered[it.SKUID]; ok {
				p.CoveredItems++
			}
		}
		p.IsActive = p.TotalItems > 0 && p.CoveredItems == p.TotalItems
		out = append(out, p)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
