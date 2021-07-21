package testhelpers

import (
	portainer "github.com/portainer/portainer/api"
)

type ldapService struct{}

// NewGitService creates new mock for portainer.GitService.
func NewLDAPService() *ldapService {
	return &ldapService{}
}

// AuthenticateUser is used to authenticate a user against a LDAP/AD.
func (service *ldapService) AuthenticateUser(username, password string, settings *portainer.LDAPSettings) error {
	return nil
}

// GetUserGroups is used to retrieve user groups from LDAP/AD.
func (service *ldapService) GetUserGroups(username string, settings *portainer.LDAPSettings) ([]string, error) {
	return []string{"stuff"}, nil
}

func (service *ldapService) GetUserAdminGroups(username string, settings *portainer.LDAPSettings) ([]string, error) {
	return []string{"manager", "lead"}, nil
}

// SearchGroups searches for groups with the specified settings
func (service *ldapService) SearchAdminGroups(settings *portainer.LDAPSettings) ([]string, error) {
	return nil, nil
}

func (service *ldapService) TestConnectivity(settings *portainer.LDAPSettings) error {
	return nil
}
