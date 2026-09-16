// Package migrations embeds the reviewed SQL files in the API binary.
package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"fmt"
	"sort"
)

//go:embed *.up.sql
var files embed.FS

// Up applies each migration atomically, recording its checksum. A transaction-level
// advisory lock serializes deployments; existing SQL is never silently rewritten.
func Up(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(473246531)"); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS crm_schema_migrations (name text PRIMARY KEY, checksum text NOT NULL, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	entries, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		body, err := files.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		sum := fmt.Sprintf("%x", sha256.Sum256(body))
		var recorded string
		err = tx.QueryRowContext(ctx, "SELECT checksum FROM crm_schema_migrations WHERE name = $1", entry.Name()).Scan(&recorded)
		if err == nil {
			if recorded != sum {
				return fmt.Errorf("migration %s changed after application", entry.Name())
			}
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			return fmt.Errorf("migration %s: %w (existing databases require an explicit baseline review before adoption)", entry.Name(), err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO crm_schema_migrations(name, checksum) VALUES ($1, $2)", entry.Name(), sum); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func Ready(ctx context.Context, db *sql.DB) error {
	entries, err := files.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		body, err := files.ReadFile(entry.Name())
		if err != nil {
			return err
		}
		var checksum string
		if err := db.QueryRowContext(ctx, "SELECT checksum FROM crm_schema_migrations WHERE name = $1", entry.Name()).Scan(&checksum); err != nil {
			return err
		}
		if checksum != fmt.Sprintf("%x", sha256.Sum256(body)) {
			return fmt.Errorf("migration checksum mismatch: %s", entry.Name())
		}
	}
	return nil
}
