package webhooks

import (
	"context"
	"errors"
	"net/http"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
)

// @summary Execute a webhook
// @description Acts on a passed in token UUID to restart the docker service
// @tags webhooks
// @accept json
// @produce json
// @param token path string true "Webhook token"
// @success 202 "Webhook executed"
// @failure 400
// @failure 500
// @router /webhooks/{token} [post]
func (handler *Handler) webhookExecute(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {

	webhookToken, err := request.RetrieveRouteVariableValue(r, "token")

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Invalid service id parameter", err}
	}

	webhook, err := handler.DataStore.Webhook().WebhookByToken(webhookToken)

	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a webhook with this token", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve webhook from the database", err}
	}

	resourceID := webhook.ResourceID
	endpointID := webhook.EndpointID
	webhookType := webhook.WebhookType

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an endpoint with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an endpoint with the specified identifier inside the database", err}
	}

	imageTag, _ := request.RetrieveQueryParameter(r, "tag", true)

	switch webhookType {
	case portainer.ServiceWebhook:
		return handler.executeServiceWebhook(w, endpoint, resourceID, imageTag)
	default:
		return &httperror.HandlerError{http.StatusInternalServerError, "Unsupported webhook type", errors.New("Webhooks for this resource are not currently supported")}
	}
}

func (handler *Handler) executeServiceWebhook(w http.ResponseWriter, endpoint *portainer.Endpoint, resourceID string, imageTag string) *httperror.HandlerError {
	dockerClient, err := handler.DockerClientFactory.CreateClient(endpoint, "")
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error creating docker client", err}
	}
	defer dockerClient.Close()

	service, _, err := dockerClient.ServiceInspectWithRaw(context.Background(), resourceID, dockertypes.ServiceInspectOptions{InsertDefaults: true})
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error looking up service", err}
	}

	service.Spec.TaskTemplate.ForceUpdate++

	if imageTag != "" {
		service.Spec.TaskTemplate.ContainerSpec.Image = string(service.Spec.TaskTemplate.ContainerSpec.Image[:strings.LastIndex(service.Spec.TaskTemplate.ContainerSpec.Image, ":")]) + ":" + imageTag
	} else {
		service.Spec.TaskTemplate.ContainerSpec.Image = strings.Split(service.Spec.TaskTemplate.ContainerSpec.Image, "@sha")[0]
	}

	_, err = dockerClient.ServiceUpdate(context.Background(), resourceID, service.Version, service.Spec, dockertypes.ServiceUpdateOptions{QueryRegistry: true})

	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Error updating service", err}
	}
	return response.Empty(w)
}
