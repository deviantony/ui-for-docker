package bolt

import (
	"io"
	"log"
	"path"
	"time"

	"github.com/boltdb/bolt"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/customtemplate"
	"github.com/portainer/portainer/api/bolt/dockerhub"
	"github.com/portainer/portainer/api/bolt/edgegroup"
	"github.com/portainer/portainer/api/bolt/edgejob"
	"github.com/portainer/portainer/api/bolt/edgestack"
	"github.com/portainer/portainer/api/bolt/endpoint"
	"github.com/portainer/portainer/api/bolt/endpointgroup"
	"github.com/portainer/portainer/api/bolt/endpointrelation"
	"github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/bolt/extension"
	"github.com/portainer/portainer/api/bolt/internal"
	"github.com/portainer/portainer/api/bolt/migrator"
	"github.com/portainer/portainer/api/bolt/registry"
	"github.com/portainer/portainer/api/bolt/resourcecontrol"
	"github.com/portainer/portainer/api/bolt/role"
	"github.com/portainer/portainer/api/bolt/schedule"
	"github.com/portainer/portainer/api/bolt/settings"
	"github.com/portainer/portainer/api/bolt/ssl"
	"github.com/portainer/portainer/api/bolt/stack"
	"github.com/portainer/portainer/api/bolt/tag"
	"github.com/portainer/portainer/api/bolt/team"
	"github.com/portainer/portainer/api/bolt/teammembership"
	"github.com/portainer/portainer/api/bolt/tunnelserver"
	"github.com/portainer/portainer/api/bolt/user"
	"github.com/portainer/portainer/api/bolt/version"
	"github.com/portainer/portainer/api/bolt/webhook"
	"github.com/portainer/portainer/api/internal/authorization"
)

const (
	databaseFileName = "portainer.db"
)

// Store defines the implementation of portainer.DataStore using
// BoltDB as the storage system.
type Store struct {
	path                    string
	connection              *internal.DbConnection
	isNew                   bool
	fileService             portainer.FileService
	CustomTemplateService   *customtemplate.Service
	DockerHubService        *dockerhub.Service
	EdgeGroupService        *edgegroup.Service
	EdgeJobService          *edgejob.Service
	EdgeStackService        *edgestack.Service
	EndpointGroupService    *endpointgroup.Service
	EndpointService         *endpoint.Service
	EndpointRelationService *endpointrelation.Service
	ExtensionService        *extension.Service
	RegistryService         *registry.Service
	ResourceControlService  *resourcecontrol.Service
	RoleService             *role.Service
	ScheduleService         *schedule.Service
	SettingsService         *settings.Service
	SSLSettingsService      *ssl.Service
	StackService            *stack.Service
	TagService              *tag.Service
	TeamMembershipService   *teammembership.Service
	TeamService             *team.Service
	TunnelServerService     *tunnelserver.Service
	UserService             *user.Service
	VersionService          *version.Service
	WebhookService          *webhook.Service
}

func (store *Store) edition() portainer.SoftwareEdition {
	edition, err := store.VersionService.Edition()
	if err == errors.ErrObjectNotFound {
		edition = portainer.PortainerCE
	}
	return edition
}

// NewStore initializes a new Store and the associated services
func NewStore(storePath string, fileService portainer.FileService) (*Store, error) {
	store := &Store{
		path:        storePath,
		fileService: fileService,
		isNew:       true,
		connection:  &internal.DbConnection{},
	}

	databasePath := path.Join(storePath, databaseFileName)
	databaseFileExists, err := fileService.FileExists(databasePath)
	if err != nil {
		return nil, err
	}

	if databaseFileExists {
		store.isNew = false
	}

	return store, nil
}

// Open opens and initializes the BoltDB database.
func (store *Store) Open() error {
	databasePath := path.Join(store.path, databaseFileName)
	db, err := bolt.Open(databasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	store.connection.DB = db

	return store.initServices()
}

// Close closes the BoltDB database.
// Safe to being called multiple times.
func (store *Store) Close() error {
	if store.connection.DB != nil {
		return store.connection.Close()
	}
	return nil
}

// IsNew returns true if the database was just created and false if it is re-using
// existing data.
func (store *Store) IsNew() bool {
	return store.isNew
}

// CheckCurrentEdition checks if current edition is community edition
func (store *Store) CheckCurrentEdition() error {
	if store.edition() != portainer.PortainerCE {
		return errors.ErrWrongDBEdition
	}
	return nil
}

// MigrateData automatically migrate the data based on the DBVersion.
// This process is only triggered on an existing database, not if the database was just created.
// if force is true, then migrate regardless.
func (store *Store) MigrateData(force bool) error {
	if store.isNew && !force {
		return store.VersionService.StoreDBVersion(portainer.DBVersion)
	}

	version, err := store.VersionService.DBVersion()
	if err == errors.ErrObjectNotFound {
		version = 0
	} else if err != nil {
		return err
	}

	if version < portainer.DBVersion {
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
		migrator := migrator.NewMigrator(migratorParams)

		log.Printf("Migrating database from version %v to %v.\n", version, portainer.DBVersion)
		err = migrator.Migrate()
		if err != nil {
			log.Printf("An error occurred during database migration: %s\n", err)
			return err
		}
	}

	return nil
}

// BackupTo backs up db to a provided writer.
// It does hot backup and doesn't block other database reads and writes
func (store *Store) BackupTo(w io.Writer) error {
	return store.connection.View(func(tx *bolt.Tx) error {
		_, err := tx.WriteTo(w)
		return err
	})
}
