package edgejobs

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id EdgeJobList
// @summary Fetch EdgeJobs list
// @description
// @tags edge_jobs
// @security jwt
// @accept json
// @produce json
// @success 200 {array} portainer.EdgeJob
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs [get]
// GET request on /api/edge_jobs
func (handler *Handler) edgeJobList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobs, err := handler.DataStore.EdgeJob().EdgeJobs()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Edge jobs from the database", err}
	}

	return response.JSON(w, edgeJobs)
}
