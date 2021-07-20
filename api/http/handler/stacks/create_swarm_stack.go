package stacks

import (
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

type swarmStackFromFileContentPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Swarm cluster identifier
	SwarmID string `example:"jpofkc0i9uo9wtx1zesuk649w" validate:"required"`
	// Content of the Stack file
	StackFileContent string `example:"version: 3\n services:\n web:\n image:nginx" validate:"required"`
	// A list of environment variables used during stack deployment
	Env []portainer.Pair
}

func (payload *swarmStackFromFileContentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.SwarmID) {
		return errors.New("Invalid Swarm ID")
	}
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	return nil
}

func (handler *Handler) createSwarmStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {
	var payload swarmStackFromFileContentPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueName(endpoint, payload.Name, 0, true)
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
		Type:         portainer.DockerSwarmStack,
		SwarmID:      payload.SwarmID,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		Env:          payload.Env,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Compose file on disk", err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createSwarmDeployConfig(r, stack, endpoint, false)
	if configErr != nil {
		return configErr
	}

	err = handler.deploySwarmStack(config)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, err.Error(), err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the stack inside the database", err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type swarmStackFromGitRepositoryPayload struct {
	// Name of the stack
	Name string `example:"myStack" validate:"required"`
	// Swarm cluster identifier
	SwarmID string `example:"jpofkc0i9uo9wtx1zesuk649w" validate:"required"`
	// A list of environment variables used during stack deployment
	Env []portainer.Pair

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
}

func (payload *swarmStackFromGitRepositoryPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Name) {
		return errors.New("Invalid stack name")
	}
	if govalidator.IsNull(payload.SwarmID) {
		return errors.New("Invalid Swarm ID")
	}
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && (govalidator.IsNull(payload.RepositoryUsername) || govalidator.IsNull(payload.RepositoryPassword)) {
		return errors.New("Invalid repository credentials. Username and password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.ComposeFilePathInRepository) {
		payload.ComposeFilePathInRepository = filesystem.ComposeFileDefaultName
	}
	return nil
}

func (handler *Handler) createSwarmStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {
	var payload swarmStackFromGitRepositoryPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueName(endpoint, payload.Name, 0, true)
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
		Type:         portainer.DockerSwarmStack,
		SwarmID:      payload.SwarmID,
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

	config, configErr := handler.createSwarmDeployConfig(r, stack, endpoint, false)
	if configErr != nil {
		return configErr
	}

	err = handler.deploySwarmStack(config)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, err.Error(), err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the stack inside the database", err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type swarmStackFromFileUploadPayload struct {
	Name             string
	SwarmID          string
	StackFileContent []byte
	Env              []portainer.Pair
}

func (payload *swarmStackFromFileUploadPayload) Validate(r *http.Request) error {
	name, err := request.RetrieveMultiPartFormValue(r, "Name", false)
	if err != nil {
		return errors.New("Invalid stack name")
	}
	payload.Name = name

	swarmID, err := request.RetrieveMultiPartFormValue(r, "SwarmID", false)
	if err != nil {
		return errors.New("Invalid Swarm ID")
	}
	payload.SwarmID = swarmID

	composeFileContent, _, err := request.RetrieveMultiPartFormFile(r, "file")
	if err != nil {
		return errors.New("Invalid Compose file. Ensure that the Compose file is uploaded correctly")
	}
	payload.StackFileContent = composeFileContent

	var env []portainer.Pair
	err = request.RetrieveMultiPartFormJSONValue(r, "Env", &env, true)
	if err != nil {
		return errors.New("Invalid Env parameter")
	}
	payload.Env = env
	return nil
}

func (handler *Handler) createSwarmStackFromFileUpload(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint, userID portainer.UserID) *httperror.HandlerError {
	payload := &swarmStackFromFileUploadPayload{}
	err := payload.Validate(r)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	payload.Name = handler.SwarmStackManager.NormalizeStackName(payload.Name)

	isUnique, err := handler.checkUniqueName(endpoint, payload.Name, 0, true)
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
		Type:         portainer.DockerSwarmStack,
		SwarmID:      payload.SwarmID,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ComposeFileDefaultName,
		Env:          payload.Env,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist Compose file on disk", err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	config, configErr := handler.createSwarmDeployConfig(r, stack, endpoint, false)
	if configErr != nil {
		return configErr
	}

	err = handler.deploySwarmStack(config)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, err.Error(), err}
	}

	stack.CreatedBy = config.user.Username

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist the stack inside the database", err}
	}

	doCleanUp = false
	return handler.decorateStackResponse(w, stack, userID)
}

type swarmStackDeploymentConfig struct {
	stack      *portainer.Stack
	endpoint   *portainer.Endpoint
	registries []portainer.Registry
	prune      bool
	isAdmin    bool
	user       *portainer.User
}

func (handler *Handler) createSwarmDeployConfig(r *http.Request, stack *portainer.Stack, endpoint *portainer.Endpoint, prune bool) (*swarmStackDeploymentConfig, *httperror.HandlerError) {
	securityContext, err := security.RetrieveRestrictedRequestContext(r)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve info from request context", err}
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to load user information from the database", err}
	}

	registries, err := handler.DataStore.Registry().Registries()
	if err != nil {
		return nil, &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve registries from the database", err}
	}
	filteredRegistries := security.FilterRegistries(registries, user, securityContext.UserMemberships, endpoint.ID)

	config := &swarmStackDeploymentConfig{
		stack:      stack,
		endpoint:   endpoint,
		registries: filteredRegistries,
		prune:      prune,
		isAdmin:    securityContext.IsAdmin,
		user:       user,
	}

	return config, nil
}

func (handler *Handler) deploySwarmStack(config *swarmStackDeploymentConfig) error {
	isAdminOrEndpointAdmin, err := handler.userIsAdminOrEndpointAdmin(config.user, config.endpoint.ID)
	if err != nil {
		return err
	}

	settings := &config.endpoint.SecuritySettings

	if !settings.AllowBindMountsForRegularUsers && !isAdminOrEndpointAdmin {
		composeFilePath := path.Join(config.stack.ProjectPath, config.stack.EntryPoint)

		stackContent, err := handler.FileService.GetFileContent(composeFilePath)
		if err != nil {
			return err
		}

		err = handler.isValidStackFile(stackContent, settings)
		if err != nil {
			return err
		}
	}

	handler.stackCreationMutex.Lock()
	defer handler.stackCreationMutex.Unlock()

	handler.SwarmStackManager.Login(config.registries, config.endpoint)

	err = handler.SwarmStackManager.Deploy(config.stack, config.prune, config.endpoint)
	if err != nil {
		return err
	}

	err = handler.SwarmStackManager.Logout(config.endpoint)
	if err != nil {
		return err
	}

	return nil
}
