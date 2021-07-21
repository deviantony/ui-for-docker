package cli

import (
	"fmt"
)

const (
	defaultNamespace                        = "default"
	portainerNamespace                      = "portainer"
	portainerUserCRName                     = "portainer-cr-user"
	portainerUserCRBName                    = "portainer-crb-user"
	portainerClusterAdminServiceAccountName = "portainer-sa-clusteradmin"
	portainerUserServiceAccountPrefix       = "portainer-sa-user"
	portainerRBPrefix                       = "portainer-rb"
	portainerConfigMapName                  = "portainer-config"
	portainerConfigMapAccessPoliciesKey     = "NamespaceAccessPolicies"
	portainerShellPodPrefix                 = "portainer-pod-kubectl-shell"
)

func userServiceAccountName(userID int, instanceID string) string {
	return fmt.Sprintf("%s-%s-%d", portainerUserServiceAccountPrefix, instanceID, userID)
}

func userServiceAccountTokenSecretName(serviceAccountName string, instanceID string) string {
	return fmt.Sprintf("%s-%s-secret", instanceID, serviceAccountName)
}

func namespaceClusterRoleBindingName(namespace string, instanceID string) string {
	return fmt.Sprintf("%s-%s-%s", portainerRBPrefix, instanceID, namespace)
}

func userShellPodPrefix(serviceAccountName string) string {
	return fmt.Sprintf("%s-%s-", portainerShellPodPrefix, serviceAccountName)
}
