package websocket

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	httperror "github.com/portainer/libhttp/error"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/kubernetes/cli"
)

// Handler is the HTTP handler used to handle websocket operations.
type Handler struct {
	*mux.Router
	DataStore               portainer.DataStore
	SignatureService        portainer.DigitalSignatureService
	ReverseTunnelService    portainer.ReverseTunnelService
	KubernetesClientFactory *cli.ClientFactory
	requestBouncer          *security.RequestBouncer
	connectionUpgrader      websocket.Upgrader
}

// NewHandler creates a handler to manage websocket operations.
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router:             mux.NewRouter(),
		connectionUpgrader: websocket.Upgrader{},
		requestBouncer:     bouncer,
	}
	h.PathPrefix("/websocket/exec").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketExec)))
	h.PathPrefix("/websocket/attach").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketAttach)))
	h.PathPrefix("/websocket/pod").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketPodExec)))
	h.PathPrefix("/websocket/kubernetes-shell").Handler(
		bouncer.AuthenticatedAccess(httperror.LoggerHandler(h.websocketShellPodExec)))
	return h
}
