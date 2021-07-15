package ssl

import (
	"net/http"

	"github.com/gorilla/mux"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/ssl"
)

// Handler is the HTTP handler used to handle MOTD operations.
type Handler struct {
	*mux.Router
	SSLService *ssl.Service
}

// NewHandler returns a new Handler
func NewHandler(bouncer *security.RequestBouncer) *Handler {
	h := &Handler{
		Router: mux.NewRouter(),
	}
	h.Handle("/ssl",
		bouncer.AdminAccess(httperror.LoggerHandler(h.sslInspect))).Methods(http.MethodGet)
	h.Handle("/ssl",
		bouncer.AdminAccess(httperror.LoggerHandler(h.sslUpdate))).Methods(http.MethodPut)

	return h
}
