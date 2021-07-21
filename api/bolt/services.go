package bolt

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/customtemplate"
	"github.com/portainer/portainer/api/bolt/dockerhub"
	"github.com/portainer/portainer/api/bolt/edgegroup"
	"github.com/portainer/portainer/api/bolt/edgejob"
	"github.com/portainer/portainer/api/bolt/edgestack"
	"github.com/portainer/portainer/api/bolt/endpoint"
	"github.com/portainer/portainer/api/bolt/endpointgroup"
	"github.com/portainer/portainer/api/bolt/endpointrelation"
	"github.com/portainer/portainer/api/bolt/extension"
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
)

func (store *Store) initServices() error {
	authorizationsetService, err := role.NewService(store.connection)
	if err != nil {
		return err
	}
	store.RoleService = authorizationsetService

	customTemplateService, err := customtemplate.NewService(store.connection)
	if err != nil {
		return err
	}
	store.CustomTemplateService = customTemplateService

	dockerhubService, err := dockerhub.NewService(store.connection)
	if err != nil {
		return err
	}
	store.DockerHubService = dockerhubService

	edgeStackService, err := edgestack.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeStackService = edgeStackService

	edgeGroupService, err := edgegroup.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeGroupService = edgeGroupService

	edgeJobService, err := edgejob.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EdgeJobService = edgeJobService

	endpointgroupService, err := endpointgroup.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EndpointGroupService = endpointgroupService

	endpointService, err := endpoint.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EndpointService = endpointService

	endpointRelationService, err := endpointrelation.NewService(store.connection)
	if err != nil {
		return err
	}
	store.EndpointRelationService = endpointRelationService

	extensionService, err := extension.NewService(store.connection)
	if err != nil {
		return err
	}
	store.ExtensionService = extensionService

	registryService, err := registry.NewService(store.connection)
	if err != nil {
		return err
	}
	store.RegistryService = registryService

	resourcecontrolService, err := resourcecontrol.NewService(store.connection)
	if err != nil {
		return err
	}
	store.ResourceControlService = resourcecontrolService

	settingsService, err := settings.NewService(store.connection)
	if err != nil {
		return err
	}
	store.SettingsService = settingsService

	sslSettingsService, err := ssl.NewService(store.connection)
	if err != nil {
		return err
	}
	store.SSLSettingsService = sslSettingsService

	stackService, err := stack.NewService(store.connection)
	if err != nil {
		return err
	}
	store.StackService = stackService

	tagService, err := tag.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TagService = tagService

	teammembershipService, err := teammembership.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TeamMembershipService = teammembershipService

	teamService, err := team.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TeamService = teamService

	tunnelServerService, err := tunnelserver.NewService(store.connection)
	if err != nil {
		return err
	}
	store.TunnelServerService = tunnelServerService

	userService, err := user.NewService(store.connection)
	if err != nil {
		return err
	}
	store.UserService = userService

	versionService, err := version.NewService(store.connection)
	if err != nil {
		return err
	}
	store.VersionService = versionService

	webhookService, err := webhook.NewService(store.connection)
	if err != nil {
		return err
	}
	store.WebhookService = webhookService

	scheduleService, err := schedule.NewService(store.connection)
	if err != nil {
		return err
	}
	store.ScheduleService = scheduleService

	return nil
}

// CustomTemplate gives access to the CustomTemplate data management layer
func (store *Store) CustomTemplate() portainer.CustomTemplateService {
	return store.CustomTemplateService
}

// EdgeGroup gives access to the EdgeGroup data management layer
func (store *Store) EdgeGroup() portainer.EdgeGroupService {
	return store.EdgeGroupService
}

// EdgeJob gives access to the EdgeJob data management layer
func (store *Store) EdgeJob() portainer.EdgeJobService {
	return store.EdgeJobService
}

// EdgeStack gives access to the EdgeStack data management layer
func (store *Store) EdgeStack() portainer.EdgeStackService {
	return store.EdgeStackService
}

// Endpoint gives access to the Endpoint data management layer
func (store *Store) Endpoint() portainer.EndpointService {
	return store.EndpointService
}

// EndpointGroup gives access to the EndpointGroup data management layer
func (store *Store) EndpointGroup() portainer.EndpointGroupService {
	return store.EndpointGroupService
}

// EndpointRelation gives access to the EndpointRelation data management layer
func (store *Store) EndpointRelation() portainer.EndpointRelationService {
	return store.EndpointRelationService
}

// Registry gives access to the Registry data management layer
func (store *Store) Registry() portainer.RegistryService {
	return store.RegistryService
}

// ResourceControl gives access to the ResourceControl data management layer
func (store *Store) ResourceControl() portainer.ResourceControlService {
	return store.ResourceControlService
}

// Role gives access to the Role data management layer
func (store *Store) Role() portainer.RoleService {
	return store.RoleService
}

// Settings gives access to the Settings data management layer
func (store *Store) Settings() portainer.SettingsService {
	return store.SettingsService
}

// SSLSettings gives access to the SSL Settings data management layer
func (store *Store) SSLSettings() portainer.SSLSettingsService {
	return store.SSLSettingsService
}

// Stack gives access to the Stack data management layer
func (store *Store) Stack() portainer.StackService {
	return store.StackService
}

// Tag gives access to the Tag data management layer
func (store *Store) Tag() portainer.TagService {
	return store.TagService
}

// TeamMembership gives access to the TeamMembership data management layer
func (store *Store) TeamMembership() portainer.TeamMembershipService {
	return store.TeamMembershipService
}

// Team gives access to the Team data management layer
func (store *Store) Team() portainer.TeamService {
	return store.TeamService
}

// TunnelServer gives access to the TunnelServer data management layer
func (store *Store) TunnelServer() portainer.TunnelServerService {
	return store.TunnelServerService
}

// User gives access to the User data management layer
func (store *Store) User() portainer.UserService {
	return store.UserService
}

// Version gives access to the Version data management layer
func (store *Store) Version() portainer.VersionService {
	return store.VersionService
}

// Webhook gives access to the Webhook data management layer
func (store *Store) Webhook() portainer.WebhookService {
	return store.WebhookService
}
