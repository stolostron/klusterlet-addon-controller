// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
// Licensed Materials - Property of IBM
//
// Copyright (c) 2020 Red Hat, Inc.

// Package inspect provide information and utils about the cluster itself
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
