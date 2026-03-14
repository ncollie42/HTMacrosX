package database

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenMigratesUserTimezoneColumn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")

	legacyDB, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	_, err = legacyDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			hashed_password TEXT NOT NULL,
			target_calories REAL NOT NULL DEFAULT 1751.6,
			target_fat REAL NOT NULL DEFAULT 44.8,
			target_carb REAL NOT NULL DEFAULT 247.1,
			target_fiber REAL NOT NULL DEFAULT 32.0,
			target_protein REAL NOT NULL DEFAULT 90.0
		);
		INSERT INTO users (username, hashed_password, target_calories, target_fat, target_carb, target_fiber, target_protein)
		VALUES ('legacy-user', 'hash', 1751.6, 44.8, 247.1, 32.0, 90.0);
	`)
	if err != nil {
		t.Fatalf("seed legacy schema: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("legacyDB.Close: %v", err)
	}

	Open(path)
	t.Cleanup(func() {
		if sqlDB != nil {
			_ = sqlDB.Close()
			sqlDB = nil
		}
	})

	var timezone string
	if err := sqlDB.QueryRow(`SELECT timezone FROM users WHERE username = ?`, "legacy-user").Scan(&timezone); err != nil {
		t.Fatalf("query migrated timezone: %v", err)
	}
	if timezone != "" {
		t.Fatalf("timezone = %q, want empty", timezone)
	}
}

func TestUserTimezoneCRUD(t *testing.T) {
	setupTestDB(t)

	userID := createTestUser(t, "tz-user")
	if err := UpdateUserTimezone(userID, "America/New_York"); err != nil {
		t.Fatalf("UpdateUserTimezone valid: %v", err)
	}
	got, err := GetUserTimezone(userID)
	if err != nil {
		t.Fatalf("GetUserTimezone: %v", err)
	}
	if got != "America/New_York" {
		t.Fatalf("timezone = %q, want %q", got, "America/New_York")
	}

	if err := UpdateUserTimezone(userID, "PST"); err == nil {
		t.Fatalf("UpdateUserTimezone invalid zone unexpectedly succeeded")
	}
	got, err = GetUserTimezone(userID)
	if err != nil {
		t.Fatalf("GetUserTimezone after invalid update: %v", err)
	}
	if got != "America/New_York" {
		t.Fatalf("timezone after invalid update = %q, want unchanged", got)
	}
}
