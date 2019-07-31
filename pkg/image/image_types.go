/*
 * IBM Confidential
 * OCO Source Materials
 * 5737-E67
 * (C) Copyright IBM Corporation 2018 All Rights Reserved
 * The source code for this program is not published or otherwise divested of its trade secrets, irrespective of what has been deposited with the U.S. Copyright Office.
 */

package image

import (
	corev1 "k8s.io/api/core/v1"
)

// Image defines the image to pull for a container
// +k8s:openapi-gen=true
type Image struct {
	// +kubebuilder:validation:MinLength=1
	Repository string `json:"repository,omitempty"`
	// +kubebuilder:validation:MinLength=1
	Tag string `json:"tag,omitempty"`
	// +kubebuilder:validation:Enum=Always,Never,IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}
