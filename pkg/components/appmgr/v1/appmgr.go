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
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/stolostron/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// const of appmgr
const (
	ApplicationManager      = "klusterlet-addon-appmgr"
	AppMgr                  = "appmgr"
	RequiresHubKubeConfig   = true
	managedClusterAddOnName = "application-manager"
	addonNameEnv            = "APPMGR_NAME"
)

var log = logf.Log.WithName("appmgr")

type AddonAppMgr struct{}

// IsEnabled - check whether appmgr is enabled
func (addon AddonAppMgr) IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.ApplicationManagerConfig.Enabled
}

func (addon AddonAppMgr) CheckHubKubeconfigRequired() bool {
	return RequiresHubKubeConfig
}

func (addon AddonAppMgr) GetAddonName() string {
	return AppMgr
}

func (addon AddonAppMgr) NewAddonCR(addonAgentConfig *agentv1.AddonAgentConfig, namespace string) (runtime.Object, error) {
	return newApplicationManagerCR(addonAgentConfig, namespace)
}

func (addon AddonAppMgr) GetManagedClusterAddOnName() string {
	if n := os.Getenv(addonNameEnv); len(n) != 0 {
		return n
	}
	log.Info("failed to get addon name from env var " + addonNameEnv)
	return managedClusterAddOnName
}

// newApplicationManagerCR - create CR for component application manager
func newApplicationManagerCR(
	addonAgentConfig *agentv1.AddonAgentConfig,
	namespace string,
) (*agentv1.ApplicationManager, error) {
	labels := map[string]string{
		"app": addonAgentConfig.ClusterName,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: addonAgentConfig.ImagePullPolicy,
		ImagePullSecret: addonAgentConfig.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 2),
		NodeSelector:    addonAgentConfig.NodeSelector,
	}
	switch addonAgentConfig.KlusterletAddonConfig.Spec.ApplicationManagerConfig.ProxyPolicy {
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

	imageRepository, err := addonAgentConfig.GetImage("multicluster_operators_subscription")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "subscription")
		return nil, err
	}
	gv.ImageOverrides["multicluster_operators_subscription"] = imageRepository

	return &agentv1.ApplicationManager{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "ApplicationManager",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ApplicationManager,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: agentv1.ApplicationManagerSpec{
			FullNameOverride:    ApplicationManager,
			HubKubeconfigSecret: managedClusterAddOnName + "-hub-kubeconfig",
			ClusterName:         addonAgentConfig.ClusterName,
			ClusterNamespace:    addonAgentConfig.ClusterName,
			GlobalValues:        gv,
		},
	}, nil
}
