package stacks

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
)

var (
	errStackAlreadyExists = errors.New("A stack already exists with this name")
	errStackNotExternal   = errors.New("Not an external stack")
)

// Handler is the HTTP handler used to handle stack operations.
type Handler struct {
	stackCreationMutex *sync.Mutex
	stackDeletionMutex *sync.Mutex
	requestBouncer     *security.RequestBouncer
	*mux.Router
	DataStore           portainer.DataStore
	FileService         portainer.FileService
	GitService          portainer.GitService
	SwarmStackManager   portainer.SwarmStackManager
	ComposeStackManager portainer.ComposeStackManager
	KubernetesDeployer  portainer.KubernetesDeployer
}

// NewHandler creates a handler to manage stack operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:             mux.NewRouter(),
		stackCreationMutex: &sync.Mutex{},
		stackDeletionMutex: &sync.Mutex{},
		requestBouncer:     bouncer,
	}
	h.Handle("/stacks",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.stackCreate))).Methods(http.MethodPost)
	h.Handle("/stacks",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.stackList))).Methods(http.MethodGet)
	h.Handle("/stacks/{id}",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.stackInspect))).Methods(http.MethodGet)
	h.Handle("/stacks/{id}",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.stackDelete))).Methods(http.MethodDelete)
	h.Handle("/stacks/{id}",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.stackUpdate))).Methods(http.MethodPut)
	h.Handle("/stacks/{id}/file",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.stackFile))).Methods(http.MethodGet)
	h.Handle("/stacks/{id}/migrate",
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.stackMigrate))).Methods(http.MethodPost)
	return h
}

func (handler *Handler) userCanAccessStack(securityContext *security.RestrictedRequestContext, endpointID portainer.EndpointID, resourceControl *portainer.ResourceControl) (bool, error) {
	if securityContext.IsAdmin {
		return true, nil
	}

	userTeamIDs := make([]portainer.TeamID, 0)
	for _, membership := range securityContext.UserMemberships {
		userTeamIDs = append(userTeamIDs, membership.TeamID)
	}

	if resourceControl != nil && authorization.UserCanAccessResource(securityContext.UserID, userTeamIDs, resourceControl) {
		return true, nil
	}

	_, err := handler.DataStore.Extension().Extension(portainer.RBACExtension)
	if err == bolterrors.ErrObjectNotFound {
		return false, nil
	} else if err != nil && err != bolterrors.ErrObjectNotFound {
		return false, err
	}

	user, err := handler.DataStore.User().User(securityContext.UserID)
	if err != nil {
		return false, err
	}

	_, ok := user.EndpointAuthorizations[endpointID][portainer.EndpointResourcesAccess]
	if ok {
		return true, nil
	}
	return false, nil
}

func (handler *Handler) userIsAdminOrEndpointAdmin(user *portainer.User, endpointID portainer.EndpointID) (bool, error) {
	isAdmin := user.Role == portainer.AdministratorRole
	if isAdmin {
		return true, nil
	}

	rbacExtension, err := handler.DataStore.Extension().Extension(portainer.RBACExtension)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return false, errors.New("Unable to verify if RBAC extension is loaded")
	}

	if rbacExtension == nil {
		return false, nil
	}

	_, endpointResourceAccess := user.EndpointAuthorizations[portainer.EndpointID(endpointID)][portainer.EndpointResourcesAccess]

	return endpointResourceAccess, nil
}
