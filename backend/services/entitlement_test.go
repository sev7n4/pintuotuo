package services

import (
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntitlementEnforcementStrict(t *testing.T) {
	t.Setenv("ENTITLEMENT_ENFORCEMENT", "")
	SetEntitlementEnforcementForTest("")
	assert.False(t, EntitlementEnforcementStrict())

	SetEntitlementEnforcementForTest("strict")
	assert.True(t, EntitlementEnforcementStrict())
	SetEntitlementEnforcementForTest("")
}

func TestResolveChosenPricingVersion_EmptyModel(t *testing.T) {
	v, _, ok, err := ResolveChosenPricingVersion(nil, 1, "openai", "")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, 0, v)
}

func TestResolveChosenPricingVersion_NoMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`WITH candidates AS`).
		WithArgs(1, "openai", "gpt-4o").
		WillReturnError(sql.ErrNoRows)

	v, _, ok, err := ResolveChosenPricingVersion(db, 1, "openai", "gpt-4o")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, 0, v)
}

func TestResolveChosenPricingVersion_Hit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ts := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery(`WITH candidates AS`).
		WithArgs(1, "openai", "gpt-4o").
		WillReturnRows(sqlmock.NewRows([]string{"pvid", "anchor_t"}).AddRow(3, ts))

	v, anchor, ok, err := ResolveChosenPricingVersion(db, 1, "openai", "gpt-4o")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 3, v)
	assert.True(t, anchor.Equal(ts.UTC()))
}
