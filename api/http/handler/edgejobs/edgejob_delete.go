package edgejobs

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

// @id EdgeJobDelete
// @summary Delete an EdgeJob
// @description
// @tags edge_jobs
// @security jwt
// @accept json
// @produce json
// @param id path string true "EdgeJob Id"
// @success 204
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id} [delete]
func (handler *Handler) edgeJobDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
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

	edgeJobFolder := handler.FileService.GetEdgeJobFolder(strconv.Itoa(edgeJobID))
	err = handler.FileService.RemoveDirectory(edgeJobFolder)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove the files associated to the Edge job on the filesystem", err}
	}

	handler.ReverseTunnelService.RemoveEdgeJob(edgeJob.ID)

	err = handler.DataStore.EdgeJob().DeleteEdgeJob(edgeJob.ID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove the Edge job from the database", err}
	}

	return response.Empty(w)
}
