package bolt

import (
	"fmt"
	"os"
	"path"
	"time"

	plog "github.com/portainer/portainer/api/bolt/log"
)

var backupDefaults = struct {
	backupDir        string
	commonDir        string
	databaseFileName string
}{
	"backups",
	"common",
	databaseFileName,
}

var backupLog = plog.NewScopedLog("bolt, backup")

//
// Backup Helpers
//

// createBackupFolders create initial folders for backups
func (store *Store) createBackupFolders() {
	// create common dir
	commonDir := store.commonBackupDir()
	if exists, _ := store.fileService.FileExists(commonDir); !exists {
		if err := os.MkdirAll(commonDir, 0700); err != nil {
			backupLog.Error("Error while creating common backup folder", err)
		}
	}
}

func (store *Store) databasePath() string {
	return path.Join(store.path, databaseFileName)
}

func (store *Store) commonBackupDir() string {
	return path.Join(store.path, backupDefaults.backupDir, backupDefaults.commonDir)
}

func (store *Store) copyDBFile(from string, to string) error {
	backupLog.Info(fmt.Sprintf("Copying db file from %s to %s", from, to))
	err := store.fileService.Copy(from, to, true)
	if err != nil {
		backupLog.Error("Failed", err)
	}
	return err
}

// BackupOptions provide a helper to inject backup options
type BackupOptions struct {
	Version        int
	BackupDir      string
	BackupFileName string
	BackupPath     string
}

func (store *Store) setupOptions(options *BackupOptions) *BackupOptions {
	if options == nil {
		options = &BackupOptions{}
	}
	if options.Version == 0 {
		options.Version, _ = store.version()
	}
	if options.BackupDir == "" {
		options.BackupDir = store.commonBackupDir()
	}
	if options.BackupFileName == "" {
		options.BackupFileName = fmt.Sprintf("%s.%s.%s", backupDefaults.databaseFileName, fmt.Sprintf("%03d", options.Version), time.Now().Format("20060102150405"))
	}
	if options.BackupPath == "" {
		options.BackupPath = path.Join(options.BackupDir, options.BackupFileName)
	}
	return options
}

// BackupWithOptions backup current database with options
func (store *Store) BackupWithOptions(options *BackupOptions) (string, error) {
	backupLog.Info("creating db backup")
	store.createBackupFolders()

	options = store.setupOptions(options)

	return options.BackupPath, store.copyDBFile(store.databasePath(), options.BackupPath)
}

// RestoreWithOptions previously saved backup for the current Edition  with options
// Restore strategies:
// - default: restore latest from current edition
// - restore a specific
func (store *Store) RestoreWithOptions(options *BackupOptions) error {
	options = store.setupOptions(options)

	// Check if backup file exist before restoring
	_, err := os.Stat(options.BackupPath)
	if os.IsNotExist(err) {
		backupLog.Error(fmt.Sprintf("Backup file to restore does not exist %s", options.BackupPath), err)
		return err
	}

	err = store.Close()
	if err != nil {
		backupLog.Error("Error while closing store before restore", err)
		return err
	}

	backupLog.Info("Restoring db backup")
	err = store.copyDBFile(options.BackupPath, store.databasePath())
	if err != nil {
		return err
	}

	return store.Open()
}

// RemoveWithOptions removes backup database based on supplied options
func (store *Store) RemoveWithOptions(options *BackupOptions) error {
	backupLog.Info("Removing db backup")

	options = store.setupOptions(options)
	_, err := os.Stat(options.BackupPath)

	if os.IsNotExist(err) {
		backupLog.Error(fmt.Sprintf("Backup file to remove does not exist %s", options.BackupPath), err)
		return err
	}

	backupLog.Info(fmt.Sprintf("Removing db file at %s", options.BackupPath))
	err = os.Remove(options.BackupPath)
	if err != nil {
		backupLog.Error("Failed", err)
		return err
	}

	return nil
}
