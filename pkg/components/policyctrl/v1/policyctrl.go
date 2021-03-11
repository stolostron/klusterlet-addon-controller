// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	"fmt"
	"os"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	addonoperator "github.com/open-cluster-management/klusterlet-addon-controller/pkg/components/addon-operator/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// constants for policy controller
const (
	PolicyController        = "klusterlet-addon-policyctrl"
	PolicyCtrl              = "policyctrl"
	RequiresHubKubeConfig   = true
	managedClusterAddOnName = "policy-controller"
	addonClusterRoleEnv     = "POLICYCTRL_CLUSTERROLE_NAME"
)

var log = logf.Log.WithName("policyctrl")

type AddonPolicyCtrl struct{}

func (addon AddonPolicyCtrl) IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.PolicyController.Enabled
}

func (addon AddonPolicyCtrl) CheckHubKubeconfigRequired() bool {
	return RequiresHubKubeConfig
}

func (addon AddonPolicyCtrl) GetAddonName() string {
	return PolicyCtrl
}

func (addon AddonPolicyCtrl) NewAddonCR(
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (runtime.Object, error) {
	return newPolicyControllerCR(instance, namespace)
}

func (addon AddonPolicyCtrl) GetManagedClusterAddOnName() string {
	return managedClusterAddOnName
}

func (addon AddonPolicyCtrl) GetClusterRoleName() string {
	if n := os.Getenv(addonClusterRoleEnv); len(n) != 0 {
		return n
	}
	log.Error(fmt.Errorf("env var %s not found", addonClusterRoleEnv),
		"failed to get clusterrole name")
	return addon.GetManagedClusterAddOnName()
}

// newPolicyControllerCR - create CR for component poliicy controller
func newPolicyControllerCR(
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (*agentv1.PolicyController, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
	}

	imageRepository, err := instance.GetImage("config_policy_controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "policy-controller")
		return nil, err
	}

	gv.ImageOverrides["config_policy_controller"] = imageRepository

	imageRepository, err = instance.GetImage("governance_policy_spec_sync")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "governance-policy-spec-sync")
		return nil, err
	}

	gv.ImageOverrides["governance_policy_spec_sync"] = imageRepository

	imageRepository, err = instance.GetImage("governance_policy_status_sync")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "governance-policy-status-sync")
		return nil, err
	}

	gv.ImageOverrides["governance_policy_status_sync"] = imageRepository

	imageRepository, err = instance.GetImage("governance_policy_template_sync")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "governance-policy-template-sync")
		return nil, err
	}
	gv.ImageOverrides["governance_policy_template_sync"] = imageRepository

	if imageRepositoryLease, err := instance.GetImage("klusterlet_addon_lease_controller"); err != nil {
		log.Error(err, "Fail to get Image", "Image.Key", "klusterlet_addon_lease_controller")
	} else {
		gv.ImageOverrides["klusterlet_addon_lease_controller"] = imageRepositoryLease
	}

	return &agentv1.PolicyController{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "PolicyController",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      PolicyController,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: agentv1.PolicyControllerSpec{
			FullNameOverride:            PolicyController,
			ClusterName:                 instance.Spec.ClusterName,
			ClusterNamespace:            instance.Spec.ClusterNamespace,
			HubKubeconfigSecret:         PolicyCtrl + "-hub-kubeconfig",
			GlobalValues:                gv,
			DeployedOnHub:               false,
			PostDeleteJobServiceAccount: addonoperator.KlusterletAddonOperator,
		},
	}, nil
}
