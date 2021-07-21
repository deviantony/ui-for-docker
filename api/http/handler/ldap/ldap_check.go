package ldap

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
)

type checkPayload struct {
	LDAPSettings portainer.LDAPSettings
}

func (payload *checkPayload) Validate(r *http.Request) error {
	if len(payload.LDAPSettings.URL) == 0 {
		return errors.New("Invalid LDAP URL")
	}

	return nil
}

// POST request on /ldap/check
func (handler *Handler) ldapCheck(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload checkPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	settings := &payload.LDAPSettings

	err = handler.prefillSettings(settings)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to fetch default settings", Err: err}
	}

	err = handler.LDAPService.TestConnectivity(settings)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to connect to LDAP server", Err: err}
	}

	return response.Empty(w)
}
