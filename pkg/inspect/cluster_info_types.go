// Package inspect provide information and utils about the cluster itself
// IBM Confidential
// OCO Source Materials
// 5737-E67
// (C) Copyright IBM Corporation 2019 All Rights Reserved
// The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
package inspect

import (
	"k8s.io/apimachinery/pkg/version"
)

// Info is the package global variable to store the ClusterInfo
var Info ClusterInfo

// ClusterInfo contains information about the Kubernetes Cluster
type ClusterInfo struct {
	KubeVersion version.Info
	KubeVendor  KubeVendor
	CloudVendor CloudVendor
}

// KubeVendor describe the kubernetes provider of the cluster
type KubeVendor string

const (
	// KubeVendorOpenShift OpenShift
	KubeVendorOpenShift KubeVendor = "OpenShift"
	// KubeVendorAKS Azure Kuberentes Service
	KubeVendorAKS KubeVendor = "AKS"
	// KubeVendorEKS Elastic Kubernetes Service
	KubeVendorEKS KubeVendor = "EKS"
	// KubeVendorGKE Google Kubernetes Engine
	KubeVendorGKE KubeVendor = "GKE"
	// KubeVendorICP IBM Cloud Private
	KubeVendorICP KubeVendor = "ICP"
	// KubeVendorIKS IBM Kubernetes Service
	KubeVendorIKS KubeVendor = "IKS"
	// KubeVendorOther other (unable to auto detect)
	KubeVendorOther KubeVendor = "Other"
)

// CloudVendor describe the cloud provider for the cluster
type CloudVendor string

const (
	// CloudVendorIBM IBM
	CloudVendorIBM CloudVendor = "IBM"
	// CloudVendorAWS Amazon
	CloudVendorAWS CloudVendor = "Amazon"
	// CloudVendorAzure Azure
	CloudVendorAzure CloudVendor = "Azure"
	// CloudVendorGoogle Google
	CloudVendorGoogle CloudVendor = "Google"
	// CloudVendorOther other (unable to auto detect)
	CloudVendorOther CloudVendor = "Other"
)
