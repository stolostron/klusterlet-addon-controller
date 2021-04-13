// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package csr

import (
	"fmt"
	"strings"

	addons "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components"
	certificatesv1 "k8s.io/api/certificates/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func newCSRPredicate() predicate.Predicate {
	return predicate.Predicate(predicate.Funcs{
		GenericFunc: func(e event.GenericEvent) bool { return false },
		CreateFunc: func(e event.CreateEvent) bool {
			labels := e.Meta.GetLabels()

			// ignore csr without any label
			if len(labels) == 0 {
				return false
			}

			// ignore csr without cluster name label
			clusterName, ok := labels[clusterNameLabel]
			if !ok {
				return false
			}

			// ignore csr without addon name label
			managedClusterAddonName := labels[managedClusterAddonNameLabel]
			addon, err := addons.GetAddonFromManagedClusterAddonName(managedClusterAddonName)
			if err != nil {
				return false
			}

			// ignore csr if the addon does not need hub kubeconfig
			if required := addon.CheckHubKubeconfigRequired(); !required {
				return false
			}

			csr, ok := e.Object.(*certificatesv1.CertificateSigningRequest)
			if !ok {
				return false
			}

			// ignore csr whose signer is not "kubernetes.io/kube-apiserver-client"
			if csr.Spec.SignerName != certificatesv1.KubeAPIServerClientSignerName {
				return false
			}

			// ignore csr which is not requested by registration agent
			requestorPrefix := fmt.Sprintf("system:open-cluster-management:%s:", clusterName)
			return strings.HasPrefix(csr.Spec.Username, requestorPrefix)
		},
		DeleteFunc: func(e event.DeleteEvent) bool { return false },
		UpdateFunc: func(e event.UpdateEvent) bool { return false },
	})
}
