// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	"os"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	addonoperator "github.com/stolostron/klusterlet-addon-controller/pkg/components/addon-operator/v1"
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
	addonNameEnv            = "POLICYCTRL_NAME"
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
	addonAgentConfig *agentv1.AddonAgentConfig,
	namespace string,
) (runtime.Object, error) {
	return newPolicyControllerCR(addonAgentConfig, namespace)
}

func (addon AddonPolicyCtrl) GetManagedClusterAddOnName() string {
	if n := os.Getenv(addonNameEnv); len(n) != 0 {
		return n
	}
	log.Info("failed to get addon name from env var " + addonNameEnv)
	return managedClusterAddOnName
}

// newPolicyControllerCR - create CR for component poliicy controller
func newPolicyControllerCR(
	addonAgentConfig *agentv1.AddonAgentConfig,
	namespace string,
) (*agentv1.PolicyController, error) {
	labels := map[string]string{
		"app": addonAgentConfig.ClusterName,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: addonAgentConfig.ImagePullPolicy,
		ImagePullSecret: addonAgentConfig.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
		NodeSelector:    addonAgentConfig.NodeSelector,
	}

	switch addonAgentConfig.KlusterletAddonConfig.Spec.PolicyController.ProxyPolicy {
	case agentv1.ProxyPolicyOCPGlobalProxy:
		gv.ProxyConfig = map[string]string{
			agentv1.HTTPProxy:  addonAgentConfig.KlusterletAddonConfig.Status.OCPGlobalProxy.HTTPProxy,
			agentv1.HTTPSProxy: addonAgentConfig.KlusterletAddonConfig.Status.OCPGlobalProxy.HTTPSProxy,
			agentv1.NoProxy:    addonAgentConfig.KlusterletAddonConfig.Status.OCPGlobalProxy.NoProxy,
		}
	case agentv1.ProxyPolicyCustomProxy:
		gv.ProxyConfig = map[string]string{
			agentv1.HTTPProxy:  addonAgentConfig.KlusterletAddonConfig.Spec.ProxyConfig.HTTPProxy,
			agentv1.HTTPSProxy: addonAgentConfig.KlusterletAddonConfig.Spec.ProxyConfig.HTTPSProxy,
			agentv1.NoProxy:    addonAgentConfig.KlusterletAddonConfig.Spec.ProxyConfig.NoProxy,
		}
	}

	imageRepository, err := addonAgentConfig.GetImage("config_policy_controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "policy-controller")
		return nil, err
	}

	gv.ImageOverrides["config_policy_controller"] = imageRepository

	imageRepository, err = addonAgentConfig.GetImage("governance_policy_spec_sync")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "governance-policy-spec-sync")
		return nil, err
	}

	gv.ImageOverrides["governance_policy_spec_sync"] = imageRepository

	imageRepository, err = addonAgentConfig.GetImage("governance_policy_status_sync")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "governance-policy-status-sync")
		return nil, err
	}

	gv.ImageOverrides["governance_policy_status_sync"] = imageRepository

	imageRepository, err = addonAgentConfig.GetImage("governance_policy_template_sync")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "governance-policy-template-sync")
		return nil, err
	}
	gv.ImageOverrides["governance_policy_template_sync"] = imageRepository

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
			ClusterName:                 addonAgentConfig.ClusterName,
			ClusterNamespace:            addonAgentConfig.ClusterName,
			HubKubeconfigSecret:         managedClusterAddOnName + "-hub-kubeconfig",
			GlobalValues:                gv,
			DeployedOnHub:               false,
			PostDeleteJobServiceAccount: addonoperator.KlusterletAddonOperator,
		},
	}, nil
}
