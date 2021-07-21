package bolt

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	portainer "github.com/portainer/portainer/api"
)

// isFileExist is helper function to check for file existence
func isFileExist(path string) bool {
	matches, err := filepath.Glob(path)
	if err != nil {
		return false
	}
	return len(matches) > 0
}

func TestCreateBackupFolders(t *testing.T) {
	store, teardown := MustNewTestStore(false)
	defer teardown()

	backupPath := path.Join(store.path, backupDefaults.backupDir)

	if exists, _ := store.fileService.FileExists(backupPath); exists {
		t.Error("Expect backups folder to not exist")
	}

	store.createBackupFolders()
	if exists, _ := store.fileService.FileExists(backupPath); !exists {
		t.Error("Expect backups folder to exist")
	}
}

func TestStoreCreation(t *testing.T) {
	store, teardown := MustNewTestStore(true)
	defer teardown()

	if store == nil {
		t.Error("Expect to create a store")
	}

	if store.edition() != portainer.PortainerCE {
		t.Error("Expect to get CE Edition")
	}
}

func TestBackup(t *testing.T) {
	store, teardown := MustNewTestStore(true)
	defer teardown()

	t.Run("Backup should create default db backup", func(t *testing.T) {
		store.VersionService.StoreDBVersion(portainer.DBVersion)
		store.BackupWithOptions(nil)

		backupFileName := path.Join(store.path, "backups", "common", fmt.Sprintf("portainer.db.%03d.*", portainer.DBVersion))
		if !isFileExist(backupFileName) {
			t.Errorf("Expect backup file to be created %s", backupFileName)
		}
	})

	t.Run("BackupWithOption should create a name specific backup at common path", func(t *testing.T) {
		store.BackupWithOptions(&BackupOptions{
			BackupFileName: beforePortainerVersionUpgradeBackup,
			BackupDir:      store.commonBackupDir(),
		})
		backupFileName := path.Join(store.path, "backups", "common", beforePortainerVersionUpgradeBackup)
		if !isFileExist(backupFileName) {
			t.Errorf("Expect backup file to be created %s", backupFileName)
		}
	})
}

func TestRemoveWithOptions(t *testing.T) {
	store, teardown := MustNewTestStore(true)
	defer teardown()

	t.Run("successfully removes file if existent", func(t *testing.T) {
		store.createBackupFolders()
		options := &BackupOptions{
			BackupDir:      store.commonBackupDir(),
			BackupFileName: "test.txt",
		}

		filePath := path.Join(options.BackupDir, options.BackupFileName)
		f, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("file should be created; err=%s", err)
		}
		f.Close()

		err = store.RemoveWithOptions(options)
		if err != nil {
			t.Errorf("RemoveWithOptions should successfully remove file; err=%w", err)
		}

		if isFileExist(f.Name()) {
			t.Errorf("RemoveWithOptions should successfully remove file; file=%s", f.Name())
		}
	})

	t.Run("fails to removes file if non-existent", func(t *testing.T) {
		options := &BackupOptions{
			BackupDir:      store.commonBackupDir(),
			BackupFileName: "test.txt",
		}

		err := store.RemoveWithOptions(options)
		if err == nil {
			t.Error("RemoveWithOptions should fail for non-existent file")
		}
	})
}
