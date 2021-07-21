package stacks

import (
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
)

const defaultReferenceName = "refs/heads/master"

type kubernetesStringDeploymentPayload struct {
	ComposeFormat    bool
	Namespace        string
	StackFileContent string
}

type kubernetesGitDeploymentPayload struct {
	ComposeFormat            bool
	Namespace                string
	RepositoryURL            string
	RepositoryReferenceName  string
	RepositoryAuthentication bool
	RepositoryUsername       string
	RepositoryPassword       string
	FilePathInRepository     string
}

func (payload *kubernetesStringDeploymentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.StackFileContent) {
		return errors.New("Invalid stack file content")
	}
	if govalidator.IsNull(payload.Namespace) {
		return errors.New("Invalid namespace")
	}
	return nil
}

func (payload *kubernetesGitDeploymentPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Namespace) {
		return errors.New("Invalid namespace")
	}
	if govalidator.IsNull(payload.RepositoryURL) || !govalidator.IsURL(payload.RepositoryURL) {
		return errors.New("Invalid repository URL. Must correspond to a valid URL format")
	}
	if payload.RepositoryAuthentication && govalidator.IsNull(payload.RepositoryPassword) {
		return errors.New("Invalid repository credentials. Password must be specified when authentication is enabled")
	}
	if govalidator.IsNull(payload.FilePathInRepository) {
		return errors.New("Invalid file path in repository")
	}
	if govalidator.IsNull(payload.RepositoryReferenceName) {
		payload.RepositoryReferenceName = defaultReferenceName
	}
	return nil
}

type createKubernetesStackResponse struct {
	Output string `json:"Output"`
}

func (handler *Handler) createKubernetesStackFromFileContent(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint) *httperror.HandlerError {
	var payload kubernetesStringDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:           portainer.StackID(stackID),
		Type:         portainer.KubernetesStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   filesystem.ManifestFileDefaultName,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	stackFolder := strconv.Itoa(int(stack.ID))
	projectPath, err := handler.FileService.StoreStackFileFromBytes(stackFolder, stack.EntryPoint, []byte(payload.StackFileContent))
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist Kubernetes manifest file on disk", Err: err}
	}
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	output, err := handler.deployKubernetesStack(r, endpoint, payload.StackFileContent, payload.ComposeFormat, payload.Namespace)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to deploy Kubernetes stack", Err: err}
	}

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the Kubernetes stack inside the database", Err: err}
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}

	return response.JSON(w, resp)
}

func (handler *Handler) createKubernetesStackFromGitRepository(w http.ResponseWriter, r *http.Request, endpoint *portainer.Endpoint) *httperror.HandlerError {
	var payload kubernetesGitDeploymentPayload
	if err := request.DecodeAndValidateJSONPayload(r, &payload); err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	stackID := handler.DataStore.Stack().GetNextIdentifier()
	stack := &portainer.Stack{
		ID:           portainer.StackID(stackID),
		Type:         portainer.KubernetesStack,
		EndpointID:   endpoint.ID,
		EntryPoint:   payload.FilePathInRepository,
		Status:       portainer.StackStatusActive,
		CreationDate: time.Now().Unix(),
	}

	projectPath := handler.FileService.GetStackProjectPath(strconv.Itoa(int(stack.ID)))
	stack.ProjectPath = projectPath

	doCleanUp := true
	defer handler.cleanUp(stack, &doCleanUp)

	stackFileContent, err := handler.cloneManifestContentFromGitRepo(&payload, stack.ProjectPath)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Failed to process manifest from Git repository", Err: err}
	}

	output, err := handler.deployKubernetesStack(r, endpoint, stackFileContent, payload.ComposeFormat, payload.Namespace)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to deploy Kubernetes stack", Err: err}
	}

	err = handler.DataStore.Stack().CreateStack(stack)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist the stack inside the database", Err: err}
	}

	resp := &createKubernetesStackResponse{
		Output: output,
	}
	return response.JSON(w, resp)
}

func (handler *Handler) deployKubernetesStack(request *http.Request, endpoint *portainer.Endpoint, stackConfig string, composeFormat bool, namespace string) (string, error) {
	handler.stackCreationMutex.Lock()
	defer handler.stackCreationMutex.Unlock()

	if composeFormat {
		convertedConfig, err := handler.KubernetesDeployer.ConvertCompose(stackConfig)
		if err != nil {
			return "", err
		}
		stackConfig = string(convertedConfig)
	}

	return handler.KubernetesDeployer.Deploy(request, endpoint, stackConfig, namespace)

}

func (handler *Handler) cloneManifestContentFromGitRepo(gitInfo *kubernetesGitDeploymentPayload, projectPath string) (string, error) {
	repositoryUsername := gitInfo.RepositoryUsername
	repositoryPassword := gitInfo.RepositoryPassword
	if !gitInfo.RepositoryAuthentication {
		repositoryUsername = ""
		repositoryPassword = ""
	}

	err := handler.GitService.CloneRepository(projectPath, gitInfo.RepositoryURL, gitInfo.RepositoryReferenceName, repositoryUsername, repositoryPassword)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadFile(filepath.Join(projectPath, gitInfo.FilePathInRepository))
	if err != nil {
		return "", err
	}
	return string(content), nil
}
