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

// constants for search collector
const (
	SearchCollector = "klusterlet-addon-search"
	Search          = "search"
)

var log = logf.Log.WithName("search")

// IsEnabled - check whether search is enabled
func IsEnabled(instance *agentv1.KlusterletAddonConfig) bool {
	return instance.Spec.SearchCollectorConfig.Enabled
}

// NewSearchCollectorCR - create CR for component search collector
func NewSearchCollectorCR(instance *agentv1.KlusterletAddonConfig, namespace string) (*agentv1.SearchCollector, error) {
	labels := map[string]string{
		"app": instance.Name,
	}

	gv := agentv1.GlobalValues{
		ImagePullPolicy: instance.Spec.ImagePullPolicy,
		ImagePullSecret: instance.Spec.ImagePullSecret,
		ImageOverrides:  make(map[string]string, 1),
	}

	imageKey, imageRepository, err := instance.GetImage("search-collector")
	if err != nil {
		log.Error(err, "Fail to get Image", "Component.Name", "search-collector")
		return nil, err
	}

	gv.ImageOverrides[imageKey] = imageRepository

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
			ClusterName:         instance.Spec.ClusterName,
			ClusterNamespace:    instance.Spec.ClusterNamespace,
			HubKubeconfigSecret: Search + "-hub-kubeconfig",
			GlobalValues:        gv,
		},
	}, err
}
