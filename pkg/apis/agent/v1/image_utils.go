// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	"context"
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stolostron/cluster-lifecycle-api/helpers/imageregistry"
	"github.com/stolostron/klusterlet-addon-controller/version"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
)

//	var defaultComponentImageKeyMap = map[string]string{
//		"cert-policy-controller":          "cert_policy_controller",
//		"addon-operator":                  "endpoint_component_operator",
//		"coredns":                         "coredns",
//		"deployable":                      "multicluster_operators_deployable",
//		"policy-controller":               "config_policy_controller",
//		"governance-policy-spec-sync":     "governance_policy_spec_sync",
//		"governance-policy-status-sync":   "governance_policy_status_sync",
//		"governance-policy-template-sync": "governance_policy_template_sync",
//		"router":                          "management_ingress",
//		"search-collector":                "search_collector",
//		"service-registry":                "multicloud_manager",
//		"subscription":                    "multicluster_operators_subscription",
//		"work-manager":                    "multicloud_manager",
//	}
const ocmVersionLabel = "ocm-release-version"

// Manifest contains the manifest.
// The Manifest is loaded using the LoadManifest method.

var versionList []string

type manifest struct {
	Images map[string]string
}

var manifests map[string]manifest

// GetImage returns the image.  for the specified component return error if information not found
func (config *AddonAgentConfig) GetImage(component string) (imageRepository string, err error) {
	m, err := getManifest(version.Version)
	if err != nil {
		return "", err
	}

	image := m.Images[component]
	if image == "" {
		return "", fmt.Errorf("addon image not found")
	}

	return imageregistry.OverrideImageByAnnotation(config.ManagedCluster.GetAnnotations(), image)
}

// GetImage returns the image.  for the specified component return error if information not found
func GetImage(managedCluster *clusterv1.ManagedCluster, component string) (string, error) {
	m, err := getManifest(version.Version)
	if err != nil {
		return "", err
	}

	image := m.Images[component]
	if image == "" {
		return "", fmt.Errorf("addon image not found")
	}

	return imageregistry.OverrideImageByAnnotation(managedCluster.GetAnnotations(), image)
}

// getManifest returns the manifest that is best matching the required version
func getManifest(version string) (*manifest, error) {
	if len(versionList) == 0 || manifests == nil {
		return nil, fmt.Errorf("image manifest not loaded")
	}

	if m, ok := manifests[version]; ok {
		return &m, nil
	}

	return nil, fmt.Errorf("version %s not supported", version)
}

// LoadImages - loads image manifests from configmap, if configmap is not found get from env
func LoadImages(k8s client.Client) error {
	manifests = make(map[string]manifest)
	m := manifest{Images: make(map[string]string)}
	configmapList := &corev1.ConfigMapList{}

	err := k8s.List(context.TODO(), configmapList, client.MatchingLabels{"ocm-configmap-type": "image-manifest"})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if len(configmapList.Items) == 0 {
		for envImageName, imageName := range EnvImageNameMap {
			image := os.Getenv(envImageName)
			if image != "" {
				m.Images[imageName] = image
			}
		}
		if len(m.Images) == 0 {
			// the images are used only in image override case. so no need to return error here
			klog.Warningf("no image manifest loaded from configmap or Env")
			return nil
		}
		manifests[version.Version] = m
		versionList = append(versionList, version.Version)
		return nil
	}

	for _, cm := range configmapList.Items {
		omcVersion := cm.Labels[ocmVersionLabel]
		m.Images = cm.Data
		manifests[omcVersion] = m
		versionList = append(versionList, omcVersion)
	}
	return nil
}
