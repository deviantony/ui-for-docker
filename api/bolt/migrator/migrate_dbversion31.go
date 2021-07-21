package migrator

import (
	"fmt"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	endpointutils "github.com/portainer/portainer/api/internal/endpoint"
	snapshotutils "github.com/portainer/portainer/api/internal/snapshot"
)

func (m *Migrator) migrateDBVersionToDB32() error {
	err := m.updateRegistriesToDB32()
	if err != nil {
		return err
	}

	err = m.updateDockerhubToDB32()
	if err != nil {
		return err
	}

	err = m.migrateAdminGroupSearchSettings()
	if err != nil {
		return err
	}

	err = m.updateVolumeResourceControlToDB32()
	if err != nil {
		return err
	}

	return nil
}

func (m *Migrator) updateRegistriesToDB32() error {
	registries, err := m.registryService.Registries()
	if err != nil {
		return err
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, registry := range registries {

		registry.RegistryAccesses = portainer.RegistryAccesses{}

		for _, endpoint := range endpoints {

			filteredUserAccessPolicies := portainer.UserAccessPolicies{}
			for userId, registryPolicy := range registry.UserAccessPolicies {
				if _, found := endpoint.UserAccessPolicies[userId]; found {
					filteredUserAccessPolicies[userId] = registryPolicy
				}
			}

			filteredTeamAccessPolicies := portainer.TeamAccessPolicies{}
			for teamId, registryPolicy := range registry.TeamAccessPolicies {
				if _, found := endpoint.TeamAccessPolicies[teamId]; found {
					filteredTeamAccessPolicies[teamId] = registryPolicy
				}
			}

			registry.RegistryAccesses[endpoint.ID] = portainer.RegistryAccessPolicies{
				UserAccessPolicies: filteredUserAccessPolicies,
				TeamAccessPolicies: filteredTeamAccessPolicies,
				Namespaces:         []string{},
			}
		}
		m.registryService.UpdateRegistry(registry.ID, &registry)
	}
	return nil
}

func (m *Migrator) updateDockerhubToDB32() error {
	dockerhub, err := m.dockerhubService.DockerHub()
	if err == errors.ErrObjectNotFound {
		return nil
	} else if err != nil {
		return err
	}

	if !dockerhub.Authentication {
		return nil
	}

	registry := &portainer.Registry{
		Type:             portainer.DockerHubRegistry,
		Name:             "Dockerhub (authenticated - migrated)",
		URL:              "docker.io",
		Authentication:   true,
		Username:         dockerhub.Username,
		Password:         dockerhub.Password,
		RegistryAccesses: portainer.RegistryAccesses{},
	}

	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return err
	}

	for _, endpoint := range endpoints {

		if endpoint.Type != portainer.KubernetesLocalEnvironment &&
			endpoint.Type != portainer.AgentOnKubernetesEnvironment &&
			endpoint.Type != portainer.EdgeAgentOnKubernetesEnvironment {

			userAccessPolicies := portainer.UserAccessPolicies{}
			for userId := range endpoint.UserAccessPolicies {
				if _, found := endpoint.UserAccessPolicies[userId]; found {
					userAccessPolicies[userId] = portainer.AccessPolicy{
						RoleID: 0,
					}
				}
			}

			teamAccessPolicies := portainer.TeamAccessPolicies{}
			for teamId := range endpoint.TeamAccessPolicies {
				if _, found := endpoint.TeamAccessPolicies[teamId]; found {
					teamAccessPolicies[teamId] = portainer.AccessPolicy{
						RoleID: 0,
					}
				}
			}

			registry.RegistryAccesses[endpoint.ID] = portainer.RegistryAccessPolicies{
				UserAccessPolicies: userAccessPolicies,
				TeamAccessPolicies: teamAccessPolicies,
				Namespaces:         []string{},
			}
		}
	}

	return m.registryService.CreateRegistry(registry)
}

func (m *Migrator) updateVolumeResourceControlToDB32() error {
	endpoints, err := m.endpointService.Endpoints()
	if err != nil {
		return fmt.Errorf("failed fetching endpoints: %w", err)
	}

	resourceControls, err := m.resourceControlService.ResourceControls()
	if err != nil {
		return fmt.Errorf("failed fetching resource controls: %w", err)
	}

	toUpdate := map[portainer.ResourceControlID]string{}
	volumeResourceControls := map[string]*portainer.ResourceControl{}

	for i := range resourceControls {
		resourceControl := resourceControls[i]
		if resourceControl.Type == portainer.VolumeResourceControl {
			volumeResourceControls[resourceControl.ResourceID] = &resourceControl
		}
	}

	for _, endpoint := range endpoints {
		if !endpointutils.IsDockerEndpoint(&endpoint) {
			continue
		}

		totalSnapshots := len(endpoint.Snapshots)
		if totalSnapshots == 0 {
			continue
		}

		snapshot := endpoint.Snapshots[totalSnapshots-1]

		endpointDockerID, err := snapshotutils.FetchDockerID(snapshot)
		if err != nil {
			return fmt.Errorf("failed fetching endpoint docker id: %w", err)
		}

		if volumesData, done := snapshot.SnapshotRaw.Volumes.(map[string]interface{}); done {
			if volumesData["Volumes"] == nil {
				continue
			}

			findResourcesToUpdateForDB32(endpointDockerID, volumesData, toUpdate, volumeResourceControls)
		}
	}

	for _, resourceControl := range volumeResourceControls {
		if newResourceID, ok := toUpdate[resourceControl.ID]; ok {
			resourceControl.ResourceID = newResourceID
			err := m.resourceControlService.UpdateResourceControl(resourceControl.ID, resourceControl)
			if err != nil {
				return fmt.Errorf("failed updating resource control %d: %w", resourceControl.ID, err)
			}

		} else {
			err := m.resourceControlService.DeleteResourceControl(resourceControl.ID)
			if err != nil {
				return fmt.Errorf("failed deleting resource control %d: %w", resourceControl.ID, err)
			}

		}
	}

	return nil
}

func findResourcesToUpdateForDB32(dockerID string, volumesData map[string]interface{}, toUpdate map[portainer.ResourceControlID]string, volumeResourceControls map[string]*portainer.ResourceControl) {
	volumes := volumesData["Volumes"].([]interface{})
	for _, volumeMeta := range volumes {
		volume := volumeMeta.(map[string]interface{})
		volumeName := volume["Name"].(string)
		oldResourceID := fmt.Sprintf("%s%s", volumeName, volume["CreatedAt"].(string))
		resourceControl, ok := volumeResourceControls[oldResourceID]

		if ok {
			toUpdate[resourceControl.ID] = fmt.Sprintf("%s_%s", volumeName, dockerID)
		}
	}
}

func (m *Migrator) migrateAdminGroupSearchSettings() error {
	legacySettings, err := m.settingsService.Settings()
	if err != nil {
		return err
	}
	if legacySettings.LDAPSettings.AdminGroupSearchSettings == nil {
		legacySettings.LDAPSettings.AdminGroupSearchSettings = []portainer.LDAPGroupSearchSettings{}
	}
	return m.settingsService.UpdateSettings(legacySettings)
}
