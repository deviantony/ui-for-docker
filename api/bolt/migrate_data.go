package bolt

import (
	"fmt"

	"github.com/portainer/portainer/api/cli"

	werrors "github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	plog "github.com/portainer/portainer/api/bolt/log"
	"github.com/portainer/portainer/api/bolt/migrator"
	"github.com/portainer/portainer/api/internal/authorization"
)

const beforePortainerVersionUpgradeBackup = "portainer.db.bak"

var migrateLog = plog.NewScopedLog("bolt, migrate")

// FailSafeMigrate backup and restore DB if migration fail
func (store *Store) FailSafeMigrate(migrator *migrator.Migrator) error {
	defer func() {
		if err := recover(); err != nil {
			migrateLog.Info(fmt.Sprintf("Error during migration, recovering [%v]", err))
			store.Rollback(true)
		}
	}()
	return migrator.MigrateCE()
}

// MigrateData automatically migrate the data based on the DBVersion.
// This process is only triggered on an existing database, not if the database was just created.
// if force is true, then migrate regardless.
func (store *Store) MigrateData(force bool) error {
	if store.isNew && !force {
		return store.VersionService.StoreDBVersion(portainer.DBVersion)
	}

	migrator, err := store.newMigrator()
	if err != nil {
		return err
	}

	// backup db file before upgrading DB to support rollback
	isUpdating, err := store.VersionService.IsUpdating()
	if err != nil && err != errors.ErrObjectNotFound {
		return err
	}

	if !isUpdating && migrator.Version() != portainer.DBVersion {
		err = store.backupVersion(migrator)
		if err != nil {
			return werrors.Wrapf(err, "failed to backup database")
		}
	}

	if migrator.Version() < portainer.DBVersion {
		migrator, err := store.newMigrator()
		if err != nil {
			return err
		}

		migrateLog.Info(fmt.Sprintf("Migrating database from version %v to %v.\n", migrator.Version(), portainer.DBVersion))
		err = store.FailSafeMigrate(migrator)
		if err != nil {
			migrateLog.Error("An error occurred during database migration", err)
			return err
		}
	}

	return nil
}

func (store *Store) newMigrator() (*migrator.Migrator, error) {
	version, err := store.version()
	if err != nil {
		return nil, err
	}

	migratorParams := &migrator.Parameters{
		DB:                      store.connection.DB,
		DatabaseVersion:         version,
		EndpointGroupService:    store.EndpointGroupService,
		EndpointService:         store.EndpointService,
		EndpointRelationService: store.EndpointRelationService,
		ExtensionService:        store.ExtensionService,
		RegistryService:         store.RegistryService,
		ResourceControlService:  store.ResourceControlService,
		RoleService:             store.RoleService,
		ScheduleService:         store.ScheduleService,
		SettingsService:         store.SettingsService,
		StackService:            store.StackService,
		TagService:              store.TagService,
		TeamMembershipService:   store.TeamMembershipService,
		UserService:             store.UserService,
		VersionService:          store.VersionService,
		FileService:             store.fileService,
		DockerhubService:        store.DockerHubService,
		AuthorizationService:    authorization.NewService(store),
	}
	return migrator.NewMigrator(migratorParams), nil
}

// getBackupRestoreOptions returns options to store db at common backup dir location; used by:
// - db backup prior to version upgrade
// - db rollback
func getBackupRestoreOptions(store *Store) *BackupOptions {
	return &BackupOptions{
		BackupDir:      store.commonBackupDir(),
		BackupFileName: beforePortainerVersionUpgradeBackup,
	}
}

// backupVersion will backup the database or panic if any errors occur
func (store *Store) backupVersion(migrator *migrator.Migrator) error {
	migrateLog.Info("Backing up database prior to version upgrade...")

	options := getBackupRestoreOptions(store)

	_, err := store.BackupWithOptions(options)
	if err != nil {
		migrateLog.Error("An error occurred during database backup", err)
		store.RemoveWithOptions(options)
		return err
	}

	return nil
}

// Rollback to a pre-upgrade backup copy/snapshot of portainer.db
func (store *Store) Rollback(force bool) error {

	if !force {
		confirmed, err := cli.Confirm("Are you sure you want to rollback your database to the previous backup?")
		if err != nil || !confirmed {
			return err
		}
	}

	options := getBackupRestoreOptions(store)

	err := store.RestoreWithOptions(options)
	if err != nil {
		return err
	}

	return store.Close()
}
