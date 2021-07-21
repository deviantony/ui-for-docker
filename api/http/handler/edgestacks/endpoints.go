package edgestacks

import (
	"fmt"

	portainer "github.com/portainer/portainer/api"
	endpointutils "github.com/portainer/portainer/api/internal/endpoint"
)

func hasKubeEndpoint(endpointService portainer.EndpointService, endpointIDs []portainer.EndpointID) (bool, error) {
	return hasEndpointPredicate(endpointService, endpointIDs, endpointutils.IsKubernetesEndpoint)
}

func hasDockerEndpoint(endpointService portainer.EndpointService, endpointIDs []portainer.EndpointID) (bool, error) {
	return hasEndpointPredicate(endpointService, endpointIDs, endpointutils.IsDockerEndpoint)
}

func hasEndpointPredicate(endpointService portainer.EndpointService, endpointIDs []portainer.EndpointID, predicate func(*portainer.Endpoint) bool) (bool, error) {
	for _, endpointID := range endpointIDs {
		endpoint, err := endpointService.Endpoint(endpointID)
		if err != nil {
			return false, fmt.Errorf("failed to retrieve endpoint from database: %w", err)
		}

		if predicate(endpoint) {
			return true, nil
		}
	}

	return false, nil
}

type endpointRelationsConfig struct {
	endpoints      []portainer.Endpoint
	endpointGroups []portainer.EndpointGroup
	edgeGroups     []portainer.EdgeGroup
}

func fetchEndpointRelationsConfig(dataStore portainer.DataStore) (*endpointRelationsConfig, error) {
	endpoints, err := dataStore.Endpoint().Endpoints()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve endpoints from database: %w", err)
	}

	endpointGroups, err := dataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve endpoint groups from database: %w", err)
	}

	edgeGroups, err := dataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve edge groups from database: %w", err)
	}

	return &endpointRelationsConfig{
		endpoints:      endpoints,
		endpointGroups: endpointGroups,
		edgeGroups:     edgeGroups,
	}, nil
}
