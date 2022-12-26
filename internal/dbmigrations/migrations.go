package dbmigrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Apply(db *sql.DB) (err error) {
	goose.SetLogger(logger{})
	goose.SetBaseFS(migrations)

	if err = goose.SetDialect("sqlite3"); err != nil {
		err = fmt.Errorf("goose setting dialect: %w", err)
		return
	}

	var version int64
	version, err = goose.EnsureDBVersion(db)
	if err != nil {
		err = fmt.Errorf("goose ensuring db version table: %w", err)
		return
	}
	log.Debug().Int64("version", version).Msgf("goose current version")

	if err = goose.Up(db, "migrations"); err != nil {
		err = fmt.Errorf("goose up: %w", err)
		return
	}

	return nil
}

// Rollback a single version of the database schema.
func Rollback(db *sql.DB) (err error) {
	goose.SetLogger(logger{})
	goose.SetBaseFS(migrations)

	if err = goose.SetDialect("sqlite3"); err != nil {
		err = fmt.Errorf("goose setting dialect: %w", err)
		return
	}

	if err = goose.Down(db, "migrations"); err != nil {
		err = fmt.Errorf("goose down: %w", err)
		return
	}

	return nil
}
