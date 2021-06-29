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
	"github.com/portainer/portainer/api/internal/authorization"
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

func (handler *Handler) authenticateOAuth(code string, settings *portainer.OAuthSettings) (*portainer.OAuthInfo, error) {
	if code == "" {
		return nil, errors.New("Invalid OAuth authorization code")
	}

	if settings == nil {
		return nil, errors.New("Invalid OAuth configuration")
	}

	authInfo, err := handler.OAuthService.Authenticate(code, settings)
	if err != nil {
		return nil, err
	}

	return authInfo, nil
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
func (handler *Handler) validateOAuth(w http.ResponseWriter, r *http.Request) (*authMiddlewareResponse, *httperror.HandlerError) {
	resp := &authMiddlewareResponse{
		Method: portainer.AuthenticationOAuth,
	}

	var payload oauthPayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request payload",
			Err:        err,
		}
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to retrieve settings from the database",
			Err:        err,
		}
	}

	if settings.AuthenticationMethod != 3 {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusForbidden,
			Message:    "OAuth authentication is not enabled",
			Err:        errors.New("OAuth authentication is not enabled"),
		}
	}

	authInfo, err := handler.authenticateOAuth(payload.Code, &settings.OAuthSettings)
	if err != nil {
		log.Printf("[DEBUG] - OAuth authentication error: %s", err)
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to authenticate through OAuth",
			Err:        httperrors.ErrUnauthorized,
		}
	}

	resp.Username = authInfo.Username

	user, err := handler.DataStore.User().UserByUsername(authInfo.Username)
	if err != nil && err != bolterrors.ErrObjectNotFound {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Unable to retrieve a user with the specified username from the database",
			Err:        err,
		}
	}

	if user == nil && !settings.OAuthSettings.OAuthAutoMapTeamMemberships && !settings.OAuthSettings.OAuthAutoCreateUsers {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusForbidden,
			Message:    "Account not created beforehand in Portainer and automatic user provisioning not enabled",
			Err:        httperrors.ErrUnauthorized,
		}
	}

	if user == nil && !settings.OAuthSettings.AdminAutoPopulate && !settings.OAuthSettings.OAuthAutoCreateUsers {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusForbidden,
			Message:    "Auto OAuth admin population failed: user not created beforehand in Portainer and automatic user provisioning not enabled",
			Err:        httperrors.ErrUnauthorized,
		}
	}

	autoOAuthAdmin := false

	if settings.OAuthSettings.AdminAutoPopulate {
		isValid, err := validateClaimWithRegex(settings.OAuthSettings, authInfo.Teams)
		if err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to validate OAuth teams with pre-set claim regexs",
				Err:        err,
			}
		}
		if user != nil && isValid {
			user.Role = portainer.AdministratorRole
			if err := handler.DataStore.User().UpdateUser(user.ID, user); err != nil {
				return resp, &httperror.HandlerError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Unable to persist user changes inside the database",
					Err:        err,
				}
			}
			if err := handler.AuthorizationService.UpdateUsersAuthorizations(); err != nil {
				return resp, &httperror.HandlerError{
					StatusCode: http.StatusInternalServerError,
					Message:    "Unable to update user authorizations",
					Err:        err,
				}
			}

		}
		autoOAuthAdmin = isValid
	}

	if user == nil {
		user = &portainer.User{
			Username:                authInfo.Username,
			Role:                    portainer.StandardUserRole,
			PortainerAuthorizations: authorization.DefaultPortainerAuthorizations(),
		}

		if autoOAuthAdmin {
			user.Role = portainer.AdministratorRole
		}

		err = handler.DataStore.User().CreateUser(user)
		if err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to persist user inside the database",
				Err:        err,
			}
		}

		if settings.OAuthSettings.DefaultTeamID != 0 {
			membership := &portainer.TeamMembership{
				UserID: user.ID,
				TeamID: settings.OAuthSettings.DefaultTeamID,
				Role:   portainer.TeamMember,
			}

			err = handler.DataStore.TeamMembership().CreateTeamMembership(membership)
			if err != nil {
				return &authMiddlewareResponse{
						Method: portainer.AuthenticationOAuth,
					}, &httperror.HandlerError{
						StatusCode: http.StatusInternalServerError,
						Message:    "Unable to persist team membership inside the database",
						Err:        err,
					}
			}
		}

		err = handler.AuthorizationService.UpdateUsersAuthorizations()
		if err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to update user authorizations",
				Err:        err,
			}
		}
	}

	if settings.OAuthSettings.OAuthAutoMapTeamMemberships {
		if settings.OAuthSettings.TeamMemberships.OAuthClaimName == "" {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to process user oauth team memberships",
				Err:        errors.New("empty value set for oauth team membership Claim name"),
			}
		}

		err = updateOAuthTeamMemberships(handler.DataStore, settings.OAuthSettings.TeamMemberships.OAuthClaimMappings, *user, authInfo.Teams)
		if err != nil {
			return resp, &httperror.HandlerError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Unable to update user oauth team memberships",
				Err:        err,
			}
		}
	}

	info := handler.LicenseService.Info()

	if user.Role != portainer.AdministratorRole && !info.Valid {
		return resp, &httperror.HandlerError{
			StatusCode: http.StatusForbidden,
			Message:    "License is not valid",
			Err:        httperrors.ErrNoValidLicense,
		}
	}

	return handler.writeTokenForOAuth(w, user, authInfo.ExpiryTime, portainer.AuthenticationOAuth)
}
