package ssl

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

func (handler *Handler) sslInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	settings, err := handler.SSLService.GetSSLSettings()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Failed to fetch certificate info", err}
	}

	return response.JSON(w, settings)
}
