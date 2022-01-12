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

// constants for cert policy controller
const (
	CertPolicyController    = "klusterlet-addon-certpolicyctrl"
	CertPolicyCtrl          = "certpolicyctrl"
	RequiresHubKubeConfig   = true
	managedClusterAddOnName = "cert-policy-controller"
	addonNameEnv            = "CERTPOLICYCTRL_NAME"
)

var log = logf.Log.WithName("certpolicyctrl")

type AddonCertPolicyCtrl struct{}

func (addon AddonCertPolicyCtrl) IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.CertPolicyControllerConfig.Enabled
}

func (addon AddonCertPolicyCtrl) CheckHubKubeconfigRequired() bool {
	return RequiresHubKubeConfig
}

func (addon AddonCertPolicyCtrl) GetAddonName() string {
	return CertPolicyCtrl
}

func (addon AddonCertPolicyCtrl) NewAddonCR(
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (runtime.Object, error) {
	return newCertPolicyControllerCR(instance, namespace)
}

func (addon AddonCertPolicyCtrl) GetManagedClusterAddOnName() string {
	if n := os.Getenv(addonNameEnv); len(n) != 0 {
		return n
	}
	log.Info("failed to get addon name from env var " + addonNameEnv)
	return managedClusterAddOnName
}

// newCertPolicyControllerCR - create CR for component cert policy controller
func newCertPolicyControllerCR(
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (*agentv1.CertPolicyController, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
	}

	imageRepository, err := instance.GetImage("cert_policy_controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cert-policy")
		return nil, err
	}
	gv.ImageOverrides["cert_policy_controller"] = imageRepository

	return &agentv1.CertPolicyController{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "CertPolicyController",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      CertPolicyController,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: agentv1.CertPolicyControllerSpec{
			FullNameOverride:    CertPolicyController,
			HubKubeconfigSecret: managedClusterAddOnName + "-hub-kubeconfig",
			ClusterName:         instance.Spec.ClusterName,
			ClusterNamespace:    instance.Spec.ClusterNamespace,
			GlobalValues:        gv,
		},
	}, err
}
