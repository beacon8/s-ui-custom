package migration

import "testing"

// A fresh install creates the clients table before the version row, so the
// 1.2 migration sees an empty table. Saving an empty slice makes GORM return
// "empty slice found" and aborts the whole install.
func TestMigrateClientsHandlesEmptyTable(t *testing.T) {
	db := openLegacyDB(t)
	defer closeDB(t, db)

	if err := migrateClients(db); err != nil {
		t.Fatalf("migrateClients on an empty clients table: %v", err)
	}
}
