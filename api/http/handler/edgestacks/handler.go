package edgestacks

import (
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/http/security"
)

// Handler is the HTTP handler used to handle endpoint group operations.
type Handler struct {
	*mux.Router
	requestBouncer     *security.RequestBouncer
	DataStore          portainer.DataStore
	FileService        portainer.FileService
	GitService         portainer.GitService
	KubernetesDeployer portainer.KubernetesDeployer
}

// NewHandler creates a handler to manage endpoint group operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:         mux.NewRouter(),
		requestBouncer: bouncer,
	}
	h.Handle("/edge_stacks",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackCreate)))).Methods(http.MethodPost)
	h.Handle("/edge_stacks",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackList)))).Methods(http.MethodGet)
	h.Handle("/edge_stacks/{id}",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackInspect)))).Methods(http.MethodGet)
	h.Handle("/edge_stacks/{id}",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackUpdate)))).Methods(http.MethodPut)
	h.Handle("/edge_stacks/{id}",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackDelete)))).Methods(http.MethodDelete)
	h.Handle("/edge_stacks/{id}/file",
		bouncer.AdminAccess(bouncer.EdgeComputeOperation(httperror.LoggerHandler(h.edgeStackFile)))).Methods(http.MethodGet)
	h.Handle("/edge_stacks/{id}/status",
		bouncer.PublicAccess(httperror.LoggerHandler(h.edgeStackStatusUpdate))).Methods(http.MethodPut)
	return h
}

func (handler *Handler) convertAndStoreKubeManifestIfNeeded(edgeStack *portainer.EdgeStack, relatedEndpointIds []portainer.EndpointID) error {
	hasKubeEndpoint, err := hasKubeEndpoint(handler.DataStore.Endpoint(), relatedEndpointIds)
	if err != nil {
		return fmt.Errorf("unable to check if edge stack has kube endpoints: %w", err)
	}

	if !hasKubeEndpoint {
		return nil
	}

	composeConfig, err := handler.FileService.GetFileContent(path.Join(edgeStack.ProjectPath, edgeStack.EntryPoint))
	if err != nil {
		return fmt.Errorf("unable to retrieve Compose file from disk: %w", err)
	}

	kompose, err := handler.KubernetesDeployer.ConvertCompose(string(composeConfig))
	if err != nil {
		return fmt.Errorf("failed converting compose file to kubernetes manifest: %w", err)
	}

	komposeFileName := filesystem.ManifestFileDefaultName
	_, err = handler.FileService.StoreEdgeStackFileFromBytes(strconv.Itoa(int(edgeStack.ID)), komposeFileName, kompose)
	if err != nil {
		return fmt.Errorf("failed to store kube manifest file: %w", err)
	}

	edgeStack.ManifestPath = komposeFileName

	return nil
}
