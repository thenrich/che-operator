//
// Copyright (c) 2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//
package identity_provider

import (
	"fmt"

	"github.com/eclipse-che/che-operator/pkg/deploy"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

const KeckoakSA = "che-keycloak"
const KeycalokSAClusterRoleTemplate = "%s-che-keycloak"
const KeycloakSAClusterRoleBindingTemplate = "%s-che-keycloak"
const ClusterPermissionsKeycloakFinalizer = "che-keycloak-endpoints-monitor.finalizers.che.eclipse.org"

func getEndpointMonitorPolicies() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"services", "endpoints", "pods"},
			Verbs:     []string{"get", "list"},
		},
	}
}

func getClusterRoleName(deployContext *deploy.DeployContext) string {
	return fmt.Sprintf(KeycalokSAClusterRoleTemplate, deployContext.CheCluster.Namespace)
}

func getClusterRoleBindingName(deployContext *deploy.DeployContext) string {
	return fmt.Sprintf(KeycloakSAClusterRoleBindingTemplate, deployContext.CheCluster.Namespace)
}

func delegateEndpointMonitorPermissions(deployContext *deploy.DeployContext) (bool, error) {
	done, err := deploy.SyncClusterRoleToCluster(deployContext, getClusterRoleName(deployContext), getEndpointMonitorPolicies())
	if !done {
		return false, err
	}

	done, err = deploy.SyncClusterRoleBindingToCluster(deployContext, getClusterRoleBindingName(deployContext), KeckoakSA, getClusterRoleName(deployContext))
	if !done {
		return false, err
	}

	err = deploy.AppendFinalizer(deployContext, ClusterPermissionsKeycloakFinalizer)
	return err == nil, err
}

func removeEndpointMonitorPermissions(deployContext *deploy.DeployContext) (bool, error) {
	done, err := deploy.Delete(deployContext, types.NamespacedName{Name: getClusterRoleName(deployContext)}, &rbacv1.ClusterRole{})
	if !done {
		return false, err
	}

	done, err = deploy.Delete(deployContext, types.NamespacedName{Name: getClusterRoleBindingName(deployContext)}, &rbacv1.ClusterRoleBinding{})
	if !done {
		return false, err
	}

	err = deploy.DeleteFinalizer(deployContext, ClusterPermissionsKeycloakFinalizer)
	return err == nil, err
}
