package settings

import (
	"errors"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/filesystem"
)

type settingsUpdatePayload struct {
	// URL to a logo that will be displayed on the login page as well as on top of the sidebar. Will use default Portainer logo when value is empty string
	LogoURL *string `example:"https://mycompany.mydomain.tld/logo.png"`
	// A list of label name & value that will be used to hide containers when querying containers
	BlackListedLabels []portainer.Pair
	// Active authentication method for the Portainer instance. Valid values are: 1 for internal, 2 for LDAP, or 3 for oauth
	AuthenticationMethod *int                     `example:"1"`
	LDAPSettings         *portainer.LDAPSettings  `example:""`
	OAuthSettings        *portainer.OAuthSettings `example:""`
	// The interval in which endpoint snapshots are created
	SnapshotInterval *string `example:"5m"`
	// URL to the templates that will be displayed in the UI when navigating to App Templates
	TemplatesURL *string `example:"https://raw.githubusercontent.com/portainer/templates/master/templates.json"`
	// The default check in interval for edge agent (in seconds)
	EdgeAgentCheckinInterval *int `example:"5"`
	// Whether edge compute features are enabled
	EnableEdgeComputeFeatures *bool `example:"true"`
	// The duration of a user session
	UserSessionTimeout *string `example:"5m"`
	// Whether telemetry is enabled
	EnableTelemetry *bool `example:"false"`
}

func (payload *settingsUpdatePayload) Validate(r *http.Request) error {
	if payload.AuthenticationMethod != nil && *payload.AuthenticationMethod != 1 && *payload.AuthenticationMethod != 2 && *payload.AuthenticationMethod != 3 {
		return errors.New("Invalid authentication method value. Value must be one of: 1 (internal), 2 (LDAP/AD) or 3 (OAuth)")
	}
	if payload.LogoURL != nil && *payload.LogoURL != "" && !govalidator.IsURL(*payload.LogoURL) {
		return errors.New("Invalid logo URL. Must correspond to a valid URL format")
	}
	if payload.TemplatesURL != nil && *payload.TemplatesURL != "" && !govalidator.IsURL(*payload.TemplatesURL) {
		return errors.New("Invalid external templates URL. Must correspond to a valid URL format")
	}
	if payload.UserSessionTimeout != nil {
		_, err := time.ParseDuration(*payload.UserSessionTimeout)
		if err != nil {
			return errors.New("Invalid user session timeout")
		}
	}
	if payload.AuthenticationMethod != nil && *payload.AuthenticationMethod == 2 {
		if payload.LDAPSettings == nil {
			return errors.New("Invalid LDAP Configuration")
		}
		if payload.LDAPSettings.URL == "" {
			return errors.New("Invalid LDAP URL")
		}
		if payload.LDAPSettings.AdminAutoPopulate && len(payload.LDAPSettings.AdminGroupSearchSettings) == 0 {
			return errors.New("Invalid AdminGroupSearchSettings")
		}
		if !payload.LDAPSettings.AdminAutoPopulate && len(payload.LDAPSettings.AdminGroups) > 0 {
			payload.LDAPSettings.AdminGroups = []string{}
		}
	}

	return nil
}

// @id SettingsUpdate
// @summary Update Portainer settings
// @description Update Portainer settings.
// @description **Access policy**: administrator
// @tags settings
// @security jwt
// @accept json
// @produce json
// @param body body settingsUpdatePayload true "New settings"
// @success 200 {object} portainer.Settings "Success"
// @failure 400 "Invalid request"
// @failure 500 "Server error"
// @router /settings [put]
func (handler *Handler) settingsUpdate(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	var payload settingsUpdatePayload
	err := request.DecodeAndValidateJSONPayload(r, &payload)
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid request payload", err}
	}

	settings, err := handler.DataStore.Settings().Settings()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve the settings from the database", err}
	}

	if payload.AuthenticationMethod != nil {
		settings.AuthenticationMethod = portainer.AuthenticationMethod(*payload.AuthenticationMethod)
	}

	if payload.LogoURL != nil {
		settings.LogoURL = *payload.LogoURL
	}

	if payload.TemplatesURL != nil {
		settings.TemplatesURL = *payload.TemplatesURL
	}

	if payload.BlackListedLabels != nil {
		settings.BlackListedLabels = payload.BlackListedLabels
	}

	if payload.LDAPSettings != nil {
		ldapReaderDN := settings.LDAPSettings.ReaderDN
		ldapPassword := settings.LDAPSettings.Password
		if payload.LDAPSettings.ReaderDN != "" {
			ldapReaderDN = payload.LDAPSettings.ReaderDN
		}
		if payload.LDAPSettings.Password != "" {
			ldapPassword = payload.LDAPSettings.Password
		}
		settings.LDAPSettings = *payload.LDAPSettings
		settings.LDAPSettings.ReaderDN = ldapReaderDN
		settings.LDAPSettings.Password = ldapPassword
	}

	if payload.OAuthSettings != nil {
		clientSecret := payload.OAuthSettings.ClientSecret
		if clientSecret == "" {
			clientSecret = settings.OAuthSettings.ClientSecret
		}
		settings.OAuthSettings = *payload.OAuthSettings
		settings.OAuthSettings.ClientSecret = clientSecret
	}

	if payload.EnableEdgeComputeFeatures != nil {
		settings.EnableEdgeComputeFeatures = *payload.EnableEdgeComputeFeatures
	}

	if payload.SnapshotInterval != nil && *payload.SnapshotInterval != settings.SnapshotInterval {
		err := handler.updateSnapshotInterval(settings, *payload.SnapshotInterval)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to update snapshot interval", err}
		}
	}

	if payload.EdgeAgentCheckinInterval != nil {
		settings.EdgeAgentCheckinInterval = *payload.EdgeAgentCheckinInterval
	}

	if payload.UserSessionTimeout != nil {
		settings.UserSessionTimeout = *payload.UserSessionTimeout

		userSessionDuration, _ := time.ParseDuration(*payload.UserSessionTimeout)

		handler.JWTService.SetUserSessionDuration(userSessionDuration)
	}

	if payload.EnableTelemetry != nil {
		settings.EnableTelemetry = *payload.EnableTelemetry
	}

	tlsError := handler.updateTLS(settings)
	if tlsError != nil {
		return tlsError
	}

	err = handler.DataStore.Settings().UpdateSettings(settings)
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to persist settings changes inside the database", err}
	}

	return response.JSON(w, settings)
}

func (handler *Handler) updateSnapshotInterval(settings *portainer.Settings, snapshotInterval string) error {
	settings.SnapshotInterval = snapshotInterval

	err := handler.SnapshotService.SetSnapshotInterval(snapshotInterval)
	if err != nil {
		return err
	}

	return nil
}

func (handler *Handler) updateTLS(settings *portainer.Settings) *httperror.HandlerError {
	if (settings.LDAPSettings.TLSConfig.TLS || settings.LDAPSettings.StartTLS) && !settings.LDAPSettings.TLSConfig.TLSSkipVerify {
		caCertPath, _ := handler.FileService.GetPathForTLSFile(filesystem.LDAPStorePath, portainer.TLSFileCA)
		settings.LDAPSettings.TLSConfig.TLSCACertPath = caCertPath
	} else {
		settings.LDAPSettings.TLSConfig.TLSCACertPath = ""
		err := handler.FileService.DeleteTLSFiles(filesystem.LDAPStorePath)
		if err != nil {
			return &httperror.HandlerError{http.StatusInternalServerError, "Unable to remove TLS files from disk", err}
		}
	}
	return nil
}
