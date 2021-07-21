package stacks

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/filesystem"
	httperrors "github.com/portainer/portainer/api/http/errors"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/stackutils"
)

type stackGitRedployPayload struct {
	RepositoryReferenceName  string
	RepositoryAuthentication bool
	RepositoryUsername       string
	RepositoryPassword       string
	Env                      []portainer.Pair
}

func (payload *stackGitRedployPayload) Validate(r *http.Request) error {
	if payload.RepositoryAuthentication && (govalidator.IsNull(payload.RepositoryUsername) || govalidator.IsNull(payload.RepositoryPassword)) {
		return errors.New("Invalid repository credentials. Username and password must be specified when authentication is enabled")
	}
	return nil
}

// PUT request on /api/stacks/:id/git?endpointId=<endpointId>
func (handler *Handler) stackGitRedeploy(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid stack identifier route variable", Err: err}
	}

	stack, err := handler.DataStore.Stack().Stack(portainer.StackID(stackID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find a stack with the specified identifier inside the database", Err: err}
	}

	if stack.GitConfig == nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Stack is not created from git", Err: err}
	}

	// TODO: this is a work-around for stacks created with Portainer version >= 1.17.1
	// The EndpointID property is not available for these stacks, this API endpoint
	// can use the optional EndpointID query parameter to associate a valid endpoint identifier to the stack.
	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", true)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid query parameter: endpointId", Err: err}
	}
	if endpointID != int(stack.EndpointID) {
		stack.EndpointID = portainer.EndpointID(endpointID)
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(stack.EndpointID)
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusNotFound, Message: "Unable to find the endpoint associated to the stack inside the database", Err: err}
	} else if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to find the endpoint associated to the stack inside the database", Err: err}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Permission denied to access endpoint", Err: err}
	}

	resourceControl, err := handler.DataStore.ResourceControl().ResourceControlByResourceIDAndType(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portainer.StackResourceControl)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve a resource control associated to the stack", Err: err}
	}

	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve info from request context", Err: err}
	}

	access, err := handler.userCanAccessStack(securityContext, endpoint.ID, resourceControl)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to verify user authorizations to validate stack access", Err: err}
	}
	if !access {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Access denied to resource", Err: httperrors.ErrResourceAccessDenied}
	}

	var payload stackGitRedployPayload
	err = request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	stack.GitConfig.ReferenceName = payload.RepositoryReferenceName
	stack.Env = payload.Env

	backupProjectPath := fmt.Sprintf("%s-old", stack.ProjectPath)
	err = filesystem.MoveDirectory(stack.ProjectPath, backupProjectPath)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to move git repository directory", Err: err}
	}

	repositoryUsername := payload.RepositoryUsername
	repositoryPassword := payload.RepositoryPassword
	if !payload.RepositoryAuthentication {
		repositoryUsername = ""
		repositoryPassword = ""
	}

	err = handler.GitService.CloneRepository(stack.ProjectPath, stack.GitConfig.URL, payload.RepositoryReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		restoreError := filesystem.MoveDirectory(backupProjectPath, stack.ProjectPath)
		if restoreError != nil {
			log.Printf("[WARN] [http,stacks,git] [error: %s] [message: failed restoring backup folder]", restoreError)
		}

		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to clone git repository", Err: err}
	}

	defer func() {
		err = handler.FileService.RemoveDirectory(backupProjectPath)
		if err != nil {
			log.Printf("[WARN] [http,stacks,git] [error: %s] [message: unable to remove git repository directory]", err)
		}
	}()

	httpErr := handler.deployStack(r, stack, endpoint)
	if httpErr != nil {
		return httpErr
	}

	err = handler.DataStore.Stack().UpdateStack(stack.ID, stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack changes inside the database", Err: err}
	}

	return response.JSON(w, stack)
}

func (handler *Handler) deployStack(r *http.Request, stack *portainer.Stack, endpoint *portainer.Endpoint) *httperror.HandlerError {
	if stack.Type == portainer.DockerSwarmStack {
		config, httpErr := handler.createSwarmDeployConfig(r, stack, endpoint, false)
		if httpErr != nil {
			return httpErr
		}

		err := handler.deploySwarmStack(config)
		if err != nil {
			return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
		}

		stack.UpdateDate = time.Now().Unix()
		stack.UpdatedBy = config.user.Username
		stack.Status = portainer.StackStatusActive

		return nil
	}

	config, httpErr := handler.createComposeDeployConfig(r, stack, endpoint)
	if httpErr != nil {
		return httpErr
	}

	err := handler.deployComposeStack(config)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
	}

	stack.UpdateDate = time.Now().Unix()
	stack.UpdatedBy = config.user.Username
	stack.Status = portainer.StackStatusActive

	return nil
}
