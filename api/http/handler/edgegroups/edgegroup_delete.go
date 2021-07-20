package edgegroups

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

// @id EdgeGroupDelete
// @summary Deletes an EdgeGroup
// @description
// @tags edge_groups
// @security jwt
// @accept json
// @produce json
// @param id path int true "EdgeGroup Id"
// @success 204
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_groups/{id} [delete]
func (handler *Handler) edgeGroupDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Edge group identifier route variable", err}
	}

	_, err = handler.DataStore.EdgeGroup().EdgeGroup(portainer.EdgeGroupID(edgeGroupID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an Edge group with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an Edge group with the specified identifier inside the database", err}
	}

	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Edge stacks from the database", err}
	}

	for _, edgeStack := range edgeStacks {
		for _, groupID := range edgeStack.EdgeGroups {
			if groupID == portainer.EdgeGroupID(edgeGroupID) {
				return &httperror.HandlerError{http.StatusForbidden, "Edge group is used by an Edge stack", errors.New("Edge group is used by an Edge stack")}
			}
		}
	}

	err = handler.DataStore.EdgeGroup().DeleteEdgeGroup(portainer.EdgeGroupID(edgeGroupID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove the Edge group from the database", err}
	}

	return response.Empty(w)

}
