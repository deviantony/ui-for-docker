package bolt

import (
	"fmt"
	"log"
	"strings"
	"testing"

	portainer "github.com/portainer/portainer/api"
)

// testVersion is a helper which tests current store version against wanted version
func testVersion(store *Store, versionWant int, t *testing.T) {
	if v, _ := store.version(); v != versionWant {
		t.Errorf("Expect store version to be %d but was %d", versionWant, v)
	}
}

func TestMigrateData(t *testing.T) {
	t.Run("MigrateData for New Store", func(t *testing.T) {
		store, teardown := MustNewTestStore(false)
		defer teardown()
		store.MigrateData(false)
		testVersion(store, portainer.DBVersion, t)
		store.Close()
	})

	tests := []struct {
		version         int
		expectedVersion int
	}{
		{version: 2, expectedVersion: portainer.DBVersion},
		{version: 21, expectedVersion: portainer.DBVersion},
	}
	for _, tc := range tests {
		store, teardown := MustNewTestStore(true)
		defer teardown()

		// Setup data
		store.VersionService.StoreDBVersion(tc.version)

		// Required roles by migrations 22.2
		store.RoleService.CreateRole(&portainer.Role{ID: 1})
		store.RoleService.CreateRole(&portainer.Role{ID: 2})
		store.RoleService.CreateRole(&portainer.Role{ID: 3})
		store.RoleService.CreateRole(&portainer.Role{ID: 4})

		t.Run(fmt.Sprintf("MigrateData for version %d", tc.version), func(t *testing.T) {
			store.MigrateData(true)
			testVersion(store, tc.expectedVersion, t)
		})

		t.Run(fmt.Sprintf("Restoring DB after migrateData for version %d", tc.version), func(t *testing.T) {
			store.Rollback(true)
			store.Open()
			testVersion(store, tc.version, t)
		})
	}

	t.Run("Error in MigrateData should restore backup before MigrateData", func(t *testing.T) {
		store, teardown := MustNewTestStore(false)
		defer teardown()

		version := 2
		store.VersionService.StoreDBVersion(version)

		store.MigrateData(true)

		testVersion(store, version, t)
	})

	t.Run("MigrateData should create backup file upon update", func(t *testing.T) {
		store, teardown := MustNewTestStore(false)
		defer teardown()
		store.VersionService.StoreDBVersion(0)

		store.MigrateData(true)

		options := store.setupOptions(getBackupRestoreOptions(store))

		if !isFileExist(options.BackupPath) {
			t.Errorf("Backup file should exist; file=%s", options.BackupPath)
		}
	})

	t.Run("MigrateData should fail to create backup if database file is set to updating", func(t *testing.T) {
		store, teardown := MustNewTestStore(false)
		defer teardown()

		store.VersionService.StoreIsUpdating(true)

		store.MigrateData(true)

		options := store.setupOptions(getBackupRestoreOptions(store))

		if isFileExist(options.BackupPath) {
			t.Errorf("Backup file should not exist for dirty database; file=%s", options.BackupPath)
		}
	})

	t.Run("MigrateData should not create backup on startup if portainer version matches db", func(t *testing.T) {
		store, teardown := MustNewTestStore(false)
		defer teardown()

		store.MigrateData(true)

		options := store.setupOptions(getBackupRestoreOptions(store))

		if isFileExist(options.BackupPath) {
			t.Errorf("Backup file should not exist for dirty database; file=%s", options.BackupPath)
		}
	})

}

func Test_getBackupRestoreOptions(t *testing.T) {
	store, teardown := MustNewTestStore(false)
	defer teardown()

	options := getBackupRestoreOptions(store)

	wantDir := store.commonBackupDir()
	if !strings.HasSuffix(options.BackupDir, wantDir) {
		log.Fatalf("incorrect backup dir; got=%s, want=%s", options.BackupDir, wantDir)
	}

	wantFilename := "portainer.db.bak"
	if options.BackupFileName != wantFilename {
		log.Fatalf("incorrect backup file; got=%s, want=%s", options.BackupFileName, wantFilename)
	}
}

func TestRollback(t *testing.T) {
	t.Run("Rollback should restore upgrade after backup", func(t *testing.T) {
		version := 21
		store, teardown := MustNewTestStore(false)
		defer teardown()
		store.VersionService.StoreDBVersion(version)

		_, err := store.BackupWithOptions(getBackupRestoreOptions(store))
		if err != nil {
			log.Fatal(err)
		}

		// Change the current edition
		err = store.VersionService.StoreDBVersion(version + 10)
		if err != nil {
			log.Fatal(err)
		}

		err = store.Rollback(true)
		if err != nil {
			t.Logf("Rollback failed: %s", err)
			t.Fail()
			return
		}

		store.Open()
		testVersion(store, version, t)
	})
}
