package ssl

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type sslUpdatePayload struct {
	Cert        *string
	Key         *string
	HTTPEnabled *bool
}

func (payload *sslUpdatePayload) Validate(r *http.Request) error {
	if (payload.Cert == nil || payload.Key == nil) && payload.Cert != payload.Key {
		return errors.New("both certificate and key files should be provided")
	}

	return nil
}

// @id SSLUpdate
// @summary Update the ssl settings
// @description Update the ssl settings.
// @description **Access policy**: administrator
// @tags ssl
// @security jwt
// @accept json
// @produce json
// @param body body sslUpdatePayload true "SSL Settings"
// @success 204 "Success"
// @failure 400 "Invalid request"
// @failure 403 "Permission denied to access settings"
// @failure 500 "Server error"
// @router /ssl [put]
func (handler *Handler) sslUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload sslUpdatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	if payload.Cert != nil {
		err = handler.SSLService.SetCertificates([]byte(*payload.Cert), []byte(*payload.Key))
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Failed to save certificate", err}
		}
	}

	if payload.HTTPEnabled != nil {
		err = handler.SSLService.SetHTTPEnabled(*payload.HTTPEnabled)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Failed to force https", err}
		}
	}

	return response.Empty(w)
}
