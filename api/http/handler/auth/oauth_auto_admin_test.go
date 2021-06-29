package auth

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
)

func Test_validateClaimWithRegex(t *testing.T) {
	t.Run("returns false if no regex match occurs", func(t *testing.T) {
		oAuthSettings := portainer.OAuthSettings{
			AdminGroupClaimsRegexList: []string{"@"},
		}
		oAuthTeams := []string{"#portainer"}

		isValid, err := validateClaimWithRegex(oAuthSettings, oAuthTeams)
		if err != nil {
			t.Errorf("failed to validate, error: %v", err)
		}
		if isValid {
			t.Errorf("should be invalid when matching AdminGroupClaimsRegexList: %v and OAuth team: %v", oAuthSettings.AdminGroupClaimsRegexList, oAuthTeams)
		}
	})

	t.Run("returns true if regex match - single element in AdminGroupClaimsRegexList", func(t *testing.T) {
		oAuthSettings := portainer.OAuthSettings{
			AdminGroupClaimsRegexList: []string{"@"},
		}
		oAuthTeams := []string{"@portainer"}

		isValid, err := validateClaimWithRegex(oAuthSettings, oAuthTeams)
		if err != nil {
			t.Errorf("failed to validate, error: %v", err)
		}
		if !isValid {
			t.Errorf("should be valid when matching AdminGroupClaimsRegexList: %v and OAuth team: %v", oAuthSettings.AdminGroupClaimsRegexList, oAuthTeams)
		}
	})

	t.Run("returns true if regex match - multiple elements in AdminGroupClaimsRegexList and oAuthTeams", func(t *testing.T) {
		oAuthSettings := portainer.OAuthSettings{
			AdminGroupClaimsRegexList: []string{"@", "#"},
		}
		oAuthTeams := []string{"portainer", "@portainer"}

		isValid, err := validateClaimWithRegex(oAuthSettings, oAuthTeams)
		if err != nil {
			t.Errorf("failed to validate, error: %v", err)
		}
		if !isValid {
			t.Errorf("should be valid when matching AdminGroupClaimsRegexList: %v and OAuth team: %v", oAuthSettings.AdminGroupClaimsRegexList, oAuthTeams)
		}
	})
}
