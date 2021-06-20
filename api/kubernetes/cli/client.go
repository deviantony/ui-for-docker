package cli

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	cmap "github.com/orcaman/concurrent-map"

	portainer "github.com/portainer/portainer/api"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type (
	// ClientFactory is used to create Kubernetes clients
	ClientFactory struct {
		dataStore            portainer.DataStore
		reverseTunnelService portainer.ReverseTunnelService
		signatureService     portainer.DigitalSignatureService
		instanceID           string
		endpointClients      cmap.ConcurrentMap
	}

	// KubeClient represent a service used to execute Kubernetes operations
	KubeClient struct {
		cli        kubernetes.Interface
		instanceID string
	}
)

// NewClientFactory returns a new instance of a ClientFactory
func NewClientFactory(signatureService portainer.DigitalSignatureService, reverseTunnelService portainer.ReverseTunnelService, instanceID string, dataStore portainer.DataStore) *ClientFactory {
	return &ClientFactory{
		dataStore:            dataStore,
		signatureService:     signatureService,
		reverseTunnelService: reverseTunnelService,
		instanceID:           instanceID,
		endpointClients:      cmap.New(),
	}
}

// Remove the cached kube client so a new one can be created
func (factory *ClientFactory) RemoveKubeClient(endpoint *portainer.Endpoint) {
	factory.endpointClients.Remove(strconv.Itoa(int(endpoint.ID)))
}

// GetKubeClient checks if an existing client is already registered for the endpoint and returns it if one is found.
// If no client is registered, it will create a new client, register it, and returns it.
func (factory *ClientFactory) GetKubeClient(endpoint *portainer.Endpoint) (portainer.KubeClient, error) {
	key := strconv.Itoa(int(endpoint.ID))
	client, ok := factory.endpointClients.Get(key)
	if !ok {
		client, err := factory.createKubeClient(endpoint)
		if err != nil {
			return nil, err
		}

		factory.endpointClients.Set(key, client)
		return client, nil
	}

	return client.(portainer.KubeClient), nil
}

func (factory *ClientFactory) createKubeClient(endpoint *portainer.Endpoint) (portainer.KubeClient, error) {
	cli, err := factory.CreateClient(endpoint)
	if err != nil {
		return nil, err
	}

	kubecli := &KubeClient{
		cli:        cli,
		instanceID: factory.instanceID,
	}

	return kubecli, nil
}

// CreateClient returns a pointer to a new Clientset instance
func (factory *ClientFactory) CreateClient(endpoint *portainer.Endpoint) (*kubernetes.Clientset, error) {
	switch endpoint.Type {
	case portainer.KubernetesLocalEnvironment:
		return buildLocalClient()
	case portainer.AgentOnKubernetesEnvironment:
		return factory.buildAgentClient(endpoint)
	case portainer.EdgeAgentOnKubernetesEnvironment:
		return factory.buildEdgeClient(endpoint)
	}

	return nil, errors.New("unsupported endpoint type")
}

type agentHeaderRoundTripper struct {
	signatureHeader string
	publicKeyHeader string

	roundTripper http.RoundTripper
}

// RoundTrip is the implementation of the http.RoundTripper interface.
// It decorates the request with specific agent headers
func (rt *agentHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add(portainer.PortainerAgentPublicKeyHeader, rt.publicKeyHeader)
	req.Header.Add(portainer.PortainerAgentSignatureHeader, rt.signatureHeader)

	return rt.roundTripper.RoundTrip(req)
}

func (factory *ClientFactory) buildAgentClient(endpoint *portainer.Endpoint) (*kubernetes.Clientset, error) {
	endpointURL := fmt.Sprintf("https://%s/kubernetes", endpoint.URL)
	signature, err := factory.signatureService.CreateSignature(portainer.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.BuildConfigFromFlags(endpointURL, "")
	if err != nil {
		return nil, err
	}
	config.Insecure = true

	config.Wrap(func(rt http.RoundTripper) http.RoundTripper {
		return &agentHeaderRoundTripper{
			signatureHeader: signature,
			publicKeyHeader: factory.signatureService.EncodedPublicKey(),
			roundTripper:    rt,
		}
	})

	return kubernetes.NewForConfig(config)
}

func (factory *ClientFactory) buildEdgeClient(endpoint *portainer.Endpoint) (*kubernetes.Clientset, error) {
	tunnel := factory.reverseTunnelService.GetTunnelDetails(endpoint.ID)

	if tunnel.Status == portainer.EdgeAgentIdle {
		err := factory.reverseTunnelService.SetTunnelStatusToRequired(endpoint.ID)
		if err != nil {
			return nil, fmt.Errorf("failed opening tunnel to endpoint: %w", err)
		}

		if endpoint.EdgeCheckinInterval == 0 {
			settings, err := factory.dataStore.Settings().Settings()
			if err != nil {
				return nil, fmt.Errorf("failed fetching settings from db: %w", err)
			}

			endpoint.EdgeCheckinInterval = settings.EdgeAgentCheckinInterval
		}

		waitForAgentToConnect := time.Duration(endpoint.EdgeCheckinInterval) * time.Second
		time.Sleep(waitForAgentToConnect * 2)

		tunnel = factory.reverseTunnelService.GetTunnelDetails(endpoint.ID)
	}

	endpointURL := fmt.Sprintf("http://127.0.0.1:%d/kubernetes", tunnel.Port)

	config, err := clientcmd.BuildConfigFromFlags(endpointURL, "")
	if err != nil {
		return nil, err
	}
	config.Insecure = true

	return kubernetes.NewForConfig(config)
}

func buildLocalClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
