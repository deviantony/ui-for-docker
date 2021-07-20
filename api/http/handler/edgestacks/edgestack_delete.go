package edgestacks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/internal/edge"
)

// @id EdgeStackDelete
// @summary Delete an EdgeStack
// @description
// @tags edge_stacks
// @security jwt
// @accept json
// @produce json
// @param id path string true "EdgeStack Id"
// @success 204
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks/{id} [delete]
func (handler *Handler) edgeStackDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid edge stack identifier route variable", err}
	}

	edgeStack, err := handler.DataStore.EdgeStack().EdgeStack(portainer.EdgeStackID(edgeStackID))
	if err == errors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an edge stack with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an edge stack with the specified identifier inside the database", err}
	}

	err = handler.DataStore.EdgeStack().DeleteEdgeStack(portainer.EdgeStackID(edgeStackID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove the edge stack from the database", err}
	}

	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoints from database", err}
	}

	endpointGroups, err := handler.DataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoint groups from database", err}
	}

	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge groups from database", err}
	}

	relatedEndpoints, err := edge.EdgeStackRelatedEndpoints(edgeStack.EdgeGroups, endpoints, endpointGroups, edgeGroups)

	for _, endpointID := range relatedEndpoints {
		relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpointID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find endpoint relation in database", err}
		}

		delete(relation.EdgeStacks, edgeStack.ID)

		err = handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpointID, relation)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist endpoint relation in database", err}
		}
	}

	return response.Empty(w)
}
