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

// constants for component CRs
const (
	IAMPolicyController = "klusterlet-addon-iampolicyctrl"
	IAMPolicyCtrl       = "iampolicyctrl"
)

var log = logf.Log.WithName("iampolicyctrl")

// IsEnabled - check whether iampolicyctrl is enabled
func IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.IAMPolicyControllerConfig.Enabled
}

// NewIAMPolicyControllerCR - create CR for component iam poliicy controller
func NewIAMPolicyControllerCR(
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

	imageKey, imageRepository, err := instance.GetImage("iam-policy-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "iam-policy")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

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
			FullNameOverride: IAMPolicyController,
			ClusterName:      instance.Spec.ClusterName,
			ClusterNamespace: instance.Spec.ClusterNamespace,
			GlobalValues:     gv,
		},
	}, err
}
