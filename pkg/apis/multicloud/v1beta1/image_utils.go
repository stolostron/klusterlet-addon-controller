// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

package v1beta1

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghodss/yaml"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/open-cluster-management/endpoint-operator/pkg/image"
	"github.com/open-cluster-management/endpoint-operator/version"
)

var defaultComponentImageMap = map[string]string{
	"cert-policy-controller":       "cert-policy-controller",
	"cis-controller-controller":    "cis-controller",
	"cis-controller-crawler":       "cis-crawler",
	"cis-controller-drishti":       "drishti-cis",
	"cis-controller-minio":         "img-minio",
	"cis-controller-minio-cleaner": "img-minio-mc",
	"component-operator":           "endpoint-component-operator",
	"connection-manager":           "multicloud-manager",
	"coredns":                      "coredns",
	"deployable":                   "multicluster-operators-deployable",
	"iam-policy-controller":        "iam-policy-controller",
	"policy-controller":            "mcm-compliance",
	"router":                       "management-ingress",
	"search-collector":             "search-collector",
	"service-registry":             "multicloud-manager",
	"subscription":                 "multicluster-operators-subscription",
	"work-manager":                 "multicloud-manager",
}

var defaultComponentTagMap = map[string]string{
	"cert-policy-controller":       "3.4.0",
	"cis-controller-controller":    "3.6.0",
	"cis-controller-crawler":       "3.6.0",
	"cis-controller-drishti":       "3.4.0",
	"cis-controller-minio":         "RELEASE.2019-04-09T01-22-30Z.3",
	"cis-controller-minio-cleaner": "RELEASE.2019-04-03T17-59-57Z.3",
	"component-operator":           "1.0.0",
	"connection-manager":           "0.0.1",
	"coredns":                      "1.2.6.1",
	"deployable":                   "1.0.0",
	"iam-policy-controller":        "1.0.0",
	"policy-controller":            "3.6.0",
	"router":                       "1.0.0",
	"search-collector":             "3.5.0",
	"service-registry":             "0.0.1",
	"subscription":                 "1.0.0",
	"work-manager":                 "0.0.1",
}

//Manifest contains the manifest.
//The Manifest is loaded using the LoadManifest method.
var Manifest manifest

var log = logf.Log.WithName("image_utils")

type manifest struct {
	Images []imageManifest `json:"inline"`
}

type imageManifest struct {
	Name           string `json:"name,omitempty"`
	ManifestSha256 string `json:"manifest-sha256,omitempty"`
}

func init() {
	err := LoadManifest()
	if err != nil {
		log.Error(err, "Error while reading the manifest")
	}
}

// GetImage returns the image.Image,  for the specified component return error if information not found
func (instance Endpoint) GetImage(component string,
	imageShaDigestIn map[string]string,
) (img image.Image, imageShaDigest map[string]string, err error) {
	imageShaDigest = imageShaDigestIn
	img = image.Image{
		PullPolicy: instance.Spec.ImagePullPolicy,
	}

	if imageName, ok := defaultComponentImageMap[component]; ok {
		img.Name = imageName
	} else {
		return img, imageShaDigest, fmt.Errorf("unable to locate default image name for component %s", component)
	}

	if instance.Spec.ImageRegistry != "" {
		img.Repository = instance.Spec.ImageRegistry
	}

	if instance.Spec.ImageNamePostfix != "" {
		img.Name = img.Name + instance.Spec.ImageNamePostfix
	}

	if len(instance.Spec.ComponentsImagesTag) > 0 {
		if tag, ok := instance.Spec.ComponentsImagesTag[component]; ok {
			img.Tag = tag
		} // else {
		// TODO how to log - WARN("unable to locate tag for component %s", component)
		// don't want to add new dependencies to other projects importing this package
		//}
	}
	if img.Tag == "" {
		if tag, ok := defaultComponentTagMap[component]; ok {
			img.Tag = tag
		} else {
			return img, imageShaDigest, fmt.Errorf("unable to locate default tag for component %s", component)
		}
	}
	img.TagPostfix = os.Getenv("IMAGE_TAG_POSTFIX")
	useSHA := os.Getenv("USE_SHA_MANIFEST")
	// fmt.Println("useSHA:" + useSHA)
	if strings.ToLower(useSHA) == "true" {
		im, err := getImageManifest(img.Name)
		if err != nil {
			return img, imageShaDigest, err
		}
		if im != nil {
			shaDigestKey := strings.ReplaceAll(img.Name, "-", "_")
			// fmt.Println("shaDigestKey:" + shaDigestKey)
			imageShaDigest[shaDigestKey] = im.ManifestSha256
			// fmt.Println(("sha:" + imageShaDigest[shaDigestKey]))
			return img, imageShaDigest, nil
		}
	}
	return img, imageShaDigest, nil
}

//getImageManifest returns the *imageManifest and nil if not found
//Return an error only if the manifest is malformed
func getImageManifest(imageName string) (*imageManifest, error) {
	for i, im := range Manifest.Images {
		if im.Name == imageName {
			return &Manifest.Images[i], nil
		}
	}
	return nil, nil
}

//LoadManifest returns the *imageManifest and nil if not found
//Return an error only if the manifest is malformed
func LoadManifest() error {
	Manifest.Images = make([]imageManifest, 0)
	filePath := filepath.Join("image-manifests", version.Version+".json")
	homeDir := os.Getenv("HOME")
	if homeDir != "" {
		filePath = filepath.Join(homeDir, filePath)
	}
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(b, &Manifest.Images)
	if err != nil {
		return err
	}
	return nil
}
