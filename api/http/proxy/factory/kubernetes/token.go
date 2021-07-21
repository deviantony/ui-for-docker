package kubernetes

import (
	"io/ioutil"
	"sync"

	portainer "github.com/portainer/portainer/api"
)

const defaultServiceAccountTokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

type tokenManager struct {
	tokenCache *tokenCache
	kubecli    portainer.KubeClient
	dataStore  portainer.DataStore
	mutex      sync.Mutex
	adminToken string
}

// NewTokenManager returns a pointer to a new instance of tokenManager.
// If the useLocalAdminToken parameter is set to true, it will search for the local admin service account
// and associate it to the manager.
func NewTokenManager(kubecli portainer.KubeClient, dataStore portainer.DataStore, cache *tokenCache, setLocalAdminToken bool) (*tokenManager, error) {
	tokenManager := &tokenManager{
		tokenCache: cache,
		kubecli:    kubecli,
		dataStore:  dataStore,
		mutex:      sync.Mutex{},
		adminToken: "",
	}

	if setLocalAdminToken {
		token, err := ioutil.ReadFile(defaultServiceAccountTokenFile)
		if err != nil {
			return nil, err
		}

		tokenManager.adminToken = string(token)
	}

	return tokenManager, nil
}

func (manager *tokenManager) getAdminServiceAccountToken() string {
	return manager.adminToken
}

func (manager *tokenManager) getUserServiceAccountToken(userID int, endpointID portainer.EndpointID) (string, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	token, ok := manager.tokenCache.getToken(userID)
	if !ok {
		memberships, err := manager.dataStore.TeamMembership().TeamMembershipsByUserID(portainer.UserID(userID))
		if err != nil {
			return "", err
		}

		teamIds := make([]int, 0)
		for _, membership := range memberships {
			teamIds = append(teamIds, int(membership.TeamID))
		}

		endpoint, err := manager.dataStore.Endpoint().Endpoint(endpointID)
		if err != nil {
			return "", err
		}

		restrictDefaultNamespace := endpoint.Kubernetes.Configuration.RestrictDefaultNamespace
		err = manager.kubecli.SetupUserServiceAccount(userID, teamIds, restrictDefaultNamespace)
		if err != nil {
			return "", err
		}

		serviceAccountToken, err := manager.kubecli.GetServiceAccountBearerToken(userID)
		if err != nil {
			return "", err
		}

		manager.tokenCache.addToken(userID, serviceAccountToken)
		token = serviceAccountToken
	}

	return token, nil
}
