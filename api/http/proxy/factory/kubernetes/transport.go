package kubernetes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/portainer/portainer/api/http/security"

	portainer "github.com/portainer/portainer/api"
)

type baseTransport struct {
	httpTransport *http.Transport
	tokenManager  *tokenManager
	endpoint      *portainer.Endpoint
	dataStore     portainer.DataStore
}

func newBaseTransport(httpTransport *http.Transport, tokenManager *tokenManager, endpoint *portainer.Endpoint, dataStore portainer.DataStore) *baseTransport {
	return &baseTransport{
		httpTransport: httpTransport,
		tokenManager:  tokenManager,
		endpoint:      endpoint,
		dataStore:     dataStore,
	}
}

func (transport *baseTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	apiVersionRe := regexp.MustCompile(`^(/kubernetes)?/api/v[0-9](\.[0-9])?`)
	requestPath := apiVersionRe.ReplaceAllString(request.URL.Path, "")

	switch {
	case strings.EqualFold(requestPath, "/namespaces"):
		return transport.executeKubernetesRequest(request, true)
	case strings.HasPrefix(requestPath, "/namespaces"):
		return transport.proxyNamespacedRequest(request, requestPath)
	default:
		return transport.executeKubernetesRequest(request, true)
	}
}

func (transport *baseTransport) proxyNamespacedRequest(request *http.Request, fullRequestPath string) (*http.Response, error) {
	requestPath := strings.TrimPrefix(fullRequestPath, "/namespaces/")
	split := strings.SplitN(requestPath, "/", 2)
	namespace := split[0]

	requestPath = ""
	if len(split) > 1 {
		requestPath = split[1]
	}

	switch {
	case requestPath == "" && request.Method == "DELETE":
		return transport.deleteNamespaceRequest(request, namespace)
	default:
		return transport.executeKubernetesRequest(request, true)
	}
}

func (transport *baseTransport) executeKubernetesRequest(request *http.Request, shouldLog bool) (*http.Response, error) {
	return transport.httpTransport.RoundTrip(request)
}

var (
	namespaceRegex = regexp.MustCompile(`^/namespaces/([^/]*)$`)
)

func getRoundTripToken(
	request *http.Request,
	tokenManager *tokenManager,
	endpointIdentifier portainer.EndpointID,
) (string, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return "", err
	}

	var token string
	if tokenData.Role == portainer.AdministratorRole {
		token = tokenManager.GetAdminServiceAccountToken()
	} else {
		token, err = tokenManager.GetUserServiceAccountToken(int(tokenData.ID))
		if err != nil {
			log.Printf("Failed retrieving service account token: %v", err)
			return "", err
		}
	}

	return token, nil
}

func decorateAgentRequest(r *http.Request, dataStore portainer.DataStore) error {
	requestPath := strings.TrimPrefix(r.URL.Path, "/v2")

	switch {
	case strings.HasPrefix(requestPath, "/dockerhub"):
		decorateAgentDockerHubRequest(r, dataStore)
	}

	return nil
}

func decorateAgentDockerHubRequest(r *http.Request, dataStore portainer.DataStore) error {
	dockerhub, err := dataStore.DockerHub().DockerHub()
	if err != nil {
		return err
	}

	newBody, err := json.Marshal(dockerhub)
	if err != nil {
		return err
	}

	r.Method = http.MethodPost

	r.Body = ioutil.NopCloser(bytes.NewReader(newBody))
	r.ContentLength = int64(len(newBody))

	return nil
}
