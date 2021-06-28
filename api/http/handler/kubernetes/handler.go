package kubernetes

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/authorization"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

// Handler is the HTTP handler which will natively deal with to external endpoints.
type Handler struct {
	*mux.Router
	DataStore               portainer.DataStore
	KubernetesClientFactory *cli.ClientFactory
	authorizationService    *authorization.Service
}

// NewHandler creates a handler to process pre-proxied requests to external APIs.
func NewHandler(bouncer *security.RequestBouncer, authorizationService *authorization.Service) *Handler {
	h := &Handler{
		Router:               mux.NewRouter(),
		authorizationService: authorizationService,
	}
	h.PathPrefix("/kubernetes/{id}/nodes_limits").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.getKubernetesNodesLimits))).Methods(http.MethodGet)
	return h
}
