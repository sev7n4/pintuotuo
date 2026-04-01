package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashUserAPIKey returns the SHA256 hex digest used to store user platform API keys (api_keys.key_hash).
func HashUserAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
