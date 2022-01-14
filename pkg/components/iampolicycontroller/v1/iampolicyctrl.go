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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
)

// constants for component CRs
const (
	IAMPolicyController     = "klusterlet-addon-iampolicyctrl"
	IAMPolicyCtrl           = "iampolicyctrl"
	RequiresHubKubeConfig   = true
	managedClusterAddOnName = "iam-policy-controller"
	addonNameEnv            = "IAMPOLICYCTRL_NAME"
)

var log = logf.Log.WithName("iampolicyctrl")

type AddonIAMPolicyCtrl struct{}

func (addon AddonIAMPolicyCtrl) IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.IAMPolicyControllerConfig.Enabled
}
func (addon AddonIAMPolicyCtrl) CheckHubKubeconfigRequired() bool {
	return RequiresHubKubeConfig
}

func (addon AddonIAMPolicyCtrl) GetAddonName() string {
	return IAMPolicyCtrl
}

func (addon AddonIAMPolicyCtrl) NewAddonCR(
	addonAgentConfig *agentv1.AddonAgentConfig,
	namespace string,
) (runtime.Object, error) {
	return newIAMPolicyControllerCR(addonAgentConfig, namespace)
}

func (addon AddonIAMPolicyCtrl) GetManagedClusterAddOnName() string {
	if n := os.Getenv(addonNameEnv); len(n) != 0 {
		return n
	}
	log.Info("failed to get addon name from env var " + addonNameEnv)
	return managedClusterAddOnName
}

// newIAMPolicyControllerCR - create CR for component iam poliicy controller
func newIAMPolicyControllerCR(
	addonAgentConfig *agentv1.AddonAgentConfig,
	namespace string,
) (*agentv1.IAMPolicyController, error) {
	labels := map[string]string{
		"app": addonAgentConfig.ClusterName,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: addonAgentConfig.ImagePullPolicy,
		ImagePullSecret: addonAgentConfig.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
		NodeSelector:    addonAgentConfig.NodeSelector,
	}
	switch addonAgentConfig.KlusterletAddonConfig.Spec.IAMPolicyControllerConfig.ProxyPolicy {
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

	imageRepository, err := addonAgentConfig.GetImage("iam_policy_controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "iam-policy")
		return nil, err
	}
	gv.ImageOverrides["iam_policy_controller"] = imageRepository

	return &agentv1.IAMPolicyController{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "IAMPolicyController",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      IAMPolicyController,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: agentv1.IAMPolicyControllerSpec{
			FullNameOverride:    IAMPolicyController,
			HubKubeconfigSecret: managedClusterAddOnName + "-hub-kubeconfig",
			ClusterName:         addonAgentConfig.ClusterName,
			ClusterNamespace:    addonAgentConfig.ClusterName,
			GlobalValues:        gv,
		},
	}, err
}
