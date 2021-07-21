package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/http/client"
	"github.com/portainer/portainer/api/internal/endpointutils"
)

type dockerhubStatusResponse struct {
	Remaining int `json:"remaining"`
	Limit     int `json:"limit"`
}

// GET request on /api/endpoints/{id}/dockerhub/{registryId}
func (handler *Handler) endpointDockerhubStatus(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid endpoint identifier route variable", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an endpoint with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an endpoint with the specified identifier inside the database", err}
	}

	if !endpointutils.IsLocalEndpoint(endpoint) {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid environment type", errors.New("Invalid environment type")}
	}

	registryID, err := request.RetrieveNumericRouteVariableValue(r, "registryId")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid registry identifier route variable", err}
	}

	var registry *portainer.Registry

	if registryID == 0 {
		registry = &portainer.Registry{}
	} else {
		registry, err = handler.DataStore.Registry().Registry(portainer.RegistryID(registryID))
		if err == bolterrors.ErrObjectNotFound {
			return &httperror.HandlerError{http.StatusNotFound, "Unable to find a registry with the specified identifier inside the database", err}
		} else if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a registry with the specified identifier inside the database", err}
		}

		if registry.Type != portainer.DockerHubRegistry {
			return &httperror.HandlerError{http.StatusBadRequest, "Invalid registry type", errors.New("Invalid registry type")}
		}
	}

	httpClient := client.NewHTTPClient()
	token, err := getDockerHubToken(httpClient, registry)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve DockerHub token from DockerHub", err}
	}

	resp, err := getDockerHubLimits(httpClient, token)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve DockerHub rate limits from DockerHub", err}
	}

	return response.JSON(w, resp)
}

func getDockerHubToken(httpClient *client.HTTPClient, registry *portainer.Registry) (string, error) {
	type dockerhubTokenResponse struct {
		Token string `json:"token"`
	}

	requestURL := "https://auth.docker.io/token?service=registry.docker.io&scope=repository:ratelimitpreview/test:pull"

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return "", err
	}

	if registry.Authentication {
		req.SetBasicAuth(registry.Username, registry.Password)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed fetching dockerhub token")
	}

	var data dockerhubTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return data.Token, nil
}

func getDockerHubLimits(httpClient *client.HTTPClient, token string) (*dockerhubStatusResponse, error) {
	requestURL := "https://registry-1.docker.io/v2/ratelimitpreview/test/manifests/latest"

	req, err := http.NewRequest(http.MethodHead, requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed fetching dockerhub limits")
	}

	// An error with rateLimit-Limit or RateLimit-Remaining is likely for dockerhub pro accounts where there is no rate limit.
	// In that specific case the headers will not be present.  Don't bubble up the error as its normal
	// See: https://docs.docker.com/docker-hub/download-rate-limit/
	rateLimit, err := parseRateLimitHeader(resp.Header, "RateLimit-Limit")
	if err != nil {
		return nil, nil
	}

	rateLimitRemaining, err := parseRateLimitHeader(resp.Header, "RateLimit-Remaining")
	if err != nil {
		return nil, nil
	}

	return &dockerhubStatusResponse{
		Limit:     rateLimit,
		Remaining: rateLimitRemaining,
	}, nil
}

func parseRateLimitHeader(headers http.Header, headerKey string) (int, error) {
	headerValue := headers.Get(headerKey)
	if headerValue == "" {
		return 0, fmt.Errorf("Missing %s header", headerKey)
	}

	matches := strings.Split(headerValue, ";")
	value, err := strconv.Atoi(matches[0])
	if err != nil {
		return 0, err
	}

	return value, nil
}
