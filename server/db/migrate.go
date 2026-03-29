package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/pressly/goose/v3"
)

const embeddedMigrationsDir = "migrations"

//go:embed migrations/*.sql
var embeddedMigrations embed.FS

func applyMigrations(ctx context.Context, conn *sql.DB) error {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	goose.SetBaseFS(embeddedMigrations)
	defer goose.SetBaseFS(nil)

	if err := goose.UpContext(ctx, conn, embeddedMigrationsDir); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func RunMigrationCommand(ctx context.Context, rawPath string, args []string) error {
	if len(args) == 0 {
		return errors.New("migration command is required")
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	goose.SetBaseFS(nil)
	goose.SetSequential(true)

	dir, err := migrationsDirPath()
	if err != nil {
		return err
	}

	switch args[0] {
	case "create":
		if len(args) < 2 {
			return errors.New("create requires a migration name")
		}
		migrationType := "sql"
		if len(args) >= 3 {
			migrationType = args[2]
		}
		return goose.Create(nil, dir, args[1], migrationType)
	case "fix":
		return goose.Fix(dir)
	case "validate":
		return goose.RunContext(ctx, args[0], nil, dir, args[1:]...)
	}

	resolvedPath, err := resolveSQLitePath(rawPath)
	if err != nil {
		return err
	}

	conn, err := openSQLite(resolvedPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	return goose.RunContext(ctx, args[0], conn, dir, args[1:]...)
}

func migrationsDirPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("resolve migrations directory")
	}
	return filepath.Join(filepath.Dir(filename), embeddedMigrationsDir), nil
}
