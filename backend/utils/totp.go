package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// ValidateTOTP 校验 RFC 6238 标准 30 秒步长、6 位数字、SHA1（与 Google Authenticator 兼容）。
func ValidateTOTP(code string, secretBase32 string) bool {
	key, err := decodeBase32Secret(secretBase32)
	if err != nil {
		return false
	}
	c := strings.TrimSpace(code)
	if len(c) != 6 {
		return false
	}
	now := time.Now().Unix()
	for _, skew := range []int64{-1, 0, 1} {
		if hotpMatch(key, uint64(now/30+skew), c) {
			return true
		}
	}
	return false
}

func decodeBase32Secret(s string) ([]byte, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	s = strings.TrimRight(s, "=")
	for len(s)%8 != 0 {
		s += "="
	}
	return base32.StdEncoding.DecodeString(s)
}

func hotpMatch(key []byte, counter uint64, expect string) bool {
	h := hmac.New(sha1.New, key)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)
	_, _ = h.Write(buf)
	digest := h.Sum(nil)
	offset := digest[len(digest)-1] & 0x0f
	v := binary.BigEndian.Uint32(digest[offset:offset+4]) & 0x7fffffff
	otp := fmt.Sprintf("%06d", v%1000000)
	return subtle.ConstantTimeCompare([]byte(otp), []byte(expect)) == 1
}
