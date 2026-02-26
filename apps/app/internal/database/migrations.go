package database

import (
	"fmt"
	"log"
	"time"
)

// migration describes a single schema migration.
type migration struct {
	name string
	sql  string
}

// migrations is the ordered list of all schema changes.
// Never remove or reorder — only append.
var allMigrations = []migration{
	{
		name: "001_create_departments",
		sql: `CREATE TABLE IF NOT EXISTS departments (
	id          TEXT PRIMARY KEY,
	name        TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL DEFAULT '',
	created_at  TEXT NOT NULL,
	updated_at  TEXT NOT NULL
);`,
	},
	{
		name: "002_users_add_department_id",
		sql:  `ALTER TABLE users ADD COLUMN department_id TEXT REFERENCES departments(id);`,
	},
	{
		name: "003_policies_add_department_id",
		sql:  `ALTER TABLE policies ADD COLUMN department_id TEXT REFERENCES departments(id);`,
	},
	{
		name: "004_policies_add_visibility_type",
		sql:  `ALTER TABLE policies ADD COLUMN visibility_type TEXT NOT NULL DEFAULT 'organization';`,
	},
	{
		name: "005_roles_rename_admin_to_superadmin",
		sql:  `UPDATE users SET role = 'SuperAdmin' WHERE role = 'Admin';`,
	},
}

// Migrate runs any pending schema migrations. Safe to call on every startup.
func (db *DB) Migrate() error {
	// Create the migrations tracking table.
	_, err := db.conn.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
	name       TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL
);`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range allMigrations {
		var existing string
		err := db.conn.QueryRow(
			`SELECT name FROM schema_migrations WHERE name = ?`, m.name,
		).Scan(&existing)
		if err == nil {
			// Already applied.
			continue
		}

		log.Printf("Applying migration: %s", m.name)
		if _, err := db.conn.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %s: %w", m.name, err)
		}
		if _, err := db.conn.Exec(
			`INSERT INTO schema_migrations (name, applied_at) VALUES (?, ?)`,
			m.name, time.Now().UTC().Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
		log.Printf("  Applied: %s", m.name)
	}
	return nil
}
