package ldap

import (
	"errors"
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	httperrors "github.com/portainer/portainer/api/http/errors"
)

type testLoginPayload struct {
	LDAPSettings portainer.LDAPSettings
	Username     string
	Password     string
}

type testLoginResponse struct {
	Valid bool `json:"valid"`
}

func (payload *testLoginPayload) Validate(r *http.Request) error {
	if len(payload.LDAPSettings.URL) == 0 {
		return errors.New("Invalid LDAP URL")
	}

	return nil
}

// POST request on /ldap/test
func (handler *Handler) ldapTestLogin(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload testLoginPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	settings := &payload.LDAPSettings

	err = handler.prefillSettings(settings)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to fetch default settings", err}
	}

	err = handler.LDAPService.AuthenticateUser(payload.Username, payload.Password, settings)
	if err != nil && err != httperrors.ErrUnauthorized {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to test user authorization", err}
	}

	return response.JSON(w, &testLoginResponse{Valid: err != httperrors.ErrUnauthorized})

}
