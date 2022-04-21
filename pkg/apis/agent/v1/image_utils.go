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
	"sort"

	"github.com/stolostron/klusterlet-addon-controller/pkg/helpers/imageregistry"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Masterminds/semver"
	"github.com/stolostron/klusterlet-addon-controller/version"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// var defaultComponentImageKeyMap = map[string]string{
// 	"cert-policy-controller":          "cert_policy_controller",
// 	"addon-operator":                  "endpoint_component_operator",
// 	"coredns":                         "coredns",
// 	"deployable":                      "multicluster_operators_deployable",
// 	"iam-policy-controller":           "iam_policy_controller",
// 	"policy-controller":               "config_policy_controller",
// 	"governance-policy-spec-sync":     "governance_policy_spec_sync",
// 	"governance-policy-status-sync":   "governance_policy_status_sync",
// 	"governance-policy-template-sync": "governance_policy_template_sync",
// 	"router":                          "management_ingress",
// 	"search-collector":                "search_collector",
// 	"service-registry":                "multicloud_manager",
// 	"subscription":                    "multicluster_operators_subscription",
// 	"work-manager":                    "multicloud_manager",
// }
const ocmVersionLabel = "ocm-release-version"

// Manifest contains the manifest.
// The Manifest is loaded using the LoadManifest method.

var versionList []*semver.Version

var log = logf.Log.WithName("image_utils")

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
// if no version can match (major version), will return error
func getManifest(version string) (*manifest, error) {
	if len(versionList) == 0 || manifests == nil {
		return nil, fmt.Errorf("image manifest not loaded")
	}

	// find exact version first
	if m, ok := manifests[version]; ok {
		return &m, nil
	}
	log.Error(fmt.Errorf("Failed to find image manifest in version %s", version), "version not found")

	// find the version use ^
	currVersion, err := semver.NewVersion(version)
	if err != nil {
		log.Error(err, "not valid version "+version)
		return nil, err
	}

	versionConstraint, err := semver.NewConstraint(
		fmt.Sprintf("^%d.%d.%d", currVersion.Major(), currVersion.Minor(), currVersion.Patch()),
	)
	if err != nil {
		log.Error(err, "failed to generate semver constraint")
		return nil, err
	}
	// search for the first possible version
	// (used linear because versionList is very short)
	for _, v := range versionList {
		if isValid := versionConstraint.Check(v); isValid {
			if m, ok := manifests[v.Original()]; ok {
				return &m, nil
			}
		}
	}

	return nil, fmt.Errorf("version %s not supported", version)
}

// LoadConfigmaps - loads pre-release image manifests
func LoadConfigmaps(k8s client.Client) error {
	manifests = make(map[string]manifest)
	configmapList := &corev1.ConfigMapList{}

	err := k8s.List(context.TODO(), configmapList, client.MatchingLabels{"ocm-configmap-type": "image-manifest"})
	if err != nil {
		return err
	}

	for _, cm := range configmapList.Items {
		version := cm.Labels[ocmVersionLabel]
		v, err := semver.NewVersion(version)
		if err != nil {
			log.Error(err, "Invalid semantic version found in image-manifests")
			continue
		}
		m := manifest{}
		m.Images = make(map[string]string)
		m.Images = cm.Data
		manifests[v.Original()] = m

		versionList = append(versionList, v)
	}
	sort.Sort(semver.Collection(versionList))
	return nil
}
