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

// @id RegistryUpdate
// @summary Update a registry
// @description Update a registry
// @description **Access policy**: administrator
// @tags registries
// @security jwt
// @accept json
// @produce json
// @param id path int true "Registry identifier"
// @param body body registryUpdatePayload true "Registry details"
// @success 200 {object} portainer.Registry "Success"
// @failure 400 "Invalid request"
// @failure 404 "Registry not found"
// @failure 409 "Another registry with the same URL already exists"
// @failure 500 "Server error"
// @router /registries/{id} [put]

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
