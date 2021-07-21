package ldap

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/crypto"
	httperrors "github.com/portainer/portainer/api/http/errors"
)

var (
	// errUserNotFound defines an error raised when the user is not found via LDAP search
	// or that too many entries (> 1) are returned.
	errUserNotFound = errors.New("User not found or too many entries returned")
)

// Service represents a service used to authenticate users against a LDAP/AD.
type Service struct{}

func searchUser(username string, conn *ldap.Conn, settings []portainer.LDAPSearchSettings) (string, error) {
	var userDN string
	found := false
	usernameEscaped := ldap.EscapeFilter(username)

	for _, searchSettings := range settings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.BaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(%s=%s))", searchSettings.Filter, searchSettings.UserNameAttribute, usernameEscaped),
			[]string{"dn"},
			nil,
		)

		// Deliberately skip errors on the search request so that we can jump to other search settings
		// if any issue arise with the current one.
		sr, err := conn.Search(searchRequest)
		if err != nil {
			continue
		}

		if len(sr.Entries) == 1 {
			found = true
			userDN = sr.Entries[0].DN
			break
		}
	}

	if !found {
		return "", errUserNotFound
	}

	return userDN, nil
}

func createConnection(settings *portainer.LDAPSettings) (*ldap.Conn, error) {

	if settings.TLSConfig.TLS || settings.StartTLS {
		config, err := crypto.CreateTLSConfigurationFromDisk(settings.TLSConfig.TLSCACertPath, settings.TLSConfig.TLSCertPath, settings.TLSConfig.TLSKeyPath, settings.TLSConfig.TLSSkipVerify)
		if err != nil {
			return nil, err
		}
		config.ServerName = strings.Split(settings.URL, ":")[0]

		if settings.TLSConfig.TLS {
			return ldap.DialTLS("tcp", settings.URL, config)
		}

		conn, err := ldap.Dial("tcp", settings.URL)
		if err != nil {
			return nil, err
		}

		err = conn.StartTLS(config)
		if err != nil {
			return nil, err
		}

		return conn, nil
	}

	return ldap.Dial("tcp", settings.URL)
}

// AuthenticateUser is used to authenticate a user against a LDAP/AD.
func (*Service) AuthenticateUser(username, password string, settings *portainer.LDAPSettings) error {

	connection, err := createConnection(settings)
	if err != nil {
		return err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return err
		}
	}

	userDN, err := searchUser(username, connection, settings.SearchSettings)
	if err != nil {
		return err
	}

	err = connection.Bind(userDN, password)
	if err != nil {
		return httperrors.ErrUnauthorized
	}

	return nil
}

// GetUserGroups is used to retrieve user groups from LDAP/AD.
func (*Service) GetUserGroups(username string, settings *portainer.LDAPSettings) ([]string, error) {
	connection, err := createConnection(settings)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return nil, err
		}
	}

	userDN, err := searchUser(username, connection, settings.SearchSettings)
	if err != nil {
		return nil, err
	}

	userGroups := getGroups(userDN, connection, settings.GroupSearchSettings)

	return userGroups, nil
}

// GetUserGroups is used to retrieve user groups from LDAP/AD.
func (*Service) GetUserAdminGroups(username string, settings *portainer.LDAPSettings) ([]string, error) {
	connection, err := createConnection(settings)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return nil, err
		}
	}

	userDN, err := searchUser(username, connection, settings.SearchSettings)
	if err != nil {
		return nil, err
	}

	userGroups := getGroupsByUser(userDN, connection, settings.AdminGroupSearchSettings)

	return userGroups, nil
}

// SearchGroups searches for groups with the specified settings
func (*Service) SearchAdminGroups(settings *portainer.LDAPSettings) ([]string, error) {
	type groupSet map[string]bool

	connection, err := createConnection(settings)
	if err != nil {
		return nil, err
	}
	defer connection.Close()

	if !settings.AnonymousMode {
		err = connection.Bind(settings.ReaderDN, settings.Password)
		if err != nil {
			return nil, err
		}
	}

	userGroups := map[string]groupSet{}

	for _, searchSettings := range settings.AdminGroupSearchSettings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			searchSettings.GroupFilter,
			[]string{"cn", searchSettings.GroupAttribute},
			nil,
		)

		sr, err := connection.Search(searchRequest)
		if err != nil {
			return nil, err
		}

		for _, entry := range sr.Entries {
			members := entry.GetAttributeValues(searchSettings.GroupAttribute)
			for _, username := range members {
				_, ok := userGroups[username]
				if !ok {
					userGroups[username] = groupSet{}
				}
				userGroups[username][entry.GetAttributeValue("cn")] = true
			}
		}
	}

	groupsMap := make(map[string]bool)

	for _, groups := range userGroups {
		for group := range groups {
			groupsMap[group] = true
		}
	}

	groups := make([]string, 0, len(groupsMap))
	for group := range groupsMap {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	return groups, nil
}

// TestConnectivity is used to test a connection against the LDAP server using the credentials
// specified in the LDAPSettings.
func (*Service) TestConnectivity(settings *portainer.LDAPSettings) error {

	connection, err := createConnection(settings)
	if err != nil {
		return err
	}
	defer connection.Close()

	err = connection.Bind(settings.ReaderDN, settings.Password)
	if err != nil {
		return err
	}
	return nil
}

// Get a list of group names for specified user from LDAP/AD
func getGroups(userDN string, conn *ldap.Conn, settings []portainer.LDAPGroupSearchSettings) []string {
	groups := make([]string, 0)
	userDNEscaped := ldap.EscapeFilter(userDN)

	for _, searchSettings := range settings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(%s=%s))", searchSettings.GroupFilter, searchSettings.GroupAttribute, userDNEscaped),
			[]string{"cn"},
			nil,
		)

		// Deliberately skip errors on the search request so that we can jump to other search settings
		// if any issue arise with the current one.
		sr, err := conn.Search(searchRequest)
		if err != nil {
			continue
		}

		for _, entry := range sr.Entries {
			for _, attr := range entry.Attributes {
				groups = append(groups, attr.Values[0])
			}
		}
	}

	return groups
}

// Get a list of group names for specified user from LDAP/AD
func getGroupsByUser(userDN string, conn *ldap.Conn, settings []portainer.LDAPGroupSearchSettings) []string {
	groups := make([]string, 0)
	userDNEscaped := ldap.EscapeFilter(userDN)

	for _, searchSettings := range settings {
		searchRequest := ldap.NewSearchRequest(
			searchSettings.GroupBaseDN,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(&%s(%s=%s))", searchSettings.GroupFilter, searchSettings.GroupAttribute, userDNEscaped),
			[]string{"cn"},
			nil,
		)

		// Deliberately skip errors on the search request so that we can jump to other search settings
		// if any issue arise with the current one.
		sr, err := conn.Search(searchRequest)
		if err != nil {
			continue
		}

		for _, entry := range sr.Entries {
			for _, attr := range entry.Attributes {
				groups = append(groups, attr.Values[0])
			}
		}
	}

	return groups
}
