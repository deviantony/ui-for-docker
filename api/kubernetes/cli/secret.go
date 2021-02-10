package cli

import (
	"context"
	"errors"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kcl *KubeClient) createServiceAccountToken(serviceAccountName string) error {
	serviceAccountSecretName := userServiceAccountTokenSecretName(serviceAccountName, kcl.instanceID)

	serviceAccountSecret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountSecretName,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": serviceAccountName,
			},
		},
		Type: "kubernetes.io/service-account-token",
	}

	_, err := kcl.cli.CoreV1().Secrets(portainerNamespace).Create(context.TODO(), serviceAccountSecret, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (kcl *KubeClient) getServiceAccountToken(serviceAccountName string) (string, error) {
	serviceAccountSecretName := userServiceAccountTokenSecretName(serviceAccountName, kcl.instanceID)

	secret, err := kcl.cli.CoreV1().Secrets(portainerNamespace).Get(context.TODO(), serviceAccountSecretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// API token secret is populated asynchronously.
	// Is it created by the controller and will depend on the environment/secret-store:
	// https://github.com/kubernetes/kubernetes/issues/67882#issuecomment-422026204
	// as a work-around, we wait for up to 5 seconds for the secret to be populated.
	timeout := time.After(5 * time.Second)
	searchingForSecret := true
	for searchingForSecret {
		select {
		case <-timeout:
			return "", errors.New("unable to find secret token associated to user service account (timeout)")
		default:
			secret, err = kcl.cli.CoreV1().Secrets(portainerNamespace).Get(context.TODO(), serviceAccountSecretName, metav1.GetOptions{})
			if err != nil {
				return "", err
			}

			if len(secret.Data) > 0 {
				searchingForSecret = false
				break
			}

			time.Sleep(1 * time.Second)
		}
	}

	secretTokenData, ok := secret.Data["token"]
	if ok {
		return string(secretTokenData), nil
	}

	return "", errors.New("unable to find secret token associated to user service account")
}
