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
	"os"

	"github.com/open-cluster-management/endpoint-operator/pkg/image"
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

// GetImage returns the image.Image for the specified component return error if information not found
func (instance Endpoint) GetImage(component string) (image.Image, error) {
	img := image.Image{
		PullPolicy: instance.Spec.ImagePullPolicy,
	}

	if instance.Spec.ImageRegistry != "" {
		img.Repository = instance.Spec.ImageRegistry + "/"
	}

	if imageName, ok := defaultComponentImageMap[component]; ok {
		img.Repository = img.Repository + imageName
	} else {
		return img, fmt.Errorf("unable to locate default image name for component %s", component)
	}

	if instance.Spec.ImageNamePostfix != "" {
		img.Repository = img.Repository + instance.Spec.ImageNamePostfix
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
			return img, fmt.Errorf("unable to locate default tag for component %s", component)
		}
		imageTagPostfix := os.Getenv("IMAGE_TAG_POSTFIX")
		img.Tag = img.Tag + imageTagPostfix
	}

	return img, nil
}
