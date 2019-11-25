// Package v1beta1 of apis contain the API type definition for the components
// IBM Confidential
// OCO Source Materials
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package v1beta1

import (
	"fmt"
	"strings"

	"github.ibm.com/IBMPrivateCloud/ibm-klusterlet-operator/pkg/image"
)

var componentImageNameMap32x = map[string]string{
	"search-collector":               "search-collector",
	"weave":                          "mcm-weave-scope",
	"topology-collector":             "weave-collector",
	"router":                         "icp-management-ingress",
	"tiller":                         "tiller",
	"work-manager":                   "mcm-klusterlet",
	"deployable":                     "deployable",
	"connection-manager":             "mcm-operator",
	"cert-manager-controller":        "icp-cert-manager-controller",
	"cert-manager-acmesolver":        "icp-cert-manager-acmesolver",
	"service-registry":               "mcm-service-registry",
	"coredns":                        "coredns",
	"component-operator":             "klusterlet-component-operator",
	"policy-controller":              "mcm-compliance",
	"metering-reader":                "metering-data-manager",
	"metering-sender":                "metering-data-manager",
	"metering-dm":                    "metering-data-manager",
	"curl":                           "curl",
	"mongodb":                        "icp-mongodb",
	"mongodb-install":                "icp-mongodb-install",
	"mongodb-metrics":                "icp-mongodb-exporter",
	"prometheus":                     "prometheus",
	"configmap-reload":               "configmap-reload",
	"alertrule-controller":           "alert-rule-controller",
	"prometheus-operator-controller": "prometheus-operator-controller",
	"prometheus-operator":            "prometheus-operator",
	"prometheus-config-reloader":     "prometheus-config-reloader",
	"subscription":                   "subscription",
	"helmcrd":                        "helm-crd-controller",
	"helmcrd-admission-controller":   "helm-crd-admission-controller",
}

var componentImageNameMap = map[string]string{
	"search-collector":               "search-collector",
	"weave":                          "mcm-weave-scope",
	"topology-collector":             "weave-collector",
	"router":                         "icp-management-ingress",
	"tiller":                         "tiller",
	"work-manager":                   "mcm-klusterlet",
	"deployable":                     "deployable",
	"connection-manager":             "mcm-operator",
	"cert-manager-controller":        "icp-cert-manager-controller",
	"cert-manager-acmesolver":        "icp-cert-manager-acmesolver",
	"service-registry":               "mcm-service-registry",
	"coredns":                        "coredns",
	"component-operator":             "klusterlet-component-operator",
	"policy-controller":              "mcm-compliance",
	"metering-reader":                "metering-data-manager",
	"metering-sender":                "metering-data-manager",
	"metering-dm":                    "metering-data-manager",
	"curl":                           "curl",
	"mongodb":                        "ibm-mongodb",
	"mongodb-install":                "ibm-mongodb-install",
	"mongodb-metrics":                "ibm-mongodb-exporter",
	"subscription":                   "subscription",
	"helmcrd":                        "helm-crd-controller",
	"helmcrd-admission-controller":   "helm-crd-admission-controller",
	"prometheus":                     "prometheus",
	"configmap-reload":               "configmap-reload",
	"alertrule-controller":           "alert-rule-controller",
	"prometheus-operator-controller": "prometheus-operator-controller",
	"prometheus-operator":            "prometheus-operator",
	"prometheus-config-reloader":     "prometheus-config-reloader",
}

var versionComponentImageNameMap = map[string]map[string]string{
	"3.2.0":      componentImageNameMap32x,
	"3.2.1":      componentImageNameMap32x,
	"3.2.1.1910": componentImageNameMap32x,
	"3.3.0":      componentImageNameMap,
	"latest":     componentImageNameMap,
}

var versionComponentTagMap = map[string]map[string]string{
	"3.2.0": map[string]string{
		"search-collector":             "3.2.0",
		"weave":                        "3.2.0",
		"topology-collector":           "3.2.0",
		"router":                       "2.3.0",
		"tiller":                       "v2.12.3-icp-3.2.0",
		"work-manager":                 "3.2.0",
		"deployable":                   "3.2.0",
		"connection-manager":           "3.2.0",
		"cert-manager-controller":      "0.7.0",
		"cert-manager-acmesolver":      "0.7.0",
		"service-registry":             "3.2.0",
		"coredns":                      "1.2.6",
		"policy-controller":            "3.2.0",
		"component-operator":           "3.2.0",
		"metering-reader":              "3.2.0",
		"metering-sender":              "3.2.0",
		"metering-dm":                  "3.2.0",
		"curl":                         "4.2.0-f3",
		"mongodb":                      "4.0.6-f1",
		"mongodb-install":              "3.2.0",
		"mongodb-metrics":              "3.2.0",
		"subscription":                 "3.2.0",
		"helmcrd":                      "3.2.0",
		"helmcrd-admission-controller": "3.2.0",
	},
	"3.2.1": map[string]string{
		"search-collector":               "3.2.1",
		"weave":                          "3.2.1",
		"topology-collector":             "3.2.1",
		"router":                         "2.4.0",
		"tiller":                         "v2.12.3-icp-3.2.1",
		"work-manager":                   "3.2.1",
		"deployable":                     "3.2.1",
		"connection-manager":             "3.2.1",
		"cert-manager-controller":        "0.7.0.1",
		"cert-manager-acmesolver":        "0.7.0.1",
		"service-registry":               "3.2.1",
		"coredns":                        "1.2.6.1",
		"policy-controller":              "3.2.1",
		"component-operator":             "3.2.1",
		"metering-reader":                "3.2.1",
		"metering-sender":                "3.2.1",
		"metering-dm":                    "3.2.1",
		"curl":                           "4.2.0-f4",
		"mongodb":                        "4.0.12",
		"mongodb-install":                "3.2.1",
		"mongodb-metrics":                "3.2.1",
		"prometheus":                     "v2.8.0-f1",
		"configmap-reload":               "v0.2.2-f4",
		"alertrule-controller":           "v1.1.0-f1",
		"prometheus-operator-controller": "v1.0.0",
		"prometheus-operator":            "v0.31",
		"prometheus-config-reloader":     "v0.31",
		"subscription":                   "3.2.1",
		"helmcrd":                        "3.2.1",
		"helmcrd-admission-controller":   "3.2.1",
	},
	"3.2.1.1910": map[string]string{
		"search-collector":               "3.2.1",
		"weave":                          "3.2.1",
		"topology-collector":             "3.2.1",
		"router":                         "2.4.0.1910",
		"tiller":                         "v2.12.3-icp-3.2.1",
		"work-manager":                   "3.2.1",
		"deployable":                     "3.2.1",
		"connection-manager":             "3.2.1",
		"cert-manager-controller":        "0.7.0.1",
		"cert-manager-acmesolver":        "0.7.0.1",
		"service-registry":               "3.2.1",
		"coredns":                        "1.2.6.1",
		"policy-controller":              "3.2.1",
		"component-operator":             "3.2.1",
		"metering-reader":                "3.2.1",
		"metering-sender":                "3.2.1",
		"metering-dm":                    "3.2.1",
		"curl":                           "4.2.0-f4",
		"mongodb":                        "4.0.12",
		"mongodb-install":                "3.2.1",
		"mongodb-metrics":                "3.2.1",
		"prometheus":                     "v2.8.0-f1",
		"configmap-reload":               "v0.2.2-f4",
		"alertrule-controller":           "v1.1.0-f1",
		"prometheus-operator-controller": "v1.0.0",
		"prometheus-operator":            "v0.31",
		"prometheus-config-reloader":     "v0.31",
		"subscription":                   "3.2.1",
		"helmcrd":                        "3.2.1",
		"helmcrd-admission-controller":   "3.2.1",
	},
	"3.3.0": map[string]string{
		"search-collector":               "3.3.0",
		"weave":                          "3.3.0",
		"topology-collector":             "3.3.0",
		"router":                         "2.5.0",
		"tiller":                         "v2.12.3-icp-3.2.2",
		"work-manager":                   "3.3.0",
		"deployable":                     "3.3.0",
		"connection-manager":             "3.3.0",
		"cert-manager-controller":        "0.7.0.1",
		"cert-manager-acmesolver":        "0.7.0.1",
		"service-registry":               "3.3.0",
		"coredns":                        "1.2.6.1",
		"policy-controller":              "3.4.0",
		"component-operator":             "3.3.0",
		"metering-reader":                "3.3.0",
		"metering-sender":                "3.3.0",
		"metering-dm":                    "3.3.0",
		"curl":                           "4.2.0-f4",
		"mongodb":                        "4.0.12-build.1",
		"mongodb-install":                "3.3.0",
		"mongodb-metrics":                "3.3.0",
		"prometheus":                     "v2.8.0-f1",
		"configmap-reload":               "v0.2.2-f4",
		"alertrule-controller":           "v1.1.0-f1",
		"prometheus-operator-controller": "v1.0.0",
		"prometheus-operator":            "v0.31",
		"prometheus-config-reloader":     "v0.31",
		"subscription":                   "3.3.0",
		"helmcrd":                        "3.2.1",
		"helmcrd-admission-controller":   "3.2.1",
	},
	"latest": map[string]string{
		"search-collector":               "3.3.0",
		"weave":                          "3.3.0",
		"topology-collector":             "3.3.0",
		"router":                         "2.5.0",
		"tiller":                         "v2.12.3-icp-3.2.2",
		"work-manager":                   "3.3.0",
		"deployable":                     "3.3.0",
		"connection-manager":             "3.3.0",
		"cert-manager-controller":        "0.7.0.1",
		"cert-manager-acmesolver":        "0.7.0.1",
		"service-registry":               "3.3.0",
		"coredns":                        "1.2.6.1",
		"policy-controller":              "3.4.0",
		"component-operator":             "3.3.0",
		"metering-reader":                "3.3.0",
		"metering-sender":                "3.3.0",
		"metering-dm":                    "3.3.0",
		"curl":                           "4.2.0-f4",
		"mongodb":                        "4.0.12-build.1",
		"mongodb-install":                "3.3.0",
		"mongodb-metrics":                "3.3.0",
		"prometheus":                     "v2.8.0-f1",
		"configmap-reload":               "v0.2.2-f4",
		"alertrule-controller":           "v1.1.0-f1",
		"prometheus-operator-controller": "v1.0.0",
		"prometheus-operator":            "v0.31",
		"prometheus-config-reloader":     "v0.31",
		"subscription":                   "3.3.0",
		"helmcrd":                        "3.2.1",
		"helmcrd-admission-controller":   "3.2.1",
	},
}

// GetImage returns the image.Image for the specified component return error if information not found
func (instance Endpoint) GetImage(name string) (image.Image, error) {
	img := image.Image{}

	versionSplit := strings.Split(instance.Spec.Version, "-")
	if len(versionSplit) == 0 || len(versionSplit) > 2 {
		return img, fmt.Errorf("invalid version %s", instance.Spec.Version)
	}

	if componentImageMap, ok := versionComponentImageNameMap[versionSplit[0]]; ok {
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
		return img, fmt.Errorf("unable to locate image name for version %s", versionSplit[0])
	}

	if instance.Spec.ImageNamePostfix != "" {
		img.Repository = img.Repository + instance.Spec.ImageNamePostfix
	}

	if componentTagMap, ok := versionComponentTagMap[versionSplit[0]]; ok {
		if tag, ok := componentTagMap[name]; ok {
			img.Tag = tag
		} else {
			return img, fmt.Errorf("unable to locate image tag for component %s", name)
		}
	} else {
		return img, fmt.Errorf("unable to locate image name for version %s", versionSplit[0])
	}

	if len(versionSplit) == 2 {
		img.Tag = img.Tag + "-" + versionSplit[1]
	}

	img.PullPolicy = instance.Spec.ImagePullPolicy

	return img, nil
}
