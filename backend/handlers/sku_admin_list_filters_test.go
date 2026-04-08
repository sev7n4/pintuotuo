package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminSKUListFilters_Sellable(t *testing.T) {
	where, args := adminSKUListFilters("sellable", "", "", "", "", false, "", "")
	assert.Contains(t, where, "s.status = 'active'")
	assert.Contains(t, where, "sp.status = 'active'")
	assert.Empty(t, args)
}

func TestAdminSKUListFilters_AllWithSKUStatus(t *testing.T) {
	where, args := adminSKUListFilters("all", "inactive", "", "", "", false, "", "")
	assert.Contains(t, where, "s.status = $1")
	assert.Equal(t, []interface{}{"inactive"}, args)
}

func TestAdminSKUListFilters_Misaligned(t *testing.T) {
	where, _ := adminSKUListFilters("all", "", "", "", "", true, "", "")
	assert.Contains(t, where, "sp.status = 'inactive'")
}

func TestAdminSKUListFilters_SearchQ(t *testing.T) {
	where, args := adminSKUListFilters("all", "", "", "", "glm", false, "", "")
	assert.Contains(t, where, "ILIKE")
	assert.Len(t, args, 3)
}

func TestResolveAdminSKUListScope_DefaultSellable(t *testing.T) {
	scope, st := ResolveAdminSKUListScope("", "")
	assert.Equal(t, "sellable", scope)
	assert.Equal(t, "", st)
}

func TestResolveAdminSKUListScope_StatusImpliesAll(t *testing.T) {
	scope, st := ResolveAdminSKUListScope("", "active")
	assert.Equal(t, "all", scope)
	assert.Equal(t, "active", st)
}

func TestResolveAdminSKUListScope_StatusAllMeansFullList(t *testing.T) {
	scope, st := ResolveAdminSKUListScope("", "all")
	assert.Equal(t, "all", scope)
	assert.Equal(t, "", st)
}

func TestResolveAdminSKUListScope_ExplicitSellableIgnoresStatusForScope(t *testing.T) {
	scope, st := ResolveAdminSKUListScope("sellable", "inactive")
	assert.Equal(t, "sellable", scope)
	assert.Equal(t, "inactive", st)
}
