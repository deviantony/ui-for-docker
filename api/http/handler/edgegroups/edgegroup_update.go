package edgegroups

import (
	"errors"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/internal/edge"
)

type edgeGroupUpdatePayload struct {
	Name         string
	Dynamic      bool
	TagIDs       []portainer.TagID
	Endpoints    []portainer.EndpointID
	PartialMatch *bool
}

func (payload *edgeGroupUpdatePayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid Edge group name")
	}
	if payload.Dynamic && (payload.TagIDs == nil || len(payload.TagIDs) == 0) {
		return errors.New("TagIDs is mandatory for a dynamic Edge group")
	}
	if !payload.Dynamic && (payload.Endpoints == nil || len(payload.Endpoints) == 0) {
		return errors.New("Endpoints is mandatory for a static Edge group")
	}
	return nil
}

// @id EgeGroupUpdate
// @summary Updates an EdgeGroup
// @description
// @tags edge_groups
// @security jwt
// @accept json
// @produce json
// @param id path int true "EdgeGroup Id"
// @param body body edgeGroupUpdatePayload true "EdgeGroup data"
// @success 200 {object} portainer.EdgeGroup
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_groups/{id} [put]
func (handler *Handler) edgeGroupUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeGroupID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Edge group identifier route variable", err}
	}

	var payload edgeGroupUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	edgeGroup, err := handler.DataStore.EdgeGroup().EdgeGroup(portainer.EdgeGroupID(edgeGroupID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an Edge group with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an Edge group with the specified identifier inside the database", err}
	}

	if payload.Name != "" {
		edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Edge groups from the database", err}
		}
		for _, edgeGroup := range edgeGroups {
			if edgeGroup.Name == payload.Name && edgeGroup.ID != portainer.EdgeGroupID(edgeGroupID) {
				return &httperror.HandlerError{http.StatusBadRequest, "Edge group name must be unique", errors.New("Edge group name must be unique")}
			}
		}

		edgeGroup.Name = payload.Name
	}
	endpoints, err := handler.DataStore.Endpoint().Endpoints()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoints from database", err}
	}

	endpointGroups, err := handler.DataStore.EndpointGroup().EndpointGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoint groups from database", err}
	}

	oldRelatedEndpoints := edge.EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)

	edgeGroup.Dynamic = payload.Dynamic
	if edgeGroup.Dynamic {
		edgeGroup.TagIDs = payload.TagIDs
	} else {
		endpointIDs := []portainer.EndpointID{}
		for _, endpointID := range payload.Endpoints {
			endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoint from the database", err}
			}

			if endpoint.Type == portainer.EdgeAgentOnDockerEnvironment || endpoint.Type == portainer.EdgeAgentOnKubernetesEnvironment {
				endpointIDs = append(endpointIDs, endpoint.ID)
			}
		}
		edgeGroup.Endpoints = endpointIDs
	}

	if payload.PartialMatch != nil {
		edgeGroup.PartialMatch = *payload.PartialMatch
	}

	err = handler.DataStore.EdgeGroup().UpdateEdgeGroup(edgeGroup.ID, edgeGroup)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Edge group changes inside the database", err}
	}

	newRelatedEndpoints := edge.EdgeGroupRelatedEndpoints(edgeGroup, endpoints, endpointGroups)
	endpointsToUpdate := append(newRelatedEndpoints, oldRelatedEndpoints...)

	for _, endpointID := range endpointsToUpdate {
		err = handler.updateEndpoint(endpointID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Endpoint relation changes inside the database", err}
		}
	}

	return response.JSON(w, edgeGroup)
}

func (handler *Handler) updateEndpoint(endpointID portainer.EndpointID) error {
	relation, err := handler.DataStore.EndpointRelation().EndpointRelation(endpointID)
	if err != nil {
		return err
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
	if err != nil {
		return err
	}

	endpointGroup, err := handler.DataStore.EndpointGroup().EndpointGroup(endpoint.GroupID)
	if err != nil {
		return err
	}

	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return err
	}

	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return err
	}

	edgeStackSet := map[portainer.EdgeStackID]bool{}

	endpointEdgeStacks := edge.EndpointRelatedEdgeStacks(endpoint, endpointGroup, edgeGroups, edgeStacks)
	for _, edgeStackID := range endpointEdgeStacks {
		edgeStackSet[edgeStackID] = true
	}

	relation.EdgeStacks = edgeStackSet

	return handler.DataStore.EndpointRelation().UpdateEndpointRelation(endpoint.ID, relation)
}
