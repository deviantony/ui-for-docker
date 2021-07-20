package endpoints

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

// @id EndpointAssociationDelete
// @summary De-association an edge endpoint
// @description De-association an edge endpoint.
// @description **Access policy**: administrator
// @security jwt
// @tags endpoints
// @produce json
// @param id path int true "Endpoint identifier"
// @success 200 {object} portainer.Endpoint "Success"
// @failure 400 "Invalid request"
// @failure 404 "Endpoint not found"
// @failure 500 "Server error"
// @router /api/endpoints/:id/association [put]
func (handler *Handler) endpointAssociationDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
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

	if endpoint.Type != portainer.EdgeAgentOnKubernetesEnvironment && endpoint.Type != portainer.EdgeAgentOnDockerEnvironment {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid endpoint type", errors.New("Invalid endpoint type")}
	}

	endpoint.EdgeID = ""
	endpoint.Snapshots = []portainer.DockerSnapshot{}
	endpoint.Kubernetes.Snapshots = []portainer.KubernetesSnapshot{}

	err = handler.DataStore.Endpoint().UpdateEndpoint(portainer.EndpointID(endpointID), endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Failed persisting endpoint in database", err}
	}

	handler.ReverseTunnelService.SetTunnelStatusToIdle(endpoint.ID)

	return response.JSON(w, endpoint)
}
