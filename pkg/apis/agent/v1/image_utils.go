// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package v1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var defaultComponentImageKeyMap = map[string]string{
	"cert-policy-controller":          "cert_policy_controller",
	"addon-operator":                  "endpoint_component_operator",
	"coredns":                         "coredns",
	"deployable":                      "multicluster_operators_deployable",
	"iam-policy-controller":           "iam_policy_controller",
	"policy-controller":               "config_policy_controller",
	"governance-policy-spec-sync":     "governance_policy_spec_sync",
	"governance-policy-status-sync":   "governance_policy_status_sync",
	"governance-policy-template-sync": "governance_policy_template_sync",
	"router":                          "management_ingress",
	"search-collector":                "search_collector",
	"service-registry":                "multicloud_manager",
	"subscription":                    "multicluster_operators_subscription",
	"work-manager":                    "multicloud_manager",
}

//Manifest contains the manifest.
//The Manifest is loaded using the LoadManifest method.

var versionList []*semver.Version

var log = logf.Log.WithName("image_utils")

type manifest struct {
	Images []manifestElement `json:"inline"`
}

var manifests map[string]manifest

type manifestElement struct {
	ImageKey        string `json:"image-key,omitempty"`
	ImageName       string `json:"image-name,omitempty"`
	ImageVersion    string `json:"image-version,omitempty"`
	ImageTag        string `json:"image-tag,omitempty"`
	ImageDigest     string `json:"image-digest,omitempty"`
	ImageRepository string `json:"image-remote,omitempty"`
	GitSha256       string `json:"git-sha256,omitempty"`
	GitRepository   string `json:"git-repository,omitempty"`
}

func init() {
	manifests = make(map[string]manifest)

	manifestDir := "image-manifests"
	parentDir := os.Getenv("IMAGE_MANIFEST_PATH")

	if parentDir != "" {
		manifestDir = filepath.Join(parentDir, "image-manifests")
	}

	err := LoadManifests(manifestDir)
	if err != nil {
		log.Error(err, "Error while getting version lists")
	}
}

// GetImage returns the image.Image,  for the specified component return error if information not found
func (instance KlusterletAddonConfig) GetImage(component string) (imageKey, imageRepository string, err error) {

	if v, ok := defaultComponentImageKeyMap[component]; ok {
		imageKey = v
	} else {
		return "", "", fmt.Errorf("unable to locate default image name for component %s", component)
	}

	imageManifest, err := getImageManifestElement(instance.Spec.Version, imageKey)
	if err != nil {
		return "", "", err
	}

	imageKey = imageManifest.ImageKey

	if instance.Spec.ImageRegistry != "" {
		imageRepository = instance.Spec.ImageRegistry
	} else {
		imageRepository = imageManifest.ImageRepository
	}

	imageRepository = imageRepository + "/" + imageManifest.ImageName + "@" + imageManifest.ImageDigest

	return imageKey, imageRepository, nil
}

//getImageManifestElement returns the *manifestElement and nil if not found
//Return an error only if the manifest is malformed
func getImageManifestElement(version, imageKey string) (*manifestElement, error) {
	m, err := getManifest(version)
	if err != nil {
		return nil, err
	}

	for i, im := range m.Images {
		if im.ImageKey == imageKey {
			return &m.Images[i], nil
		}
	}
	return nil, fmt.Errorf("ImageManifest not found for %s", imageKey)
}

//readManifestFile returns the *manifest and nil if not found
func readManifestFile(manifestPath string) (*manifest, error) {
	b, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	m := manifest{}
	m.Images = make([]manifestElement, 0)
	err = yaml.Unmarshal(b, &m.Images)
	if err != nil {
		return nil, err
	}

	return &m, nil
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

// LoadManifests loads the available version list of klusterlet
func LoadManifests(manifestDirPath string) error {
	files, err := ioutil.ReadDir(manifestDirPath)
	if err != nil {
		log.Error(err, "Fail to read manifest directory", "path", manifestDirPath)
		return err
	}

	c, err := semver.NewConstraint(">= 2.0.0")
	if err != nil {
		log.Error(err, "Invalid semantic constraint")
	}

	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), ".json") {
			manifestFileName := filepath.Join(manifestDirPath, file.Name())
			version := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			v, err := semver.NewVersion(version)
			if err != nil {
				log.Error(err, "Invalid semantic version found in image-manifests")
				return err
			}
			if !c.Check(v) {
				continue
			}
			// load manifest.json
			if m, err := readManifestFile(manifestFileName); err != nil {
				log.Error(err, "Failed to read image-manifests "+manifestFileName)
				return err
			} else {
				manifests[v.Original()] = *m
				versionList = append(versionList, v)
			}
		}
	}
	sort.Sort(semver.Collection(versionList))

	return nil
}

// GetAvailableVersions returns the available version list of klusterlet
func (instance KlusterletAddonConfig) GetAvailableVersions() ([]*semver.Version, error) {
	if len(versionList) == 0 {
		return nil, fmt.Errorf("Version list is empty")
	}

	return versionList, nil
}
