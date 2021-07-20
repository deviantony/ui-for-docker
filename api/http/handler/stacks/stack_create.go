package stacks

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	gittypes "github.com/portainer/portainer/api/git/types"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/internal/endpointutils"
	"github.com/portainer/portainer/api/internal/stackutils"
)

func (handler *Handler) cleanUp(stack *portainer.Stack, doCleanUp *bool) error {
	if !*doCleanUp {
		return nil
	}

	err := handler.FileService.RemoveDirectory(stack.ProjectPath)
	if err != nil {
		log.Printf("http error: Unable to cleanup stack creation (err=%s)\n", err)
	}
	return nil
}

// @id StackCreate
// @summary Deploy a new stack
// @description Deploy a new stack into a Docker environment specified via the endpoint identifier.
// @description **Access policy**: restricted
// @tags stacks
// @security jwt
// @accept json,multipart/form-data
// @produce json
// @param type query int true "Stack deployment type. Possible values: 1 (Swarm stack) or 2 (Compose stack)." Enums(1,2)
// @param method query string true "Stack deployment method. Possible values: file, string or repository." Enums(string, file, repository)
// @param endpointId query int true "Identifier of the endpoint that will be used to deploy the stack"
// @param body_swarm_string body swarmStackFromFileContentPayload false "Required when using method=string and type=1"
// @param body_swarm_repository body swarmStackFromGitRepositoryPayload false "Required when using method=repository and type=1"
// @param body_compose_string body composeStackFromFileContentPayload false "Required when using method=string and type=2"
// @param body_compose_repository body composeStackFromGitRepositoryPayload false "Required when using method=repository and type=2"
// @param Name formData string false "Name of the stack. required when method is file"
// @param SwarmID formData string false "Swarm cluster identifier. Required when method equals file and type equals 1. required when method is file"
// @param Env formData string false "Environment variables passed during deployment, represented as a JSON array [{'name': 'name', 'value': 'value'}]. Optional, used when method equals file and type equals 1."
// @param file formData file false "Stack file. required when method is file"
// @success 200 {object} portainer.CustomTemplate
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /stacks [post]
func (handler *Handler) stackCreate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	stackType, err := request.RetrieveNumericQueryParameter(r, "type", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: type", err}
	}

	method, err := request.RetrieveQueryParameter(r, "method", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: method", err}
	}

	endpointID, err := request.RetrieveNumericQueryParameter(r, "endpointId", false)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid query parameter: endpointId", err}
	}

	endpoint, err := handler.DataStore.Endpoint().Endpoint(portainer.EndpointID(endpointID))
	if err == bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{http.StatusNotFound, "Unable to find an endpoint with the specified identifier inside the database", err}
	} else if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to find an endpoint with the specified identifier inside the database", err}
	}

	if endpointutils.IsDockerEndpoint(endpoint) && !endpoint.SecuritySettings.AllowStackManagementForRegularUsers {
		securityContext, err := security.RetrieveRestrictedRequestContext(r)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user info from request context", err}
		}

		canCreate, err := handler.userCanCreateStack(securityContext, portainer.EndpointID(endpointID))

		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to verify user authorizations to validate stack creation", err}
		}

		if !canCreate {
			errMsg := "Stack creation is disabled for non-admin users"
			return &httperror.HandlerError{http.StatusForbidden, errMsg, errors.New(errMsg)}
		}
	}

	err = handler.requestBouncer.AuthorizedEndpointOperation(r, endpoint)
	if err != nil {
		return &httperror.HandlerError{http.StatusForbidden, "Permission denied to access endpoint", err}
	}

	tokenData, err := security.RetrieveTokenData(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve user details from authentication token", err}
	}

	switch portainer.StackType(stackType) {
	case portainer.DockerSwarmStack:
		return handler.createSwarmStack(w, r, method, endpoint, tokenData.ID)
	case portainer.DockerComposeStack:
		return handler.createComposeStack(w, r, method, endpoint, tokenData.ID)
	case portainer.KubernetesStack:
		return handler.createKubernetesStack(w, r, method, endpoint)
	}

	return &httperror.HandlerError{http.StatusBadRequest, "Invalid value for query parameter: type. Value must be one of: 1 (Swarm stack) or 2 (Compose stack)", errors.New(request.ErrInvalidQueryParameter)}
}

func (handler *Handler) createComposeStack(w http.ResponseWriter, r *http.Request, method string, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {

	switch method {
	case "string":
		return handler.createComposeStackFromFileContent(w, r, endpoint, userID)
	case "repository":
		return handler.createComposeStackFromGitRepository(w, r, endpoint, userID)
	case "file":
		return handler.createComposeStackFromFileUpload(w, r, endpoint, userID)
	}

	return &httperror.HandlerError{http.StatusBadRequest, "Invalid value for query parameter: method. Value must be one of: string, repository or file", errors.New(request.ErrInvalidQueryParameter)}
}

func (handler *Handler) createSwarmStack(w http.ResponseWriter, r *http.Request, method string, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {
	switch method {
	case "string":
		return handler.createSwarmStackFromFileContent(w, r, endpoint, userID)
	case "repository":
		return handler.createSwarmStackFromGitRepository(w, r, endpoint, userID)
	case "file":
		return handler.createSwarmStackFromFileUpload(w, r, endpoint, userID)
	}

	return &httperror.HandlerError{http.StatusBadRequest, "Invalid value for query parameter: method. Value must be one of: string, repository or file", errors.New(request.ErrInvalidQueryParameter)}
}

func (handler *Handler) createKubernetesStack(w http.ResponseWriter, r *http.Request, method string, endpoint *portainer.Endpoint) *httperror.HandlerError {
	switch method {
	case "string":
		return handler.createKubernetesStackFromFileContent(w, r, endpoint)
	case "repository":
		return handler.createKubernetesStackFromGitRepository(w, r, endpoint)
	}
	return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid value for query parameter: method. Value must be one of: string or repository", Err: errors.New(request.ErrInvalidQueryParameter)}
}

func (handler *Handler) isValidStackFile(stackFileContent []byte, securitySettings *portainer.EndpointSecuritySettings) error {
	composeConfigYAML, err := loader.ParseYAML(stackFileContent)
	if err != nil {
		return err
	}

	composeConfigFile := types.ConfigFile{
		Config: composeConfigYAML,
	}

	composeConfigDetails := types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{composeConfigFile},
		Environment: map[string]string{},
	}

	composeConfig, err := loader.Load(composeConfigDetails, func(options *loader.Options) {
		options.SkipValidation = true
		options.SkipInterpolation = true
	})
	if err != nil {
		return err
	}

	for key := range composeConfig.Services {
		service := composeConfig.Services[key]
		if !securitySettings.AllowBindMountsForRegularUsers {
			for _, volume := range service.Volumes {
				if volume.Type == "bind" {
					return errors.New("bind-mount disabled for non administrator users")
				}
			}
		}

		if !securitySettings.AllowPrivilegedModeForRegularUsers && service.Privileged == true {
			return errors.New("privileged mode disabled for non administrator users")
		}

		if !securitySettings.AllowHostNamespaceForRegularUsers && service.Pid == "host" {
			return errors.New("pid host disabled for non administrator users")
		}

		if !securitySettings.AllowDeviceMappingForRegularUsers && service.Devices != nil && len(service.Devices) > 0 {
			return errors.New("device mapping disabled for non administrator users")
		}

		if !securitySettings.AllowSysctlSettingForRegularUsers && service.Sysctls != nil && len(service.Sysctls) > 0 {
			return errors.New("sysctl setting disabled for non administrator users")
		}

		if !securitySettings.AllowContainerCapabilitiesForRegularUsers && (len(service.CapAdd) > 0 || len(service.CapDrop) > 0) {
			return errors.New("container capabilities disabled for non administrator users")
		}
	}

	return nil
}

func (handler *Handler) decorateStackResponse(w http.ResponseWriter, stack *portainer.Stack, userID portainer.UserID) *httperror.HandlerError {
	var resourceControl *portainer.ResourceControl

	isAdmin, err := handler.userIsAdmin(userID)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to load user information from the database", err}
	}

	if isAdmin {
		resourceControl = authorization.NewAdministratorsOnlyResourceControl(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portainer.StackResourceControl)
	} else {
		resourceControl = authorization.NewPrivateResourceControl(stackutils.ResourceControlID(stack.EndpointID, stack.Name), portainer.StackResourceControl, userID)
	}

	err = handler.DataStore.ResourceControl().CreateResourceControl(resourceControl)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist resource control inside the database", err}
	}

	stack.ResourceControl = resourceControl
	return response.JSON(w, stack)
}

func (handler *Handler) cloneAndSaveConfig(stack *portainer.Stack, projectPath, repositoryURL, refName, configFilePath string, auth bool, username, password string) error {
	if !auth {
		username = ""
		password = ""
	}

	err := handler.GitService.CloneRepository(projectPath, repositoryURL, refName, username, password)
	if err != nil {
		return fmt.Errorf("unable to clone git repository: %w", err)
	}

	stack.GitConfig = &gittypes.RepoConfig{
		URL:            repositoryURL,
		ReferenceName:  refName,
		ConfigFilePath: configFilePath,
	}
	return nil
}
