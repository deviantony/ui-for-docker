package endpoint

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

type isEndpointTypeTest struct {
	endpointType portainer.EndpointType
	expected     bool
}

func Test_IsDockerEndpoint(t *testing.T) {
	tests := []isEndpointTypeTest{
		{endpointType: portainer.DockerEnvironment, expected: true},
		{endpointType: portainer.AgentOnDockerEnvironment, expected: true},
		{endpointType: portainer.AzureEnvironment, expected: false},
		{endpointType: portainer.EdgeAgentOnDockerEnvironment, expected: true},
		{endpointType: portainer.KubernetesLocalEnvironment, expected: false},
		{endpointType: portainer.AgentOnKubernetesEnvironment, expected: false},
		{endpointType: portainer.EdgeAgentOnKubernetesEnvironment, expected: false},
	}

	for _, test := range tests {
		ans := IsDockerEndpoint(&portainer.Endpoint{Type: test.endpointType})
		assert.Equal(t, test.expected, ans)
	}
}

func Test_IsKubernetesEndpoint(t *testing.T) {
	tests := []isEndpointTypeTest{
		{endpointType: portainer.DockerEnvironment, expected: false},
		{endpointType: portainer.AgentOnDockerEnvironment, expected: false},
		{endpointType: portainer.AzureEnvironment, expected: false},
		{endpointType: portainer.EdgeAgentOnDockerEnvironment, expected: false},
		{endpointType: portainer.KubernetesLocalEnvironment, expected: true},
		{endpointType: portainer.AgentOnKubernetesEnvironment, expected: true},
		{endpointType: portainer.EdgeAgentOnKubernetesEnvironment, expected: true},
	}

	for _, test := range tests {
		ans := IsKubernetesEndpoint(&portainer.Endpoint{Type: test.endpointType})
		assert.Equal(t, test.expected, ans)
	}
}
