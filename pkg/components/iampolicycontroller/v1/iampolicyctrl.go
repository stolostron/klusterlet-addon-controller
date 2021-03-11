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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/open-cluster-management/klusterlet-addon-controller/pkg/apis/agent/v1"
)

// constants for component CRs
const (
	IAMPolicyController     = "klusterlet-addon-iampolicyctrl"
	IAMPolicyCtrl           = "iampolicyctrl"
	RequiresHubKubeConfig   = true
	managedClusterAddOnName = "iam-policy-controller"
	addonClusterRoleEnv     = "IAMPOLICYCTRL_CLUSTERROLE_NAME"
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
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (runtime.Object, error) {
	return newIAMPolicyControllerCR(instance, namespace)
}

func (addon AddonIAMPolicyCtrl) GetManagedClusterAddOnName() string {
	return managedClusterAddOnName
}

func (addon AddonIAMPolicyCtrl) GetClusterRoleName() string {
	if n := os.Getenv(addonClusterRoleEnv); len(n) == 0 {
		return n
	}
	log.Error(fmt.Errorf("env var %s not found", addonClusterRoleEnv),
		"failed to get clusterrole name")
	return addon.GetManagedClusterAddOnName()
}

// newIAMPolicyControllerCR - create CR for component iam poliicy controller
func newIAMPolicyControllerCR(
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (*agentv1.IAMPolicyController, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
	}

	imageRepository, err := instance.GetImage("iam_policy_controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "iam-policy")
		return nil, err
	}
	gv.ImageOverrides["iam_policy_controller"] = imageRepository

	if imageRepositoryLease, err := instance.GetImage("klusterlet_addon_lease_controller"); err != nil {
		log.Error(err, "Fail to get Image", "Image.Key", "klusterlet_addon_lease_controller")
	} else {
		gv.ImageOverrides["klusterlet_addon_lease_controller"] = imageRepositoryLease
	}

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
			HubKubeconfigSecret: IAMPolicyCtrl + "-hub-kubeconfig",
			ClusterName:         instance.Spec.ClusterName,
			ClusterNamespace:    instance.Spec.ClusterNamespace,
			GlobalValues:        gv,
		},
	}, err
}
