package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migrate runs all *.sql files in the given directory in lexicographic order.
// It is intentionally simple (no rollback, no versioning table) — use a proper
// migration tool like golang-migrate/migrate in production.
//
// Files must be named NNN_description.sql so they sort correctly, e.g.:
//   001_users.sql
//   002_assets.sql
func Migrate(ctx context.Context, pool *Pool, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory %q: %w", migrationsDir, err)
	}

	// Collect .sql files and sort them.
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(migrationsDir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("reading migration %q: %w", f, err)
		}

		if _, err = pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("applying migration %q: %w", f, err)
		}

		fmt.Printf("✓ applied migration: %s\n", filepath.Base(f))
	}

	return nil
}
