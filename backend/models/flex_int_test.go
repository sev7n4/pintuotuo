package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlexInt_UnmarshalJSON(t *testing.T) {
	var s struct {
		N FlexInt `json:"n"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"n":9}`), &s))
	assert.Equal(t, FlexInt(9), s.N)

	require.NoError(t, json.Unmarshal([]byte(`{"n":9.9}`), &s))
	assert.Equal(t, FlexInt(9), s.N)

	require.NoError(t, json.Unmarshal([]byte(`{"n":"9.9"}`), &s))
	assert.Equal(t, FlexInt(9), s.N)

	require.NoError(t, json.Unmarshal([]byte(`{"n":-1}`), &s))
	assert.Equal(t, FlexInt(-1), s.N)

	require.NoError(t, json.Unmarshal([]byte(`{"n":null}`), &s))
	assert.Equal(t, FlexInt(0), s.N)
}
