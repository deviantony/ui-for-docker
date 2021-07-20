package edgejobs

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
)

// @id EdgeJobCreate
// @summary Create an EdgeJob
// @description
// @tags edge_jobs
// @security jwt
// @accept json
// @produce json
// @param method query string true "Creation Method" Enums(file, string)
// @param body_string body edgeJobCreateFromFileContentPayload true "EdgeGroup data when method is string"
// @param body_file body edgeJobCreateFromFilePayload true "EdgeGroup data when method is file"
// @success 200 {object} portainer.EdgeGroup
// @failure 503 "Edge compute features are disabled"
// @failure 500
// @router /edge_jobs [post]
func (handler *Handler) edgeJobCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: method. Valid values are: file or string", err}
	}

	switch method {
	case "string":
		return handler.createEdgeJobFromFileContent(w, r)
	case "file":
		return handler.createEdgeJobFromFile(w, r)
	default:
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: method. Valid values are: file or string", errors.New(request.ErrInvalidQueryParameter)}
	}
}

type edgeJobCreateFromFileContentPayload struct {
	Name           string
	CronExpression string
	Recurring      bool
	Endpoints      []portainer.EndpointID
	FileContent    string
}

func (payload *edgeJobCreateFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid Edge job name")
	}

	if !govalidator.Matches(payload.Name, `^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`) {
		return errors.New("Invalid Edge job name format. Allowed characters are: [a-zA-Z0-9_.-]")
	}

	if govalidator.IsNull(payload.CronExpression) {
		return errors.New("Invalid cron expression")
	}

	if payload.Endpoints == nil || len(payload.Endpoints) == 0 {
		return errors.New("Invalid endpoints payload")
	}

	if govalidator.IsNull(payload.FileContent) {
		return errors.New("Invalid script file content")
	}

	return nil
}

func (handler *Handler) createEdgeJobFromFileContent(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload edgeJobCreateFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	edgeJob := handler.createEdgeJobObjectFromFileContentPayload(&payload)

	err = handler.addAndPersistEdgeJob(edgeJob, []byte(payload.FileContent))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to schedule Edge job", err}
	}

	return response.JSON(w, edgeJob)
}

type edgeJobCreateFromFilePayload struct {
	Name           string
	CronExpression string
	Recurring      bool
	Endpoints      []portainer.EndpointID
	File           []byte
}

func (payload *edgeJobCreateFromFilePayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return errors.New("Invalid Edge job name")
	}

	if !govalidator.Matches(name, `^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`) {
		return errors.New("Invalid Edge job name format. Allowed characters are: [a-zA-Z0-9_.-]")
	}
	payload.Name = name

	cronExpression, err := request.RetrieveMultiPartFormValue(r, "CronExpression", false)
	if err != nil {
		return errors.New("Invalid cron expression")
	}
	payload.CronExpression = cronExpression

	var endpoints []portainer.EndpointID
	err = request.RetrieveMultiPartFormJSONValue(r, "Endpoints", &endpoints, false)
	if err != nil {
		return errors.New("Invalid endpoints")
	}
	payload.Endpoints = endpoints

	file, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return errors.New("Invalid script file. Ensure that the file is uploaded correctly")
	}
	payload.File = file

	return nil
}

func (handler *Handler) createEdgeJobFromFile(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	payload := &edgeJobCreateFromFilePayload{}
	err := payload.Validate(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	edgeJob := handler.createEdgeJobObjectFromFilePayload(payload)

	err = handler.addAndPersistEdgeJob(edgeJob, payload.File)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to schedule Edge job", err}
	}

	return response.JSON(w, edgeJob)
}

func (handler *Handler) createEdgeJobObjectFromFilePayload(payload *edgeJobCreateFromFilePayload) *portainer.EdgeJob {
	edgeJobIdentifier := portainer.EdgeJobID(handler.DataStore.EdgeJob().GetNextIdentifier())

	endpoints := convertEndpointsToMetaObject(payload.Endpoints)

	edgeJob := &portainer.EdgeJob{
		ID:             edgeJobIdentifier,
		Name:           payload.Name,
		CronExpression: payload.CronExpression,
		Recurring:      payload.Recurring,
		Created:        time.Now().Unix(),
		Endpoints:      endpoints,
		Version:        1,
	}

	return edgeJob
}

func (handler *Handler) createEdgeJobObjectFromFileContentPayload(payload *edgeJobCreateFromFileContentPayload) *portainer.EdgeJob {
	edgeJobIdentifier := portainer.EdgeJobID(handler.DataStore.EdgeJob().GetNextIdentifier())

	endpoints := convertEndpointsToMetaObject(payload.Endpoints)

	edgeJob := &portainer.EdgeJob{
		ID:             edgeJobIdentifier,
		Name:           payload.Name,
		CronExpression: payload.CronExpression,
		Recurring:      payload.Recurring,
		Created:        time.Now().Unix(),
		Endpoints:      endpoints,
		Version:        1,
	}

	return edgeJob
}

func (handler *Handler) addAndPersistEdgeJob(edgeJob *portainer.EdgeJob, file []byte) error {
	edgeCronExpression := strings.Split(edgeJob.CronExpression, " ")
	if len(edgeCronExpression) == 6 {
		edgeCronExpression = edgeCronExpression[1:]
	}
	edgeJob.CronExpression = strings.Join(edgeCronExpression, " ")

	for ID := range edgeJob.Endpoints {
		endpoint, err := handler.DataStore.Endpoint().Endpoint(ID)
		if err != nil {
			return err
		}

		if endpoint.Type != portainer.EdgeAgentOnDockerEnvironment && endpoint.Type != portainer.EdgeAgentOnKubernetesEnvironment {
			delete(edgeJob.Endpoints, ID)
		}
	}

	if len(edgeJob.Endpoints) == 0 {
		return errors.New("Endpoints are mandatory for an Edge job")
	}

	scriptPath, err := handler.FileService.StoreEdgeJobFileFromBytes(strconv.Itoa(int(edgeJob.ID)), file)
	if err != nil {
		return err
	}
	edgeJob.ScriptPath = scriptPath

	for endpointID := range edgeJob.Endpoints {
		handler.ReverseTunnelService.AddEdgeJob(endpointID, edgeJob)
	}

	return handler.DataStore.EdgeJob().CreateEdgeJob(edgeJob)
}

func convertEndpointsToMetaObject(endpoints []portainer.EndpointID) map[portainer.EndpointID]portainer.EdgeJobEndpointMeta {
	endpointsMap := map[portainer.EndpointID]portainer.EdgeJobEndpointMeta{}

	for _, endpointID := range endpoints {
		endpointsMap[endpointID] = portainer.EdgeJobEndpointMeta{}
	}

	return endpointsMap
}
