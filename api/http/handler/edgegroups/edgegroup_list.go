package edgegroups

import (
	"fmt"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
)

type decoratedEdgeGroup struct {
	portainer.EdgeGroup
	HasEdgeStack  bool `json:"HasEdgeStack"`
	EndpointTypes []portainer.EndpointType
}

// @id EdgeGroupList
// @summary list EdgeGroups
// @description
// @tags edge_groups
// @security jwt
// @accept json
// @produce json
// @success 200 {array} portainer.EdgeGroup{HasEdgeStack=bool} "EdgeGroups"
// @failure 500
// @failure 503 Edge compute features are disabled
// @router /edge_groups [get]
func (handler *Handler) edgeGroupList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeGroups, err := handler.DataStore.EdgeGroup().EdgeGroups()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Edge groups from the database", err}
	}

	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve Edge stacks from the database", err}
	}

	usedEdgeGroups := make(map[portainer.EdgeGroupID]bool)

	for _, stack := range edgeStacks {
		for _, groupID := range stack.EdgeGroups {
			usedEdgeGroups[groupID] = true
		}
	}

	decoratedEdgeGroups := []decoratedEdgeGroup{}
	for _, orgEdgeGroup := range edgeGroups {
		edgeGroup := decoratedEdgeGroup{
			EdgeGroup:     orgEdgeGroup,
			EndpointTypes: []portainer.EndpointType{},
		}
		if edgeGroup.Dynamic {
			endpointIDs, err := handler.getEndpointsByTags(edgeGroup.TagIDs, edgeGroup.PartialMatch)
			if err != nil {
				return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoints and endpoint groups for Edge group", err}
			}

			edgeGroup.Endpoints = endpointIDs
		}

		endpointTypes, err := getEndpointTypes(handler.DataStore.Endpoint(), edgeGroup.Endpoints)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve endpoint types for Edge group", err}
		}

		edgeGroup.EndpointTypes = endpointTypes

		edgeGroup.HasEdgeStack = usedEdgeGroups[edgeGroup.ID]

		decoratedEdgeGroups = append(decoratedEdgeGroups, edgeGroup)
	}

	return response.JSON(w, decoratedEdgeGroups)
}

func getEndpointTypes(endpointService portainer.EndpointService, endpointIds []portainer.EndpointID) ([]portainer.EndpointType, error) {
	typeSet := map[portainer.EndpointType]bool{}
	for _, endpointID := range endpointIds {
		endpoint, err := endpointService.Endpoint(endpointID)
		if err != nil {
			return nil, fmt.Errorf("failed fetching endpoint: %w", err)
		}

		typeSet[endpoint.Type] = true
	}

	endpointTypes := make([]portainer.EndpointType, 0, len(typeSet))
	for endpointType := range typeSet {
		endpointTypes = append(endpointTypes, endpointType)
	}

	return endpointTypes, nil
}
