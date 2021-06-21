package cli

import (
	"context"
	"fmt"

	portainer "github.com/portainer/portainer/api"
	clientV1 "k8s.io/client-go/tools/clientcmd/api/v1"
)

// GetKubeConfig returns kubeconfig for the current user based on:
// - portainer server url
// - portainer user bearer token
// - portainer token data - which maps to k8s service account
func (kcl *KubeClient) GetKubeConfig(ctx context.Context, apiServerURL string, bearerToken string, tokenData *portainer.TokenData) (*clientV1.Config, error) {
	serviceAccount, err := kcl.GetServiceAccount(tokenData)
	if err != nil {
		errText := fmt.Sprintf("unable to find serviceaccount associated with user; username=%s", tokenData.Username)
		return nil, fmt.Errorf("%s; err=%w", errText, err)
	}

	kubeconfig := generateKubeconfig(apiServerURL, bearerToken, serviceAccount.Name)

	return kubeconfig, nil
}

// generateKubeconfig will generate and return kubeconfig resource - usable by `kubectl` cli
// which will allow the client to connect directly to k8s server endpoint via portainer (proxy)
func generateKubeconfig(apiServerURL, bearerToken, serviceAccountName string) *clientV1.Config {
	const (
		KubeConfigPortainerContext = "portainer-ctx"
		KubeConfigPortainerCluster = "portainer-cluster"
	)

	return &clientV1.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		CurrentContext: KubeConfigPortainerContext,
		Contexts: []clientV1.NamedContext{
			{
				Name: KubeConfigPortainerContext,
				Context: clientV1.Context{
					AuthInfo: serviceAccountName,
					Cluster:  KubeConfigPortainerCluster,
				},
			},
		},
		Clusters: []clientV1.NamedCluster{
			{
				Name: KubeConfigPortainerCluster,
				Cluster: clientV1.Cluster{
					Server:                apiServerURL,
					InsecureSkipTLSVerify: true,
				},
			},
		},
		AuthInfos: []clientV1.NamedAuthInfo{
			{
				Name: serviceAccountName,
				AuthInfo: clientV1.AuthInfo{
					Token: bearerToken,
				},
			},
		},
	}
}
