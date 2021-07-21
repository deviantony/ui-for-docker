package kubernetes

import (
	"crypto/tls"
	"net/http"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

type agentTransport struct {
	*baseTransport
	signatureService portainer.DigitalSignatureService
}

// NewAgentTransport returns a new transport that can be used to send signed requests to a Portainer agent
func NewAgentTransport(signatureService portainer.DigitalSignatureService, tlsConfig *tls.Config, tokenManager *tokenManager, endpoint *portainer.Endpoint, k8sClientFactory *cli.ClientFactory, dataStore portainer.DataStore) *agentTransport {
	transport := &agentTransport{
		baseTransport: newBaseTransport(
			&http.Transport{
				TLSClientConfig: tlsConfig,
			},
			tokenManager,
			endpoint,
			k8sClientFactory,
			dataStore,
		),
		signatureService: signatureService,
	}

	return transport
}

// RoundTrip is the implementation of the the http.RoundTripper interface
func (transport *agentTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	token, err := transport.getRoundTripToken(request, transport.tokenManager)
	if err != nil {
		return nil, err
	}

	request.Header.Set(portainer.PortainerAgentKubernetesSATokenHeader, token)

	if strings.HasPrefix(request.URL.Path, "/v2") {
		err := decorateAgentRequest(request, transport.dataStore)
		if err != nil {
			return nil, err
		}
	}

	signature, err := transport.signatureService.CreateSignature(portainer.PortainerAgentSignatureMessage)
	if err != nil {
		return nil, err
	}

	request.Header.Set(portainer.PortainerAgentPublicKeyHeader, transport.signatureService.EncodedPublicKey())
	request.Header.Set(portainer.PortainerAgentSignatureHeader, signature)

	return transport.baseTransport.RoundTrip(request)
}
