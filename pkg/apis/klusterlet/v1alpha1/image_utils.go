/*
 * IBM Confidential
 * OCO Source Materials
 * 5737-E67
 * (C) Copyright IBM Corporation 2018 All Rights Reserved
 * The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
 */

package v1alpha1

import (
	"fmt"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
)

var versionComponentImageMap = map[string]map[string]string{
	"3.2.0": map[string]string{
		"search-collector":        "ibmcom/search-collector",
		"weave":                   "ibmcom/mcm-weave-scope",
		"collector":               "ibmcom/weave-collector",
		"router":                  "ibmcom/icp-management-ingress",
		"tiller":                  "ibmcom/tiller",
		"work-manager":            "ibmcom/mcm-klusterlet",
		"deployable":              "ibmcom/deployable",
		"connection-manager":      "ibmcom/mcm-operator",
		"cert-manager-controller": "ibmcom/icp-cert-manager-controller",
		"cert-manager-acmesolver": "ibmcom/icp-cert-manager-acmesolver",
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
	},
}

// GetImage returns the image.Image for the specified component return error if information not found
func (instance KlusterletService) GetImage(name string) (image.Image, error) {
	img := image.Image{}

	if componentImageMap, ok := versionComponentImageMap[instance.Spec.Version]; ok {
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
