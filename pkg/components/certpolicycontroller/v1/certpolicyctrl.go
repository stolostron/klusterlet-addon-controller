// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
)

// constants for cert policy controller
const (
	CertPolicyController  = "klusterlet-addon-certpolicyctrl"
	CertPolicyCtrl        = "certpolicyctrl"
	RequiresHubKubeConfig = false
)

var log = logf.Log.WithName("certpolicyctrl")

// IsEnabled - check whether certpolicyctrl is enabled
func IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.CertPolicyControllerConfig.Enabled
}

// NewCertPolicyControllerCR - create CR for component cert policy controller
func NewCertPolicyControllerCR(
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

	imageKey, imageRepository, err := instance.GetImage("cert-policy-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cert-policy")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

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
			FullNameOverride: CertPolicyController,
			ClusterName:      instance.Spec.ClusterName,
			ClusterNamespace: instance.Spec.ClusterNamespace,
			GlobalValues:     gv,
		},
	}, err
}
