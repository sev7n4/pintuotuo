package handlers

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pintuotuo/backend/models"
)

// 出站白名单：将 IDE/CLI（Claude Code、Kilo 等）随请求带入的协议头复制到上游，便于 Anthropic 兼容厂商
//（GLM / Kimi / MiniMax 等）识别 beta、版本号或组织维度；鉴权类头永不从此路径复制。
//
// 关闭全部透传：API_PROXY_FORWARD_CLIENT_HEADERS=0|false
// 追加允许名（逗号分隔，大小写不敏感）：API_PROXY_FORWARD_EXTRA_HEADERS=X-My-Gateway-Id,Foo-Bar

const (
	envProxyForwardClientHeaders = "API_PROXY_FORWARD_CLIENT_HEADERS"
	envProxyForwardExtraHeaders  = "API_PROXY_FORWARD_EXTRA_HEADERS"
)

var (
	proxyForwardHeaderBlocklist = map[string]struct{}{}
	proxyForwardBlocklistOnce   sync.Once
)

func initProxyForwardHeaderBlocklist() {
	proxyForwardBlocklistOnce.Do(func() {
		for _, h := range []string{
			"Authorization",
			"X-Api-Key",
			"Cookie",
			"Host",
			"Content-Length",
			"Accept-Encoding",
			"Connection",
			"Transfer-Encoding",
			"Upgrade",
			"Proxy-Authorization",
			"Te",
			"Trailer",
			"X-Request-Id",
			"M-Span-Id",
			"M-Trace-Id",
		} {
			proxyForwardHeaderBlocklist[http.CanonicalHeaderKey(h)] = struct{}{}
		}
	})
}

// envLooksExplicitlyDisabled 识别环境变量中常见的显式关闭取值（集中字面量以满足 goconst）。
func envLooksExplicitlyDisabled(raw string) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "0", "false", "off", "no":
		return true
	default:
		return false
	}
}

func proxyClientHeaderForwardEnabled() bool {
	return !envLooksExplicitlyDisabled(os.Getenv(envProxyForwardClientHeaders))
}

func defaultProxyForwardHeaderNames() []string {
	return []string{
		"Anthropic-Beta",
		"Anthropic-Version",
		"Openai-Beta",
		"Openai-Organization",
	}
}

func parseProxyForwardExtraHeaderNames() []string {
	raw := strings.TrimSpace(os.Getenv(envProxyForwardExtraHeaders))
	if raw == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		t := strings.TrimSpace(p)
		if t == "" {
			continue
		}
		ck := http.CanonicalHeaderKey(t)
		if ck == "" {
			continue
		}
		out = append(out, ck)
	}
	return out
}

func proxyForwardHeaderAllowlist() []string {
	initProxyForwardHeaderBlocklist()
	seen := make(map[string]struct{})
	var list []string
	for _, h := range defaultProxyForwardHeaderNames() {
		ck := http.CanonicalHeaderKey(h)
		if _, blocked := proxyForwardHeaderBlocklist[ck]; blocked {
			continue
		}
		if _, ok := seen[ck]; ok {
			continue
		}
		seen[ck] = struct{}{}
		list = append(list, ck)
	}
	for _, ck := range parseProxyForwardExtraHeaderNames() {
		if _, blocked := proxyForwardHeaderBlocklist[ck]; blocked {
			continue
		}
		if _, ok := seen[ck]; ok {
			continue
		}
		seen[ck] = struct{}{}
		list = append(list, ck)
	}
	return list
}

// applyProxyWhitelistClientHeaders 将入站请求中白名单内的头复制到出站上游请求（多值用 Add 保留）。
// 须在设置平台鉴权头之后调用；Anthropic-Version 缺省由调用方在透传后再补默认 2023-06-01。
func applyProxyWhitelistClientHeaders(c *gin.Context, dst *http.Request) {
	if c == nil || dst == nil || !proxyClientHeaderForwardEnabled() {
		return
	}
	for _, canon := range proxyForwardHeaderAllowlist() {
		vals := c.Request.Header.Values(canon)
		if len(vals) == 0 {
			continue
		}
		dst.Header.Del(canon)
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			dst.Header.Add(canon, v)
		}
	}
}

// applyProxyOutboundAuthHeaders 设置 Content-Type、上游鉴权（Anthropic x-api-key 或 Bearer），
// 再应用追踪头与白名单透传；若 useAnthropicKeyAuth 且透传后仍无 Anthropic-Version，则写入 2023-06-01。
func applyProxyOutboundAuthHeaders(
	c *gin.Context,
	hreq *http.Request,
	requestID string,
	useAnthropicKeyAuth bool,
	dk string,
	pk *models.MerchantAPIKey,
	setSSEAccept bool,
) {
	hreq.Header.Set("Content-Type", "application/json")
	if useAnthropicKeyAuth {
		hreq.Header.Set("X-Api-Key", dk)
	} else {
		authToken := resolveAuthTokenFromRouteMode(resolveRouteMode(pk), dk)
		hreq.Header.Set("Authorization", "Bearer "+authToken)
	}
	if setSSEAccept {
		hreq.Header.Set("Accept", "text/event-stream")
	}
	applyProxyUpstreamHeaders(c, hreq, requestID)
	applyProxyWhitelistClientHeaders(c, hreq)
	if useAnthropicKeyAuth {
		if strings.TrimSpace(hreq.Header.Get("Anthropic-Version")) == "" {
			hreq.Header.Set("Anthropic-Version", "2023-06-01")
		}
	}
}
