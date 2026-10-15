package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

func Migrate(ctx context.Context, db *sql.DB, args []string) error {

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	migrationsDir := filepath.Join(dir, "internal", "database", "migrations")

	if len(args) == 0 {
		return fmt.Errorf("%s <up|create> [name]", args[0])
	}

	switch args[1] {
	case "up":
		return goose.Up(db, migrationsDir)
	case "create":
		if len(args) < 3 {
			return fmt.Errorf("usage: %s create <name>", args[0])
		}

		return goose.Create(db, migrationsDir, args[2], "sql")
	}

	return fmt.Errorf("'%s' is not a valid command", args[1])
}
