package middleware

import (
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	apperrors "github.com/pintuotuo/backend/errors"
)

var (
	adminAllowParsedOnce sync.Once
	adminAllowNets       []*net.IPNet
	adminAllowParseErr   error
)

func parseAdminAllowedNets() ([]*net.IPNet, error) {
	raw := strings.TrimSpace(os.Getenv("ADMIN_ALLOWED_CIDRS"))
	if raw == "" {
		return nil, nil
	}
	var nets []*net.IPNet
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if !strings.Contains(part, "/") {
			if ip := net.ParseIP(part); ip != nil {
				if ip.To4() != nil {
					part += "/32"
				} else {
					part += "/128"
				}
			}
		}
		_, n, err := net.ParseCIDR(part)
		if err != nil {
			return nil, err
		}
		nets = append(nets, n)
	}
	return nets, nil
}

// AdminIPAllowlistMiddleware 当环境变量 ADMIN_ALLOWED_CIDRS 非空时，仅允许匹配 CIDR 的客户端访问后续路由（通常配合 /admin）。
func AdminIPAllowlistMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		adminAllowParsedOnce.Do(func() {
			adminAllowNets, adminAllowParseErr = parseAdminAllowedNets()
		})
		if adminAllowParseErr != nil {
			RespondWithError(c, apperrors.NewAppError(
				"ADMIN_IP_CONFIG_INVALID",
				"服务器 ADMIN_ALLOWED_CIDRS 配置无效",
				http.StatusInternalServerError,
				adminAllowParseErr,
			))
			c.Abort()
			return
		}
		if len(adminAllowNets) == 0 {
			c.Next()
			return
		}
		ip := net.ParseIP(c.ClientIP())
		if ip == nil {
			RespondWithError(c, apperrors.NewAppError("FORBIDDEN", "无法解析客户端 IP", http.StatusForbidden, nil))
			c.Abort()
			return
		}
		for _, n := range adminAllowNets {
			if n.Contains(ip) {
				c.Next()
				return
			}
		}
		RespondWithError(c, apperrors.NewAppError("ADMIN_IP_NOT_ALLOWED", "当前网络不允许访问管理接口", http.StatusForbidden, nil))
		c.Abort()
	}
}
