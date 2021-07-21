package stacks

import (
	"log"
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/portainer/libhttp/response"

	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/stacks"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
)

func (handler *Handler) webhookInvoke(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	webhookID, err := retrieveUUIDRouteVariableValue(r, "webhookID")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid webhook identifier route variable", Err: err}
	}

	stack, err := handler.DataStore.Stack().StackByWebhookID(webhookID.String())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == bolterrors.ErrObjectNotFound {
			statusCode = http.StatusNotFound
		}
		return &httperror.HandlerError{StatusCode: statusCode, Message: "Unable to find the stack by webhook ID", Err: err}
	}

	if err = stacks.RedeployWhenChanged(stack.ID, handler.StackDeployer, handler.DataStore, handler.GitService); err != nil {
		log.Printf("[ERROR] %s\n", err)
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Failed to update the stack", Err: err}
	}

	return response.Empty(w)
}

func retrieveUUIDRouteVariableValue(r *http.Request, name string) (uuid.UUID, error) {
	webhookID, err := request.RetrieveRouteVariableValue(r, name)
	if err != nil {
		return uuid.Nil, err
	}

	uid, err := uuid.FromString(webhookID)

	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}
