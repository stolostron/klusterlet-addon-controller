// Package v1beta1 of apis contain the API type definition for the components
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"fmt"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
)

var versionComponentImageNameMap = map[string]map[string]string{
	"3.2.0": map[string]string{
		"search-collector":        "search-collector",
		"weave":                   "mcm-weave-scope",
		"collector":               "weave-collector",
		"router":                  "icp-management-ingress",
		"tiller":                  "tiller",
		"work-manager":            "mcm-klusterlet",
		"deployable":              "deployable",
		"connection-manager":      "mcm-operator",
		"cert-manager-controller": "icp-cert-manager-controller",
		"cert-manager-acmesolver": "icp-cert-manager-acmesolver",
		"service-registry":        "mcm-service-registry",
		"coredns":                 "coredns",
	},
}

var versionComponentTagMap = map[string]map[string]string{
	"3.2.0": map[string]string{
		"search-collector":        "3.2.0",
		"weave":                   "3.2.0",
		"collector":               "3.2.0",
		"router":                  "2.3.0",
		"tiller":                  "v2.12.3-icp-3.2.0",
		"work-manager":            "3.2.0",
		"deployable":              "3.2.0",
		"connection-manager":      "3.2.0",
		"cert-manager-controller": "0.7.0",
		"cert-manager-acmesolver": "0.7.0",
		"service-registry":        "3.2.0",
		"coredns":                 "1.2.6",
	},
}

// GetImage returns the image.Image for the specified component return error if information not found
func (instance Endpoint) GetImage(name string) (image.Image, error) {
	img := image.Image{}

	if componentImageMap, ok := versionComponentImageNameMap[instance.Spec.Version]; ok {
		if imageName, ok := componentImageMap[name]; ok {
			if instance.Spec.ImageRegistry != "" {
				img.Repository = instance.Spec.ImageRegistry + "/" + imageName
			} else {
				img.Repository = imageName
			}
		} else {
			return img, fmt.Errorf("unable to locate image name for component %s", name)
		}
	} else {
		return img, fmt.Errorf("unable to locate image name for version %s", instance.Spec.Version)
	}

	if instance.Spec.ImageNamePostfix != "" {
		img.Repository = img.Repository + instance.Spec.ImageNamePostfix
	}

	if componentTagMap, ok := versionComponentTagMap[instance.Spec.Version]; ok {
		if tag, ok := componentTagMap[name]; ok {
			img.Tag = tag
		} else {
			return img, fmt.Errorf("unable to locate image tag for component %s", name)
		}
	} else {
		return img, fmt.Errorf("unable to locate image name for version %s", instance.Spec.Version)
	}

	img.PullPolicy = instance.Spec.ImagePullPolicy

	return img, nil

}
