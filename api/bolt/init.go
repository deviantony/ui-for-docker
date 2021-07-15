package bolt

import (
	"github.com/gofrs/uuid"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
)

// Init creates the default data set.
func (store *Store) Init() error {
	instanceID, err := store.VersionService.InstanceID()
	if err == errors.ErrObjectNotFound {
		uid, err := uuid.NewV4()
		if err != nil {
			return err
		}

		instanceID = uid.String()
		err = store.VersionService.StoreInstanceID(instanceID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	_, err = store.SettingsService.Settings()
	if err == errors.ErrObjectNotFound {
		defaultSettings := &portainer.Settings{
			AuthenticationMethod: portainer.AuthenticationInternal,
			BlackListedLabels:    make([]portainer.Pair, 0),
			LDAPSettings: portainer.LDAPSettings{
				AnonymousMode:   true,
				AutoCreateUsers: true,
				TLSConfig:       portainer.TLSConfiguration{},
				SearchSettings: []portainer.LDAPSearchSettings{
					portainer.LDAPSearchSettings{},
				},
				GroupSearchSettings: []portainer.LDAPGroupSearchSettings{
					portainer.LDAPGroupSearchSettings{},
				},
			},
			OAuthSettings: portainer.OAuthSettings{},

			EdgeAgentCheckinInterval: portainer.DefaultEdgeAgentCheckinIntervalInSeconds,
			TemplatesURL:             portainer.DefaultTemplatesURL,
			UserSessionTimeout:       portainer.DefaultUserSessionTimeout,
		}

		err = store.SettingsService.UpdateSettings(defaultSettings)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	_, err = store.SSLSettings().Settings()
	if err != nil {
		if err != errors.ErrObjectNotFound {
			return err
		}

		defaultSSLSettings := &portainer.SSLSettings{
			HTTPEnabled: true,
		}

		err = store.SSLSettings().UpdateSettings(defaultSSLSettings)
		if err != nil {
			return err
		}
	}

	groups, err := store.EndpointGroupService.EndpointGroups()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		unassignedGroup := &portainer.EndpointGroup{
			Name:               "Unassigned",
			Description:        "Unassigned endpoints",
			Labels:             []portainer.Pair{},
			UserAccessPolicies: portainer.UserAccessPolicies{},
			TeamAccessPolicies: portainer.TeamAccessPolicies{},
			TagIDs:             []portainer.TagID{},
		}

		err = store.EndpointGroupService.CreateEndpointGroup(unassignedGroup)
		if err != nil {
			return err
		}
	}

	return nil
}
