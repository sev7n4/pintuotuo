package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashUserAPIKey returns the SHA256 hex digest used to store user platform API keys (api_keys.key_hash).
func HashUserAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

const platformKeyPrefix = "ptd_"

// PlatformAPIKeyPreview builds a non-secret hint for UI lists (first 4 + last 4 hex chars).
func PlatformAPIKeyPreview(full string) string {
	if !strings.HasPrefix(full, platformKeyPrefix) || len(full) < len(platformKeyPrefix)+8 {
		return platformKeyPrefix + "…"
	}
	body := full[len(platformKeyPrefix):]
	if len(body) < 8 {
		return platformKeyPrefix + "…"
	}
	return platformKeyPrefix + body[:4] + "…" + body[len(body)-4:]
}
