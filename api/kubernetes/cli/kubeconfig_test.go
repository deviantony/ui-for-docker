package cli

import (
	"context"
	"errors"
	"testing"

	portainer "github.com/portainer/portainer/api"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kfake "k8s.io/client-go/kubernetes/fake"
)

func Test_GetKubeConfig(t *testing.T) {

	t.Run("returns error if SA non-existent", func(t *testing.T) {
		k := &KubeClient{
			cli:        kfake.NewSimpleClientset(),
			instanceID: "test",
		}

		tokenData := &portainer.TokenData{
			ID:       1,
			Role:     portainer.AdministratorRole,
			Username: portainerClusterAdminServiceAccountName,
		}

		_, err := k.GetKubeConfig(context.Background(), "localhost", "abc", tokenData)

		if err == nil {
			t.Error("GetKubeConfig should fail as service account does not exist")
		}
		if k8sErr := errors.Unwrap(err); !k8serrors.IsNotFound(k8sErr) {
			t.Error("GetKubeConfig should fail with service account not found k8s error")
		}
	})

	t.Run("successfully obtains kubeconfig for cluster admin", func(t *testing.T) {
		k := &KubeClient{
			cli:        kfake.NewSimpleClientset(),
			instanceID: "test",
		}

		tokenData := &portainer.TokenData{
			Role:     portainer.AdministratorRole,
			Username: portainerClusterAdminServiceAccountName,
		}
		serviceAccount := &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: tokenData.Username},
		}

		k.cli.CoreV1().ServiceAccounts(portainerNamespace).Create(serviceAccount)
		defer k.cli.CoreV1().ServiceAccounts(portainerNamespace).Delete(serviceAccount.Name, nil)

		_, err := k.GetKubeConfig(context.Background(), "localhost", "abc", tokenData)

		if err != nil {
			t.Errorf("GetKubeConfig should succeed; err=%s", err)
		}
	})

	t.Run("successfully obtains kubeconfig for standard user", func(t *testing.T) {
		k := &KubeClient{
			cli:        kfake.NewSimpleClientset(),
			instanceID: "test",
		}

		tokenData := &portainer.TokenData{
			ID:   1,
			Role: portainer.StandardUserRole,
		}
		nonAdminUserName := userServiceAccountName(int(tokenData.ID), k.instanceID)
		serviceAccount := &v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{Name: nonAdminUserName},
		}

		k.cli.CoreV1().ServiceAccounts(portainerNamespace).Create(serviceAccount)
		defer k.cli.CoreV1().ServiceAccounts(portainerNamespace).Delete(serviceAccount.Name, nil)

		_, err := k.GetKubeConfig(context.Background(), "localhost", "abc", tokenData)

		if err != nil {
			t.Errorf("GetKubeConfig should succeed; err=%s", err)
		}
	})
}

func Test_generateKubeconfig(t *testing.T) {
	apiServerURL, bearerToken, serviceAccountName := "localhost", "test-token", "test-user"

	t.Run("generates Config resource kind", func(t *testing.T) {
		config := generateKubeconfig(apiServerURL, bearerToken, serviceAccountName)
		want := "Config"
		if config.Kind != want {
			t.Errorf("generateKubeconfig resource kind should be %s", want)
		}
	})

	t.Run("generates v1 version", func(t *testing.T) {
		config := generateKubeconfig(apiServerURL, bearerToken, serviceAccountName)
		want := "v1"
		if config.APIVersion != want {
			t.Errorf("generateKubeconfig api version should be %s", want)
		}
	})

	t.Run("generates single entry context cluster and authinfo", func(t *testing.T) {
		config := generateKubeconfig(apiServerURL, bearerToken, serviceAccountName)
		if len(config.Contexts) != 1 {
			t.Error("generateKubeconfig should generate single context configuration")
		}
		if len(config.Clusters) != 1 {
			t.Error("generateKubeconfig should generate single cluster configuration")
		}
		if len(config.AuthInfos) != 1 {
			t.Error("generateKubeconfig should generate single user configuration")
		}
	})

	t.Run("sets default context appropriately", func(t *testing.T) {
		config := generateKubeconfig(apiServerURL, bearerToken, serviceAccountName)
		want := "portainer-ctx"
		if config.CurrentContext != want {
			t.Errorf("generateKubeconfig set cluster to be %s", want)
		}
	})

	t.Run("generates cluster with InsecureSkipTLSVerify to be set to true", func(t *testing.T) {
		config := generateKubeconfig(apiServerURL, bearerToken, serviceAccountName)
		if config.Clusters[0].Cluster.InsecureSkipTLSVerify != true {
			t.Error("generateKubeconfig default cluster InsecureSkipTLSVerify should be true")
		}
	})

	t.Run("should contain passed in value", func(t *testing.T) {
		config := generateKubeconfig(apiServerURL, bearerToken, serviceAccountName)
		if config.Clusters[0].Cluster.Server != apiServerURL {
			t.Errorf("generateKubeconfig default cluster server url should be %s", apiServerURL)
		}

		if config.AuthInfos[0].Name != serviceAccountName {
			t.Errorf("generateKubeconfig default authinfo name should be %s", serviceAccountName)
		}

		if config.AuthInfos[0].AuthInfo.Token != bearerToken {
			t.Errorf("generateKubeconfig default authinfo user token should be %s", bearerToken)
		}
	})
}
