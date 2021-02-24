// (c) Copyright IBM Corporation 2019, 2020. All Rights Reserved.
// Note to U.S. Government Users Restricted Rights:
// U.S. Government Users Restricted Rights - Use, duplication or disclosure restricted by GSA ADP Schedule
// Contract with IBM Corp.
//
// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package v1

import (
	corev1 "k8s.io/api/core/v1"
)

// GlobalValues defines the global values
// +k8s:openapi-gen=true
type GlobalValues struct {
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	ImagePullSecret string            `json:"imagePullSecret,omitempty"`
	ImageOverrides  map[string]string `json:"imageOverrides,omitempty"`
}
