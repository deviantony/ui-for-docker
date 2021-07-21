package stacks

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/http/security"
)

type composeStackFromFileContentPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx" validate:"required"`
	// A list of environment variables used during stack deployment
	Env []portainer.Pair `example:""`
}

func (payload *composeStackFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}

	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

func (handler *Handler) createComposeStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {
	var payload composeStackFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	payload.Name = handler.ComposeStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueName(endpoint, payload.Name, 0, false)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to check for name collision", err}
	}
	if !isUnique {
		errorMessage := fmt.Sprintf("A stack with the name '%s' is already running", payload.Name)
		return &httperror.HandlerError{http.StatusConflict, errorMessage, errors.New(errorMessage)}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:           portainer.StackID(stackID),
		Name:         payload.Name,
		Type:         portainer.DockerComposeStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		Env:          payload.Env,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist Compose file on disk", Err: err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createComposeDeployConfig(r, stack, endpoint)
	if configErr != nil {
		return configErr
	}

	err = handler.deployComposeStack(config)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack inside the database", Err: err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type composeStackFromGitRepositoryPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`

	// URL of a Git repository hosting the Stack file
	RepositoryURL string `example:"https://github.com/openfaas/faas" validate:"required"`
	// Reference name of a Git repository hosting the Stack file
	RepositoryReferenceName string `example:"refs/heads/master"`
	// Use basic authentication to clone the Git repository
	RepositoryAuthentication bool `example:"true"`
	// Username used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryUsername string `example:"myGitUsername"`
	// Password used in basic authentication. Required when RepositoryAuthentication is true.
	RepositoryPassword string `example:"myGitPassword"`
	// Path to the Stack file inside the Git repository
	ComposeFilePathInRepository string `example:"docker-compose.yml" default:"docker-compose.yml"`

	// A list of environment variables used during stack deployment
	Env []portainer.Pair
}

func (payload *composeStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}

	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && (govalidator.IsNull(payload.RepositoryUsername) || govalidator.IsNull(payload.RepositoryPassword)) {
		return errors.New("Invalid repository credentials. Username and password must be specified when authentication is enabled")
	}

	return nil
}

func (handler *Handler) createComposeStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {
	var payload composeStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	payload.Name = handler.ComposeStackManager.NormalizeStackName(payload.Name)
	if payload.ComposeFilePathInRepository == "" {
		payload.ComposeFilePathInRepository = filesystem.ComposeFileDefaultName
	}

	isUnique, err := handler.checkUniqueName(endpoint, payload.Name, 0, false)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to check for name collision", err}
	}
	if !isUnique {
		errorMessage := fmt.Sprintf("A stack with the name '%s' already exists", payload.Name)
		return &httperror.HandlerError{http.StatusConflict, errorMessage, errors.New(errorMessage)}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:           portainer.StackID(stackID),
		Name:         payload.Name,
		Type:         portainer.DockerComposeStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   payload.ComposeFilePathInRepository,
		Env:          payload.Env,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	projectPath := handler.FileService.GetStackProjectPath(strconv.Itoa(int(stack.ID)))
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	err = handler.cloneAndSaveConfig(stack, projectPath, payload.RepositoryURL, payload.RepositoryReferenceName, payload.ComposeFilePathInRepository, payload.RepositoryAuthentication, payload.RepositoryUsername, payload.RepositoryPassword)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to clone git repository", Err: err}
	}

	config, configErr := handler.createComposeDeployConfig(r, stack, endpoint)
	if configErr != nil {
		return configErr
	}

	err = handler.deployComposeStack(config)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack inside the database", Err: err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type composeStackFromFileUploadPayload struct {
	Name             string
	StackFileContent []byte
	Env              []portainer.Pair
}

func decodeRequestForm(r *http.Request) (*composeStackFromFileUploadPayload, error) {
	payload := &composeStackFromFileUploadPayload{}
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return nil, errors.New("Invalid stack name")
	}
	payload.Name = name

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return nil, errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.StackFileContent = composeFileContent

	var env []portainer.Pair
	err = request.RetrieveMultiPartFormJSONValue(r, "Env", &env, true)
	if err != nil {
		return nil, errors.New("Invalid Env parameter")
	}
	payload.Env = env
	return payload, nil
}

func (handler *Handler) createComposeStackFromFileUpload(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {
	payload, err := decodeRequestForm(r)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	payload.Name = handler.ComposeStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueName(endpoint, payload.Name, 0, false)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to check for name collision", Err: err}
	}
	if !isUnique {
		errorMessage := fmt.Sprintf("A stack with the name '%s' already exists", payload.Name)
		return &httperror.HandlerError{StatusCode: http.StatusConflict, Message: errorMessage, Err: errors.New(errorMessage)}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:           portainer.StackID(stackID),
		Name:         payload.Name,
		Type:         portainer.DockerComposeStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		Env:          payload.Env,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, payload.StackFileContent)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist Compose file on disk", Err: err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createComposeDeployConfig(r, stack, endpoint)
	if configErr != nil {
		return configErr
	}

	err = handler.deployComposeStack(config)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: err.Error(), Err: err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack inside the database", Err: err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type composeStackDeploymentConfig struct {
	stack      *portainer.Stack
	endpoint   *portainer.Endpoint
	registries []portainer.Registry
	isAdmin    bool
	user       *portainer.User
}

func (handler *Handler) createComposeDeployConfig(r *http.Request, stack *portainer.Stack, endpoint *portainer.Endpoint) (*composeStackDeploymentConfig, *httperror.HandlerError) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve info from request context", Err: err}
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to load user information from the database", Err: err}
	}

	registries, err := handler.DataStore.Registry().Registries()
	if err != nil {
		return nil, &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve registries from the database", Err: err}
	}
	filteredRegistries := security.FilterRegistries(registries, user, securityContext.UserMemberships, endpoint.ID)

	config := &composeStackDeploymentConfig{
		stack:      stack,
		endpoint:   endpoint,
		registries: filteredRegistries,
		isAdmin:    securityContext.IsAdmin,
		user:       user,
	}

	return config, nil
}

// TODO: libcompose uses credentials store into a config.json file to pull images from
// private registries. Right now the only solution is to re-use the embedded Docker binary
// to login/logout, which will generate the required data in the config.json file and then
// clean it. Hence the use of the mutex.
// We should contribute to libcompose to support authentication without using the config.json file.
func (handler *Handler) deployComposeStack(config *composeStackDeploymentConfig) error {
	isAdminOrEndpointAdmin, err := handler.userIsAdminOrEndpointAdmin(config.user, config.endpoint.ID)
	if err != nil {
		return err
	}

	securitySettings := &config.endpoint.SecuritySettings

	if (!securitySettings.AllowBindMountsForRegularUsers ||
		!securitySettings.AllowPrivilegedModeForRegularUsers ||
		!securitySettings.AllowHostNamespaceForRegularUsers ||
		!securitySettings.AllowDeviceMappingForRegularUsers ||
		!securitySettings.AllowSysctlSettingForRegularUsers ||
		!securitySettings.AllowContainerCapabilitiesForRegularUsers) &&
		!isAdminOrEndpointAdmin {

		composeFilePath := path.Join(config.stack.ProjectPath, config.stack.EntryPoint)
		stackContent, err := handler.FileService.GetFileContent(composeFilePath)
		if err != nil {
			return err
		}

		err = handler.isValidStackFile(stackContent, securitySettings)
		if err != nil {
			return err
		}
	}

	handler.stackCreationMutex.Lock()
	defer handler.stackCreationMutex.Unlock()

	handler.SwarmStackManager.Login(config.registries, config.endpoint)

	err = handler.ComposeStackManager.Up(context.TODO(), config.stack, config.endpoint)
	if err != nil {
		return err
	}

	return handler.SwarmStackManager.Logout(config.endpoint)
}
