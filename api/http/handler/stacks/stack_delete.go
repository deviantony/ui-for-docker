package stacks

import (
	"errors"
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/stackutils"
)

// @id StackDelete
// @summary Remove a stack
// @description Remove a stack.
// @description **Access policy**: restricted
// @tags stacks
// @security jwt
// @param id path int true "Stack identifier"
// @param external query boolean false "Set to true to delete an external stack. Only external Swarm stacks are supported"
// @param endpointId query int false "Endpoint identifier used to remove an external stack (required when external is set to true)"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied"
// @failure 404 "Not found"
// @failure 500 "Server error"
// @router /stacks/{id} [delete]
func (handler *Handler) stackDelete(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid stack identifier route variable", err}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve info from request context", err}
	}

	externalStack, _ := request.RetrieveBooleanQueryParameter(r, "external", true)
	if externalStack {
		return handler.deleteExternalStack(r, w, stackID, securityContext)
	}

	id, err := strconv.Atoi(stackID)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid stack identifier route variable", err}
	}

	stack, err := handler.DataStore.Stack().Stack(portainer.StackID(id))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find a stack with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find a stack with the specified identifier inside the database", err}
	}

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", true)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: endpointId", err}
	}

	isOrphaned := portainer.EndpointID(endpointID) != stack.EndpointID

	if isOrphaned && !securityContext.IsAdmin {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to remove orphaned stack", errors.New("Permission denied to remove orphaned stack")}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find the endpoint associated to the stack inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find the endpoint associated to the stack inside the database", err}
	}

	resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portainer.StackResourceControl)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve a resource control associated to the stack", err}
	}

	if !isOrphaned {
		err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint)
		if err != nil {
			return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access endpoint", err}
		}

		access, err := handler.userCanAccessStack(securityContext, endpoint.ID, resourceControl)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to verify user authorizations to validate stack access", err}
		}
		if !access {
			return &httperror.HandlerError{http.StatusForbidden, "Access denied to resource", httperrors.ErrResourceAccessDenied}
		}
	}

	err = handler.deleteStack(stack, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, err.Error(), err}
	}

	err = handler.DataStore.Stack().DeleteStack(portainer.StackID(id))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove the stack from the database", err}
	}

	if resourceControl != nil {
		err = handler.DataStore.ResourceControl().DeleteResourceControl(resourceControl.ID)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove the associated resource control from the database", err}
		}
	}

	err = handler.FileService.RemoveDirectory(stack.ProjectPath)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove stack files from disk", err}
	}

	return response.Empty(w)
}

func (handler *Handler) deleteExternalStack(r *http.Request, w http.ResponseWriter, stackName string, securityContext *security.RestrictedRequestContext) *httperror.HandlerError {
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: endpointId", err}
	}

	if !securityContext.IsAdmin {
		return &httperror.HandlerError{http.StatusUnauthorized, "Permission denied to delete the stack", httperrors.ErrUnauthorized}
	}

	stack, err := handler.DataStore.Stack().StackByName(stackName)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to check for stack existence inside the database", err}
	}
	if stack != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "A stack with this name exists inside the database. Cannot use external delete method", errors.New("A tag already exists with this name")}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find the endpoint associated to the stack inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find the endpoint associated to the stack inside the database", err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access endpoint", err}
	}

	stack = &portainer.Stack{
		Name: stackName,
		Type: portainer.DockerSwarmStack,
	}

	err = handler.deleteStack(stack, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to delete stack", err}
	}

	return response.Empty(w)
}

func (handler *Handler) deleteStack(stack *portainer.Stack, endpoint *portainer.Endpoint) error {
	if stack.Type == portainer.DockerSwarmStack {
		return handler.SwarmStackManager.Remove(stack, endpoint)
	}

	return handler.ComposeStackManager.Down(stack, endpoint)
}
