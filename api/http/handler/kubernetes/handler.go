package kubernetes

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

// Handler is the HTTP handler which will natively deal with to external endpoints.
type Handler struct {
	*mux.Router
	DataStore               portainer.DataStore
	KubernetesClientFactory *cli.ClientFactory
}

// NewHandler creates a handler to process pre-proxied requests to external APIs.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}
	h.PathPrefix("/kubernetes/{id}/config").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.getKubernetesConfig))).Methods(http.MethodGet)
	return h
}
