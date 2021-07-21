package ssl

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id SSLInspect
// @summary Inspect the ssl settings
// @description Retrieve the ssl settings.
// @description **Access policy**: administrator
// @tags ssl
// @security jwt
// @produce json
// @success 200 {object} portainer.SSLSettings "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access settings"
// @failure 500 "Server error"
// @router /ssl [get]
func (handler *Handler) sslInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	settings, err := handler.SSLService.GetSSLSettings()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Failed to fetch certificate info", err}
	}

	return response.JSON(w, settings)
}
