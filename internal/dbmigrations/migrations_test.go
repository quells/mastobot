package dbmigrations

import (
	"database/sql"
	"io/fs"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

func TestApply(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "must create database connection")
	require.NoError(t, Apply(db), "must run database migrations")
}

func TestRollback(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err, "must create database connection")
	require.NoError(t, Apply(db), "must run database migrations")

	var entries []fs.DirEntry
	entries, err = migrations.ReadDir("migrations")
	require.NoError(t, err, "must list files")

	for _, entry := range entries {
		require.NoError(t, Rollback(db), "must apply rollback %s", entry.Name())
	}

	require.Error(t, Rollback(db), "expected error after all rollbacks have been applied")
}
