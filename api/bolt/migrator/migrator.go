package migrator

import (
	"github.com/boltdb/bolt"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/endpoint"
	"github.com/portainer/portainer/api/bolt/endpointgroup"
	"github.com/portainer/portainer/api/bolt/endpointrelation"
	"github.com/portainer/portainer/api/bolt/extension"
	"github.com/portainer/portainer/api/bolt/registry"
	"github.com/portainer/portainer/api/bolt/resourcecontrol"
	"github.com/portainer/portainer/api/bolt/role"
	"github.com/portainer/portainer/api/bolt/schedule"
	"github.com/portainer/portainer/api/bolt/settings"
	"github.com/portainer/portainer/api/bolt/stack"
	"github.com/portainer/portainer/api/bolt/tag"
	"github.com/portainer/portainer/api/bolt/teammembership"
	"github.com/portainer/portainer/api/bolt/user"
	"github.com/portainer/portainer/api/bolt/version"
	"github.com/portainer/portainer/api/internal/authorization"
)

type (
	// Migrator defines a service to migrate data after a Portainer version update.
	Migrator struct {
		db                      *bolt.DB
		currentDBVersion        int
		endpointGroupService    *endpointgroup.Service
		endpointService         *endpoint.Service
		endpointRelationService *endpointrelation.Service
		extensionService        *extension.Service
		registryService         *registry.Service
		resourceControlService  *resourcecontrol.Service
		roleService             *role.Service
		scheduleService         *schedule.Service
		settingsService         *settings.Service
		stackService            *stack.Service
		tagService              *tag.Service
		teamMembershipService   *teammembership.Service
		userService             *user.Service
		versionService          *version.Service
		fileService             portainer.FileService
		authorizationService    *authorization.Service
	}

	// Parameters represents the required parameters to create a new Migrator instance.
	Parameters struct {
		DB                      *bolt.DB
		DatabaseVersion         int
		EndpointGroupService    *endpointgroup.Service
		EndpointService         *endpoint.Service
		EndpointRelationService *endpointrelation.Service
		ExtensionService        *extension.Service
		RegistryService         *registry.Service
		ResourceControlService  *resourcecontrol.Service
		RoleService             *role.Service
		ScheduleService         *schedule.Service
		SettingsService         *settings.Service
		StackService            *stack.Service
		TagService              *tag.Service
		TeamMembershipService   *teammembership.Service
		UserService             *user.Service
		VersionService          *version.Service
		FileService             portainer.FileService
		AuthorizationService    *authorization.Service
	}
)

// NewMigrator creates a new Migrator.
func NewMigrator(parameters *Parameters) *Migrator {
	return &Migrator{
		db:                      parameters.DB,
		currentDBVersion:        parameters.DatabaseVersion,
		endpointGroupService:    parameters.EndpointGroupService,
		endpointService:         parameters.EndpointService,
		endpointRelationService: parameters.EndpointRelationService,
		extensionService:        parameters.ExtensionService,
		registryService:         parameters.RegistryService,
		resourceControlService:  parameters.ResourceControlService,
		roleService:             parameters.RoleService,
		scheduleService:         parameters.ScheduleService,
		settingsService:         parameters.SettingsService,
		tagService:              parameters.TagService,
		teamMembershipService:   parameters.TeamMembershipService,
		stackService:            parameters.StackService,
		userService:             parameters.UserService,
		versionService:          parameters.VersionService,
		fileService:             parameters.FileService,
		authorizationService:    parameters.AuthorizationService,
	}
}

// Migrate checks the database version and migrate the existing data to the most recent data model.
func (m *Migrator) Migrate() error {
	// Portainer < 1.12
	if m.currentDBVersion < 1 {
		err := m.updateAdminUserToDBVersion1()
		if err != nil {
			return err
		}
	}

	// Portainer 1.12.x
	if m.currentDBVersion < 2 {
		err := m.updateResourceControlsToDBVersion2()
		if err != nil {
			return err
		}
		err = m.updateEndpointsToDBVersion2()
		if err != nil {
			return err
		}
	}

	// Portainer 1.13.x
	if m.currentDBVersion < 3 {
		err := m.updateSettingsToDBVersion3()
		if err != nil {
			return err
		}
	}

	// Portainer 1.14.0
	if m.currentDBVersion < 4 {
		err := m.updateEndpointsToDBVersion4()
		if err != nil {
			return err
		}
	}

	// https://github.com/portainer/portainer/issues/1235
	if m.currentDBVersion < 5 {
		err := m.updateSettingsToVersion5()
		if err != nil {
			return err
		}
	}

	// https://github.com/portainer/portainer/issues/1236
	if m.currentDBVersion < 6 {
		err := m.updateSettingsToVersion6()
		if err != nil {
			return err
		}
	}

	// https://github.com/portainer/portainer/issues/1449
	if m.currentDBVersion < 7 {
		err := m.updateSettingsToVersion7()
		if err != nil {
			return err
		}
	}

	if m.currentDBVersion < 8 {
		err := m.updateEndpointsToVersion8()
		if err != nil {
			return err
		}
	}

	// https: //github.com/portainer/portainer/issues/1396
	if m.currentDBVersion < 9 {
		err := m.updateEndpointsToVersion9()
		if err != nil {
			return err
		}
	}

	// https://github.com/portainer/portainer/issues/461
	if m.currentDBVersion < 10 {
		err := m.updateEndpointsToVersion10()
		if err != nil {
			return err
		}
	}

	// https://github.com/portainer/portainer/issues/1906
	if m.currentDBVersion < 11 {
		err := m.updateEndpointsToVersion11()
		if err != nil {
			return err
		}
	}

	// Portainer 1.18.0
	if m.currentDBVersion < 12 {
		err := m.updateEndpointsToVersion12()
		if err != nil {
			return err
		}

		err = m.updateEndpointGroupsToVersion12()
		if err != nil {
			return err
		}

		err = m.updateStacksToVersion12()
		if err != nil {
			return err
		}
	}

	// Portainer 1.19.0
	if m.currentDBVersion < 13 {
		err := m.updateSettingsToVersion13()
		if err != nil {
			return err
		}
	}

	// Portainer 1.19.2
	if m.currentDBVersion < 14 {
		err := m.updateResourceControlsToDBVersion14()
		if err != nil {
			return err
		}
	}

	// Portainer 1.20.0
	if m.currentDBVersion < 15 {
		err := m.updateSettingsToDBVersion15()
		if err != nil {
			return err
		}

		err = m.updateTemplatesToVersion15()
		if err != nil {
			return err
		}
	}

	if m.currentDBVersion < 16 {
		err := m.updateSettingsToDBVersion16()
		if err != nil {
			return err
		}
	}

	// Portainer 1.20.1
	if m.currentDBVersion < 17 {
		err := m.updateExtensionsToDBVersion17()
		if err != nil {
			return err
		}
	}

	// Portainer 1.21.0
	if m.currentDBVersion < 18 {
		err := m.updateUsersToDBVersion18()
		if err != nil {
			return err
		}

		err = m.updateEndpointsToDBVersion18()
		if err != nil {
			return err
		}

		err = m.updateEndpointGroupsToDBVersion18()
		if err != nil {
			return err
		}

		err = m.updateRegistriesToDBVersion18()
		if err != nil {
			return err
		}
	}

	// Portainer 1.22.0
	if m.currentDBVersion < 19 {
		err := m.updateSettingsToDBVersion19()
		if err != nil {
			return err
		}
	}

	// Portainer 1.22.1
	if m.currentDBVersion < 20 {
		err := m.updateUsersToDBVersion20()
		if err != nil {
			return err
		}

		err = m.updateSettingsToDBVersion20()
		if err != nil {
			return err
		}

		err = m.updateSchedulesToDBVersion20()
		if err != nil {
			return err
		}
	}

	// Portainer 1.23.0
	// DBVersion 21 is missing as it was shipped as via hotfix 1.22.2
	if m.currentDBVersion < 22 {
		err := m.updateResourceControlsToDBVersion22()
		if err != nil {
			return err
		}

		err = m.updateUsersAndRolesToDBVersion22()
		if err != nil {
			return err
		}
	}

	// Portainer 1.24.0
	if m.currentDBVersion < 23 {
		err := m.updateTagsToDBVersion23()
		if err != nil {
			return err
		}

		err = m.updateEndpointsAndEndpointGroupsToDBVersion23()
		if err != nil {
			return err
		}
	}

	// Portainer 1.24.1
	if m.currentDBVersion < 24 {
		err := m.updateSettingsToDB24()
		if err != nil {
			return err
		}
	}

	// Portainer 2.0.0
	if m.currentDBVersion < 25 {
		err := m.updateSettingsToDB25()
		if err != nil {
			return err
		}

		err = m.updateStacksToDB24()
		if err != nil {
			return err
		}
	}

	// Portainer 2.1.0
	if m.currentDBVersion < 26 {
		err := m.updateEndpointSettingsToDB26()
		if err != nil {
			return err
		}
	}

	// Portainer 2.2.0
	if m.currentDBVersion < 27 {
		err := m.updateStackResourceControlToDB27()
		if err != nil {
			return err
		}
	}

	// Portainer 2.6.0
	if m.currentDBVersion < 30 {
		err := m.migrateDBVersionTo30()
		if err != nil {
			return err
		}
	}

	return m.versionService.StoreDBVersion(portainer.DBVersion)
}
