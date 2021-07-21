package edgestacks

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_hasKubeEndpoint(t *testing.T) {
	endpoints := []portainer.Endpoint{
		{ID: 1, Type: portainer.DockerEnvironment},
		{ID: 2, Type: portainer.AgentOnDockerEnvironment},
		{ID: 3, Type: portainer.AzureEnvironment},
		{ID: 4, Type: portainer.EdgeAgentOnDockerEnvironment},
		{ID: 5, Type: portainer.KubernetesLocalEnvironment},
		{ID: 6, Type: portainer.AgentOnKubernetesEnvironment},
		{ID: 7, Type: portainer.EdgeAgentOnKubernetesEnvironment},
	}

	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints(endpoints))

	tests := []struct {
		endpointIds []portainer.EndpointID
		expected    bool
	}{
		{endpointIds: []portainer.EndpointID{1}, expected: false},
		{endpointIds: []portainer.EndpointID{2}, expected: false},
		{endpointIds: []portainer.EndpointID{3}, expected: false},
		{endpointIds: []portainer.EndpointID{4}, expected: false},
		{endpointIds: []portainer.EndpointID{5}, expected: true},
		{endpointIds: []portainer.EndpointID{6}, expected: true},
		{endpointIds: []portainer.EndpointID{7}, expected: true},
		{endpointIds: []portainer.EndpointID{7, 2}, expected: true},
		{endpointIds: []portainer.EndpointID{6, 4, 1}, expected: true},
		{endpointIds: []portainer.EndpointID{1, 2, 3}, expected: false},
	}

	for _, test := range tests {

		ans, err := hasKubeEndpoint(datastore.Endpoint(), test.endpointIds)
		assert.NoError(t, err, "hasKubeEndpoint shouldn't fail")

		assert.Equal(t, test.expected, ans, "hasKubeEndpoint expected to return %b for %v, but returned %b", test.expected, test.endpointIds, ans)
	}
}

func Test_hasKubeEndpoint_failWhenEndpointDontExist(t *testing.T) {
	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints([]portainer.Endpoint{}))

	_, err := hasKubeEndpoint(datastore.Endpoint(), []portainer.EndpointID{1})
	assert.Error(t, err, "hasKubeEndpoint should fail")
}

func Test_hasDockerEndpoint(t *testing.T) {
	endpoints := []portainer.Endpoint{
		{ID: 1, Type: portainer.DockerEnvironment},
		{ID: 2, Type: portainer.AgentOnDockerEnvironment},
		{ID: 3, Type: portainer.AzureEnvironment},
		{ID: 4, Type: portainer.EdgeAgentOnDockerEnvironment},
		{ID: 5, Type: portainer.KubernetesLocalEnvironment},
		{ID: 6, Type: portainer.AgentOnKubernetesEnvironment},
		{ID: 7, Type: portainer.EdgeAgentOnKubernetesEnvironment},
	}

	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints(endpoints))

	tests := []struct {
		endpointIds []portainer.EndpointID
		expected    bool
	}{
		{endpointIds: []portainer.EndpointID{1}, expected: true},
		{endpointIds: []portainer.EndpointID{2}, expected: true},
		{endpointIds: []portainer.EndpointID{3}, expected: false},
		{endpointIds: []portainer.EndpointID{4}, expected: true},
		{endpointIds: []portainer.EndpointID{5}, expected: false},
		{endpointIds: []portainer.EndpointID{6}, expected: false},
		{endpointIds: []portainer.EndpointID{7}, expected: false},
		{endpointIds: []portainer.EndpointID{7, 2}, expected: true},
		{endpointIds: []portainer.EndpointID{6, 4, 1}, expected: true},
		{endpointIds: []portainer.EndpointID{1, 2, 3}, expected: true},
	}

	for _, test := range tests {

		ans, err := hasDockerEndpoint(datastore.Endpoint(), test.endpointIds)
		assert.NoError(t, err, "hasDockerEndpoint shouldn't fail")

		assert.Equal(t, test.expected, ans, "hasDockerEndpoint expected to return %b for %v, but returned %b", test.expected, test.endpointIds, ans)
	}
}

func Test_hasDockerEndpoint_failWhenEndpointDontExist(t *testing.T) {
	datastore := testhelpers.NewDatastore(testhelpers.WithEndpoints([]portainer.Endpoint{}))

	_, err := hasDockerEndpoint(datastore.Endpoint(), []portainer.EndpointID{1})
	assert.Error(t, err, "hasDockerEndpoint should fail")
}
