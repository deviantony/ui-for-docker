package edgegroups

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
)

// @id EdgeGroupInspect
// @summary Inspects an EdgeGroup
// @description
// @tags edge_groups
// @security jwt
// @accept json
// @produce json
// @param id path int true "EdgeGroup Id"
// @success 200 {object} portainer.EdgeGroup
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_groups/{id} [get]
func (handler *Handler) edgeGroupInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Edge group identifier route variable", err}
	}

	edgeGroup, err := handler.DataStore.EdgeGroup().EdgeGroup(portainer.EdgeGroupID(edgeGroupID))
	if err == errors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an Edge group with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an Edge group with the specified identifier inside the database", err}
	}

	if edgeGroup.Dynamic {
		endpoints, err := handler.getEndpointsByTags(edgeGroup.TagIDs, edgeGroup.PartialMatch)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoints and endpoint groups for Edge group", err}
		}

		edgeGroup.Endpoints = endpoints
	}

	return response.JSON(w, edgeGroup)
}
