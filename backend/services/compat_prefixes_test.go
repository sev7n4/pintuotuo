package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeCompatPrefixes(t *testing.T) {
	t.Parallel()
	out, err := NormalizeCompatPrefixes([]string{"  DeepSeek ", "deepseek", "GLM-"})
	require.NoError(t, err)
	require.Equal(t, []string{"deepseek", "glm-"}, out)

	_, err = NormalizeCompatPrefixes([]string{"bad prefix"})
	require.Error(t, err)

	_, err = NormalizeCompatPrefixes([]string{"-x"})
	require.Error(t, err)
}
