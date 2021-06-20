package cli

import (
	portainer "github.com/portainer/portainer/api"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServiceAccount returns the portainer ServiceAccountName associated to the specified user.
func (kcl *KubeClient) GetServiceAccount(tokenData *portainer.TokenData) (*v1.ServiceAccount, error) {
	var portainerServiceAccountName string
	if tokenData.Role == portainer.AdministratorRole {
		portainerServiceAccountName = portainerClusterAdminServiceAccountName
	} else {
		portainerServiceAccountName = userServiceAccountName(int(tokenData.ID), kcl.instanceID)
	}

	// verify name exists as service account resource within portainer namespace
	serviceAccount, err := kcl.cli.CoreV1().ServiceAccounts(portainerNamespace).Get(portainerServiceAccountName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return serviceAccount, nil
}

// GetServiceAccountBearerToken returns the ServiceAccountToken associated to the specified user.
func (kcl *KubeClient) GetServiceAccountBearerToken(userID int) (string, error) {
	serviceAccountName := userServiceAccountName(userID, kcl.instanceID)

	return kcl.getServiceAccountToken(serviceAccountName)
}

// SetupUserServiceAccount will make sure that all the required resources are created inside the Kubernetes
// cluster before creating a ServiceAccount and a ServiceAccountToken for the specified Portainer user.
//It will also create required default RoleBinding and ClusterRoleBinding rules.
func (kcl *KubeClient) SetupUserServiceAccount(userID int, teamIDs []int) error {
	serviceAccountName := userServiceAccountName(userID, kcl.instanceID)

	err := kcl.ensureRequiredResourcesExist()
	if err != nil {
		return err
	}

	err = kcl.ensureServiceAccountForUserExists(serviceAccountName)
	if err != nil {
		return err
	}

	return kcl.setupNamespaceAccesses(userID, teamIDs, serviceAccountName)
}

func (kcl *KubeClient) ensureRequiredResourcesExist() error {
	return kcl.createPortainerUserClusterRole()
}

func (kcl *KubeClient) ensureServiceAccountForUserExists(serviceAccountName string) error {
	err := kcl.createUserServiceAccount(portainerNamespace, serviceAccountName)
	if err != nil {
		return err
	}

	err = kcl.createServiceAccountToken(serviceAccountName)
	if err != nil {
		return err
	}

	err = kcl.ensureServiceAccountHasPortainerUserClusterRole(serviceAccountName)
	if err != nil {
		return err
	}

	return kcl.ensureNamespaceAccessForServiceAccount(serviceAccountName, defaultNamespace)
}

func (kcl *KubeClient) createUserServiceAccount(namespace, serviceAccountName string) error {
	serviceAccount := &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceAccountName,
		},
	}

	_, err := kcl.cli.CoreV1().ServiceAccounts(namespace).Create(serviceAccount)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (kcl *KubeClient) ensureServiceAccountHasPortainerUserClusterRole(serviceAccountName string) error {
	clusterRoleBinding, err := kcl.cli.RbacV1().ClusterRoleBindings().Get(portainerUserCRBName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		clusterRoleBinding = &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: portainerUserCRBName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      serviceAccountName,
					Namespace: portainerNamespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: portainerUserCRName,
			},
		}

		_, err := kcl.cli.RbacV1().ClusterRoleBindings().Create(clusterRoleBinding)
		return err
	} else if err != nil {
		return err
	}

	for _, subject := range clusterRoleBinding.Subjects {
		if subject.Name == serviceAccountName {
			return nil
		}
	}

	clusterRoleBinding.Subjects = append(clusterRoleBinding.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      serviceAccountName,
		Namespace: portainerNamespace,
	})

	_, err = kcl.cli.RbacV1().ClusterRoleBindings().Update(clusterRoleBinding)
	return err
}

func (kcl *KubeClient) removeNamespaceAccessForServiceAccount(serviceAccountName, namespace string) error {
	roleBindingName := namespaceClusterRoleBindingName(namespace, kcl.instanceID)

	roleBinding, err := kcl.cli.RbacV1().RoleBindings(namespace).Get(roleBindingName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	updatedSubjects := roleBinding.Subjects[:0]

	for _, subject := range roleBinding.Subjects {
		if subject.Name != serviceAccountName {
			updatedSubjects = append(updatedSubjects, subject)
		}
	}

	roleBinding.Subjects = updatedSubjects

	_, err = kcl.cli.RbacV1().RoleBindings(namespace).Update(roleBinding)
	return err
}

func (kcl *KubeClient) ensureNamespaceAccessForServiceAccount(serviceAccountName, namespace string) error {
	roleBindingName := namespaceClusterRoleBindingName(namespace, kcl.instanceID)

	roleBinding, err := kcl.cli.RbacV1().RoleBindings(namespace).Get(roleBindingName, metav1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		roleBinding = &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: roleBindingName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      serviceAccountName,
					Namespace: portainerNamespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: "edit",
			},
		}

		_, err = kcl.cli.RbacV1().RoleBindings(namespace).Create(roleBinding)
		return err
	} else if err != nil {
		return err
	}

	for _, subject := range roleBinding.Subjects {
		if subject.Name == serviceAccountName {
			return nil
		}
	}

	roleBinding.Subjects = append(roleBinding.Subjects, rbacv1.Subject{
		Kind:      "ServiceAccount",
		Name:      serviceAccountName,
		Namespace: portainerNamespace,
	})

	_, err = kcl.cli.RbacV1().RoleBindings(namespace).Update(roleBinding)
	return err
}
