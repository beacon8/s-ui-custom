package database

import (
	"testing"

	"github.com/alireza0/s-ui/cmd/migration"
	"github.com/alireza0/s-ui/config"
)

// The installer runs `sui admin` first, which creates an empty database, and
// then `sui migrate`. A fresh database has no version row and no clients/TLS,
// so the 1.2 migration must tolerate empty tables and the later migrations
// must tolerate columns that only exist on older schemas.
func TestMigrateFreshInstall(t *testing.T) {
	t.Setenv("SUI_DB_FOLDER", t.TempDir())

	if err := InitDB(config.GetDBPath()); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	sqlDB, err := GetDB().DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}

	if err := migration.MigrateDb(); err != nil {
		t.Fatalf("MigrateDb on a fresh database: %v", err)
	}

	if err := OpenDB(config.GetDBPath()); err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer func() {
		sqlDB, err := GetDB().DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}()

	var version string
	if err := GetDB().Raw("SELECT value FROM settings WHERE key = ?", "version").Scan(&version).Error; err != nil {
		t.Fatal(err)
	}
	if version != config.GetVersion() {
		t.Fatalf("version = %q, want %q", version, config.GetVersion())
	}
}
