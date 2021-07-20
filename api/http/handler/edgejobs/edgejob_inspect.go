package edgejobs

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

type edgeJobInspectResponse struct {
	*portainer.EdgeJob
	Endpoints []portainer.EndpointID
}

// @id EdgeJobInspect
// @summary Inspect an EdgeJob
// @description
// @tags edge_jobs
// @security jwt
// @accept json
// @produce json
// @param id path string true "EdgeJob Id"
// @success 200 {object} portainer.EdgeJob
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id} [get]
func (handler *Handler) edgeJobInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Edge job identifier route variable", err}
	}

	edgeJob, err := handler.DataStore.EdgeJob().EdgeJob(portainer.EdgeJobID(edgeJobID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an Edge job with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an Edge job with the specified identifier inside the database", err}
	}

	endpointIDs := []portainer.EndpointID{}

	for endpointID := range edgeJob.Endpoints {
		endpointIDs = append(endpointIDs, endpointID)
	}

	responseObj := edgeJobInspectResponse{
		EdgeJob:   edgeJob,
		Endpoints: endpointIDs,
	}

	return response.JSON(w, responseObj)
}
