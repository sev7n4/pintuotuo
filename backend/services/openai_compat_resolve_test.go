package services

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveOpenAICompatModel_SlashSyntax(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	p, m := ResolveOpenAICompatModel(db, "openai/gpt-4o")
	assert.Equal(t, "openai", p)
	assert.Equal(t, "gpt-4o", m)

	p, m = ResolveOpenAICompatModel(db, "deepseek/deepseek-chat")
	assert.Equal(t, "deepseek", p)
	assert.Equal(t, "deepseek-chat", m)
}

func TestResolveOpenAICompatModel_DBPrefixes(t *testing.T) {
	cases := []struct {
		code     string
		prefixes []string
		model    string
		wantProv string
	}{
		{"deepseek", []string{"deepseek"}, "deepseek-chat", "deepseek"},
		{"anthropic", []string{"claude"}, "claude-3-sonnet-20240229", "anthropic"},
		{"google", []string{"gemini"}, "gemini-pro", "google"},
		{"zhipu", []string{"glm-", "chatglm", "cog-"}, "glm-4-air", "zhipu"},
	}
	for _, tc := range cases {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		rows := sqlmock.NewRows([]string{"code", "compat_prefixes"}).
			AddRow(tc.code, pq.StringArray(tc.prefixes))
		mock.ExpectQuery("SELECT code, compat_prefixes").WillReturnRows(rows)
		p, m := ResolveOpenAICompatModel(db, tc.model)
		assert.Equal(t, tc.wantProv, p, "model=%s", tc.model)
		assert.Equal(t, tc.model, m)
		require.NoError(t, mock.ExpectationsWereMet())
		db.Close()
	}
}

func TestResolveOpenAICompatModel_DefaultOpenAI(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	empty := sqlmock.NewRows([]string{"code", "compat_prefixes"})
	mock.ExpectQuery("SELECT code, compat_prefixes").WillReturnRows(empty)

	p, m := ResolveOpenAICompatModel(db, "gpt-3.5-turbo")
	assert.Equal(t, "openai", p)
	assert.Equal(t, "gpt-3.5-turbo", m)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestResolveOpenAICompatModel_NilDB(t *testing.T) {
	p, m := ResolveOpenAICompatModel((*sql.DB)(nil), "deepseek-chat")
	assert.Equal(t, "openai", p)
	assert.Equal(t, "deepseek-chat", m)
}
