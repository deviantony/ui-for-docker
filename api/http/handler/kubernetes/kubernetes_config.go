package kubernetes

import (
	"errors"
	"fmt"
	"strings"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
	kcli "github.com/portainer/portainer/api/kubernetes/cli"

	"net/http"
)

// @id GetKubernetesConfig
// @summary Generates kubeconfig file enabling client communication with k8s api server
// @description Generates kubeconfig file enabling client communication with k8s api server
// @description **Access policy**: authorized
// @tags kubernetes
// @security jwt
// @accept json
// @produce json
// @param id path int true "Endpoint identifier"
// @success 200 "Success"
// @failure 400 "Invalid request"
// @failure 401 "Unauthorized"
// @failure 403 "Permission denied"
// @failure 404 "Endpoint or ServiceAccount not found"
// @failure 500 "Server error"
// @router /kubernetes/{id}/config [get]
func (handler *Handler) getKubernetesConfig(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	if r.TLS == nil {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Kubernetes config generation only supported on portainer instances running with TLS",
			Err:        errors.New("missing request TLS config"),
		}
	}

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

	bearerToken, err := extractBearerToken(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusUnauthorized, "Unauthorized", err}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access endpoint", err}
	}

	cli, err := handler.KubernetesClientFactory.GetKubeClient(endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to create Kubernetes client", err}
	}

	apiServerURL := getProxyUrl(r, endpointID)

	config, err := cli.GetKubeConfig(r.Context(), apiServerURL, bearerToken, tokenData)
	if err != nil {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to generate Kubeconfig", err}
	}

	contentAcceptHeader := r.Header.Get("Accept")
	if contentAcceptHeader == "text/yaml" {
		yaml, err := kcli.GenerateYAML(config)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Failed to generate Kubeconfig", err}
		}
		w.Header().Set("Content-Disposition", `attachment; filename=config.yaml`)
		return YAML(w, yaml)
	}

	w.Header().Set("Content-Disposition", `attachment; filename="config.json"`)
	return response.JSON(w, config)
}

// extractBearerToken extracts user's portainer bearer token from request auth header
func extractBearerToken(r *http.Request) (string, error) {
	token := ""
	tokens := r.Header["Authorization"]
	if len(tokens) >= 1 {
		token = tokens[0]
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if token == "" {
		return "", httperrors.ErrUnauthorized
	}
	return token, nil
}

// getProxyUrl generates portainer proxy url which acts as proxy to k8s api server
func getProxyUrl(r *http.Request, endpointID int) string {
	return fmt.Sprintf("https://%s/api/endpoints/%d/kubernetes", r.Host, endpointID)
}

// YAML writes yaml response as string to writer. Returns a pointer to a HandlerError if encoding fails.
// This could be moved to a more useful place; but that place is most likely not in this project.
// It should actually go in https://github.com/portainer/libhttp - since that is from where we use response.JSON.
// We use `data interface{}` as parameter - since im trying to keep it as close to (or the same as) response.JSON method signature:
// https://github.com/portainer/libhttp/blob/d20481a3da823c619887c440a22fdf4fa8f318f2/response/response.go#L13
func YAML(rw http.ResponseWriter, data interface{}) *httperror.HandlerError {
	rw.Header().Set("Content-Type", "text/yaml")

	strData, ok := data.(string)
	if !ok {
		return &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to write YAML response",
			Err:        errors.New("failed to convert input to string"),
		}
	}

	fmt.Fprint(rw, strData)

	return nil
}
