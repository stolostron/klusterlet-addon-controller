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
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	agentv1 "github.com/open-cluster-management/endpoint-operator/pkg/apis/agent/v1"
)

// constants for work manager
const (
	WorkManager           = "klusterlet-addon-workmgr"
	WorkMgr               = "workmgr"
	RequiresHubKubeConfig = true
)

var log = logf.Log.WithName("workmgr")

// NewWorkManagerCR - create CR for component work manager
func NewWorkManagerCR(instance *agentv1.KlusterletAddonConfig,
	client client.Client, namespace string) (*agentv1.WorkManager, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
	}

	imageKey, imageRepository, err := instance.GetImage("work-manager")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "work-manager")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

	clusterLabels := instance.Spec.ClusterLabels

	return &agentv1.WorkManager{
		TypeMeta: metav1.TypeMeta{
			APIVersion: agentv1.SchemeGroupVersion.String(),
			Kind:       "WorkManager",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      WorkManager,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: agentv1.WorkManagerSpec{
			FullNameOverride: WorkManager,

			ClusterName:      instance.Spec.ClusterName,
			ClusterNamespace: instance.Spec.ClusterNamespace,
			ClusterLabels:    clusterLabels,

			HubKubeconfigSecret: WorkMgr + "-hub-kubeconfig",

			GlobalValues: gv,
		},
	}, nil
}
