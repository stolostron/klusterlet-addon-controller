// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// const of appmgr
const (
	ApplicationManager      = "klusterlet-addon-appmgr"
	AppMgr                  = "appmgr"
	RequiresHubKubeConfig   = true
	managedClusterAddOnName = "application-manager"
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

func (addon AddonAppMgr) NewAddonCR(instance *agentv1.KlusterletAddonConfig, namespace string) (runtime.Object, error) {
	return newApplicationManagerCR(instance, namespace)
}

func (addon AddonAppMgr) GetManagedClusterAddOnName() string {
	return managedClusterAddOnName
}

// newApplicationManagerCR - create CR for component application manager
func newApplicationManagerCR(
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (*agentv1.ApplicationManager, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 2),
	}

	imageRepository, err := instance.GetImage("multicluster_operators_subscription")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "subscription")
		return nil, err
	}
	gv.ImageOverrides["multicluster_operators_subscription"] = imageRepository

	if imageRepositoryLease, err := instance.GetImage("klusterlet_addon_lease_controller"); err != nil {
		log.Error(err, "Fail to get Image", "Image.Key", "klusterlet_addon_lease_controller")
	} else {
		gv.ImageOverrides["klusterlet_addon_lease_controller"] = imageRepositoryLease
	}

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
			HubKubeconfigSecret: AppMgr + "-hub-kubeconfig",
			ClusterName:         instance.Spec.ClusterName,
			ClusterNamespace:    instance.Spec.ClusterNamespace,
			GlobalValues:        gv,
		},
	}, nil
}
