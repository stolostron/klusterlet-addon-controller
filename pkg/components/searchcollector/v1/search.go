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

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
)

// constants for search collector
const (
	SearchCollector         = "klusterlet-addon-search"
	Search                  = "search"
	RequiresHubKubeConfig   = true
	managedClusterAddOnName = "search-collector"
	addonNameEnv            = "SEARCH_NAME"
)

var log = logf.Log.WithName("search")

type AddonSearch struct{}

func (addon AddonSearch) IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.SearchCollectorConfig.Enabled
}

func (addon AddonSearch) CheckHubKubeconfigRequired() bool {
	return RequiresHubKubeConfig
}

func (addon AddonSearch) GetAddonName() string {
	return Search
}

func (addon AddonSearch) NewAddonCR(addonAgentConfig *agentv1.AddonAgentConfig, namespace string) (runtime.Object, error) {
	return newSearchCollectorCR(addonAgentConfig, namespace)
}

func (addon AddonSearch) GetManagedClusterAddOnName() string {
	if n := os.Getenv(addonNameEnv); len(n) != 0 {
		return n
	}
	log.Info("failed to get addon name from env var " + addonNameEnv)
	return managedClusterAddOnName
}

// newSearchCollectorCR - create CR for component search collector
func newSearchCollectorCR(addonAgentConfig *agentv1.AddonAgentConfig, namespace string) (*agentv1.SearchCollector, error) {
	labels := map[string]string{
		"app": addonAgentConfig.ClusterName,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: addonAgentConfig.ImagePullPolicy,
		ImagePullSecret: addonAgentConfig.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
		NodeSelector:    addonAgentConfig.NodeSelector,
	}

	switch addonAgentConfig.KlusterletAddonConfig.Spec.SearchCollectorConfig.ProxyPolicy {
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

	imageRepository, err := addonAgentConfig.GetImage("search_collector")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "search-collector")
		return nil, err
	}
	gv.ImageOverrides["search_collector"] = imageRepository

	return &agentv1.SearchCollector{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "SearchCollector",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      SearchCollector,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: agentv1.SearchCollectorSpec{
			FullNameOverride:    SearchCollector,
			ClusterName:         addonAgentConfig.ClusterName,
			ClusterNamespace:    addonAgentConfig.ClusterName,
			HubKubeconfigSecret: managedClusterAddOnName + "-hub-kubeconfig",
			GlobalValues:        gv,
		},
	}, err
}
