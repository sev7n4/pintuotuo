package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashUserAPIKey_Deterministic(t *testing.T) {
	k := "ptd_" + "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	h1 := HashUserAPIKey(k)
	h2 := HashUserAPIKey(k)
	assert.Equal(t, h1, h2)
	assert.Len(t, h1, 64)
}
