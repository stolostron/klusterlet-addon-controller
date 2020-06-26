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

// const of cispolicyctrl
const (
	CISPolicyController = "klusterlet-addon-cispolicyctrl"
	CISPolicyCtrl       = "cispolicyctrl"
)

var log = logf.Log.WithName("cispolicyctrl")

// IsEnabled - check whether cispolicyctrl is enabled
func IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.CISPolicyControllerConfig.Enabled
}

// NewCISPolicyControllerCR - create CR for component cis policy controller
func NewCISPolicyControllerCR(
	instance *agentv1.KlusterletAddonConfig,
	namespace string,
) (*agentv1.CISPolicyController, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 5),
	}

	imageKey, imageRepository, err := instance.GetImage("cis-controller-controller")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-controller")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-crawler")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-crawler")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-drishti")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-drishti")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-minio")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-minio")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	imageKey, imageRepository, err = instance.GetImage("cis-controller-minio-cleaner")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "cis-controller-minio-cleaner")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	return &agentv1.CISPolicyController{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "CISPolicyController",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      CISPolicyController,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: agentv1.CISPolicyControllerSpec{
			FullNameOverride: CISPolicyController,
			ClusterName:      instance.Spec.ClusterName,
			ClusterNamespace: instance.Spec.ClusterNamespace,
			GlobalValues:     gv,
		},
	}, err
}
