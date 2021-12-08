//
// Copyright (c) 2019-2021 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//
package server

import (
	"github.com/eclipse-che/che-operator/pkg/deploy"
	"github.com/eclipse-che/che-operator/pkg/util"
	routev1 "github.com/openshift/api/route/v1"
	networking "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CheHostReconciler struct {
	deploy.Reconcilable
}

func NewCheHostReconciler() *CheHostReconciler {
	return &CheHostReconciler{}
}

func (s *CheHostReconciler) Reconcile(ctx *deploy.DeployContext) (reconcile.Result, bool, error) {
	done, err := s.detectDefaultCheHost(ctx)
	if !done {
		return reconcile.Result{}, false, err
	}

	done, err = s.syncCheService(ctx)
	if !done {
		return reconcile.Result{}, false, err
	}

	done, err = s.exposeCheEndpoint(ctx)
	if !done {
		return reconcile.Result{}, false, err
	}

	return reconcile.Result{}, true, nil
}

func (s *CheHostReconciler) Finalize(ctx *deploy.DeployContext) error {
	return nil
}

func (s *CheHostReconciler) detectDefaultCheHost(ctx *deploy.DeployContext) (bool, error) {
	// only for OpenShift
	if !util.IsOpenShift || ctx.DefaultCheHost != "" {
		return true, nil
	}

	done, err := deploy.SyncRouteToCluster(
		ctx,
		getComponentName(ctx),
		"",
		"/",
		getServerExposingServiceName(ctx.CheCluster),
		8080,
		ctx.CheCluster.Spec.Server.CheServerRoute,
		getComponentName(ctx))
	if !done {
		return false, err
	}

	route := &routev1.Route{}
	exists, err := deploy.GetNamespacedObject(ctx, getComponentName(ctx), route)
	if !exists {
		return false, err
	}

	ctx.DefaultCheHost = route.Spec.Host
	return true, nil
}

func (s *CheHostReconciler) syncCheService(ctx *deploy.DeployContext) (bool, error) {
	portName := []string{"http"}
	portNumber := []int32{8080}

	if ctx.CheCluster.Spec.Metrics.Enable {
		portName = append(portName, "metrics")
		portNumber = append(portNumber, deploy.DefaultCheMetricsPort)
	}

	if ctx.CheCluster.Spec.Server.CheDebug == "true" {
		portName = append(portName, "debug")
		portNumber = append(portNumber, deploy.DefaultCheDebugPort)
	}

	spec := deploy.GetServiceSpec(ctx, deploy.CheServiceName, portName, portNumber, getComponentName(ctx))
	return deploy.Sync(ctx, spec, deploy.ServiceDefaultDiffOpts)
}

func (s CheHostReconciler) exposeCheEndpoint(ctx *deploy.DeployContext) (bool, error) {
	cheHost := ""
	exposedServiceName := getServerExposingServiceName(ctx.CheCluster)

	if !util.IsOpenShift {
		_, done, err := deploy.SyncIngressToCluster(
			ctx,
			getComponentName(ctx),
			ctx.CheCluster.Spec.Server.CheHost,
			"",
			exposedServiceName,
			8080,
			ctx.CheCluster.Spec.Server.CheServerIngress,
			getComponentName(ctx))
		if !done {
			return false, err
		}

		ingress := &networking.Ingress{}
		exists, err := deploy.GetNamespacedObject(ctx, getComponentName(ctx), ingress)
		if !exists {
			return false, err
		}

		cheHost = ingress.Spec.Rules[0].Host
	} else {
		customHost := ctx.CheCluster.Spec.Server.CheHost
		if ctx.DefaultCheHost == customHost {
			// let OpenShift set a hostname by itself since it requires a routes/custom-host permissions
			customHost = ""
		}

		done, err := deploy.SyncRouteToCluster(
			ctx,
			getComponentName(ctx),
			customHost,
			"/",
			exposedServiceName,
			8080,
			ctx.CheCluster.Spec.Server.CheServerRoute,
			getComponentName(ctx))
		if !done {
			return false, err
		}

		route := &routev1.Route{}
		exists, err := deploy.GetNamespacedObject(ctx, getComponentName(ctx), route)
		if !exists {
			return false, err
		}

		if customHost == "" {
			ctx.DefaultCheHost = route.Spec.Host
		}
		cheHost = route.Spec.Host
	}

	if ctx.CheCluster.Spec.Server.CheHost != cheHost {
		ctx.CheCluster.Spec.Server.CheHost = cheHost
		err := deploy.UpdateCheCRSpec(ctx, "CheHost URL", cheHost)
		return err == nil, err
	}

	return true, nil
}