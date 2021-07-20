package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	portainer "github.com/portainer/portainer/api"
	bolterrors "github.com/portainer/portainer/api/bolt/errors"
	httperrors "github.com/portainer/portainer/api/http/errors"
)

type oauthPayload struct {
	// OAuth code returned from OAuth Provided
	Code string
}

func (payload *oauthPayload) Validate(r *http.Request) error {
	if govalidator.IsNull(payload.Code) {
		return errors.New("Invalid OAuth authorization code")
	}
	return nil
}

func (handler *Handler) authenticateOAuth(code string, settings *portainer.OAuthSettings) (string, error) {
	if code == "" {
		return "", errors.New("Invalid OAuth authorization code")
	}

	if settings == nil {
		return "", errors.New("Invalid OAuth configuration")
	}

	username, err := handler.OAuthService.Authenticate(code, settings)
	if err != nil {
		return "", err
	}

	return username, nil
}

// @id ValidateOAuth
// @summary Authenticate with OAuth
// @tags auth
// @accept json
// @produce json
// @param body body oauthPayload true "OAuth Credentials used for authentication"
// @success 200 {object} authenticateResponse "Success"
// @failure 400 "Invalid request"
// @failure 422 "Invalid Credentials"
// @failure 500 "Server error"
// @router /auth/oauth/validate [post]
func (handler *Handler) validateOAuth(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload oauthPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "Invalid request payload", Err: err}
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve settings from the database", Err: err}
	}

	if settings.AuthenticationMethod != portainer.AuthenticationOAuth {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "OAuth authentication is not enabled", Err: errors.New("OAuth authentication is not enabled")}
	}

	username, err := handler.authenticateOAuth(payload.Code, &settings.OAuthSettings)
	if err != nil {
		log.Printf("[DEBUG] - OAuth authentication error: %s", err)
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to authenticate through OAuth", Err: httperrors.ErrUnauthorized}
	}

	user, err := handler.DataStore.User().UserByUsername(username)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to retrieve a user with the specified username from the database", Err: err}
	}

	if user == nil && !settings.OAuthSettings.OAuthAutoCreateUsers {
		return &httperror.HandlerError{StatusCode: http.StatusForbidden, Message: "Account not created beforehand in Portainer and automatic user provisioning not enabled", Err: httperrors.ErrUnauthorized}
	}

	if user == nil {
		user = &portainer.User{
			Username: username,
			Role:     portainer.StandardUserRole,
		}

		err = handler.DataStore.User().CreateUser(user)
		if err != nil {
			return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist user inside the database", Err: err}
		}

		if settings.OAuthSettings.DefaultTeamID != 0 {
			membership := &portainer.TeamMembership{
				UserID: user.ID,
				TeamID: settings.OAuthSettings.DefaultTeamID,
				Role:   portainer.TeamMember,
			}

			err = handler.DataStore.TeamMembership().CreateTeamMembership(membership)
			if err != nil {
				return &httperror.HandlerError{StatusCode: http.StatusInternalServerError, Message: "Unable to persist team membership inside the database", Err: err}
			}
		}

	}

	return handler.writeToken(w, user)
}
