package services

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	compatPrefixMaxCount  = 32
	compatPrefixMaxLen    = 64
	compatPrefixPatternRe = `^[a-z0-9][a-z0-9._-]*$`
)

var compatPrefixPattern = regexp.MustCompile(compatPrefixPatternRe)

// NormalizeCompatPrefixes trims, lowercases, dedupes (case-insensitive), validates each segment.
// Empty input returns empty slice (stored as {} in DB).
func NormalizeCompatPrefixes(prefixes []string) ([]string, error) {
	if len(prefixes) == 0 {
		return nil, nil
	}
	if len(prefixes) > compatPrefixMaxCount {
		return nil, fmt.Errorf("at most %d prefixes allowed", compatPrefixMaxCount)
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(prefixes))
	for _, raw := range prefixes {
		p := strings.ToLower(strings.TrimSpace(raw))
		if p == "" {
			continue
		}
		if len(p) > compatPrefixMaxLen {
			return nil, fmt.Errorf("prefix too long (max %d): %q", compatPrefixMaxLen, raw)
		}
		if !compatPrefixPattern.MatchString(p) {
			return nil, fmt.Errorf("invalid prefix %q: use lowercase letters, digits, . _ - only; must start with letter or digit", raw)
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	if len(out) > compatPrefixMaxCount {
		return nil, fmt.Errorf("at most %d unique prefixes allowed", compatPrefixMaxCount)
	}
	return out, nil
}
