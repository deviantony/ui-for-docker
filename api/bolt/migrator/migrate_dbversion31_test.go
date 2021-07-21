package migrator

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/portainer/portainer/api/bolt/internal"
	"github.com/portainer/portainer/api/bolt/settings"
)

func TestMigrateAdminGroupSearchSettings(t *testing.T) {
	testingDBStorePath, _ = os.Getwd()
	testingDBFileName = "portainer-ee-mig-32.db"
	databasePath := path.Join(testingDBStorePath, testingDBFileName)
	dbConn, err := bolt.Open(databasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Errorf("failed to init testing DB connection: %v", err)
	}
	defer dbConn.Close()
	defer os.Remove(testingDBFileName)
	internalDBConn := &internal.DbConnection{
		DB: dbConn,
	}
	settingsService, err := settings.NewService(internalDBConn)
	if err != nil {
		t.Errorf("failed to init testing settings service: %v", err)
	}
	dummySettingsObj := map[string]interface{}{
		"LogoURL": "example.com",
	}
	if err := internal.UpdateObject(internalDBConn, "settings", []byte("SETTINGS"), dummySettingsObj); err != nil {
		t.Errorf("failed to create mock settings: %v", err)
	}
	m := &Migrator{
		db:              dbConn,
		settingsService: settingsService,
	}
	if err := m.migrateAdminGroupSearchSettings(); err != nil {
		t.Errorf("failed to update settings: %v", err)
	}
	updatedSettings, err := m.settingsService.Settings()
	if err != nil {
		t.Errorf("failed to retrieve the updated settings: %v", err)
	}
	if updatedSettings.LDAPSettings.AdminGroupSearchSettings == nil {
		t.Error("LDAP AdminGroupSearchSettings should not be nil")
	}

}
