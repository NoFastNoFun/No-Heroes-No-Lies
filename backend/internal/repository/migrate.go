package repository

import (
	"context"
	"embed"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(ctx context.Context, pool *pgxpool.Pool, fs embed.FS, dir string) error {
	if err := ensureMigrationsTable(ctx, pool); err != nil {
		return err
	}
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		applied, err := isMigrationApplied(ctx, pool, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		body, err := fs.ReadFile(path.Join(dir, name))
		if err != nil {
			return err
		}
		if _, err := pool.Exec(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", name); err != nil {
			return err
		}
	}
	return nil
}

func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY
		)
	`)
	return err
}

func isMigrationApplied(ctx context.Context, pool *pgxpool.Pool, version string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
	return exists, err
}
