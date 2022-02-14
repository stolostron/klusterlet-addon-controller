// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

// Package klusterletaddon contains the main reconcile function & related functions for klusterletAddonConfigs
package klusterletaddon

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	addonoperator "github.com/stolostron/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	"github.com/stolostron/klusterlet-addon-controller/pkg/utils"
	manifestworkv1 "open-cluster-management.io/api/work/v1"
)

// const for addon operator
const (
	KlusterletAddonOperatorPostfix = "-klusterlet-addon-operator"
)

// createManifestWorkComponentOperator - creates manifest work for klusterlet addon operator
func createManifestWorkComponentOperator(
	addonAgentConfig *agentv1.AddonAgentConfig,
	r *ReconcileKlusterletAddon) error {

	var manifests []manifestworkv1.Manifest

	// create namespace
	klusterletaddonNamespace := addonoperator.NewNamespace()

	// Create Component Operator ClusteRole
	clusterRole := addonoperator.NewClusterRole(addonAgentConfig)

	// create cluster role binding
	clusterRoleBinding := addonoperator.NewClusterRoleBinding(addonAgentConfig)

	// create service account
	serviceAccount := addonoperator.NewServiceAccount(addonAgentConfig, addonoperator.KlusterletAddonNamespace)

	// create imagePullSecret
	imagePullSecret, err := addonoperator.NewImagePullSecret(addonAgentConfig.ImagePullSecretNamespace,
		addonAgentConfig.ImagePullSecret, r.client)
	if err != nil {
		log.Error(err, "Fail to create imagePullSecret")
		return err
	}

	// create deployment for klusterlet addon operator
	deployment, err := addonoperator.NewDeployment(addonAgentConfig, addonoperator.KlusterletAddonNamespace)
	if err != nil {
		log.Error(err, "Fail to create desired klusterlet addon operator deployment")
		return err
	}
	// add namespace, clusterrole, clusterrolebinding, serviceaccount
	nsManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: klusterletaddonNamespace}}
	crManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRole}}
	crbManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: clusterRoleBinding}}
	saManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: serviceAccount}}
	manifests = append(manifests, nsManifest, crManifest, crbManifest, saManifest)
	// add imagePullSecret
	if imagePullSecret != nil {
		ipsManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: imagePullSecret}}
		manifests = append(manifests, ipsManifest)
	}
	// add deployment
	dplManifest := manifestworkv1.Manifest{RawExtension: runtime.RawExtension{Object: deployment}}
	manifests = append(manifests, dplManifest)

	manifestWork := &manifestworkv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonAgentConfig.ClusterName + KlusterletAddonOperatorPostfix,
			Namespace: addonAgentConfig.ClusterName,
		},
		Spec: manifestworkv1.ManifestWorkSpec{
			DeleteOption: &manifestworkv1.DeleteOption{
				PropagationPolicy: manifestworkv1.DeletePropagationPolicyTypeSelectivelyOrphan,
				SelectivelyOrphan: &manifestworkv1.SelectivelyOrphan{
					OrphaningRules: []manifestworkv1.OrphaningRule{
						{
							Resource: "namespaces",
							Name:     agentv1.KlusterletAddonNamespace,
						},
					},
				},
			},
			Workload: manifestworkv1.ManifestsTemplate{
				Manifests: manifests,
			},
		},
	}

	if err := utils.CreateOrUpdateManifestWork(manifestWork, r.client, addonAgentConfig.KlusterletAddonConfig, r.scheme); err != nil {
		log.Error(err, "Failed to create manifest work for component")
		return err
	}

	return nil
}
