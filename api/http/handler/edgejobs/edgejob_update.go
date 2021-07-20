package edgejobs

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

type edgeJobUpdatePayload struct {
	Name           *string
	CronExpression *string
	Recurring      *bool
	Endpoints      []portainer.EndpointID
	FileContent    *string
}

func (payload *edgeJobUpdatePayload) Validate(r *http.Request) error {
	if payload.Name != nil && !govalidator.Matches(*payload.Name, `^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`) {
		return errors.New("Invalid Edge job name format. Allowed characters are: [a-zA-Z0-9_.-]")
	}
	return nil
}

// @id EdgeJobUpdate
// @summary Update an EdgeJob
// @description
// @tags edge_jobs
// @security jwt
// @accept json
// @produce json
// @param id path string true "EdgeJob Id"
// @param body body edgeJobUpdatePayload true "EdgeGroup data"
// @success 200 {object} portainer.EdgeJob
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id} [post]
func (handler *Handler) edgeJobUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Edge job identifier route variable", err}
	}

	var payload edgeJobUpdatePayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	edgeJob, err := handler.DataStore.EdgeJob().EdgeJob(portainer.EdgeJobID(edgeJobID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an Edge job with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an Edge job with the specified identifier inside the database", err}
	}

	err = handler.updateEdgeSchedule(edgeJob, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update Edge job", err}
	}

	err = handler.DataStore.EdgeJob().UpdateEdgeJob(edgeJob.ID, edgeJob)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Edge job changes inside the database", err}
	}

	return response.JSON(w, edgeJob)
}

func (handler *Handler) updateEdgeSchedule(edgeJob *portainer.EdgeJob, payload *edgeJobUpdatePayload) error {
	if payload.Name != nil {
		edgeJob.Name = *payload.Name
	}

	if payload.Endpoints != nil {
		endpointsMap := map[portainer.EndpointID]portainer.EdgeJobEndpointMeta{}

		for _, endpointID := range payload.Endpoints {
			endpoint, err := handler.DataStore.Endpoint().Endpoint(endpointID)
			if err != nil {
				return err
			}

			if endpoint.Type != portainer.EdgeAgentOnDockerEnvironment && endpoint.Type != portainer.EdgeAgentOnKubernetesEnvironment {
				continue
			}

			if meta, ok := edgeJob.Endpoints[endpointID]; ok {
				endpointsMap[endpointID] = meta
			} else {
				endpointsMap[endpointID] = portainer.EdgeJobEndpointMeta{}
			}
		}

		edgeJob.Endpoints = endpointsMap
	}

	updateVersion := false
	if payload.CronExpression != nil {
		edgeJob.CronExpression = *payload.CronExpression
		updateVersion = true
	}

	if payload.FileContent != nil {
		_, err := handler.FileService.StoreEdgeJobFileFromBytes(strconv.Itoa(int(edgeJob.ID)), []byte(*payload.FileContent))
		if err != nil {
			return err
		}

		updateVersion = true
	}

	if payload.Recurring != nil {
		edgeJob.Recurring = *payload.Recurring
		updateVersion = true
	}

	if updateVersion {
		edgeJob.Version++
	}

	for endpointID := range edgeJob.Endpoints {
		handler.ReverseTunnelService.AddEdgeJob(endpointID, edgeJob)
	}

	return nil
}
