// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019, 2020 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.

package v1beta1

import (
	"fmt"
	"os"

	"github.com/open-cluster-management/endpoint-operator/pkg/image"
)

var defaultComponentImageMap = map[string]string{
	"cert-manager-acmesolver":        "icp-cert-manager-acmesolver",
	"cert-manager-controller":        "icp-cert-manager-controller",
	"cert-policy":                    "cert-policy-controller",
	"component-operator":             "endpoint-component-operator",
	"configmap-reload":               "configmap-reload",
	"connection-manager":             "multicloud-manager",
	"coredns":                        "coredns",
	"curl":                           "curl",
	"deployable":                     "deployable",
	"policy-controller":              "mcm-compliance",
	"prometheus":                     "prometheus",
	"prometheus-config-reloader":     "prometheus-config-reloader",
	"prometheus-operator":            "prometheus-operator",
	"prometheus-operator-controller": "prometheus-controller",
	"router":                         "icp-management-ingress",
	"search-collector":               "search-collector",
	"service-registry":               "mcm-service-registry",
	"subscription":                   "subscription",
	"topology-collector":             "weave-collector",
	"work-manager":                   "multicloud-manager",
	"weave":                          "mcm-weave-scope",
}

var defaultComponentTagMap = map[string]string{
	"cert-manager-acmesolver":        "0.10.0",
	"cert-manager-controller":        "0.10.0",
	"cert-policy":                    "3.4.0",
	"component-operator":             "3.3.0",
	"configmap-reload":               "v0.2.2-build.1",
	"connection-manager":             "latest",
	"coredns":                        "1.2.6.1",
	"curl":                           "4.2.0-build.2",
	"deployable":                     "3.3.0",
	"policy-controller":              "3.4.0",
	"prometheus":                     "v2.8.0-build.1",
	"prometheus-config-reloader":     "v0.31-build.1",
	"prometheus-operator":            "v0.31-build.1",
	"prometheus-operator-controller": "v1.1.0",
	"router":                         "2.5.0",
	"search-collector":               "3.3.0",
	"service-registry":               "3.3.0",
	"subscription":                   "3.3.0",
	"topology-collector":             "3.3.0",
	"work-manager":                   "latest",
	"weave":                          "3.3.0",
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
		if len(imageTagPostfix) > 0 {
			img.Tag = img.Tag + "-" + imageTagPostfix
		}
	}

	return img, nil
}
