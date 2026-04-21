package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/pintuotuo/backend/config"
	apperrors "github.com/pintuotuo/backend/errors"
	"github.com/pintuotuo/backend/middleware"
)

// EntitlementPackageStat aggregates social + sales counts for storefront cards.
type EntitlementPackageStat struct {
	PackageID     int   `json:"package_id"`
	FavoriteCount int64 `json:"favorite_count"`
	LikeCount     int64 `json:"like_count"`
	SalesCount    int64 `json:"sales_count"`
	ReviewCount   int64 `json:"review_count"`
	UserFavorited bool  `json:"user_favorited,omitempty"`
	UserLiked     bool  `json:"user_liked,omitempty"`
	UserReviewed  bool  `json:"user_reviewed,omitempty"`
}

func parseEntitlementPackageIDsQuery(raw string) ([]int, *apperrors.AppError) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]int, 0, len(parts))
	seen := map[int]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil || n <= 0 {
			return nil, apperrors.NewAppError("INVALID_IDS", "无效的套餐 id 列表", http.StatusBadRequest, nil)
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		ids = append(ids, n)
	}
	return ids, nil
}

// BatchGetEntitlementPackageStats GET /entitlement-packages/stats?ids=1,2,3
func BatchGetEntitlementPackageStats(c *gin.Context) {
	ids, perr := parseEntitlementPackageIDsQuery(c.Query("ids"))
	if perr != nil {
		middleware.RespondWithError(c, perr)
		return
	}
	if len(ids) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": []EntitlementPackageStat{}})
		return
	}

	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	rows, err := db.Query(
		`SELECT u.id::int,
			(SELECT COUNT(*)::bigint FROM entitlement_package_favorites f WHERE f.package_id = u.id),
			(SELECT COUNT(*)::bigint FROM entitlement_package_likes l WHERE l.package_id = u.id),
			(SELECT COUNT(*)::bigint FROM orders o WHERE o.entitlement_package_id = u.id AND o.status IN ('paid', 'completed')),
			(SELECT COUNT(*)::bigint FROM entitlement_package_reviews r WHERE r.package_id = u.id)
		 FROM unnest($1::int[]) AS u(id)`,
		pq.Array(ids),
	)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	defer rows.Close()

	out := make([]EntitlementPackageStat, 0, len(ids))
	for rows.Next() {
		var s EntitlementPackageStat
		if scanErr := rows.Scan(&s.PackageID, &s.FavoriteCount, &s.LikeCount, &s.SalesCount, &s.ReviewCount); scanErr != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}

	uid, ok := middleware.OptionalJWTUserID(c)
	if ok && len(ids) > 0 {
		fav := map[int]struct{}{}
		r1, err := db.Query(
			`SELECT package_id FROM entitlement_package_favorites WHERE user_id = $1 AND package_id = ANY($2::int[])`,
			uid, pq.Array(ids),
		)
		if err == nil {
			defer r1.Close()
			for r1.Next() {
				var pid int
				if err := r1.Scan(&pid); err == nil {
					fav[pid] = struct{}{}
				}
			}
		}
		likes := map[int]struct{}{}
		r2, err := db.Query(
			`SELECT package_id FROM entitlement_package_likes WHERE user_id = $1 AND package_id = ANY($2::int[])`,
			uid, pq.Array(ids),
		)
		if err == nil {
			defer r2.Close()
			for r2.Next() {
				var pid int
				if err := r2.Scan(&pid); err == nil {
					likes[pid] = struct{}{}
				}
			}
		}
		rev := map[int]struct{}{}
		r3, err := db.Query(
			`SELECT package_id FROM entitlement_package_reviews WHERE user_id = $1 AND package_id = ANY($2::int[])`,
			uid, pq.Array(ids),
		)
		if err == nil {
			defer r3.Close()
			for r3.Next() {
				var pid int
				if err := r3.Scan(&pid); err == nil {
					rev[pid] = struct{}{}
				}
			}
		}
		for i := range out {
			if _, ok := fav[out[i].PackageID]; ok {
				out[i].UserFavorited = true
			}
			if _, ok := likes[out[i].PackageID]; ok {
				out[i].UserLiked = true
			}
			if _, ok := rev[out[i].PackageID]; ok {
				out[i].UserReviewed = true
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": out})
}

func entitlementPackageExists(db *sql.DB, packageID int) (bool, error) {
	var one int
	err := db.QueryRow(`SELECT 1 FROM entitlement_packages WHERE id = $1`, packageID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// AddEntitlementPackageFavorite POST /entitlement-packages/:id/favorite
func AddEntitlementPackageFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	pkgID, err := strconv.Atoi(c.Param("id"))
	if err != nil || pkgID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	ok, err := entitlementPackageExists(db, pkgID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !ok {
		middleware.RespondWithError(c, apperrors.NewAppError("ENTITLEMENT_PACKAGE_NOT_FOUND", "套餐不存在", http.StatusNotFound, nil))
		return
	}
	if _, err := db.Exec(
		`INSERT INTO entitlement_package_favorites (user_id, package_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		userID, pkgID,
	); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	var cnt int64
	_ = db.QueryRow(`SELECT COUNT(*)::bigint FROM entitlement_package_favorites WHERE package_id = $1`, pkgID).Scan(&cnt)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"favorited": true, "favorite_count": cnt}})
}

// RemoveEntitlementPackageFavorite DELETE /entitlement-packages/:id/favorite
func RemoveEntitlementPackageFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	pkgID, err := strconv.Atoi(c.Param("id"))
	if err != nil || pkgID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if _, err := db.Exec(`DELETE FROM entitlement_package_favorites WHERE user_id = $1 AND package_id = $2`, userID, pkgID); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	var cnt int64
	_ = db.QueryRow(`SELECT COUNT(*)::bigint FROM entitlement_package_favorites WHERE package_id = $1`, pkgID).Scan(&cnt)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"favorited": false, "favorite_count": cnt}})
}

// ToggleEntitlementPackageLike POST /entitlement-packages/:id/like — toggles like for current user
func ToggleEntitlementPackageLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	pkgID, err := strconv.Atoi(c.Param("id"))
	if err != nil || pkgID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	ok, err := entitlementPackageExists(db, pkgID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !ok {
		middleware.RespondWithError(c, apperrors.NewAppError("ENTITLEMENT_PACKAGE_NOT_FOUND", "套餐不存在", http.StatusNotFound, nil))
		return
	}
	res, err := db.Exec(`DELETE FROM entitlement_package_likes WHERE user_id = $1 AND package_id = $2`, userID, pkgID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	n, _ := res.RowsAffected()
	liked := false
	if n == 0 {
		if _, err := db.Exec(`INSERT INTO entitlement_package_likes (user_id, package_id) VALUES ($1, $2)`, userID, pkgID); err != nil {
			middleware.RespondWithError(c, apperrors.ErrDatabaseError)
			return
		}
		liked = true
	}
	var cnt int64
	_ = db.QueryRow(`SELECT COUNT(*)::bigint FROM entitlement_package_likes WHERE package_id = $1`, pkgID).Scan(&cnt)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"liked": liked, "like_count": cnt}})
}

type entitlementPackageReviewReq struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}

// UpsertEntitlementPackageReview POST /entitlement-packages/:id/reviews
func UpsertEntitlementPackageReview(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		middleware.RespondWithError(c, apperrors.ErrInvalidToken)
		return
	}
	pkgID, err := strconv.Atoi(c.Param("id"))
	if err != nil || pkgID <= 0 {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	var req entitlementPackageReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondWithError(c, apperrors.ErrInvalidRequest)
		return
	}
	db := config.GetDB()
	if db == nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	ok, err := entitlementPackageExists(db, pkgID)
	if err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	if !ok {
		middleware.RespondWithError(c, apperrors.NewAppError("ENTITLEMENT_PACKAGE_NOT_FOUND", "套餐不存在", http.StatusNotFound, nil))
		return
	}
	comment := strings.TrimSpace(req.Comment)
	if _, err := db.Exec(
		`INSERT INTO entitlement_package_reviews (user_id, package_id, rating, comment)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id, package_id) DO UPDATE SET
		   rating = EXCLUDED.rating,
		   comment = EXCLUDED.comment,
		   updated_at = CURRENT_TIMESTAMP`,
		userID, pkgID, req.Rating, nullIfEmpty(comment),
	); err != nil {
		middleware.RespondWithError(c, apperrors.ErrDatabaseError)
		return
	}
	var rc int64
	_ = db.QueryRow(`SELECT COUNT(*)::bigint FROM entitlement_package_reviews WHERE package_id = $1`, pkgID).Scan(&rc)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"review_count": rc}})
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
